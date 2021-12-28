/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy_test

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/prp/policy"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("Init", func() {

	var ctl *gomock.Controller
	var patches *gomonkey.Patches
	BeforeEach(func() {
		cacheimpls.LocalPolicyCache = gocache.New(1*time.Minute, 1*time.Minute)
		cacheimpls.ChangeListCache = redis.NewMockCache("changelist", 1*time.Minute)
		cacheimpls.PolicyCache = redis.NewMockCache("policy", 1*time.Minute)

		ctl = gomock.NewController(GinkgoT())
		patches = gomonkey.NewPatches()
	})
	AfterEach(func() {
		ctl.Finish()
		patches.Reset()
	})

	It("GetPoliciesFromCache", func() {
		mockPolicyService := mock.NewMockPolicyService(ctl)
		patches.ApplyFunc(service.NewPolicyService, func() service.PolicyService {
			return mockPolicyService
		})
		mockPolicyService.EXPECT().ListAuthBySubjectAction([]int64{123, 456, 789}, int64(1)).Return(
			[]types.AuthPolicy{
				{
					PK:        1,
					SubjectPK: 123,
				},
				{
					PK:        2,
					SubjectPK: 789,
				},
				{
					PK:        3,
					SubjectPK: 123,
				},
			},
			nil,
		).AnyTimes()

		policies, err := policy.GetPoliciesFromCache("test", 1, []int64{123, 456, 789})
		assert.NoError(GinkgoT(), err)
		assert.Len(GinkgoT(), policies, 3)

	})

	It("DeleteSystemSubjectPKsFromCache", func() {
		mockActionService := mock.NewMockActionService(ctl)
		mockActionService.EXPECT().ListThinActionBySystem("test").Return(
			[]types.ThinAction{}, nil,
		).AnyTimes()

		patches.ApplyFunc(service.NewActionService, func() service.ActionService {
			return mockActionService
		})

		err := policy.DeleteSystemSubjectPKsFromCache("test", []int64{123, 456})
		assert.NoError(GinkgoT(), err)
	})

	It("BatchDeleteSystemSubjectPKsFromCache", func() {
		mockActionService := mock.NewMockActionService(ctl)
		mockActionService.EXPECT().ListThinActionBySystem(gomock.Any()).Return(
			[]types.ThinAction{}, nil,
		).AnyTimes()
		patches.ApplyFunc(service.NewActionService, func() service.ActionService {
			return mockActionService
		})

		err := policy.BatchDeleteSystemSubjectPKsFromCache([]string{"test", "test1"}, []int64{123, 456})
		assert.NoError(GinkgoT(), err)
	})

})
