/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("Database", func() {
	It("newDatabaseRetriever", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches

		ctl = gomock.NewController(GinkgoT())
		patches = gomonkey.NewPatches()

		mockPolicyService := mock.NewMockPolicyService(ctl)
		patches.ApplyFunc(service.NewPolicyService, func() service.PolicyService {
			return mockPolicyService
		})

		r := newDatabaseRetriever(1)
		assert.NotNil(GinkgoT(), r)

		ctl.Finish()
		patches.Reset()
	})

	Describe("retrieve", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var mockPolicyService *mock.MockPolicyService

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()

			mockPolicyService = mock.NewMockPolicyService(ctl)
			patches.ApplyFunc(service.NewPolicyService, func() service.PolicyService {
				return mockPolicyService
			})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("policyService.ListAuthBySubjectAction fail", func() {
			mockPolicyService.EXPECT().ListAuthBySubjectAction([]int64{123}, int64(1)).Return(
				nil, errors.New("list policy by subject_pks fail"),
			).AnyTimes()

			r := newDatabaseRetriever(1)
			_, _, err := r.retrieve([]int64{123})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "list policy by subject_pks fail", err.Error())
		})

		It("ok", func() {
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

			r := newDatabaseRetriever(1)
			expressions, missingSubjectPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Equal(GinkgoT(), int64(1), expressions[0].PK)
			assert.Equal(GinkgoT(), int64(2), expressions[1].PK)
			assert.Equal(GinkgoT(), int64(3), expressions[2].PK)
			assert.Equal(GinkgoT(), int64(123), expressions[2].SubjectPK)

			assert.Len(GinkgoT(), missingSubjectPKs, 1)
			assert.Contains(GinkgoT(), missingSubjectPKs, int64(456))
		})
	})

	Describe("getMissingPKs", func() {
		var r *databaseRetriever
		BeforeEach(func() {
			r = &databaseRetriever{}
		})

		It("all missing", func() {
			missingSubjectPKs := r.getMissingPKs(
				[]int64{123, 456},
				[]types.AuthPolicy{},
			)

			assert.Len(GinkgoT(), missingSubjectPKs, 2)
			assert.Contains(GinkgoT(), missingSubjectPKs, int64(123))
			assert.Contains(GinkgoT(), missingSubjectPKs, int64(456))
		})

		It("1 missing", func() {
			missingSubjectPKs := r.getMissingPKs(
				[]int64{123, 456},
				[]types.AuthPolicy{
					{
						SubjectPK: 456,
					},
				},
			)

			assert.Len(GinkgoT(), missingSubjectPKs, 1)
			assert.Contains(GinkgoT(), missingSubjectPKs, int64(123))
		})

		It("no missing", func() {
			missingSubjectPKs := r.getMissingPKs(
				[]int64{123, 456},
				[]types.AuthPolicy{
					{
						SubjectPK: 456,
					},
					{
						SubjectPK: 123,
					},
				},
			)

			assert.Empty(GinkgoT(), missingSubjectPKs)
		})
	})

	Describe("setMissing", func() {
		a := databaseRetriever{}
		assert.Nil(GinkgoT(), a.setMissing([]types.AuthPolicy{}, []int64{123}))
	})
})
