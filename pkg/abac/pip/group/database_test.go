/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package group

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
	Describe("newDatabaseRetriever", func() {
		It("ok", func() {
			var ctl *gomock.Controller
			var patches *gomonkey.Patches

			ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()

			mockSubjectService := mock.NewMockSubjectService(ctl)
			patches.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
				return mockSubjectService
			})

			r := newDatabaseRetriever()
			assert.NotNil(GinkgoT(), r)

			ctl.Finish()
			patches.Reset()
		})
	})

	Describe("retrieve", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var mockSubjectService *mock.MockSubjectService

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()

			mockSubjectService = mock.NewMockSubjectService(ctl)
			patches.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
				return mockSubjectService
			})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("policyService.ListExpressionByPKs fail", func() {
			mockSubjectService.EXPECT().ListEffectThinSubjectGroups([]int64{123}).Return(
				nil, errors.New("list effect thin subject groups by pks fail"),
			).AnyTimes()

			r := newDatabaseRetriever()
			_, _, err := r.retrieve([]int64{123})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "list effect thin subject groups by pks fail", err.Error())
		})

		It("ok", func() {
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

			r := newDatabaseRetriever()
			subjectGroups, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectGroups, 2)

			assert.Equal(GinkgoT(), int64(1), subjectGroups[123][0].PK)
			assert.Equal(GinkgoT(), int64(2), subjectGroups[789][0].PK)

			assert.Len(GinkgoT(), missingPKs, 1)
			assert.Contains(GinkgoT(), missingPKs, int64(456))
		})
	})

	Describe("getMissingPKs", func() {
		var r *databaseRetriever
		BeforeEach(func() {
			r = &databaseRetriever{}
		})

		It("all missing", func() {
			missingPKs := r.getMissingPKs(
				[]int64{123, 456},
				map[int64][]types.ThinSubjectGroup{},
			)

			assert.Len(GinkgoT(), missingPKs, 2)
			assert.Contains(GinkgoT(), missingPKs, int64(123))
			assert.Contains(GinkgoT(), missingPKs, int64(456))
		})

		It("1 missing", func() {
			missingPKs := r.getMissingPKs(
				[]int64{123, 456},
				map[int64][]types.ThinSubjectGroup{
					456: {
						{
							PK:              1,
							PolicyExpiredAt: 4102444800,
						},
					},
				},
			)

			assert.Len(GinkgoT(), missingPKs, 1)
			assert.Contains(GinkgoT(), missingPKs, int64(123))
		})

		It("no missing", func() {
			missingPKs := r.getMissingPKs(
				[]int64{123, 456},
				map[int64][]types.ThinSubjectGroup{
					123: {
						{
							PK:              1,
							PolicyExpiredAt: 4102444800,
						},
					},
					456: {
						{
							PK:              2,
							PolicyExpiredAt: 4102444800,
						},
					},
				},
			)

			assert.Empty(GinkgoT(), missingPKs)
		})
	})

	Describe("setMissing", func() {
		a := databaseRetriever{}
		assert.Nil(GinkgoT(), a.setMissing(map[int64][]types.ThinSubjectGroup{}, []int64{123}))
	})
})
