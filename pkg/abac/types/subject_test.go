/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
)

var _ = Describe("subject", func() {
	Describe("Subject cases", func() {
		It("NewSubject", func() {
			s := types.NewSubject()
			assert.NotNil(GinkgoT(), s)
			assert.Len(GinkgoT(), s.Attribute.Attribute, 0)
		})

		Describe("FillAttributes", func() {
			var s types.Subject
			BeforeEach(func() {
				s = types.NewSubject()
			})

			It("ok", func() {
				expectedPK := int64(123)
				expectedGroups := []types.SubjectGroup{
					{
						PK: 1,
					},
				}
				expectedDepts := []int64{1, 2, 3}
				s.FillAttributes(expectedPK, expectedGroups, expectedDepts)

				pk, err := s.Attribute.GetPK()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedPK, pk)

				gs, err := s.Attribute.GetGroups()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedGroups, gs)

				depts, err := s.Attribute.GetDepartments()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedDepts, depts)
			})
		})

		Describe("GetEffectGroupPKs", func() {
			var s types.Subject
			BeforeEach(func() {
				s = types.NewSubject()
			})

			It("error, not exists", func() {
				_, err := s.GetEffectGroupPKs()
				assert.Error(GinkgoT(), err)
			})

			It("empty", func() {
				expectedGroups := []types.SubjectGroup{}
				s.FillAttributes(123, expectedGroups, []int64{1, 2, 3})

				pks, err := s.GetEffectGroupPKs()
				assert.NoError(GinkgoT(), err)
				assert.Empty(GinkgoT(), pks)
			})

			It("all expired", func() {
				expectedGroups := []types.SubjectGroup{
					{
						PK:              1,
						PolicyExpiredAt: 0,
					},
				}
				s.FillAttributes(123, expectedGroups, []int64{1, 2, 3})

				pks, err := s.GetEffectGroupPKs()
				assert.NoError(GinkgoT(), err)
				assert.Empty(GinkgoT(), pks)
			})

			It("ok", func() {
				nowUnix := time.Now().Unix()
				expectedGroups := []types.SubjectGroup{
					{
						PK:              1,
						PolicyExpiredAt: nowUnix + 2000,
					},
					{
						PK:              2,
						PolicyExpiredAt: 0,
					},
					{
						PK:              3,
						PolicyExpiredAt: nowUnix + 2000,
					},
				}
				s.FillAttributes(123, expectedGroups, []int64{1, 2, 3})

				pks, err := s.GetEffectGroupPKs()
				assert.NoError(GinkgoT(), err)

				assert.Len(GinkgoT(), pks, 2)
				assert.Contains(GinkgoT(), pks, int64(1))
				assert.Contains(GinkgoT(), pks, int64(3))
			})
		})

		Describe("GetDepartmentPKs", func() {
			var s types.Subject
			BeforeEach(func() {
				s = types.NewSubject()
			})

			It("error, not exists", func() {
				_, err := s.GetDepartmentPKs()
				assert.Error(GinkgoT(), err)
			})

			It("empty", func() {
				expectedDepts := []int64{1, 2, 3}
				expectedGroups := []types.SubjectGroup{}
				s.FillAttributes(123, expectedGroups, expectedDepts)

				pks, err := s.GetDepartmentPKs()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedDepts, pks)
			})
		})
	})
})
