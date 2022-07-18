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
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("StringContains", func() {
	var c *StringContainsCondition
	BeforeEach(func() {
		c = &StringContainsCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{"hello", "world"},
			},
		}
	})

	It("new", func() {
		condition, err := newStringContainsCondition("ok", []interface{}{"a", "b"})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "StringContains", c.GetName())
	})

	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), c.Eval(strCtx("hello")))
			assert.True(GinkgoT(), c.Eval(strCtx("world")))
			assert.True(GinkgoT(), c.Eval(strCtx("hellothere")))
			assert.True(GinkgoT(), c.Eval(strCtx("hereworld")))
			assert.True(GinkgoT(), c.Eval(strCtx("hereworldtest")))
			assert.True(GinkgoT(), c.Eval(strCtx("hi, hello, hey")))
		})

		It("false", func() {
			assert.False(GinkgoT(), c.Eval(strCtx("c")))
			assert.False(GinkgoT(), c.Eval(strCtx("hell")))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), c.Eval(listCtx{"helloworld", "d"}))
			assert.True(GinkgoT(), c.Eval(listCtx{"d", "helloworld"}))
			assert.True(GinkgoT(), c.Eval(listCtx{"hello", "world"}))
			assert.True(GinkgoT(), c.Eval(listCtx{"hello"}))

			assert.False(GinkgoT(), c.Eval(listCtx{"e", "f"}))
		})

		It("false, attr value not string", func() {
			assert.False(GinkgoT(), c.Eval(intCtx(1)))
		})

		It("false, expr value not string", func() {
			c = &StringContainsCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{1},
				},
			}
			assert.False(GinkgoT(), c.Eval(strCtx("foobar")))
		})
	})

	Describe("Translate", func() {
		It("fail, empty value", func() {
			c, err := newStringContainsCondition("key", []interface{}{})
			assert.NoError(GinkgoT(), err)
			_, err = c.Translate(true)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, single", func() {
			expected := map[string]interface{}{
				"op":    "string_contains",
				"field": "key",
				"value": "hello",
			}
			c, err := newStringContainsCondition("key", []interface{}{"hello"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("ok, multiple, withSystem=True", func() {
			expected := map[string]interface{}{
				"op": "OR",
				"content": []map[string]interface{}{
					{
						"op":    "string_contains",
						"field": "bk_cmdb.host.path",
						"value": "hello",
					},
					{
						"op":    "string_contains",
						"field": "bk_cmdb.host.path",
						"value": "world",
					},
				},
			}

			c, err := newStringContainsCondition("bk_cmdb.host.path", []interface{}{"hello", "world"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("ok, multiple, withSystem=False", func() {
			expected := map[string]interface{}{
				"op": "OR",
				"content": []map[string]interface{}{
					{
						"op":    "string_contains",
						"field": "host.path",
						"value": "hello",
					},
					{
						"op":    "string_contains",
						"field": "host.path",
						"value": "world",
					},
				},
			}

			c, err := newStringContainsCondition("bk_cmdb.host.path", []interface{}{"hello", "world"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(false)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})
})
