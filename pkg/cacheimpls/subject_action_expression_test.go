/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"testing"
	"time"

	"iam/pkg/cache/redis"

	"github.com/stretchr/testify/assert"
)

func TestSubjectActionCacheKey_Key(t *testing.T) {
	key := SubjectActionCacheKey{
		SubjectPK: 1,
		ActionPK:  2,
	}

	assert.Equal(t, "1:2", key.Key())
}

func TestDeleteSubjectActionExpressionCache(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)
	SubjectActionExpressionCache = mockCache

	key := SubjectActionCacheKey{
		SubjectPK: 1,
		ActionPK:  2,
	}

	SubjectActionExpressionCache.Set(key, "abc", 0)

	var value string
	err := SubjectActionExpressionCache.Get(key, &value)
	assert.Nil(t, err)
	assert.Equal(t, "abc", value)

	err = SubjectActionExpressionCache.Delete(key)
	assert.Nil(t, err)

	err = SubjectActionExpressionCache.Get(key, &value)
	assert.NotNil(t, err)
}
