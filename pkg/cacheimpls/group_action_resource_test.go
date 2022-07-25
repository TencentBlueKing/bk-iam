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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/redis"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

func Test_GetGroupActionAuthorizedResource(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)

	GroupActionResourceCache = mockCache

	key := GroupActionCacheKey{
		GroupPK:  int64(1),
		ActionPK: int64(2),
	}
	GroupActionResourceCache.Set(key, map[int64][]string{
		1: {"1", "2"},
	}, 10*time.Minute)

	resources, err := GetGroupActionAuthorizedResource(int64(1), int64(2))
	assert.NoError(t, err)
	assert.Equal(t, map[int64][]string{
		1: {"1", "2"},
	}, resources)
}

func Test_retrieveGroupActionAuthorizedResource(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)

	GroupActionResourceCache = mockCache

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockGroupResourcePolicySvc := mock.NewMockGroupResourcePolicyService(ctl)
	mockGroupResourcePolicySvc.EXPECT().
		ListResourceByGroupAction(int64(1), "system", int64(2), int64(3)).
		Return([]types.Resource{
			{ResourceTypePK: 1, ResourceID: "1"},
			{ResourceTypePK: 2, ResourceID: "2"},
			{ResourceTypePK: 1, ResourceID: "3"},
		}, nil)

	patches := gomonkey.ApplyFunc(GetAction, func(_ int64) (types.ThinAction, error) {
		return types.ThinAction{
			System: "system",
			ID:     "action",
			PK:     int64(2),
		}, nil
	})
	patches.ApplyFunc(GetActionDetail, func(_, _ string) (types.ActionDetail, error) {
		return types.ActionDetail{
			ResourceTypes: []types.ThinActionResourceType{
				{System: "system", ID: "resource_type"},
			},
		}, nil
	})
	patches.ApplyFunc(GetLocalResourceTypePK, func(systemID, resourceTypeID string) (pk int64, err error) {
		return int64(3), nil
	})
	patches.ApplyFunc(service.NewGroupResourcePolicyService, func() service.GroupResourcePolicyService {
		return mockGroupResourcePolicySvc
	})
	defer patches.Reset()

	resources, err := GetGroupActionAuthorizedResource(int64(1), int64(2))
	assert.NoError(t, err)
	assert.Equal(t, map[int64][]string{
		1: {"1", "3"},
		2: {"2"},
	}, resources)
}
