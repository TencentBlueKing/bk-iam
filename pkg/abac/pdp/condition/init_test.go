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

	Describe("AndCondition", func() {
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
			assert.Equal(GinkgoT(), "AND", c.GetName())
		})

		It("Eval", func() {
			assert.True(GinkgoT(), c.Eval(strCtx("b")))

			assert.False(GinkgoT(), c.Eval(strCtx("a")))
			assert.False(GinkgoT(), c.Eval(strCtx("c")))

			assert.False(GinkgoT(), c.Eval(strCtx("d")))
		})

		It("GetKeys", func() {
			oc := AndCondition{
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

	})

	Describe("OrCondition", func() {
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
			assert.True(GinkgoT(), c.Eval(ctx(123)))

			assert.False(GinkgoT(), c.Eval(strCtx("c")))
			assert.False(GinkgoT(), c.Eval(ctx(456)))
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
	})

	Describe("AnyCondition", func() {
		var c *AnyCondition
		BeforeEach(func() {
			c = &AnyCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{"a", "b"},
				},
			}
		})

		It("new", func() {
			condition, err := newAnyCondition("ok", []interface{}{"a", "b"})
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), condition)
		})

		It("GetName", func() {
			assert.Equal(GinkgoT(), "Any", c.GetName())
		})

		It("GetKeys", func() {
			assert.Empty(GinkgoT(), c.GetKeys())
		})

		It("Eval", func() {
			assert.True(GinkgoT(), c.Eval(ctx(1)))
			assert.True(GinkgoT(), c.Eval(boolCtx(false)))
			assert.True(GinkgoT(), c.Eval(listCtx{1, 2}))
			assert.True(GinkgoT(), c.Eval(errCtx(1)))

		})

	})

	Describe("StringEqualsCondition", func() {
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

			It("attr list", func() {
				assert.True(GinkgoT(), c.Eval(listCtx{"a", "d"}))

				assert.False(GinkgoT(), c.Eval(listCtx{"e", "f"}))
			})

		})

	})

	Describe("StringPrefixCondition", func() {
		var c *StringPrefixCondition
		BeforeEach(func() {
			c = &StringPrefixCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{"/biz,1/", "/biz,2/"},
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
			})

			It("false", func() {
				assert.False(GinkgoT(), c.Eval(strCtx("c")))
			})

			It("attr list", func() {
				assert.True(GinkgoT(), c.Eval(listCtx{"/biz,1/set,2/", "d"}))

				assert.False(GinkgoT(), c.Eval(listCtx{"e", "f"}))
			})

			It("false, attr value not string", func() {
				assert.False(GinkgoT(), c.Eval(listCtx{1}))
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
						Key:   iamPath,
						Value: []interface{}{"/biz,1/set,*/"},
					},
				}

				assert.True(GinkgoT(), c.Eval(strCtx("/biz,1/set,2/")))
				assert.False(GinkgoT(), c.Eval(strCtx("/biz,1/module,2/")))

			})
		})

	})

	Describe("NumericEqualsCondition", func() {
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
				assert.True(GinkgoT(), c.Eval(ctx(1)))
				assert.True(GinkgoT(), c.Eval(ctx(2)))
			})

			It("false", func() {
				assert.False(GinkgoT(), c.Eval(ctx(3)))
			})

			It("attr list", func() {
				assert.True(GinkgoT(), c.Eval(listCtx{2, 3}))
				assert.False(GinkgoT(), c.Eval(listCtx{3, 4}))
			})

		})

	})

	Describe("BoolCondition", func() {
		var c *BoolCondition
		BeforeEach(func() {
			c = &BoolCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{true},
				},
			}
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
				assert.False(GinkgoT(), c.Eval(ctx(1)))
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

	})

})

// TODO: move to condition unittest

