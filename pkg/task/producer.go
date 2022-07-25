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

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

type redisGroupAlterEventProducer struct {
	groupAlterEventService service.GroupAlterEventService
}

// NewRedisGroupAlterEventProducer ...
func NewRedisGroupAlterEventProducer() GroupAlterEventProducer {
	return &redisGroupAlterEventProducer{
		groupAlterEventService: service.NewGroupAlterEventService(),
	}
}

// Publish ...
func (p *redisGroupAlterEventProducer) Publish(event types.GroupAlterEvent) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ConnTypeProducer, "Publish")

	oldEvents, err := p.groupAlterEventService.ListUncheckedByGroup(event.GroupPK)
	if err != nil {
		err = errorWrapf(err, "groupAlterEventService.ListByGroup groupPK=`%d` fail", event.GroupPK)
		return err
	}

	// 去重
	uniqueIDSet := set.NewStringSet()
	for _, oldEvent := range oldEvents {
		for _, m := range eventToMessages(oldEvent) {
			uniqueIDSet.Add(m.UniqueID())
		}
	}

	messages := make([]string, 0, len(event.ActionPKs)*len(event.SubjectPKs))
	for _, m := range eventToMessages(event) {
		if uniqueIDSet.Has(m.UniqueID()) {
			continue
		}

		s, err := m.String()
		if err != nil {
			log.WithError(err).Debugf("message Marshal fail, message=%+v", m)
			continue
		}

		messages = append(messages, s)
	}

	if len(messages) == 0 {
		return nil
	}

	// 批量发消息到mq
	err = queue.Publish(messages...)
	if err != nil {
		log.WithError(err).
			Errorf("task producer sendMessages messages=%v fail", messages)

		// report to sentry
		util.ReportToSentry("task producer sendMessages fail",
			map[string]interface{}{
				"layer":    ConnTypeProducer,
				"messages": messages,
				"error":    err.Error(),
			},
		)

		return errorWrapf(err, "sendMessages messages=`%+v` fail", messages)
	}

	return nil
}

func eventToMessages(event types.GroupAlterEvent) []GroupAlterMessage {
	messages := make([]GroupAlterMessage, 0, len(event.ActionPKs)*len(event.SubjectPKs))

	for _, subjectPK := range event.SubjectPKs {
		for _, actionPK := range event.ActionPKs {
			message := GroupAlterMessage{
				GroupPK:   event.GroupPK,
				ActionPK:  actionPK,
				SubjectPK: subjectPK,
			}

			messages = append(messages, message)
		}
	}

	return messages
}
