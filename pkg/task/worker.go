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

	log "github.com/sirupsen/logrus"

	"iam/pkg/task/consumer"
	"iam/pkg/task/handler"
)

// Worker ...
type Worker struct {
	stopChan chan struct{}
}

// NewWorker ...
func NewWorker() *Worker {
	return &Worker{
		stopChan: make(chan struct{}, 1),
	}
}

// Run ...
func (s *Worker) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		log.Info("I have to go...")
		log.Info("Stopping worker gracefully")
		s.Stop()
	}()

	// Start rbac event consumer
	rbacEventConsumer := consumer.NewRedisConsumer(
		connection,
		rbacEventQueue,
		handler.NewGroupAlterMessageHandler(),
	)
	go rbacEventConsumer.Run(ctx)

	s.Wait()
	log.Info("Shutting down")
}

// Stop ...
func (s *Worker) Stop() {
	defer log.Info("Server stopped")

	s.stopChan <- struct{}{}
}

// Wait blocks until server is shut down.
func (s *Worker) Wait() {
	<-s.stopChan
}
