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
	"iam/pkg/abac/types"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("attribute", func() {
	Describe("Raw Attribute", func() {
		var a types.Attribute
		BeforeEach(func() {
			a = types.Attribute{
				"a":           1,
				"b":           "two",
				"int64":       int64(64),
				"string":      "string",
				"int64_slice": []int64{1, 2, 3},
			}
		})

		It("Has", func() {
			assert.True(GinkgoT(), a.Has("a"))
			assert.False(GinkgoT(), a.Has("c"))
		})

		It("Delete", func() {
			a.Delete("a")
			a.Delete("c")
		})

		It("Keys", func() {
			s := a.Keys()
			assert.Contains(GinkgoT(), s, "a")
			assert.Contains(GinkgoT(), s, "b")

			a = types.Attribute{}
			s = a.Keys()
			assert.Empty(GinkgoT(), s)
		})

		It("Get", func() {
			v, ok := a.Get("a")
			assert.True(GinkgoT(), ok)
			assert.Equal(GinkgoT(), 1, v)

			_, ok = a.Get("c")
			assert.False(GinkgoT(), ok)
		})

		It("Set", func() {
			a.Set("c", 3)
			assert.True(GinkgoT(), a.Has("c"))
			v, ok := a.Get("c")
			assert.True(GinkgoT(), ok)
			assert.Equal(GinkgoT(), 3, v)
		})

		It("GetInt64", func() {
			_, err := a.GetInt64("not_exists")
			assert.Error(GinkgoT(), err)

			// wrong type
			_, err = a.GetInt64("string")
			assert.Error(GinkgoT(), err)

			// ok
			v, err := a.GetInt64("int64")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(64), v)
		})

		It("GetString", func() {
			_, err := a.GetString("not_exists")
			assert.Error(GinkgoT(), err)

			// wrong type
			_, err = a.GetString("int64")
			assert.Error(GinkgoT(), err)

			// ok
			v, err := a.GetString("string")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "string", v)
		})

		It("GetInt64Slice", func() {
			_, err := a.GetInt64Slice("not_exists")
			assert.Error(GinkgoT(), err)

			// wrong type
			_, err = a.GetInt64Slice("string")
			assert.Error(GinkgoT(), err)

			// ok
			v, err := a.GetInt64Slice("int64_slice")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []int64{1, 2, 3}, v)
		})
	})

	Describe("ActionAttribute", func() {
		var a *types.ActionAttribute
		BeforeEach(func() {
			a = types.NewActionAttribute()
		})

		It("New", func() {
			assert.Empty(GinkgoT(), a.Attribute)
		})

		It("GetPK", func() {
			_, err := a.GetPK()
			assert.Error(GinkgoT(), err)

			a.Attribute[types.PKAttrName] = int64(1)
			pk, err := a.GetPK()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), pk)
		})

		It("GetAuthType", func() {
			_, err := a.GetAuthType()
			assert.Error(GinkgoT(), err)

			a.Attribute[types.AuthTypeAttrName] = int64(1)
			pk, err := a.GetAuthType()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), pk)
		})

		It("SetPK", func() {
			a.SetPK(1)
			assert.True(GinkgoT(), a.Has(types.PKAttrName))
			v, err := a.GetPK()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), v)
		})

		It("SetAuthType", func() {
			a.SetAuthType(1)
			assert.True(GinkgoT(), a.Has(types.AuthTypeAttrName))
			v, err := a.GetAuthType()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), v)
		})

		It("GetResourceTypes", func() {
			_, err := a.GetResourceTypes()
			assert.Error(GinkgoT(), err)

			a.Attribute[types.ResourceTypeAttrName] = int64(1)
			_, err = a.GetResourceTypes()
			assert.Error(GinkgoT(), err)

			expectedRt := []types.ActionResourceType{
				{
					System: "bk_test",
					Type:   "job",
				},
			}

			a.Attribute[types.ResourceTypeAttrName] = expectedRt
			rt, err := a.GetResourceTypes()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expectedRt, rt)
		})

		It("SetResourceTypes", func() {
			expectedRt := []types.ActionResourceType{
				{
					System: "bk_test",
					Type:   "job",
				},
			}
			a.SetResourceTypes(expectedRt)
			assert.True(GinkgoT(), a.Has(types.ResourceTypeAttrName))

			v, err := a.GetResourceTypes()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expectedRt, v)
		})
	})

	Describe("SubjectAttribute", func() {
		var a *types.SubjectAttribute
		BeforeEach(func() {
			a = types.NewSubjectAttribute()
		})

		It("New", func() {
			assert.Empty(GinkgoT(), a.Attribute)
		})

		It("GetPK", func() {
			_, err := a.GetPK()
			assert.Error(GinkgoT(), err)

			a.Attribute[types.PKAttrName] = int64(1)
			pk, err := a.GetPK()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), pk)
		})

		It("SetPK", func() {
			a.SetPK(1)
			assert.True(GinkgoT(), a.Has(types.PKAttrName))
			v, err := a.GetPK()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), v)
		})

		It("GetDepartments", func() {
			_, err := a.GetDepartments()
			assert.Error(GinkgoT(), err)

			// invalid
			a.Attribute[types.DeptAttrName] = int64(1)
			_, err = a.GetDepartments()
			assert.Error(GinkgoT(), err)

			// valid
			a.Attribute[types.DeptAttrName] = []int64{1, 2, 3}
			v, err := a.GetDepartments()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []int64{1, 2, 3}, v)
		})

		It("SetDepartments", func() {
			expectedDepts := []int64{1, 2, 3}
			a.SetDepartments(expectedDepts)

			assert.True(GinkgoT(), a.Has(types.DeptAttrName))
			v, err := a.GetDepartments()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []int64{1, 2, 3}, v)
		})
	})
})
