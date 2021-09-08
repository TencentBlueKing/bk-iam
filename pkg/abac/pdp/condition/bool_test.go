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

var _ = Describe("Bool", func() {
	var c *BoolCondition
	BeforeEach(func() {
		c = &BoolCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{true},
			},
		}
	})

	It("New", func() {
		c := NewBoolCondition("test", false)
		assert.NotNil(GinkgoT(), c)
	})

	It("new", func() {
		condition, err := newBoolCondition("ok", []interface{}{true})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "Bool", c.GetName())
	})

	Context("Eval", func() {
		It("errCtx", func() {
			allowed := c.Eval(errCtx(1))
			assert.False(GinkgoT(), allowed)
		})

		It("true", func() {
			assert.True(GinkgoT(), c.Eval(boolCtx(true)))
		})

		It("false, multi attr values", func() {
			assert.False(GinkgoT(), c.Eval(listCtx{1, 2}))
		})

		It("false, attr value not bool", func() {
			assert.False(GinkgoT(), c.Eval(intCtx(1)))
		})

		It("false, multi expr values", func() {
			c = &BoolCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{true, true},
				},
			}
			assert.False(GinkgoT(), c.Eval(boolCtx(true)))
		})

		It("fail, exprValue not bool", func() {
			c = &BoolCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{123},
				},
			}
			assert.False(GinkgoT(), c.Eval(boolCtx(true)))
		})

	})

	Describe("Translate", func() {
		It("not support multi value", func() {
			c1, err := newBoolCondition("key", []interface{}{true, false})
			assert.NoError(GinkgoT(), err)

			_, err = c1.Translate()
			assert.Contains(GinkgoT(), err.Error(), "bool not support multi value")
		})

		It("ok", func() {
			expected := map[string]interface{}{
				"op":    "eq",
				"field": "ok",
				"value": true,
			}
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})

})
