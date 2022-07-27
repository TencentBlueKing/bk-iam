/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package producer

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/adjust/rmq/v4"

	"iam/pkg/logging"
	"iam/pkg/util"
)

const producerLayer = "producer"

// Producer ...
type Producer interface {
	Publish(messages ...string) error
}

type redisProducer struct {
	queue rmq.Queue
}

// NewRedisGroupAlterEventProducer ...
func NewRedisProducer(queue rmq.Queue) Producer {
	return &redisProducer{
		queue: queue,
	}
}

// Publish ...
func (p *redisProducer) Publish(messages ...string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(producerLayer, "Publish")

	// 批量发消息到mq
	err := p.queue.Publish(messages...)
	if err != nil {
		logger := logging.GetWorkerLogger()
		logger.WithError(err).
			Errorf("task producer sendMessages messages=%v fail", messages)

		// report to sentry
		util.ReportToSentry("task producer sendMessages fail",
			map[string]interface{}{
				"layer":    producerLayer,
				"messages": messages,
				"error":    err.Error(),
			},
		)

		return errorWrapf(err, "sendMessages messages=`%+v` fail", messages)
	}

	return nil
}
