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
	"github.com/adjust/rmq/v4"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
)

const task = "task"

var (
	connection rmq.Connection
	queue      rmq.Queue
)

func InitRmqQueue(debugMode bool) {
	var err error

	connection, err = rmq.OpenConnectionWithRedisClient("bkiam", redis.GetDefaultRedisClient(), nil)
	if err != nil {
		log.WithError(err).Error("new rmq connection fail")
		if !debugMode {
			panic(err)
		}
	}

	queue, err = connection.OpenQueue("group_subject_action")
	if err != nil {
		log.WithError(err).Error("new rmq queue fail")
		if !debugMode {
			panic(err)
		}
	}
}
