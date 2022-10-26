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
				expectedDepts := []int64{1, 2, 3}
				s.FillAttributes(expectedPK, expectedDepts)

				pk, err := s.Attribute.GetPK()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedPK, pk)

				depts, err := s.Attribute.GetDepartments()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedDepts, depts)
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
				s.FillAttributes(123, expectedDepts)

				pks, err := s.GetDepartmentPKs()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedDepts, pks)
			})
		})
	})
})
