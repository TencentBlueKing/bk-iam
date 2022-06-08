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
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_setMissingSystemSubjectGroup(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)
	SystemSubjectGroupCache = mockCache

	setMissingSystemSubjectGroup("systemID", map[int64][]types.ThinSubjectGroup{
		1: {
			{
				PK:              2,
				PolicyExpiredAt: 2,
			},
		},
	}, []int64{1, 2, 3})

	var sg []types.ThinSubjectGroup
	SystemSubjectGroupCache.GetInto(SystemSubjectPKCacheKey{
		SystemID:  "systemID",
		SubjectPK: 1,
	}, &sg, nil)

	assert.Equal(t, []types.ThinSubjectGroup{
		{
			PK:              2,
			PolicyExpiredAt: 2,
		},
	}, sg)

	SystemSubjectGroupCache.GetInto(SystemSubjectPKCacheKey{
		SystemID:  "systemID",
		SubjectPK: 2,
	}, &sg, nil)

	assert.Len(t, sg, 0)
}

func TestBatchDeleteSystemSubjectGroupCache(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)
	SystemSubjectGroupCache = mockCache

	setMissingSystemSubjectGroup("systemID", map[int64][]types.ThinSubjectGroup{
		1: {
			{
				PK:              2,
				PolicyExpiredAt: 2,
			},
		},
	}, []int64{1, 2, 3})

	err := BatchDeleteSystemSubjectGroupCache([]string{"systemID"}, []int64{1, 2, 3})
	assert.NoError(t, err)

	var sg []types.ThinSubjectGroup
	err = SystemSubjectGroupCache.Get(SystemSubjectPKCacheKey{
		SystemID:  "systemID",
		SubjectPK: 2,
	}, &sg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func Test_batchGetSystemSubjectGroups(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)
	SystemSubjectGroupCache = mockCache

	setMissingSystemSubjectGroup("systemID", map[int64][]types.ThinSubjectGroup{
		1: {
			{
				PK:              2,
				PolicyExpiredAt: 2,
			},
		},
	}, []int64{1, 2, 3})

	subjectGroups, notExistCachePKs, err := batchGetSystemSubjectGroups("systemID", []int64{1, 2, 4})
	assert.NoError(t, err)
	assert.Equal(t, []types.ThinSubjectGroup{{
		PK:              2,
		PolicyExpiredAt: 2,
	}}, subjectGroups)
	assert.Equal(t, []int64{4}, notExistCachePKs)
}

func TestListSystemSubjectEffectGroups(t *testing.T) {
	expiration := 5 * time.Minute
	mockCache := redis.NewMockCache("mockCache", expiration)
	SystemSubjectGroupCache = mockCache

	setMissingSystemSubjectGroup("systemID", map[int64][]types.ThinSubjectGroup{
		1: {
			{
				PK:              2,
				PolicyExpiredAt: 2,
			},
		},
	}, []int64{1, 2, 3})

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockService := mock.NewMockGroupService(ctl)
	mockService.EXPECT().ListEffectThinSubjectGroups("systemID", []int64{4}).Return(
		map[int64][]types.ThinSubjectGroup{4: {{
			PK:              5,
			PolicyExpiredAt: 5,
		}}}, nil).AnyTimes()

	patches := gomonkey.ApplyFunc(service.NewGroupService,
		func() service.GroupService {
			return mockService
		})
	defer patches.Reset()

	subjectGroups, err := ListSystemSubjectEffectGroups("systemID", []int64{1, 2, 4})
	assert.NoError(t, err)
	assert.Equal(t, []types.ThinSubjectGroup{{
		PK:              2,
		PolicyExpiredAt: 2,
	}, {
		PK:              5,
		PolicyExpiredAt: 5,
	}}, subjectGroups)
}
