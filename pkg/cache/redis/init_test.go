/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"iam/pkg/config"
)

func TestGetDefaultRedisClient(t *testing.T) {
	rds := GetDefaultRedisClient()
	assert.Nil(t, rds)
}

func TestInitRedisClient(t *testing.T) {
	// wrong config
	redisConfig := &config.Redis{
		ID:       ModeStandalone,
		Addr:     "1.1.1.1",
		PoolSize: 3,
	}

	InitRedisClient(true, redisConfig)

	rds := GetDefaultRedisClient()
	assert.NotNil(t, rds)

	_, err := rds.Ping(context.TODO()).Result()
	assert.Error(t, err)

	// TODO: add success init
}
