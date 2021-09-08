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

var _ = Describe("Or", func() {
	wantOrCondition := &OrCondition{
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
	}

	var c *OrCondition
	BeforeEach(func() {
		c1, _ := newStringEqualsCondition("k1", []interface{}{"a", "b"})
		c2, _ := newNumericEqualsCondition("k2", []interface{}{123})
		c = &OrCondition{
			[]Condition{
				c1,
				c2,
			},
		}
	})

	Describe("New", func() {
		It("New ok", func() {
			c := NewOrCondition([]Condition{})
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), "OR", c.GetName())
		})

		It("wrong key", func() {
			_, err := newOrCondition("wrong", []interface{}{"abc"})
			assert.Error(GinkgoT(), err)

		})

		It("ok", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				map[string]interface{}{"StringPrefix": map[string]interface{}{"path": []interface{}{"/biz,1/"}}},
			}
			c, err := newOrCondition("content", data)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), wantOrCondition, c)
		})

		It("fail", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				1,
			}
			_, err := newOrCondition("content", data)
			assert.Error(GinkgoT(), err)
		})
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "OR", c.GetName())
	})

	It("Eval", func() {
		assert.True(GinkgoT(), c.Eval(strCtx("a")))
		assert.True(GinkgoT(), c.Eval(strCtx("b")))
		assert.True(GinkgoT(), c.Eval(intCtx(123)))

		assert.False(GinkgoT(), c.Eval(strCtx("c")))
		assert.False(GinkgoT(), c.Eval(intCtx(456)))
	})

	It("GetKeys", func() {
		oc := OrCondition{
			content: []Condition{
				&NumericEqualsCondition{
					baseCondition{
						Key: "hello",
					},
				},
			},
		}

		keys := oc.GetKeys()
		assert.Len(GinkgoT(), keys, 1)
		assert.Equal(GinkgoT(), "hello", keys[0])
	})

	Describe("Translate", func() {
		It("ok, empty", func() {
			want := map[string]interface{}{
				"op":      "OR",
				"content": []map[string]interface{}{},
			}
			c, err := newOrCondition("content", []interface{}{})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})

		It("fail, wrong value", func() {
			_, err := newOrCondition("content", []interface{}{123})
			assert.Error(GinkgoT(), err)
		})

		It("fail, new error", func() {
			_, err := newOrCondition("content", []interface{}{
				map[string]interface{}{
					"NoSupportOP": "",
				},
			})
			assert.Error(GinkgoT(), err)
		})

		It("fail, translate error", func() {
			bc, err := newBoolCondition("test", []interface{}{true, false})
			assert.NoError(GinkgoT(), err)
			c := NewOrCondition([]Condition{
				bc,
			})
			_, err = c.Translate(true)
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			want := map[string]interface{}{
				"op": "OR",
				"content": []map[string]interface{}{
					{
						"op":    "eq",
						"field": "number",
						"value": 1,
					},
					{
						"op":    "in",
						"field": "os",
						"value": []interface{}{"linux", "windows"},
					},
				},
			}
			value := []interface{}{
				map[string]interface{}{
					"NumericEquals": map[string]interface{}{
						"number": []interface{}{1},
					},
				},
				map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"os": []interface{}{"linux", "windows"},
					},
				},
			}
			c, err := newOrCondition("content", value)
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})

		It("fail, singleTranslate fail", func() {
			value := []interface{}{
				map[string]interface{}{
					"NumericEquals": map[string]interface{}{
						"number": []interface{}{1},
					},
				},
				map[string]interface{}{
					"NotExists": map[string]interface{}{
						"os": []interface{}{"linux", "windows"},
					},
				},
			}
			_, err := newOrCondition("content", value)
			assert.Error(GinkgoT(), err)
		})

	})

	Describe("PartialEval", func() {
		Describe("no nested AND/OR", func() {

			Describe("single", func() {
				var c Condition
				BeforeEach(func() {
					c = NewOrCondition([]Condition{
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "host.system",
								Value: []interface{}{"linux"},
							},
						},
					})
				})

				It("error, no dot in key", func() {
					c = NewOrCondition([]Condition{
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "system",
								Value: []interface{}{"linux"},
							},
						},
					})
					allowed, nc := c.(LogicalCondition).PartialEval(HitStrCtx("windows"))
					assert.False(GinkgoT(), allowed)
					assert.Nil(GinkgoT(), nc)
				})

				It("false", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(HitStrCtx("windows"))
					assert.False(GinkgoT(), allowed)
					assert.Nil(GinkgoT(), nc)
					//assert.Equal(GinkgoT(), NewAnyCondition(), nc)
				})

				It("true", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(HitStrCtx("linux"))
					assert.True(GinkgoT(), allowed)
					assert.Equal(GinkgoT(), NewAnyCondition(), nc)
				})
				//
				It("remain", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(MissStrCtx("linux"))
					assert.True(GinkgoT(), allowed)
					assert.NotNil(GinkgoT(), nc)
					ct, err := nc.Translate(true)
					assert.NoError(GinkgoT(), err)
					got := map[string]interface{}{"field": "host.system", "op": "eq", "value": "linux"}
					assert.Equal(GinkgoT(), got, ct)
				})
			})

			Describe("two, no remain", func() {
				var c Condition
				BeforeEach(func() {
					c = NewOrCondition([]Condition{
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "host.system",
								Value: []interface{}{"linux"},
							},
						},
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "host.type",
								Value: []interface{}{"mysql", "linux"},
							},
						},
					})
				})

				It("all true", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(HitStrCtx("linux"))
					assert.True(GinkgoT(), allowed)
					assert.Equal(GinkgoT(), NewAnyCondition(), nc)
				})

				It("one true+one false", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(HitStrCtx("mysql"))
					assert.True(GinkgoT(), allowed)
					assert.Equal(GinkgoT(), NewAnyCondition(), nc)
				})

				It("all false", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(HitStrCtx("abc"))
					assert.False(GinkgoT(), allowed)
					assert.Nil(GinkgoT(), nc)
				})
			})

			Describe("two, has remain", func() {
				var c Condition
				BeforeEach(func() {
					c = NewOrCondition([]Condition{
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "host.system",
								Value: []interface{}{"linux"},
							},
						},
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "subject.type",
								Value: []interface{}{"mysql", "linux"},
							},
						},
					})
				})

				It("one true, will true", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
						"host.system": "linux",
					})
					assert.True(GinkgoT(), allowed)
					assert.Equal(GinkgoT(), NewAnyCondition(), nc)
				})
				It("one false, one remain", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
						"host.system": "windows",
					})
					assert.True(GinkgoT(), allowed)

					ct, err := nc.Translate(true)
					assert.NoError(GinkgoT(), err)
					got := map[string]interface{}{
						"field": "subject.type", "op": "in", "value": []interface{}{"mysql", "linux"},
					}
					assert.Equal(GinkgoT(), got, ct)
				})
				It("all remain", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{})
					assert.True(GinkgoT(), allowed)

					ct, err := nc.Translate(true)
					assert.NoError(GinkgoT(), err)
					got := map[string]interface{}{
						"op": "OR",
						"content": []map[string]interface{}{
							{"field": "host.system", "op": "eq", "value": "linux"},
							{"field": "subject.type", "op": "in", "value": []interface{}{"mysql", "linux"}}},
					}
					assert.Equal(GinkgoT(), got, ct)
				})
			})
		})

		Describe("Nested AND", func() {
			var c Condition
			BeforeEach(func() {
				c = NewOrCondition([]Condition{
					&StringEqualsCondition{
						baseCondition: baseCondition{
							Key:   "host.system",
							Value: []interface{}{"linux"},
						},
					},
					NewAndCondition([]Condition{
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "subject.type",
								Value: []interface{}{"mysql", "linux"},
							},
						},
					}),
				})
			})

			It("and true, return true", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"subject.type": "mysql",
				})
				assert.True(GinkgoT(), allowed)
				assert.Equal(GinkgoT(), NewAnyCondition(), nc)
			})

			It("and false, another remain", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"subject.type": "windows",
				})
				assert.True(GinkgoT(), allowed)
				ct, err := nc.Translate(true)
				assert.NoError(GinkgoT(), err)
				got := map[string]interface{}{"field": "host.system", "op": "eq", "value": "linux"}
				assert.Equal(GinkgoT(), got, ct)
			})

			It("and remain, another true", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"host.system": "linux",
				})
				assert.True(GinkgoT(), allowed)
				assert.Equal(GinkgoT(), NewAnyCondition(), nc)
			})

			It("and remain, another false", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"host.system": "windows",
				})
				assert.True(GinkgoT(), allowed)
				//assert.Nil(GinkgoT(), nc)
				ct, err := nc.Translate(true)
				assert.NoError(GinkgoT(), err)
				got := map[string]interface{}{"field": "subject.type", "op": "in", "value": []interface{}{"mysql", "linux"}}
				assert.Equal(GinkgoT(), got, ct)
			})

			It("both remain", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{})
				assert.True(GinkgoT(), allowed)

				ct, err := nc.Translate(true)
				assert.NoError(GinkgoT(), err)
				got := map[string]interface{}{
					"op": "OR",
					"content": []map[string]interface{}{
						{"field": "host.system", "op": "eq", "value": "linux"},
						{"field": "subject.type", "op": "in", "value": []interface{}{"mysql", "linux"}},
					},
				}
				assert.Equal(GinkgoT(), got, ct)
			})
		})
	})
})
