/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package task

import (
	"context"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/config"
	"iam/pkg/database"
	"iam/pkg/logging"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/task/producer"
	"iam/pkg/util"
)

const eventLimit int64 = 1000

// type Transfer ...
type Transfer struct {
	stopChan chan struct{}
}

// NewTransfer ...
func NewTransfer() *Transfer {
	return &Transfer{
		stopChan: make(chan struct{}, 1),
	}
}

// Run ...
func (t *Transfer) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		log.Info("I have to go...")
		log.Info("Stopping worker gracefully")
		t.Stop()
	}()

	// Start transfer
	go NewGroupAlterEventTransfer(
		producer.NewRedisProducer(rbacEventQueue),
	).Run()

	t.Wait()
	log.Info("Shutting down")
}

// Stop ...
func (t *Transfer) Stop() {
	defer log.Info("Server stopped")

	t.stopChan <- struct{}{}
}

// Wait blocks until server is shut down.
func (t *Transfer) Wait() {
	<-t.stopChan
}

type subjectAction struct {
	SubjectPK int64
	ActionPK  int64
}

type GroupAlterEventTransfer struct {
	service                          service.GroupAlterEventService
	subjectActionAlterMessageService service.SubjectActionAlterMessageService
	producer                         producer.Producer
}

func NewGroupAlterEventTransfer(producer producer.Producer) *GroupAlterEventTransfer {
	return &GroupAlterEventTransfer{
		service:                          service.NewGroupAlterEventService(),
		subjectActionAlterMessageService: service.NewSubjectActionAlterMessageService(),
		producer:                         producer,
	}
}

func (t *GroupAlterEventTransfer) Run() {
	logger := logging.GetWorkerLogger()

	// run every 30 seconds
	for range time.Tick(30 * time.Second) {
		logger.Info("Start transfer group alter event to subject action alter message")

		err := t.transform()
		if err != nil {
			logger.WithError(err).Error("transform fail")
		}
	}
}

func (t *GroupAlterEventTransfer) transform() error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Transer", "transform")

	createdAt := time.Now().Add(-30 * time.Second).Unix()
	events, err := t.service.ListBeforeCreateAt(createdAt, eventLimit)
	if err != nil {
		return errorWrapf(err, "service.ListBeforeCreateAt fail createdAt=`%d`", createdAt)
	}

	if len(events) == 0 {
		return nil
	}

	// 生成subject action alter message
	subjectActionAlterMessages := convertToSubjectActionAlterMessage(events)

	tx, err := database.GenerateDefaultDBTx()
	if err != nil {
		return errorWrapf(err, "database.GenerateDefaultDBTx fail", "")
	}
	defer database.RollBackWithLog(tx)

	// 生成 subject action alter message 并同时删除 group alter event
	err = t.subjectActionAlterMessageService.BulkCreateWithTx(tx, subjectActionAlterMessages)
	if err != nil {
		return errorWrapf(
			err,
			"subjectActionAlterMessageService.BulkCreateWithTx fail message=`%v`",
			subjectActionAlterMessages,
		)
	}

	eventUUIDs := make([]string, 0, len(events))
	for _, event := range events {
		eventUUIDs = append(eventUUIDs, event.UUID)
	}

	err = t.service.BulkDeleteWithTx(tx, eventUUIDs)
	if err != nil {
		return errorWrapf(err, "service.BulkDeleteWithTx fail eventUUIDs=`%v`", eventUUIDs)
	}

	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx.Commit fail", "")
	}

	messageUUIDs := make([]string, 0, len(subjectActionAlterMessages))
	for _, message := range subjectActionAlterMessages {
		messageUUIDs = append(messageUUIDs, message.UUID)
	}

	// 发送消息
	err = t.producer.Publish(messageUUIDs...)
	if err != nil {
		return errorWrapf(err, "producer.Publish fail messageUUIDs=`%v`", messageUUIDs)
	}

	// 变更消息状态为已推送
	err = t.subjectActionAlterMessageService.BulkUpdateStatus(messageUUIDs, types.SubjectActionAlterMessageStatusPushed)
	if err != nil {
		return errorWrapf(
			err,
			"subjectActionAlterMessageService.BulkUpdateStatus fail messageUUIDs=`%v`, status=`%d`",
			messageUUIDs,
			types.SubjectActionAlterMessageStatusPushed,
		)
	}

	return nil
}

func convertToSubjectActionAlterMessage(events []types.GroupAlterEvent) []types.SubjectActionAlterMessage {
	// 合并去重，所有的subject action group
	subjectActionGroupMap := mergeSubjectActionGroup(events)

	// 生成subject action alter message
	subjectActionAlterMessages := make([]types.SubjectActionAlterMessage, 0, len(subjectActionGroupMap)*2)

	count := config.MaxGenerationCountPreSubjectActionAlterMessage
	messages := make([]types.SubjectActionGroupMessage, 0, count)
	for key, groupPKSet := range subjectActionGroupMap {
		message := types.SubjectActionGroupMessage{
			SubjectPK: key.SubjectPK,
			ActionPK:  key.ActionPK,
			GroupPKs:  groupPKSet.ToSlice(),
		}

		messages = append(messages, message)
		if len(messages) == count {
			subjectActionAlterMessages = append(subjectActionAlterMessages, types.SubjectActionAlterMessage{
				UUID:     util.GenUUID4(),
				Messages: messages,
			})

			// 清空messages
			messages = make([]types.SubjectActionGroupMessage, 0, count)
		}
	}

	if len(messages) > 0 {
		subjectActionAlterMessages = append(subjectActionAlterMessages, types.SubjectActionAlterMessage{
			UUID:     util.GenUUID4(),
			Messages: messages,
		})
	}

	return subjectActionAlterMessages
}

func mergeSubjectActionGroup(events []types.GroupAlterEvent) map[subjectAction]*set.Int64Set {
	subjectActionGroupMap := make(map[subjectAction]*set.Int64Set)
	for _, event := range events {
		for _, subjectPK := range event.SubjectPKs {
			for _, actionPK := range event.ActionPKs {
				key := subjectAction{
					SubjectPK: subjectPK,
					ActionPK:  actionPK,
				}

				if _, ok := subjectActionGroupMap[key]; !ok {
					subjectActionGroupMap[key] = set.NewInt64Set()
				}

				subjectActionGroupMap[key].Add(event.GroupPK)
			}
		}
	}
	return subjectActionGroupMap
}
