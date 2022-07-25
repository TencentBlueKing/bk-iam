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
	"errors"
	"time"

	"github.com/adjust/rmq/v4"
	log "github.com/sirupsen/logrus"

	"iam/pkg/util"
)

const (
	prefetchLimit = 1000
	pollDuration  = 100 * time.Millisecond
)

/*
TODO

1. 独立redis配置
2. 消息消费限流
*/

type groupAlterMessageConsumer struct {
	GroupAlterMessageHandler
}

func NewRedisGroupAlterMessageConsumer() GroupAlterEventConsumer {
	return &groupAlterMessageConsumer{
		GroupAlterMessageHandler: NewGroupAlterMessageHandler(),
	}
}

// Run ...
func (c *groupAlterMessageConsumer) Run(ctx context.Context) {
	// start consuming
	if err := queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		log.WithError(err).Error("rmq queue start consuming fail")
		panic(err)
	}

	// consume messages
	if _, err := queue.AddConsumer(ConnTypeConsumer, c); err != nil {
		log.WithError(err).Error("rmq queue add consumer fail")
		panic(err)
	}

	<-ctx.Done()                    // wait for signal
	<-connection.StopAllConsuming() // wait for all Consume() calls to finish
}

// Consume ...
func (c *groupAlterMessageConsumer) Consume(delivery rmq.Delivery) {
	// parse message
	payload := delivery.Payload()
	message, err := NewGroupAlterMessageFromString(payload)
	if err != nil {
		log.WithError(err).Error("new message from string fail")

		// report to sentry
		util.ReportToSentry("task producer sendMessages fail",
			map[string]interface{}{
				"layer":   ConnTypeConsumer,
				"payload": payload,
				"error":   err.Error(),
			},
		)

		// remove message from unacked queue
		if err = delivery.Ack(); err != nil {
			log.WithError(err).Errorf("rmq ack payload `%s` fail", payload)
		}
		return
	}

	// handle message
	err = c.Handle(message)
	if errors.Is(err, ErrNeedRetry) {
		// 需要重试, 重新放回队列
		if err = delivery.Push(); err != nil {
			log.WithError(err).Errorf("rmq push payload `%s` fail", payload)
		}
		return
	} else if err != nil {
		log.WithError(err).Errorf("handle message `%+v` fail", message)

		// report to sentry
		util.ReportToSentry("handle message fail",
			map[string]interface{}{
				"layer":   ConnTypeConsumer,
				"message": message,
				"error":   err.Error(),
			},
		)

		// NOTE: 如果消息处理失败, 先ack掉, 等待检查逻辑重新发送
	}

	// ack
	if err = delivery.Ack(); err != nil {
		log.WithError(err).Errorf("rmq ack payload `%s` fail", payload)
	}
}
