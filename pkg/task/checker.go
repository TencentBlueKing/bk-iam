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
	"github.com/adjust/rmq/v4"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/config"
	"iam/pkg/logging"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/task/producer"
	"iam/pkg/util"
)

const checkerLayer = "Checker"

// type Checker ...
type Checker struct {
	stopChan chan struct{}
}

// NewChecker ...
func NewChecker() *Checker {
	return &Checker{
		stopChan: make(chan struct{}, 1),
	}
}

// Run ...
func (c *Checker) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		log.Info("I have to go...")
		log.Info("Stopping worker gracefully")
		c.Stop()
	}()

	// Start subject action alter event checker
	go NewSubjectActionAlterEventChecker(
		producer.NewRedisProducer(rbacEventQueue),
	).Run()

	// Start rmq cleaner
	go StartClean()

	c.Wait()
	log.Info("Shutting down")
}

// Stop ...
func (c *Checker) Stop() {
	defer log.Info("Server stopped")

	c.stopChan <- struct{}{}
}

// Wait blocks until server is shut down.
func (c *Checker) Wait() {
	<-c.stopChan
}

func StartClean() {
	logger := logging.GetWorkerLogger()

	cleaner := rmq.NewCleaner(connection)

	// run every 2 minutes
	for range time.Tick(2 * time.Minute) {
		logger.Info("Clean rmq begin")

		returned, err := cleaner.Clean()
		if err != nil {
			logger.Warnf("rmq failed to clean: %s", err)
			continue
		}
		logger.Infof("rmq cleaned %d", returned)
		logger.Info("Clean rmq end")
	}
}

func listReadyMessage() ([]string, error) {
	cli := redis.GetDefaultMQRedisClient()

	return cli.LRange(context.Background(), rbacEventQueueKey, 0, -1).Result()
}

type SubjectActionAlterEventChecker struct {
	service  service.SubjectActionAlterEventService
	producer producer.Producer

	stats stats
}

func NewSubjectActionAlterEventChecker(producer producer.Producer) *SubjectActionAlterEventChecker {
	return &SubjectActionAlterEventChecker{
		service:  service.NewSubjectActionAlterEventService(),
		producer: producer,

		stats: stats{
			startTime:           time.Now(),
			lastShowProcessTime: time.Now(),
		},
	}
}

func (c *SubjectActionAlterEventChecker) Run() {
	logger := logging.GetWorkerLogger()

	for range time.Tick(5 * time.Minute) {
		c.stats.totalCount += 1

		err := c.check()
		if err != nil {
			c.stats.failCount += 1
			logger.WithError(err).Error("check fail")

			// report to sentry
			util.ReportToSentry("SubjectActionAlterEventChecker.check fail",
				map[string]interface{}{
					"layer": checkerLayer,
					"error": err.Error(),
				},
			)
		} else {
			c.stats.successCount += 1
		}

		c.stats.lastShowProcessTime = time.Now()
		logger.Infof("checker processed total count: %d, success count: %d, fail count: %d, elapsed: %s",
			c.stats.totalCount, c.stats.successCount, c.stats.failCount, time.Since(c.stats.startTime))
	}
}

func (c *SubjectActionAlterEventChecker) check() error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(checkerLayer, "check")

	readyMessages, err := listReadyMessage()
	if err != nil {
		return errorWrapf(err, "listReadyMessage error")
	}

	// 用于避免重复发送消息
	readyMessageSet := set.NewStringSetWithValues(readyMessages)

	// 1. 查询更新时间超过30秒, status=0的记录
	updatedAt := time.Now().Add(-30 * time.Second).Unix()
	uuids, err := c.service.ListUUIDByStatusBeforeUpdatedAt(types.SubjectActionAlterEventStatusCreated, updatedAt)
	if err != nil {
		return errorWrapf(
			err,
			"service.ListUUIDByStatusBeforeUpdatedAt fail, status=`%d`, updatedAt=`%d`",
			types.SubjectActionAlterEventStatusCreated,
			updatedAt,
		)
	}

	missUUIDs := make([]string, 0, len(uuids))
	for _, uuid := range uuids {
		if !readyMessageSet.Has(uuid) {
			missUUIDs = append(missUUIDs, uuid)
		}
	}

	// 发送不在readyMessageSet中的消息
	if len(missUUIDs) > 0 {
		err = c.producer.Publish(missUUIDs...)
		if err != nil {
			return errorWrapf(err, "producer.Publish fail, uuids=`%s`", missUUIDs)
		}
	}

	if len(uuids) > 0 {
		// 更新状态为1
		err = c.service.BulkUpdateStatus(uuids, types.SubjectActionAlterEventStatusPushed)
		if err != nil {
			return errorWrapf(
				err,
				"service.BulkUpdateStatus fail, uuid=`%s`, status=`%d`",
				uuids,
				types.SubjectActionAlterEventStatusPushed,
			)
		}
	}

	// 2. 查询更新时间超过10分钟, status>0, check_count<3的记录
	updatedAt = time.Now().Add(-10 * time.Minute).Unix()
	maxCheckCount := int64(config.MaxSubjectActionAlterEventCheckCount)
	uuids, err = c.service.ListUUIDGreaterThanStatusLessThanCheckCountBeforeUpdatedAt(
		types.SubjectActionAlterEventStatusCreated,
		maxCheckCount,
		updatedAt,
	)
	if err != nil {
		return errorWrapf(
			err,
			"service.ListUUIDGreaterThanStatusLessThanCheckCountBeforeUpdatedAt fail,"+
				" status=`%d`, checkCount=`%d`, updatedAt=`%d`",
			types.SubjectActionAlterEventStatusCreated,
			maxCheckCount,
			updatedAt,
		)
	}

	missUUIDs = make([]string, 0, len(uuids))
	for _, uuid := range uuids {
		if !readyMessageSet.Has(uuid) {
			missUUIDs = append(missUUIDs, uuid)
		}
	}

	if len(missUUIDs) > 0 {
		// 发送消息
		err = c.producer.Publish(missUUIDs...)
		if err != nil {
			return errorWrapf(err, "producer.Publish fail, uuid=`%s`", missUUIDs)
		}

		// 更新check_count=check_count+1
		err = c.service.BulkIncrCheckCount(missUUIDs)
		if err != nil {
			return errorWrapf(err, "service.BulkIncrCheckCount fail, uuids=`%s`", missUUIDs)
		}
	}

	return nil
}
