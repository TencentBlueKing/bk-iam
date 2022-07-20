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

// Producer ...
type Producer interface {
	Publish(types.GroupAlterEvent) error
}

type producer struct {
	groupAlterEventService service.GroupAlterEventService
}

// NewProducer ...
func NewProducer() Producer {
	return &producer{
		groupAlterEventService: service.NewGroupAlterEventService(),
	}
}

// Publish ...
func (p *producer) Publish(event types.GroupAlterEvent) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(task, "Publish")

	oldEvents, err := p.groupAlterEventService.ListUncheckedByGroup(event.GroupPK)
	if err != nil {
		err = errorWrapf(err, "groupAlterEventService.ListByGroup groupPK=`%d` fail", event.GroupPK)
		return err
	}

	// 去重
	uniqueIDSet := set.NewStringSet()
	for _, oldEvent := range oldEvents {
		for _, subjectPK := range oldEvent.SubjectPKs {
			for _, actionPK := range oldEvent.ActionPKs {
				message := Message{
					GroupPK:   oldEvent.GroupPK,
					ActionPK:  actionPK,
					SubjectPK: subjectPK,
				}

				uniqueIDSet.Add(message.UniqueID())
			}
		}
	}

	messages := make([]Message, 0, 5)
	for _, subjectPK := range event.SubjectPKs {
		for _, actionPK := range event.ActionPKs {
			message := Message{
				GroupPK:   event.GroupPK,
				ActionPK:  actionPK,
				SubjectPK: subjectPK,
			}

			if !uniqueIDSet.Has(message.UniqueID()) {
				messages = append(messages, message)
			}
		}
	}

	if len(messages) == 0 {
		return nil
	}

	// 批量发消息到mq
	err = sendMessages(messages)
	if err != nil {
		err = errorWrapf(err, "sendMessages messages=`%+v` fail", messages)

		log.WithError(err).
			Errorf("task producer sendMessages messages=%v fail", messages)

		// report to sentry
		util.ReportToSentry("task producer sendMessages fail",
			map[string]interface{}{
				"layer":    task,
				"messages": messages,
				"error":    err.Error(),
			},
		)

		return err
	}

	return nil
}

func sendMessages(messages []Message) (err error) {
	ms := make([]string, 0, len(messages))
	for _, message := range messages {
		str, err := message.String()
		if err != nil {
			log.WithError(err).Debugf("message Marshal fail, message=%+v", message)
			continue
		}

		ms = append(ms, str)
	}

	err = queue.Publish(ms...)
	return err
}
