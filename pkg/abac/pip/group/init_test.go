/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package group_test

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pip/group"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
	svctypes "iam/pkg/service/types"
)

var _ = Describe("Init", func() {
	BeforeEach(func() {
		cacheimpls.LocalSubjectGroupsCache = gocache.New(1*time.Minute, 1*time.Minute)
		cacheimpls.ChangeListCache = redis.NewMockCache("changelist", 1*time.Minute)
		cacheimpls.SubjectGroupCache = redis.NewMockCache("subject_group", 1*time.Minute)
	})

	It("GetSubjectGroupsFromCache", func() {
		ctl := gomock.NewController(GinkgoT())
		patches := gomonkey.NewPatches()

		mockSubjectService := mock.NewMockSubjectService(ctl)
		patches.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockSubjectService
		})
		mockSubjectService.EXPECT().ListEffectThinSubjectGroups([]int64{123, 456, 789}).Return(
			map[int64][]types.ThinSubjectGroup{
				123: {
					{
						PK:              1,
						PolicyExpiredAt: 4102444800,
					},
				},
				789: {
					{
						PK:              2,
						PolicyExpiredAt: 4102444800,
					},
				},
			},
			nil,
		).AnyTimes()

		subjectGroups, err := group.GetSubjectGroupsFromCache(svctypes.DepartmentType, []int64{123, 456, 789})
		assert.NoError(GinkgoT(), err)
		// NOTE: here is 3, the 456 will map to empty slice
		assert.Len(GinkgoT(), subjectGroups, 3)

		patches.Reset()
		ctl.Finish()
	})

	It("BatchDeleteSubjectGroupsFromCache", func() {
		err := group.BatchDeleteSubjectGroupsFromCache(svctypes.DepartmentType, []int64{123})
		assert.NoError(GinkgoT(), err)
	})
})
