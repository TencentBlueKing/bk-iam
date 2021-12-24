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

var _ = Describe("StringPrefix", func() {
	var c *StringPrefixCondition
	BeforeEach(func() {
		c = &StringPrefixCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{"/biz,1/", "/biz,2/", "hello"},
			},
		}
	})

	It("new", func() {
		condition, err := newStringPrefixCondition("ok", []interface{}{"a", "b"})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "StringPrefix", c.GetName())
	})

	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), c.Eval(strCtx("/biz,1/set,2/")))
			assert.True(GinkgoT(), c.Eval(strCtx("/biz,2/set,3/")))

			assert.True(GinkgoT(), c.Eval(strCtx("hello")))
			assert.True(GinkgoT(), c.Eval(strCtx("helloworld")))
		})

		It("false", func() {
			assert.False(GinkgoT(), c.Eval(strCtx("c")))
			assert.False(GinkgoT(), c.Eval(strCtx("hell")))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), c.Eval(listCtx{"/biz,1/set,2/", "d"}))
			assert.True(GinkgoT(), c.Eval(listCtx{"hello"}))

			assert.False(GinkgoT(), c.Eval(listCtx{"e", "f"}))
		})

		It("false, attr value not string", func() {
			assert.False(GinkgoT(), c.Eval(intCtx(1)))
		})

		It("false, expr value not string", func() {
			c = &StringPrefixCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{1},
				},
			}
			assert.False(GinkgoT(), c.Eval(strCtx("/biz,1/set,2/")))
		})

		It("_bk_iam_path_", func() {
			c = &StringPrefixCondition{
				baseCondition{
					Key:   "bk_test" + iamPathSuffix,
					Value: []interface{}{"/biz,1/set,*/"},
				},
			}
			assert.True(GinkgoT(), c.Eval(strCtx("/biz,1/set,2/")))
			assert.False(GinkgoT(), c.Eval(strCtx("/biz,1/module,2/")))
		})
	})

	Describe("Translate", func() {
		It("fail, empty value", func() {
			c, err := newStringPrefixCondition("key", []interface{}{})
			assert.NoError(GinkgoT(), err)
			_, err = c.Translate(true)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, single", func() {
			expected := map[string]interface{}{
				"op":    "starts_with",
				"field": "key",
				"value": "/biz,1/set,1/",
			}
			c, err := newStringPrefixCondition("key", []interface{}{"/biz,1/set,1/"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("ok, multiple or", func() {
			expected := map[string]interface{}{
				"op": "OR",
				"content": []map[string]interface{}{
					{
						"op":    "starts_with",
						"field": "bk_cmdb.host.path",
						"value": "/biz,1/set,1/",
					},
					{
						"op":    "starts_with",
						"field": "bk_cmdb.host.path",
						"value": "/biz,2/set,2/",
					},
				},
			}

			c, err := newStringPrefixCondition("bk_cmdb.host.path", []interface{}{"/biz,1/set,1/", "/biz,2/set,2/"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("ok, multiple or, withSystem=False", func() {
			expected := map[string]interface{}{
				"op": "OR",
				"content": []map[string]interface{}{
					{
						"op":    "starts_with",
						"field": "host.path",
						"value": "/biz,1/set,1/",
					},
					{
						"op":    "starts_with",
						"field": "host.path",
						"value": "/biz,2/set,2/",
					},
				},
			}

			c, err := newStringPrefixCondition("bk_cmdb.host.path", []interface{}{"/biz,1/set,1/", "/biz,2/set,2/"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(false)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})

})
