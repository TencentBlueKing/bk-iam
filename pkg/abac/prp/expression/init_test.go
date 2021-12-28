/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package expression_test

import (
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/prp/expression"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("Init", func() {
	BeforeEach(func() {
		cacheimpls.LocalExpressionCache = gocache.New(1*time.Minute, 1*time.Minute)
		cacheimpls.ChangeListCache = redis.NewMockCache("changelist", 1*time.Minute)
		cacheimpls.ExpressionCache = redis.NewMockCache("expression", 1*time.Minute)
	})

	It("GetExpressionsFromCache", func() {
		ctl := gomock.NewController(GinkgoT())
		patches := gomonkey.NewPatches()

		mockPolicyService := mock.NewMockPolicyService(ctl)
		patches.ApplyFunc(service.NewPolicyService, func() service.PolicyService {
			return mockPolicyService
		})
		mockPolicyService.EXPECT().ListExpressionByPKs([]int64{123, 456}).Return(
			[]types.AuthExpression{
				{
					PK: 123,
				},
				{
					PK: 456,
				},
			},
			nil,
		).AnyTimes()

		expressions, err := expression.GetExpressionsFromCache(1, []int64{123, 456})
		assert.NoError(GinkgoT(), err)
		assert.Len(GinkgoT(), expressions, 2)

		patches.Reset()
		ctl.Finish()
	})

	It("BatchDeleteExpressionsFromCache", func() {
		err := expression.BatchDeleteExpressionsFromCache(map[int64][]int64{
			1: {123, 456},
		})
		assert.NoError(GinkgoT(), err)
	})

})
