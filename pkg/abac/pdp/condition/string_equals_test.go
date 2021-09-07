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

var _ = Describe("StringEquals", func() {

	var c *StringEqualsCondition
	BeforeEach(func() {
		c = &StringEqualsCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{"a", "b"},
			},
		}
	})

	It("new", func() {
		condition, err := newStringEqualsCondition("ok", []interface{}{"a", "b"})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "StringEquals", c.GetName())
	})

	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), c.Eval(strCtx("a")))
			assert.True(GinkgoT(), c.Eval(strCtx("b")))
		})

		It("false", func() {
			assert.False(GinkgoT(), c.Eval(strCtx("c")))
		})

		It("false, not string", func() {
			assert.False(GinkgoT(), c.Eval(intCtx(1)))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), c.Eval(listCtx{"a", "d"}))
			assert.False(GinkgoT(), c.Eval(listCtx{"e", "f"}))
		})

	})

	Describe("Translate", func() {
		It("fail, empty value", func() {
			c, err := newStringEqualsCondition("key", []interface{}{})
			assert.NoError(GinkgoT(), err)

			_, err = c.Translate()
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, single eq", func() {
			expected := map[string]interface{}{
				"op":    "eq",
				"field": "key",
				"value": "a",
			}
			c, err := newStringEqualsCondition("key", []interface{}{"a"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)

		})

		It("ok, multiple in", func() {
			expected := map[string]interface{}{
				"op":    "in",
				"field": "key",
				"value": []interface{}{"a", "b"},
			}
			c, err := newStringEqualsCondition("key", []interface{}{"a", "b"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})

})
