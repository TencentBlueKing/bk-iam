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
)

var _ = Describe("NumericEquals", func() {
	var c *NumericEqualsCondition
	BeforeEach(func() {
		c = &NumericEqualsCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{1, 2},
			},
		}
	})

	It("new", func() {
		condition, err := newNumericEqualsCondition("ok", []interface{}{1, 2})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "NumericEquals", c.GetName())
	})

	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), c.Eval(intCtx(1)))
			assert.True(GinkgoT(), c.Eval(intCtx(2)))
		})

		It("false", func() {
			assert.False(GinkgoT(), c.Eval(intCtx(3)))
		})

		It("not number, false", func() {
			assert.False(GinkgoT(), c.Eval(strCtx("c")))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), c.Eval(listCtx{2, 3}))
			assert.False(GinkgoT(), c.Eval(listCtx{3, 4}))
		})

	})

	Describe("Translate", func() {
		It("fail, empty value", func() {
			c1, err := newNumericEqualsCondition("key", []interface{}{})
			assert.NoError(GinkgoT(), err)

			_, err = c1.Translate(true)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, eq", func() {
			expected := map[string]interface{}{
				"op":    "eq",
				"field": "key",
				"value": 1,
			}
			c, err := newNumericEqualsCondition("key", []interface{}{1})
			assert.NoError(GinkgoT(), err)

			c1, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c1)
		})

		It("ok, in", func() {
			expected := map[string]interface{}{
				"op":    "in",
				"field": "bk_cmdb.host.id",
				"value": []interface{}{1, 2},
			}
			c, err := newNumericEqualsCondition("bk_cmdb.host.id", []interface{}{1, 2})
			assert.NoError(GinkgoT(), err)

			c1, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c1)

		})

		It("ok, in, withSystem=False", func() {
			expected := map[string]interface{}{
				"op":    "in",
				"field": "host.id",
				"value": []interface{}{1, 2},
			}
			c, err := newNumericEqualsCondition("bk_cmdb.host.id", []interface{}{1, 2})
			assert.NoError(GinkgoT(), err)

			c1, err := c.Translate(false)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c1)

		})
	})

})
