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

const (
	transferLayer = "Transfer"

	eventLimit int64 = 1000
)

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

type stats struct {
	totalCount          int64
	successCount        int64
	failCount           int64
	startTime           time.Time
	lastShowProcessTime time.Time
}

type GroupAlterEventTransfer struct {
	service                        service.GroupAlterEventService
	subjectActionAlterEventService service.SubjectActionAlterEventService
	producer                       producer.Producer

	stats stats
}

func NewGroupAlterEventTransfer(producer producer.Producer) *GroupAlterEventTransfer {
	return &GroupAlterEventTransfer{
		service:                        service.NewGroupAlterEventService(),
		subjectActionAlterEventService: service.NewSubjectActionAlterEventService(),
		producer:                       producer,
		stats: stats{
			startTime:           time.Now(),
			lastShowProcessTime: time.Now(),
		},
	}
}

func (t *GroupAlterEventTransfer) Run() {
	logger := logging.GetWorkerLogger().WithField("layer", transferLayer)

	logger.Info("Start transfer group alter event to subject action alter event")

	for {
		t.stats.totalCount += 1

		count, err := t.transform()
		if err != nil {
			t.stats.failCount += 1
			logger.WithError(err).Error("transform fail")

			// report to sentry
			util.ReportToSentry("GroupAlterEventTransfer.transform fail",
				map[string]interface{}{
					"layer": transferLayer,
					"error": err.Error(),
				},
			)

			time.Sleep(30 * time.Second)
		} else {
			t.stats.successCount += 1

			// 时间段内的消息处理完成后, 休眠30秒
			if count < int(eventLimit) {
				time.Sleep(30 * time.Second)
			}
		}

		if t.stats.totalCount%1000 == 0 || time.Since(t.stats.lastShowProcessTime) > 30*time.Second {
			t.stats.lastShowProcessTime = time.Now()
			logger.Infof("transfer processed total count: %d, success count: %d, fail count: %d, elapsed: %s",
				t.stats.totalCount, t.stats.successCount, t.stats.failCount, time.Since(t.stats.startTime))
		}
	}
}

func (t *GroupAlterEventTransfer) transform() (count int, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(transferLayer, "transform")

	createdAt := time.Now().Add(-30 * time.Second).Unix()
	events, err := t.service.ListBeforeCreateAt(createdAt, eventLimit)
	if err != nil {
		return 0, errorWrapf(err, "service.ListBeforeCreateAt fail createdAt=`%d`", createdAt)
	}

	count = len(events)
	if count == 0 {
		return count, nil
	}

	// 生成subject action alter event
	subjectActionAlterEvents := convertToSubjectActionAlterEvent(events)

	tx, err := database.GenerateDefaultDBTx()
	if err != nil {
		return 0, errorWrapf(err, "database.GenerateDefaultDBTx fail", "")
	}
	defer database.RollBackWithLog(tx)

	// 生成 subject action alter event 并同时删除 group alter event
	err = t.subjectActionAlterEventService.BulkCreateWithTx(tx, subjectActionAlterEvents)
	if err != nil {
		return 0, errorWrapf(
			err,
			"subjectActionAlterEventService.BulkCreateWithTx fail events=`%v`",
			subjectActionAlterEvents,
		)
	}

	eventUUIDs := make([]string, 0, len(events))
	for _, event := range events {
		eventUUIDs = append(eventUUIDs, event.UUID)
	}

	err = t.service.BulkDeleteWithTx(tx, eventUUIDs)
	if err != nil {
		return 0, errorWrapf(err, "service.BulkDeleteWithTx fail eventUUIDs=`%v`", eventUUIDs)
	}

	err = tx.Commit()
	if err != nil {
		return 0, errorWrapf(err, "tx.Commit fail", "")
	}

	logger := logging.GetWorkerLogger().WithField("layer", transferLayer)
	logger.Infof(
		"transform group alter event [count: %d] to subject action alter event [count: %d] success ",
		count,
		len(subjectActionAlterEvents),
	)

	messageUUIDs := make([]string, 0, len(subjectActionAlterEvents))
	for _, message := range subjectActionAlterEvents {
		messageUUIDs = append(messageUUIDs, message.UUID)
	}

	// 发送消息
	err = t.producer.Publish(messageUUIDs...)
	if err != nil {
		// NOTE: 发送消息失败, 由checker定时检查, 这里整个转换逻辑实际已经完成, 所以不再返回错误
		err = errorWrapf(err, "producer.Publish fail messageUUIDs=`%v`", messageUUIDs)
		logger.WithError(err).Warn("Publish fail")
		return count, nil
	}

	// 变更消息状态为已推送
	err = t.subjectActionAlterEventService.BulkUpdateStatus(messageUUIDs, types.SubjectActionAlterEventStatusPushed)
	if err != nil {
		// NOTE: 同上, 设置状态失败,
		err = errorWrapf(
			err,
			"subjectActionAlterEventService.BulkUpdateStatus fail messageUUIDs=`%v`, status=`%d`",
			messageUUIDs,
			types.SubjectActionAlterEventStatusPushed,
		)
		logger.WithError(err).Warn("BulkUpdateStatus fail")
		return count, nil
	}

	return count, nil
}

func convertToSubjectActionAlterEvent(events []types.GroupAlterEvent) []types.SubjectActionAlterEvent {
	// 合并去重，所有的subject action group
	subjectActionGroupMap := mergeSubjectActionGroup(events)

	// 生成subject action alter event
	subjectActionAlterEvents := make([]types.SubjectActionAlterEvent, 0, len(subjectActionGroupMap)*2)

	messages := make([]types.SubjectActionGroupMessage, 0, len(subjectActionGroupMap))
	for key, groupPKSet := range subjectActionGroupMap {
		message := types.SubjectActionGroupMessage{
			SubjectPK: key.SubjectPK,
			ActionPK:  key.ActionPK,
			GroupPKs:  groupPKSet.ToSlice(),
		}

		messages = append(messages, message)
	}

	step := config.MaxMessageGeneratedCountPreSubjectActionAlterEvent
	for _, index := range util.Chunks(len(messages), step) {
		subjectActionAlterEvents = append(subjectActionAlterEvents, types.SubjectActionAlterEvent{
			UUID:     util.GenUUID4(),
			Messages: messages[index.Begin:index.End],
		})
	}

	return subjectActionAlterEvents
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
