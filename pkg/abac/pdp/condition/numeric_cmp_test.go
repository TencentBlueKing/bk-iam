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
	"github.com/TencentBlueKing/iam-go-sdk/expression/eval"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/condition/operator"
)

var _ = Describe("NumericCompare", func() {
	var key string
	var values []interface{}

	var eqCondition Condition
	var gtCondition Condition
	var gteCondition Condition
	var ltCondition Condition
	var lteCondition Condition

	BeforeEach(func() {
		key = "ok"
		values = []interface{}{1, 2}
		eqCondition, _ = newNumericEqualsCondition(key, values)

		singleValues := []interface{}{2}
		gtCondition, _ = newNumericGreaterThanCondition(key, singleValues)
		gteCondition, _ = newNumericGreaterThanEqualsCondition(key, singleValues)
		ltCondition, _ = newNumericLessThanCondition(key, singleValues)
		lteCondition, _ = newNumericLessThanEqualsCondition(key, singleValues)
	})

	It("newNumericCompareCondition", func() {
		c, err := newNumericCompareCondition(key, values,
			"testName", "eq", "in", eval.Equal)
		assert.NoError(GinkgoT(), err)

		assert.NotNil(GinkgoT(), c)
		assert.Equal(GinkgoT(), "testName", c.GetName())
		keys := c.GetKeys()
		assert.Len(GinkgoT(), keys, 1)
		assert.Equal(GinkgoT(), key, keys[0])
	})

	Describe("New Specific compare condition", func() {
		It("eq", func() {
			c, err := newNumericEqualsCondition(key, values)
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), operator.NumericEquals, c.GetName())
		})

		It("gt", func() {
			c, err := newNumericGreaterThanCondition(key, values)
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), operator.NumericGt, c.GetName())
		})

		It("gte", func() {
			c, err := newNumericGreaterThanEqualsCondition(key, values)
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), operator.NumericGte, c.GetName())
		})

		It("lt", func() {
			c, err := newNumericLessThanCondition(key, values)
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), operator.NumericLt, c.GetName())
		})

		It("lte", func() {
			c, err := newNumericLessThanEqualsCondition(key, values)
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), operator.NumericLte, c.GetName())
		})
	})

	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), eqCondition.Eval(intCtx(1)))
			assert.True(GinkgoT(), eqCondition.Eval(intCtx(2)))

			assert.True(GinkgoT(), gtCondition.Eval(intCtx(3)))
			assert.True(GinkgoT(), gtCondition.Eval(intCtx(4)))

			assert.True(GinkgoT(), gteCondition.Eval(intCtx(2)))
			assert.True(GinkgoT(), gteCondition.Eval(intCtx(3)))
			assert.True(GinkgoT(), gteCondition.Eval(intCtx(4)))

			assert.True(GinkgoT(), ltCondition.Eval(intCtx(0)))
			assert.True(GinkgoT(), ltCondition.Eval(intCtx(1)))

			assert.True(GinkgoT(), lteCondition.Eval(intCtx(0)))
			assert.True(GinkgoT(), lteCondition.Eval(intCtx(1)))
			assert.True(GinkgoT(), lteCondition.Eval(intCtx(2)))
		})

		It("true, different type", func() {
			assert.True(GinkgoT(), eqCondition.Eval(int64Ctx(1)))
			assert.True(GinkgoT(), eqCondition.Eval(int64Ctx(2)))

			assert.True(GinkgoT(), gtCondition.Eval(int64Ctx(3)))
			assert.True(GinkgoT(), gtCondition.Eval(int64Ctx(4)))

			assert.True(GinkgoT(), gteCondition.Eval(int64Ctx(2)))
			assert.True(GinkgoT(), gteCondition.Eval(int64Ctx(3)))
			assert.True(GinkgoT(), gteCondition.Eval(int64Ctx(4)))

			assert.True(GinkgoT(), ltCondition.Eval(int64Ctx(0)))
			assert.True(GinkgoT(), ltCondition.Eval(int64Ctx(1)))

			assert.True(GinkgoT(), lteCondition.Eval(int64Ctx(0)))
			assert.True(GinkgoT(), lteCondition.Eval(int64Ctx(1)))
			assert.True(GinkgoT(), lteCondition.Eval(int64Ctx(2)))
		})

		It("false", func() {
			assert.False(GinkgoT(), eqCondition.Eval(intCtx(3)))

			assert.False(GinkgoT(), gtCondition.Eval(intCtx(0)))
			assert.False(GinkgoT(), gtCondition.Eval(intCtx(1)))
			assert.False(GinkgoT(), gtCondition.Eval(intCtx(2)))

			assert.False(GinkgoT(), gteCondition.Eval(intCtx(0)))
			assert.False(GinkgoT(), gteCondition.Eval(intCtx(1)))

			assert.False(GinkgoT(), ltCondition.Eval(intCtx(2)))
			assert.False(GinkgoT(), ltCondition.Eval(intCtx(3)))

			assert.False(GinkgoT(), lteCondition.Eval(intCtx(3)))
		})
		//
		It("not number, false", func() {
			assert.False(GinkgoT(), eqCondition.Eval(strCtx("c")))
			assert.False(GinkgoT(), gtCondition.Eval(strCtx("c")))
			assert.False(GinkgoT(), gteCondition.Eval(strCtx("c")))
			assert.False(GinkgoT(), ltCondition.Eval(strCtx("c")))
			assert.False(GinkgoT(), lteCondition.Eval(strCtx("c")))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), eqCondition.Eval(listCtx{2, 3}))
			assert.False(GinkgoT(), eqCondition.Eval(listCtx{3, 4}))

			// gt
			assert.True(GinkgoT(), gtCondition.Eval(listCtx{3, 4}))
			//  - one true, all true
			assert.True(GinkgoT(), gtCondition.Eval(listCtx{3, 1}))
			assert.False(GinkgoT(), gtCondition.Eval(listCtx{0, 1}))

			// gte
			assert.True(GinkgoT(), gteCondition.Eval(listCtx{3, 4}))
			//  - one true, all true
			assert.True(GinkgoT(), gteCondition.Eval(listCtx{1, 2}))
			//  - should all false
			assert.False(GinkgoT(), gteCondition.Eval(listCtx{0, -1}))

			// lt
			assert.True(GinkgoT(), ltCondition.Eval(listCtx{0, -1}))
			// - one true, all true
			assert.True(GinkgoT(), ltCondition.Eval(listCtx{0, 2}))
			// - should all false
			assert.False(GinkgoT(), ltCondition.Eval(listCtx{2, 3}))

			// lte
			assert.True(GinkgoT(), ltCondition.Eval(listCtx{1, 2}))
			// - one true, all true
			assert.True(GinkgoT(), ltCondition.Eval(listCtx{-1, 3}))
			// - should all false
			assert.False(GinkgoT(), ltCondition.Eval(listCtx{3, 4}))
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

		It("ok, gte", func() {
			expected := map[string]interface{}{
				"op":    "gte",
				"field": "bk_cmdb.host.id",
				"value": []interface{}{1, 2},
			}
			c, err := newNumericGreaterThanEqualsCondition("bk_cmdb.host.id", []interface{}{1, 2})
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
