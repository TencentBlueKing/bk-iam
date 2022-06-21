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
	"strconv"
	"testing"
	"time"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/redis"
	"iam/pkg/service"
	"iam/pkg/service/mock"
)

func Test_batchSetActionGroupPKs(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)

	GroupResourcePolicyCache = mockCache

	key := SystemResourceCacheKey{
		SystemID:             "test",
		ActionResourceTypePK: int64(1),
		ResourceTypePK:       int64(2),
		ResourceID:           "resource_test",
	}
	err := batchSetActionGroupPKs(key, map[int64][]int64{
		1: {1, 2, 3},
		2: {4, 5, 6},
	})
	assert.NoError(t, err)

	hashKeyField := redis.HashKeyField{
		Key:   key.Key(),
		Field: strconv.FormatInt(1, 10),
	}
	value, err := GroupResourcePolicyCache.HGet(hashKeyField)
	assert.NoError(t, err)

	var groupPKs []int64
	err = GroupResourcePolicyCache.Unmarshal(conv.StringToBytes(value), &groupPKs)
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, groupPKs)

	hashKeyField = redis.HashKeyField{
		Key:   key.Key(),
		Field: strconv.FormatInt(2, 10),
	}
	value, err = GroupResourcePolicyCache.HGet(hashKeyField)
	assert.NoError(t, err)

	err = GroupResourcePolicyCache.Unmarshal(conv.StringToBytes(value), &groupPKs)
	assert.NoError(t, err)
	assert.Equal(t, []int64{4, 5, 6}, groupPKs)
}

func Test_getResourceActionAuthorizedGroupPKsFromCache(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)

	GroupResourcePolicyCache = mockCache

	key := SystemResourceCacheKey{
		SystemID:             "test",
		ActionResourceTypePK: int64(1),
		ResourceTypePK:       int64(2),
		ResourceID:           "resource_test",
	}
	err := batchSetActionGroupPKs(key, map[int64][]int64{
		1: {1, 2, 3},
		2: {4, 5, 6},
	})
	assert.NoError(t, err)

	groupPKs, err := getResourceActionAuthorizedGroupPKsFromCache(key, 1)
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, groupPKs)

	groupPKs, err = getResourceActionAuthorizedGroupPKsFromCache(key, 2)
	assert.NoError(t, err)
	assert.Equal(t, []int64{4, 5, 6}, groupPKs)

	_, err = getResourceActionAuthorizedGroupPKsFromCache(key, 3)
	assert.Error(t, err)
}

func Test_retrieveResourceAuthorizedActionGroup(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockService := mock.NewMockGroupResourcePolicyService(ctl)
	mockService.EXPECT().GetAuthorizedActionGroupMap("test", int64(1), int64(2), "resource_test").Return(
		map[int64][]int64{
			1: {1, 2, 3},
			2: {4, 5, 6},
		}, nil).AnyTimes()

	patches := gomonkey.ApplyFunc(service.NewGroupResourcePolicyService,
		func() service.GroupResourcePolicyService {
			return mockService
		})
	defer patches.Reset()

	key := SystemResourceCacheKey{
		SystemID:             "test",
		ActionResourceTypePK: int64(1),
		ResourceTypePK:       int64(2),
		ResourceID:           "resource_test",
	}

	actionGroupPKs, err := retrieveResourceAuthorizedActionGroup(key)
	assert.NoError(t, err)
	assert.Equal(t, map[int64][]int64{
		1: {1, 2, 3},
		2: {4, 5, 6},
	}, actionGroupPKs)
}

func TestGetResourceActionAuthorizedGroupPKs(t *testing.T) {
	key := SystemResourceCacheKey{
		SystemID:             "test",
		ActionResourceTypePK: int64(1),
		ResourceTypePK:       int64(2),
		ResourceID:           "resource_test",
	}

	err := batchSetActionGroupPKs(key, map[int64][]int64{
		1: {1, 2, 3},
		2: {4, 5, 6},
	})
	assert.NoError(t, err)

	groupPKs, err := GetResourceActionAuthorizedGroupPKs("test", int64(2), int64(1), int64(2), "resource_test")
	assert.NoError(t, err)
	assert.Equal(t, []int64{4, 5, 6}, groupPKs)
}

func TestGetResourceActionAuthorizedGroupPKs_retrieve(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockService := mock.NewMockGroupResourcePolicyService(ctl)
	mockService.EXPECT().GetAuthorizedActionGroupMap("test", int64(1), int64(2), "resource_test").Return(
		map[int64][]int64{
			1: {1, 2, 3},
			2: {4, 5, 6},
		}, nil).Times(1)

	patches := gomonkey.ApplyFunc(service.NewGroupResourcePolicyService,
		func() service.GroupResourcePolicyService {
			return mockService
		})
	defer patches.Reset()

	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)

	GroupResourcePolicyCache = mockCache

	groupPKs, err := GetResourceActionAuthorizedGroupPKs("test", int64(2), int64(1), int64(2), "resource_test")
	assert.NoError(t, err)
	assert.Equal(t, []int64{4, 5, 6}, groupPKs)

	groupPKs, err = GetResourceActionAuthorizedGroupPKs("test", int64(1), int64(1), int64(2), "resource_test")
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, groupPKs)
}
