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

	"iam/pkg/logging"
	"iam/pkg/service"
	"iam/pkg/task/producer"

	"github.com/adjust/rmq/v4"
	log "github.com/sirupsen/logrus"
)

const maxCheckTimes = 3 // TODO 从配置文件中读取

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
	go NewGroupAlterEventChecker().Run()

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
	cleaner := rmq.NewCleaner(connection)

	// run every 2 minutes
	for range time.Tick(2 * time.Minute) {
		returned, err := cleaner.Clean()
		if err != nil {
			log.Printf("failed to clean: %s", err)
			continue
		}
		log.Printf("cleaned %d", returned)
	}
}

type GroupAlterEventChecker struct {
	service  service.GroupAlterEventService
	producer producer.Producer
}

func NewGroupAlterEventChecker() *GroupAlterEventChecker {
	return &GroupAlterEventChecker{
		service:  service.NewGroupAlterEventService(),
		producer: producer.NewRedisProducer(rbacEventQueue),
	}
}

func (c *GroupAlterEventChecker) Run() {
	logger := logging.GetWorkerLogger()

	// run every 5 minutes
	for range time.Tick(5 * time.Minute) {
		createdAt := time.Now().Add(-5 * time.Minute).Unix()
		pks, err := c.service.ListPKByCheckTimesBeforeCreateAt(maxCheckTimes, createdAt)
		if err != nil {
			logger.WithError(err).
				Errorf("failed to list pk by check times before create at, checkTimes=`%d`, createdAt=`%d`",
					maxCheckTimes, createdAt)
			continue
		}

		for _, pk := range pks {
			err := c.producer.Publish(strconv.FormatInt(pk, 10))
			if err != nil {
				logger.WithError(err).Errorf(
					"failed to publish pk, pk=`%d`", pk)
				continue
			}

			err = c.service.IncrCheckTimes(pk)
			if err != nil {
				logger.WithError(err).Errorf(
					"failed to incr pk, pk=`%d`", pk)
				continue
			}
		}
	}
}
