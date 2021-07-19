/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package translate

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/types"
)

var _ = Describe("Translate", func() {
	Describe("singleTranslate", func() {
		It("fail, empty", func() {
			_, err := singleTranslate(types.PolicyCondition{}, "host")
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, not OR/AND", func() {
			expected := ExprCell{
				"op":    "in",
				"field": "host.os",
				"value": []interface{}{"linux", "windows"},
			}
			expression := types.PolicyCondition{
				"StringEquals": {
					"os": []interface{}{"linux", "windows"},
				},
			}
			ec, err := singleTranslate(expression, "host")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("ok, AND", func() {
			expected := ExprCell{
				"op": "AND",
				"content": []interface{}{
					ExprCell{
						"op":    "eq",
						"field": "host.number",
						"value": 1,
					},
					ExprCell{
						"op":    "in",
						"field": "host.os",
						"value": []interface{}{"linux", "windows"},
					},
					ExprCell{
						"op":    "eq",
						"field": "host.owner",
						"value": "admin",
					},
				},
			}
			expression := types.PolicyCondition{
				"AND": {
					"content": []interface{}{
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
						map[string]interface{}{
							"StringEquals": map[string]interface{}{
								"owner": []interface{}{"admin"},
							},
						},
					},
				},
			}
			ec, err := singleTranslate(expression, "host")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("fail, wrong operation", func() {
			expression := types.PolicyCondition{
				"NotExists": {},
			}
			_, err := singleTranslate(expression, "host")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "can not support operator")
		})

	})

	Describe("andTranslate", func() {

		It("ok, empty", func() {
			want := ExprCell{
				"op":      "AND",
				"content": []interface{}{},
			}
			ec, err := andTranslate("host", []interface{}{})

			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})

		It("fail, wrong value", func() {
			_, err := andTranslate("host", []interface{}{123})
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			want := ExprCell{
				"op": "AND",
				"content": []interface{}{
					ExprCell(map[string]interface{}{
						"op":    "eq",
						"field": "host.number",
						"value": 1,
					}),
					ExprCell(map[string]interface{}{
						"op":    "in",
						"field": "host.os",
						"value": []interface{}{"linux", "windows"},
					}),
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
			ec, err := andTranslate("host", value)
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
			_, err := andTranslate("host", value)
			assert.Error(GinkgoT(), err)
		})

	})

	Describe("orTranslate", func() {
		It("ok, empty", func() {
			want := ExprCell{
				"op":      "OR",
				"content": []interface{}{},
			}
			ec, err := orTranslate("host", []interface{}{})

			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})

		It("fail, wrong value", func() {
			_, err := orTranslate("host", []interface{}{123})
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			want := ExprCell{
				"op": "OR",
				"content": []interface{}{
					ExprCell(map[string]interface{}{
						"op":    "eq",
						"field": "host.number",
						"value": 1,
					}),
					ExprCell(map[string]interface{}{
						"op":    "in",
						"field": "host.os",
						"value": []interface{}{"linux", "windows"},
					}),
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
			ec, err := orTranslate("host", value)
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
			_, err := orTranslate("host", value)
			assert.Error(GinkgoT(), err)
		})

	})

	Describe("anyTranslate", func() {
		It("ok", func() {
			expected := ExprCell{
				"op":    "any",
				"field": "key",
				"value": []interface{}{"a"},
			}
			ec, err := anyTranslate("key", []interface{}{"a"})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})

	Describe("stringEqualsTranslate", func() {
		It("fail, empty value", func() {
			_, err := stringEqualsTranslate("key", []interface{}{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "string equals value must not be empty")
		})

		It("ok, single eq", func() {
			expected := ExprCell{
				"op":    "eq",
				"field": "key",
				"value": "a",
			}
			ec, err := stringEqualsTranslate("key", []interface{}{"a"})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)

		})

		It("ok, multiple in", func() {
			expected := ExprCell{
				"op":    "in",
				"field": "key",
				"value": []interface{}{"a", "b"},
			}
			ec, err := stringEqualsTranslate("key", []interface{}{"a", "b"})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)

		})
	})
	Describe("stringPrefixTranslate", func() {
		It("fail, empty value", func() {
			_, err := stringPrefixTranslate("key", []interface{}{})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, single", func() {
			expected := ExprCell{
				"op":    "starts_with",
				"field": "key",
				"value": "/biz,1/set,1/",
			}
			ec, err := stringPrefixTranslate("key", []interface{}{"/biz,1/set,1/"})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("ok, multiple or", func() {
			expected := ExprCell{
				"op": "OR",
				"content": []map[string]interface{}{
					{
						"op":    "starts_with",
						"field": "key",
						"value": "/biz,1/set,1/",
					},
					{
						"op":    "starts_with",
						"field": "key",
						"value": "/biz,2/set,2/",
					},
				},
			}

			ec, err := stringPrefixTranslate("key", []interface{}{"/biz,1/set,1/", "/biz,2/set,2/"})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})

	Describe("numericEqualsTranslate", func() {
		It("ok, eq", func() {
			expected := ExprCell{
				"op":    "eq",
				"field": "key",
				"value": 1,
			}
			c, err := numericEqualsTranslate("key", []interface{}{1})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c)

		})

		It("ok, in", func() {
			expected := ExprCell{
				"op":    "in",
				"field": "key",
				"value": []interface{}{1, 2},
			}
			c, err := numericEqualsTranslate("key", []interface{}{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c)

		})

		It("fail, empty value", func() {
			_, err := numericEqualsTranslate("key", []interface{}{})
			assert.Error(GinkgoT(), err)

		})

	})
	Describe("boolTranslate", func() {
		It("not support multi value", func() {
			_, err := boolTranslate("key", []interface{}{true, false})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "bool not support multi valu")
		})

		It("ok", func() {
			expected := ExprCell{
				"op":    "eq",
				"field": "key",
				"value": true,
			}
			c, err := boolTranslate("key", []interface{}{true})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c)
		})
	})
	Describe("ExprCell.Op", func() {
		It("ok", func() {
			cell := ExprCell{
				"op": "AND",
			}
			assert.Equal(GinkgoT(), cell.Op(), "AND")
		})
	})
})
