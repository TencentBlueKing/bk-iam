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
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

type intCtx int

func (c intCtx) GetAttr(key string) (interface{}, error) {
	return int(c), nil
}

func (c intCtx) HasResource(_type string) bool {
	return false
}

type int64Ctx int64

func (c int64Ctx) GetAttr(key string) (interface{}, error) {
	return int64(c), nil
}

func (c int64Ctx) HasResource(_type string) bool {
	return false
}

type strCtx string

func (c strCtx) GetAttr(key string) (interface{}, error) {
	return string(c), nil
}
func (c strCtx) HasResource(_type string) bool {
	return false
}

type boolCtx bool

func (c boolCtx) GetAttr(key string) (interface{}, error) {
	return bool(c), nil
}
func (c boolCtx) HasResource(_type string) bool {
	return false
}

type listCtx []interface{}

func (c listCtx) GetAttr(key string) (interface{}, error) {
	x := []interface{}(c)
	return x, nil
}
func (c listCtx) HasResource(_type string) bool {
	return false
}

type errCtx int

func (c errCtx) GetAttr(key string) (interface{}, error) {
	return nil, errors.New("missing key")
}
func (c errCtx) HasResource(_type string) bool {
	return false
}

type HitStrCtx string

func (c HitStrCtx) GetAttr(key string) (interface{}, error) {
	return string(c), nil
}
func (c HitStrCtx) HasResource(_type string) bool {
	return true
}

type MissStrCtx string

func (c MissStrCtx) GetAttr(key string) (interface{}, error) {
	return "", nil
}
func (c MissStrCtx) HasResource(_type string) bool {
	return false
}

type MapCtx map[string]interface{}

func (c MapCtx) GetAttr(key string) (interface{}, error) {
	value, ok := c[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return value, nil
}

// host.system  has key=system
func (c MapCtx) HasResource(_type string) bool {
	for k, _ := range c {
		if strings.HasPrefix(k, _type+".") {
			return true
		}
	}
	return false
}

var _ = Describe("BaseCondition", func() {

	Describe("GetKeys", func() {
		It("ok", func() {
			expectedKey := "test"

			c := baseCondition{
				Key:   expectedKey,
				Value: nil,
			}
			assert.Equal(GinkgoT(), []string{expectedKey}, c.GetKeys())
		})
	})

	Describe("GetValues", func() {
		It("ok", func() {
			expectedValues := []interface{}{1, "ab", 3}
			c := baseCondition{
				Key:   "test",
				Value: expectedValues,
			}
			assert.Equal(GinkgoT(), expectedValues, c.GetValues())
		})
	})

	Describe("forOr", func() {
		var fn func(interface{}, interface{}) bool
		var condition *baseCondition
		BeforeEach(func() {
			fn = func(a interface{}, b interface{}) bool {
				return a == b
			}
			condition = &baseCondition{
				Key: "key",
				Value: []interface{}{
					1,
					2,
				},
			}
		})

		It("GetAttr fail", func() {
			allowed := condition.forOr(errCtx(1), fn)
			assert.False(GinkgoT(), allowed)
		})

		It("single, hit one", func() {
			assert.True(GinkgoT(), condition.forOr(intCtx(1), fn))
			assert.True(GinkgoT(), condition.forOr(intCtx(2), fn))
		})
		It("single, missing ", func() {
			assert.False(GinkgoT(), condition.forOr(intCtx(3), fn))
		})

		It("list, hit one", func() {
			assert.True(GinkgoT(), condition.forOr(listCtx{2, 3}, fn))
		})
		It("list, missing", func() {
			assert.False(GinkgoT(), condition.forOr(listCtx{3, 4}, fn))
		})

	})

})
