/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/cache/impls"
	"iam/pkg/cache/memory"
)

var _ = Describe("Policy", func() {

	Describe("GetPoliciesAttrKeys", func() {
		var resource *types.Resource
		var policies []types.AuthPolicy
		BeforeEach(func() {
			resource = &types.Resource{
				System:    "bk_test",
				Type:      "host",
				ID:        "1",
				Attribute: nil,
			}
			policies = []types.AuthPolicy{}

			impls.LocalUnmarshaledExpressionCache = memory.NewMockCache(impls.UnmarshalExpression)
		})

		It("fail", func() {
			errExpr := `[{"system": "bk_test", "type": "host", "expression": 
{"OR": {"content": [{"NotExists": {"id": []}}]}}}]`
			policies = []types.AuthPolicy{
				{
					Expression:          errExpr,
					ExpressionSignature: "e77288fd872ccc464ac610272a56e7fb",
				},
			}
			_, err := GetPoliciesAttrKeys(resource, policies)
			assert.Error(GinkgoT(), err)
		})

		It("ok, one expression", func() {
			expr := `[{"system": "bk_job", "type": "job", 
"expression": {"OR": {"content": [{"StringEquals": {"id": ["job1"]}}]}}}, 
{"system": "bk_test", "type": "host", 
"expression": {"OR": {"content": [{"StringEquals": {"id": ["192.168.1.1"]}}, 
{"StringPrefix": {"path": ["/biz,1/"]}}]}}}]`
			policies = []types.AuthPolicy{
				{
					Expression:          expr,
					ExpressionSignature: "0e4fa20b19222af3110199099907e0c0",
				},
			}
			keys, err := GetPoliciesAttrKeys(resource, policies)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), keys, 2)
			assert.Contains(GinkgoT(), keys, "id")
			assert.Contains(GinkgoT(), keys, "path")
		})

		It("ok, same resource in list, will get the first one", func() {
			expr := `[{"system": "bk_test", "type": "host", 
"expression": {"OR": {"content": [{"StringEquals": {"area": ["job1"]}}]}}}, 
{"system": "bk_test", "type": "host", 
"expression": {"OR": {"content": [{"StringEquals": {"id": ["192.168.1.1"]}}, 
{"StringPrefix": {"path": ["/biz,1/"]}}]}}}]`
			policies = []types.AuthPolicy{
				{
					Expression:          expr,
					ExpressionSignature: "69d8e41a17d42661fced58040a272337",
				},
			}
			keys, err := GetPoliciesAttrKeys(resource, policies)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), keys, 1)
			assert.Contains(GinkgoT(), keys, "area")
			//assert.Contains(GinkgoT(), keys, "path")
			//assert.Contains(GinkgoT(), keys, "area")
		})

	})

	Describe("parseResourceConditionFromPolicies", func() {
		var resource *types.Resource
		var policies []types.AuthPolicy
		BeforeEach(func() {
			resource = &types.Resource{
				System:    "bk_test",
				Type:      "host",
				ID:        "1",
				Attribute: nil,
			}
			policies = []types.AuthPolicy{}
		})

		It("empty policies", func() {
			cs, err := parseResourceConditionFromPolicies(resource, policies)
			assert.NoError(GinkgoT(), err)
			assert.Empty(GinkgoT(), cs)
		})

		It("one ok policy", func() {
			expr := `[{"system": "bk_test", "type": "host", "expression": {"OR": {"content": [{"Any": {"id": []}}]}}}]`
			policies = []types.AuthPolicy{
				{
					Expression:          expr,
					ExpressionSignature: "ca306516f261c6127a8fd4c78d4c6b47",
				},
			}
			cs, err := parseResourceConditionFromPolicies(resource, policies)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), cs, 1)
		})

		It("two ok policy", func() {
			expr := `[{"system": "bk_test", "type": "host", "expression": {"OR": {"content": [{"Any": {"id": []}}]}}}]`
			policies = []types.AuthPolicy{
				{
					Expression:          expr,
					ExpressionSignature: "ca306516f261c6127a8fd4c78d4c6b47",
				},
				{
					Expression:          expr,
					ExpressionSignature: "ca306516f261c6127a8fd4c78d4c6b47",
				},
			}
			cs, err := parseResourceConditionFromPolicies(resource, policies)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), cs, 2)
		})

		It("two policies, one error", func() {
			expr := `[{"system": "bk_test", "type": "host", "expression": {"OR": {"content": [{"Any": {"id": []}}]}}}]`
			errExpr := `[{"system": "bk_test", "type": "host", "expression": {"OR": {"content": [{"NotExists": {"id": []}}]}}}]`
			policies = []types.AuthPolicy{
				{
					Expression:          expr,
					ExpressionSignature: "ca306516f261c6127a8fd4c78d4c6b47",
				},
				{
					Expression:          errExpr,
					ExpressionSignature: "4f7c070bc6a94e69ecb7205716857af9",
				},
			}
			_, err := parseResourceConditionFromPolicies(resource, policies)
			assert.Error(GinkgoT(), err)
		})

	})

	Describe("ParseResourceConditionFromExpression", func() {
		var resource *types.Resource
		BeforeEach(func() {
			resource = &types.Resource{}

			impls.LocalUnmarshaledExpressionCache = memory.NewMockCache(impls.UnmarshalExpression)
		})

		It("unmarshal fail", func() {
			_, err := ParseResourceConditionFromExpression(resource, "", "d41d8cd98f00b204e9800998ecf8427e")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "unmarshal")
		})

		It("empty resourceExpression", func() {
			_, err := ParseResourceConditionFromExpression(resource, "[]", "d751713988987e9331980363e24189ce")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "resource not match expression")
		})

		It("fail, not match", func() {
			expr := `[{"system": "bk_test", "type": "host", "expression": {"OR": {"content": [{"Any": {"id": []}}]}}}]`
			resource = &types.Resource{
				System:    "bk_aaa",
				Type:      "host",
				ID:        "1",
				Attribute: nil,
			}
			_, err := ParseResourceConditionFromExpression(resource, expr, "ca306516f261c6127a8fd4c78d4c6b47")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "resource not match expression")
		})

		It("single, hit", func() {
			expr := `[{"system": "bk_test", "type": "host", "expression": {"OR": {"content": [{"Any": {"id": []}}]}}}]`
			resource = &types.Resource{
				System:    "bk_test",
				Type:      "host",
				ID:        "1",
				Attribute: nil,
			}
			condition, err := ParseResourceConditionFromExpression(resource, expr, "ca306516f261c6127a8fd4c78d4c6b47")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "OR", condition.GetName())
			assert.Equal(GinkgoT(), []string{}, condition.GetKeys())
		})

		It("single, hit, but condition fail", func() {
			expr := `[{"system": "bk_test", "type": "host", "expression": {"OR": {"content": [{"NotExists": {"id": []}}]}}}]`
			resource = &types.Resource{
				System:    "bk_test",
				Type:      "host",
				ID:        "1",
				Attribute: nil,
			}
			_, err := ParseResourceConditionFromExpression(resource, expr, "4f7c070bc6a94e69ecb7205716857af9")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "expression parser error")
		})

	})

})
