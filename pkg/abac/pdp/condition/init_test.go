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

	"iam/pkg/abac/pdp/types"
)

var _ = Describe("Condition", func() {

	wantAndCondition := &AndCondition{
		baseLogicalCondition{
			content: []Condition{
				&StringEqualsCondition{
					baseCondition: baseCondition{
						Key:   "system",
						Value: []interface{}{"linux"},
					},
				},
				&StringPrefixCondition{
					baseCondition: baseCondition{
						Key:   "path",
						Value: []interface{}{"/biz,1/"},
					},
				},
			},
		},
	}

	Describe("newConditionFromInterface", func() {
		It("invalid value", func() {
			_, err := newConditionFromInterface(123)
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			data := map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"id": []interface{}{"1", "2"},
				},
			}
			want := &StringEqualsCondition{
				baseCondition: baseCondition{
					Key:   "id",
					Value: []interface{}{"1", "2"},
				},
			}
			c, err := newConditionFromInterface(data)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, c)
		})

		It("ok, a policyCondition", func() {

			data := types.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
						map[string]interface{}{"StringPrefix": map[string]interface{}{"path": []interface{}{"/biz,1/"}}},
					},
				},
			}

			c, err := newConditionFromInterface(data)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), wantAndCondition, c)

		})
	})

	Describe("NewConditionFromPolicyCondition", func() {
		It("empty policy condition", func() {
			_, err := NewConditionFromPolicyCondition(types.PolicyCondition{})
			assert.Error(GinkgoT(), err)
		})

		It("error operator", func() {
			_, err := NewConditionFromPolicyCondition(types.PolicyCondition{"notExists": map[string][]interface{}{}})
			assert.Error(GinkgoT(), err)

		})

		It("ok", func() {
			data := types.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
						map[string]interface{}{"StringPrefix": map[string]interface{}{"path": []interface{}{"/biz,1/"}}},
					},
				},
			}
			c, err := NewConditionFromPolicyCondition(data)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), wantAndCondition, c)
		})

	})

	Describe("removeSystemFromKey", func() {
		It("empty", func() {
			a := removeSystemFromKey("")
			assert.Equal(GinkgoT(), "", a)
		})

		It("no dot", func() {
			a := removeSystemFromKey("abc")
			assert.Equal(GinkgoT(), "abc", a)
		})

		It("ok", func() {
			a := removeSystemFromKey("bk_cmdb.host.id")
			assert.Equal(GinkgoT(), "host.id", a)
		})

		It("ok", func() {
			a := removeSystemFromKey("host.id")
			assert.Equal(GinkgoT(), "host.id", a)
		})

	})

})
