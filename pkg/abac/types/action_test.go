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

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("action", func() {

	Describe("Action cases", func() {
		It("NewAction", func() {
			a := types.NewAction()
			assert.NotNil(GinkgoT(), a)
			assert.Len(GinkgoT(), a.Attribute.Attribute, 0)
		})

		Describe("FillAttributes", func() {
			var a types.Action
			BeforeEach(func() {
				a = types.NewAction()
			})

			It("ok", func() {
				expectedPK := int64(123)
				expectedArt := []types.ActionResourceType{
					{
						System: "bk_test",
						Type:   "obj",
					},
				}
				a.FillAttributes(expectedPK, expectedArt)

				pk, err := a.Attribute.GetPK()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedPK, pk)

				rt, err := a.Attribute.GetResourceTypes()
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expectedArt, rt)
			})
		})

		Describe("WithoutResourceType", func() {
			var a types.Action
			BeforeEach(func() {
				a = types.NewAction()
			})

			It("true", func() {
				assert.True(GinkgoT(), a.WithoutResourceType())
			})

			It("true, empty ResourceType", func() {
				a.Attribute.SetResourceTypes([]types.ActionResourceType{})
				assert.True(GinkgoT(), a.WithoutResourceType())
			})

			It("false", func() {
				a.Attribute.SetResourceTypes([]types.ActionResourceType{
					{
						System: "bk_test",
						Type:   "test",
					},
				})
				assert.False(GinkgoT(), a.WithoutResourceType())
			})

		})
	})

	Describe("ActionResourceType cases", func() {
	})
})
