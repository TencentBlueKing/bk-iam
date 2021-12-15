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

	"iam/pkg/abac/pdp/condition/operator"
)

var _ = Describe("And", func() {
	wantAndCondition := &AndCondition{
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

	var c *AndCondition
	BeforeEach(func() {
		c1, _ := newStringEqualsCondition("k1", []interface{}{"a", "b"})
		c2, _ := newNumericEqualsCondition("k1", []interface{}{"b", "c"})
		c = &AndCondition{
			[]Condition{
				c1,
				c2,
			},
		}
	})

	Describe("New", func() {
		It("ok", func() {
			c := NewAndCondition([]Condition{})
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), operator.AND, c.GetName())
		})
	})

	Describe("new", func() {
		It("wrong key", func() {
			_, err := newAndCondition("wrong", []interface{}{"abc"})
			assert.Error(GinkgoT(), err)

		})

		It("ok", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				map[string]interface{}{"StringPrefix": map[string]interface{}{"path": []interface{}{"/biz,1/"}}},
			}
			c, err := newAndCondition("content", data)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), wantAndCondition, c)
		})

		It("fail", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				1,
			}
			_, err := newAndCondition("content", data)
			assert.Error(GinkgoT(), err)
		})
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), operator.AND, c.GetName())
	})

	It("GetKeys", func() {
		oc := AndCondition{
			content: []Condition{
				&StringEqualsCondition{
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

	It("Eval", func() {
		// k1 in [a, b] AND k1 in [b, c]
		assert.True(GinkgoT(), c.Eval(strCtx("b")))

		assert.False(GinkgoT(), c.Eval(strCtx("a")))
		assert.False(GinkgoT(), c.Eval(strCtx("c")))

		assert.False(GinkgoT(), c.Eval(strCtx("d")))
	})

	Describe("Translate", func() {
		It("ok, empty", func() {
			want := map[string]interface{}{
				"op":      "AND",
				"content": []map[string]interface{}{},
			}
			c, err := newAndCondition("content", []interface{}{})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})

		It("fail, translate error", func() {
			// bool only allow one value, here true+false will error
			bc, err := newBoolCondition("test", []interface{}{true, false})
			assert.NoError(GinkgoT(), err)
			c := NewAndCondition([]Condition{
				bc,
			})
			_, err = c.Translate(true)
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			want := map[string]interface{}{
				"op": "AND",
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
			c, err := newAndCondition("content", value)
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})
	})

	Describe("PartialEval", func() {
		Describe("no nested AND/OR", func() {

			// It("no content", func() {
			//
			// })

			Describe("single", func() {
				var c Condition
				BeforeEach(func() {
					c = NewAndCondition([]Condition{
						&StringEqualsCondition{
							baseCondition: baseCondition{
								Key:   "host.system",
								Value: []interface{}{"linux"},
							},
						},
					})
				})

				It("error, no dot in key", func() {
					c = NewAndCondition([]Condition{
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
				})

				It("true", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(HitStrCtx("linux"))
					assert.True(GinkgoT(), allowed)
					assert.Equal(GinkgoT(), NewAnyCondition(), nc)
				})

				It("any", func() {
					c = NewAndCondition([]Condition{
						NewAnyCondition(),
					})
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
					c = NewAndCondition([]Condition{
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
					assert.False(GinkgoT(), allowed)
					assert.Nil(GinkgoT(), nc)
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
					c = NewAndCondition([]Condition{
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

				It("one true, one remain", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
						"host.system": "linux",
					})
					assert.True(GinkgoT(), allowed)

					ct, err := nc.Translate(true)
					assert.NoError(GinkgoT(), err)
					got := map[string]interface{}{
						"field": "subject.type", "op": "in", "value": []interface{}{"mysql", "linux"},
					}
					assert.Equal(GinkgoT(), got, ct)

				})
				It("one false, one remain", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
						"host.system": "windows",
					})
					assert.False(GinkgoT(), allowed)
					assert.Nil(GinkgoT(), nc)
				})
				It("all remain", func() {
					allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{})
					assert.True(GinkgoT(), allowed)

					ct, err := nc.Translate(true)
					assert.NoError(GinkgoT(), err)
					got := map[string]interface{}{
						"op": "AND",
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
				c = NewAndCondition([]Condition{
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

			It("and true, another remain", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"subject.type": "mysql",
				})
				assert.True(GinkgoT(), allowed)

				ct, err := nc.Translate(true)
				assert.NoError(GinkgoT(), err)
				got := map[string]interface{}{"field": "host.system", "op": "eq", "value": "linux"}
				assert.Equal(GinkgoT(), got, ct)
			})

			It("and false, another remain", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"subject.type": "windows",
				})
				assert.False(GinkgoT(), allowed)
				assert.Nil(GinkgoT(), nc)
			})

			It("and remain, another true", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"host.system": "linux",
				})
				assert.True(GinkgoT(), allowed)

				ct, err := nc.Translate(true)
				assert.NoError(GinkgoT(), err)
				got := map[string]interface{}{
					"field": "subject.type", "op": "in", "value": []interface{}{"mysql", "linux"},
				}
				assert.Equal(GinkgoT(), got, ct)
			})

			It("and remain, another false", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{
					"host.system": "windows",
				})
				assert.False(GinkgoT(), allowed)
				assert.Nil(GinkgoT(), nc)
			})

			It("both remain", func() {
				allowed, nc := c.(LogicalCondition).PartialEval(MapCtx{})
				assert.True(GinkgoT(), allowed)

				ct, err := nc.Translate(true)
				assert.NoError(GinkgoT(), err)

				got := map[string]interface{}{
					"op": "AND",
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
