/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/bsm/redislock"

	"iam/pkg/cache/redis"
)

type subjectDistributedActionLocker struct {
	locker *redislock.Client
}

func newDistributedSubjectActionLocker() *subjectDistributedActionLocker {
	return &subjectDistributedActionLocker{
		locker: redislock.New(redis.GetDefaultRedisClient()),
	}
}

func (l *subjectDistributedActionLocker) acquire(
	ctx context.Context,
	subjectPK, actionPK int64,
) (lock *redislock.Lock, err error) {
	// Retry every 100ms, for up-to 3 minutes
	backoff := redislock.LinearBackoff(100 * time.Millisecond)
	key := fmt.Sprintf("iam:%s:sub_act_loc:%d:%d", redis.CacheVersion, subjectPK, actionPK)
	// Obtain lock with retry + custom deadline, ttl = 2 minutes
	lock, err = l.locker.Obtain(ctx, key, 2*time.Minute, &redislock.Options{
		RetryStrategy: backoff,
	})
	return
}
