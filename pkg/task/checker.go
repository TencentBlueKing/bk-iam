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
	"strconv"
	"time"

	"iam/pkg/config"
	"iam/pkg/logging"
	"iam/pkg/service"
	"iam/pkg/task/producer"

	"github.com/adjust/rmq/v4"
	log "github.com/sirupsen/logrus"
)

// type Checker struct { ...
type Checker struct {
	stopChan chan struct{}
}

// NewWorker ...
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

	// Start rmq cleaner
	go StartClean()

	// Start group event checker
	go NewGroupAlterEventChecker(
		producer.NewRedisProducer(rbacEventQueue),
	).Run()

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

type GroupAlterEventChecker struct {
	service  service.GroupAlterEventService
	producer producer.Producer
}

func NewGroupAlterEventChecker(producer producer.Producer) *GroupAlterEventChecker {
	return &GroupAlterEventChecker{
		service:  service.NewGroupAlterEventService(),
		producer: producer,
	}
}

func (c *GroupAlterEventChecker) Run() {
	logger := logging.GetWorkerLogger()

	maxCheckCount := int64(config.GetMaxGroupAlterEventCheckCount())
	// run every 5 minutes
	for range time.Tick(5 * time.Minute) {
		logger.Info("Check group alter event begin")

		createdAt := time.Now().Add(-5 * time.Minute).Unix()
		pks, err := c.service.ListPKLtCheckCountBeforeCreateAt(maxCheckCount, createdAt)
		if err != nil {
			logger.WithError(err).
				Errorf("failed to list pk by check times before create at, CheckCount=`%d`, createdAt=`%d`",
					maxCheckCount, createdAt)
			continue
		}

		logger.Infof("query group alter event, pks=`%+v`", pks)

		for _, pk := range pks {
			logger.Infof("do publish group alter event, pk=`%d`", pk)

			err := c.producer.Publish(strconv.FormatInt(pk, 10))
			if err != nil {
				logger.WithError(err).Errorf(
					"failed to publish pk, pk=`%d`", pk)
				continue
			}

			logger.Infof("publish group alter event, pk=`%d` done", pk)

			logger.Infof("do incr the checkCount of event pk=`%d` done", pk)

			err = c.service.IncrCheckCount(pk)
			if err != nil {
				logger.WithError(err).Errorf(
					"failed to incr the checkCount of event pk=`%d`", pk)
				continue
			}

			logger.Infof("incr the checkCount of event pk=`%d` done", pk)
		}

		logger.Info("Check group alter event end")
	}
}