//Describe("singleTranslate", func() {
//	It("fail, empty", func() {
//		_, err := singleTranslate(types.PolicyCondition{}, "host")
//		assert.Error(GinkgoT(), err)
//		assert.Equal(GinkgoT(), errMustNotEmpty, err)
//	})
//
//	It("ok, not OR/AND", func() {
//		expected := ExprCell{
//			"op":    "in",
//			"field": "host.os",
//			"value": []interface{}{"linux", "windows"},
//		}
//		expression := types.PolicyCondition{
//			"StringEquals": {
//				"os": []interface{}{"linux", "windows"},
//			},
//		}
//		ec, err := singleTranslate(expression, "host")
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, ec)
//	})
//
//	It("ok, AND", func() {
//		expected := ExprCell{
//			"op": "AND",
//			"content": []interface{}{
//				ExprCell{
//					"op":    "eq",
//					"field": "host.number",
//					"value": 1,
//				},
//				ExprCell{
//					"op":    "in",
//					"field": "host.os",
//					"value": []interface{}{"linux", "windows"},
//				},
//				ExprCell{
//					"op":    "eq",
//					"field": "host.owner",
//					"value": "admin",
//				},
//			},
//		}
//		expression := types.PolicyCondition{
//			"AND": {
//				"content": []interface{}{
//					map[string]interface{}{
//						"NumericEquals": map[string]interface{}{
//							"number": []interface{}{1},
//						},
//					},
//					map[string]interface{}{
//						"StringEquals": map[string]interface{}{
//							"os": []interface{}{"linux", "windows"},
//						},
//					},
//					map[string]interface{}{
//						"StringEquals": map[string]interface{}{
//							"owner": []interface{}{"admin"},
//						},
//					},
//				},
//			},
//		}
//		ec, err := singleTranslate(expression, "host")
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, ec)
//	})
//
//	It("fail, wrong operation", func() {
//		expression := types.PolicyCondition{
//			"NotExists": {},
//		}
//		_, err := singleTranslate(expression, "host")
//		assert.Error(GinkgoT(), err)
//		assert.Contains(GinkgoT(), err.Error(), "can not support operator")
//	})
//
//})
//
//Describe("andTranslate", func() {
//	It("ok, empty", func() {
//		want := ExprCell{
//			"op":      "AND",
//			"content": []interface{}{},
//		}
//		ec, err := andTranslate("host", []interface{}{})
//
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), want, ec)
//	})
//
//	It("fail, wrong value", func() {
//		_, err := andTranslate("host", []interface{}{123})
//		assert.Error(GinkgoT(), err)
//	})
//
//	It("fail, singleTranslate error", func() {
//		_, err := andTranslate("host", []interface{}{
//			map[string]interface{}{
//				"NoSupportOP": "",
//			},
//		})
//		assert.Error(GinkgoT(), err)
//	})
//
//	It("ok", func() {
//		want := ExprCell{
//			"op": "AND",
//			"content": []interface{}{
//				ExprCell(map[string]interface{}{
//					"op":    "eq",
//					"field": "host.number",
//					"value": 1,
//				}),
//				ExprCell(map[string]interface{}{
//					"op":    "in",
//					"field": "host.os",
//					"value": []interface{}{"linux", "windows"},
//				}),
//			},
//		}
//		value := []interface{}{
//			map[string]interface{}{
//				"NumericEquals": map[string]interface{}{
//					"number": []interface{}{1},
//				},
//			},
//			map[string]interface{}{
//				"StringEquals": map[string]interface{}{
//					"os": []interface{}{"linux", "windows"},
//				},
//			},
//		}
//		ec, err := andTranslate("host", value)
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), want, ec)
//	})
//
//	It("fail, singleTranslate fail", func() {
//		value := []interface{}{
//			map[string]interface{}{
//				"NumericEquals": map[string]interface{}{
//					"number": []interface{}{1},
//				},
//			},
//			map[string]interface{}{
//				"NotExists": map[string]interface{}{
//					"os": []interface{}{"linux", "windows"},
//				},
//			},
//		}
//		_, err := andTranslate("host", value)
//		assert.Error(GinkgoT(), err)
//	})
//
//})
//
//Describe("orTranslate", func() {
//	It("ok, empty", func() {
//		want := ExprCell{
//			"op":      "OR",
//			"content": []interface{}{},
//		}
//		ec, err := orTranslate("host", []interface{}{})
//
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), want, ec)
//	})
//
//	It("fail, wrong value", func() {
//		_, err := orTranslate("host", []interface{}{123})
//		assert.Error(GinkgoT(), err)
//	})
//
//	It("fail, singleTranslate error", func() {
//		_, err := orTranslate("host", []interface{}{
//			map[string]interface{}{
//				"NoSupportOP": "",
//			},
//		})
//		assert.Error(GinkgoT(), err)
//	})
//
//	It("ok", func() {
//		want := ExprCell{
//			"op": "OR",
//			"content": []interface{}{
//				ExprCell(map[string]interface{}{
//					"op":    "eq",
//					"field": "host.number",
//					"value": 1,
//				}),
//				ExprCell(map[string]interface{}{
//					"op":    "in",
//					"field": "host.os",
//					"value": []interface{}{"linux", "windows"},
//				}),
//			},
//		}
//		value := []interface{}{
//			map[string]interface{}{
//				"NumericEquals": map[string]interface{}{
//					"number": []interface{}{1},
//				},
//			},
//			map[string]interface{}{
//				"StringEquals": map[string]interface{}{
//					"os": []interface{}{"linux", "windows"},
//				},
//			},
//		}
//		ec, err := orTranslate("host", value)
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), want, ec)
//	})
//
//	It("fail, singleTranslate fail", func() {
//		value := []interface{}{
//			map[string]interface{}{
//				"NumericEquals": map[string]interface{}{
//					"number": []interface{}{1},
//				},
//			},
//			map[string]interface{}{
//				"NotExists": map[string]interface{}{
//					"os": []interface{}{"linux", "windows"},
//				},
//			},
//		}
//		_, err := orTranslate("host", value)
//		assert.Error(GinkgoT(), err)
//	})
//
//})
//
//Describe("anyTranslate", func() {
//	It("ok", func() {
//		expected := ExprCell{
//			"op":    "any",
//			"field": "key",
//			"value": []interface{}{"a"},
//		}
//		ec, err := anyTranslate("key", []interface{}{"a"})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, ec)
//	})
//})
//
//Describe("stringEqualsTranslate", func() {
//	It("fail, empty value", func() {
//		_, err := stringEqualsTranslate("key", []interface{}{})
//		assert.Error(GinkgoT(), err)
//		assert.Equal(GinkgoT(), errMustNotEmpty, err)
//	})
//
//	It("ok, single eq", func() {
//		expected := ExprCell{
//			"op":    "eq",
//			"field": "key",
//			"value": "a",
//		}
//		ec, err := stringEqualsTranslate("key", []interface{}{"a"})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, ec)
//
//	})
//
//	It("ok, multiple in", func() {
//		expected := ExprCell{
//			"op":    "in",
//			"field": "key",
//			"value": []interface{}{"a", "b"},
//		}
//		ec, err := stringEqualsTranslate("key", []interface{}{"a", "b"})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, ec)
//
//	})
//})
//
//Describe("stringPrefixTranslate", func() {
//	It("fail, empty value", func() {
//		_, err := stringPrefixTranslate("key", []interface{}{})
//		assert.Error(GinkgoT(), err)
//		assert.Equal(GinkgoT(), errMustNotEmpty, err)
//	})
//
//	It("ok, single", func() {
//		expected := ExprCell{
//			"op":    "starts_with",
//			"field": "key",
//			"value": "/biz,1/set,1/",
//		}
//		ec, err := stringPrefixTranslate("key", []interface{}{"/biz,1/set,1/"})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, ec)
//	})
//
//	It("ok, multiple or", func() {
//		expected := ExprCell{
//			"op": "OR",
//			"content": []map[string]interface{}{
//				{
//					"op":    "starts_with",
//					"field": "key",
//					"value": "/biz,1/set,1/",
//				},
//				{
//					"op":    "starts_with",
//					"field": "key",
//					"value": "/biz,2/set,2/",
//				},
//			},
//		}
//
//		ec, err := stringPrefixTranslate("key", []interface{}{"/biz,1/set,1/", "/biz,2/set,2/"})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, ec)
//	})
//})
//
//Describe("numericEqualsTranslate", func() {
//	It("fail, empty value", func() {
//		_, err := numericEqualsTranslate("key", []interface{}{})
//		assert.Error(GinkgoT(), err)
//		assert.Equal(GinkgoT(), errMustNotEmpty, err)
//	})
//
//	It("ok, eq", func() {
//		expected := ExprCell{
//			"op":    "eq",
//			"field": "key",
//			"value": 1,
//		}
//		c, err := numericEqualsTranslate("key", []interface{}{1})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, c)
//
//	})
//
//	It("ok, in", func() {
//		expected := ExprCell{
//			"op":    "in",
//			"field": "key",
//			"value": []interface{}{1, 2},
//		}
//		c, err := numericEqualsTranslate("key", []interface{}{1, 2})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, c)
//
//	})
//})
//
//Describe("boolTranslate", func() {
//	It("not support multi value", func() {
//		_, err := boolTranslate("key", []interface{}{true, false})
//		assert.Error(GinkgoT(), err)
//		assert.Contains(GinkgoT(), err.Error(), "bool not support multi value")
//	})
//
//	It("ok", func() {
//		expected := ExprCell{
//			"op":    "eq",
//			"field": "key",
//			"value": true,
//		}
//		c, err := boolTranslate("key", []interface{}{true})
//		assert.NoError(GinkgoT(), err)
//		assert.Equal(GinkgoT(), expected, c)
//	})
//})
