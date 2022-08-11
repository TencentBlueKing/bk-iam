/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package consumer

import (
	"context"
	"time"

	"github.com/adjust/rmq/v4"
	log "github.com/sirupsen/logrus"

	"iam/pkg/logging"
	"iam/pkg/task/handler"
	"iam/pkg/util"
)

const (
	consumerLayer = "consumer"

	prefetchLimit = 1000
	pollDuration  = 100 * time.Millisecond
)

type stats struct {
	totalCount          int64
	successCount        int64
	failCount           int64
	startTime           time.Time
	lastShowProcessTime time.Time
}

type redisConsumer struct {
	connection rmq.Connection
	queue      rmq.Queue

	handler handler.MessageHandler
	stats   stats
}

func NewRedisConsumer(connection rmq.Connection, queue rmq.Queue, handler handler.MessageHandler) Consumer {
	return &redisConsumer{
		connection: connection,
		queue:      queue,
		handler:    handler,
		stats: stats{
			startTime:           time.Now(),
			lastShowProcessTime: time.Now(),
		},
	}
}

// Run ...
func (c *redisConsumer) Run(ctx context.Context) {
	// start consuming
	if err := c.queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		log.WithError(err).Error("rmq queue start consuming fail")
		panic(err)
	}

	// consume messages
	if _, err := c.queue.AddConsumer(consumerLayer, c); err != nil {
		log.WithError(err).Error("rmq queue add consumer fail")
		panic(err)
	}

	<-ctx.Done()                      // wait for signal
	<-c.connection.StopAllConsuming() // wait for all Consume() calls to finish
}

// Consume ...
func (c *redisConsumer) Consume(delivery rmq.Delivery) {
	logger := logging.GetWorkerLogger()
	c.stats.totalCount += 1

	// parse message
	payload := delivery.Payload()

	logger.Debugf("receive message: %s", payload)

	// handle message
	err := c.handler.Handle(payload)
	if err != nil {
		c.stats.failCount += 1
		logger.WithError(err).Errorf("handle message `%+v` fail", payload)

		// report to sentry
		util.ReportToSentry("handle message fail",
			map[string]interface{}{
				"layer":   consumerLayer,
				"message": payload,
				"error":   err.Error(),
			},
		)
	} else {
		c.stats.successCount += 1
	}

	logger.Debugf("handle message `%+v` done", payload)

	// ack
	if err = delivery.Ack(); err != nil {
		logger.WithError(err).Errorf("rmq ack payload `%s` fail", payload)
	}

	if c.stats.totalCount%1000 == 0 || time.Since(c.stats.lastShowProcessTime) > 30*time.Second {
		c.stats.lastShowProcessTime = time.Now()
		logger.Infof("consumer processed total count: %d, success count: %d, fail count: %d, elapsed: %s",
			c.stats.totalCount, c.stats.successCount, c.stats.failCount, time.Since(c.stats.startTime))
	}
}
