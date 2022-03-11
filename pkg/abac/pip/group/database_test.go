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

// var _ = Describe("Database", func() {
// 	Describe("newDatabaseRetriever", func() {
// 		It("ok", func() {
// 			var ctl *gomock.Controller
// 			var patches *gomonkey.Patches

// 			ctl = gomock.NewController(GinkgoT())
// 			patches = gomonkey.NewPatches()

// 			mockPolicyService := mock.NewMockPolicyService(ctl)
// 			patches.ApplyFunc(service.NewPolicyService, func() service.PolicyService {
// 				return mockPolicyService
// 			})

// 			r := newDatabaseRetriever()
// 			assert.NotNil(GinkgoT(), r)

// 			ctl.Finish()
// 			patches.Reset()
// 		})
// 	})

// 	Describe("retrieve", func() {
// 		var ctl *gomock.Controller
// 		var patches *gomonkey.Patches
// 		var mockPolicyService *mock.MockPolicyService

// 		BeforeEach(func() {
// 			ctl = gomock.NewController(GinkgoT())
// 			patches = gomonkey.NewPatches()

// 			mockPolicyService = mock.NewMockPolicyService(ctl)
// 			patches.ApplyFunc(service.NewPolicyService, func() service.PolicyService {
// 				return mockPolicyService
// 			})
// 		})
// 		AfterEach(func() {
// 			ctl.Finish()
// 			patches.Reset()
// 		})

// 		It("policyService.ListExpressionByPKs fail", func() {
// 			mockPolicyService.EXPECT().ListExpressionByPKs([]int64{123}).Return(
// 				nil, errors.New("list expression by pks fail"),
// 			).AnyTimes()

// 			r := newDatabaseRetriever()
// 			_, _, err := r.retrieve([]int64{123})
// 			assert.Error(GinkgoT(), err)
// 			assert.Equal(GinkgoT(), "list expression by pks fail", err.Error())
// 		})

// 		It("ok", func() {
// 			mockPolicyService.EXPECT().ListExpressionByPKs([]int64{123, 456, 789}).Return(
// 				[]types.AuthExpression{
// 					{
// 						PK: 123,
// 					},
// 					{
// 						PK: 789,
// 					},
// 				},
// 				nil,
// 			).AnyTimes()

// 			r := newDatabaseRetriever()
// 			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
// 			assert.NoError(GinkgoT(), err)
// 			assert.Len(GinkgoT(), expressions, 2)
// 			assert.Equal(GinkgoT(), int64(123), expressions[0].PK)
// 			assert.Equal(GinkgoT(), int64(789), expressions[1].PK)

// 			assert.Len(GinkgoT(), missingPKs, 1)
// 			assert.Contains(GinkgoT(), missingPKs, int64(456))
// 		})
// 	})

// 	Describe("getMissingPKs", func() {
// 		var r *databaseRetriever
// 		BeforeEach(func() {
// 			r = &databaseRetriever{}
// 		})

// 		It("all missing", func() {
// 			missingPKs := r.getMissingPKs(
// 				[]int64{123, 456},
// 				[]types.AuthExpression{},
// 			)

// 			assert.Len(GinkgoT(), missingPKs, 2)
// 			assert.Contains(GinkgoT(), missingPKs, int64(123))
// 			assert.Contains(GinkgoT(), missingPKs, int64(456))
// 		})

// 		It("1 missing", func() {
// 			missingPKs := r.getMissingPKs(
// 				[]int64{123, 456},
// 				[]types.AuthExpression{
// 					{
// 						PK: 456,
// 					},
// 				},
// 			)

// 			assert.Len(GinkgoT(), missingPKs, 1)
// 			assert.Contains(GinkgoT(), missingPKs, int64(123))
// 		})

// 		It("no missing", func() {
// 			missingPKs := r.getMissingPKs(
// 				[]int64{123, 456},
// 				[]types.AuthExpression{
// 					{
// 						PK: 456,
// 					},
// 					{
// 						PK: 123,
// 					},
// 				},
// 			)

// 			assert.Empty(GinkgoT(), missingPKs)
// 		})
// 	})

// 	Describe("setMissing", func() {
// 		a := databaseRetriever{}
// 		assert.Nil(GinkgoT(), a.setMissing([]types.AuthExpression{}, []int64{123}))
// 	})
// })
