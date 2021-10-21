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
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/condition"
)

var _ = Describe("Expression", func() {

	Describe("ConditionsTranslate", func() {
		anyExpr := map[string]interface{}{
			"op":    "any",
			"field": "",
			"value": []interface{}{},
		}

		Describe("single condition", func() {
			It("any", func() {
				expr, err := ConditionsTranslate([]condition.Condition{
					condition.NewAnyCondition(),
				})
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), anyExpr, expr)
			})

			It("not any", func() {
				expr, err := ConditionsTranslate([]condition.Condition{
					condition.NewBoolCondition("test", false),
				})
				want := map[string]interface{}{"field": "test", "op": "eq", "value": false}
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), want, expr)

			})

		})

		Describe("two condition", func() {
			It("one any", func() {
				expr, err := ConditionsTranslate([]condition.Condition{
					condition.NewBoolCondition("test", false),
					condition.NewAnyCondition(),
				})
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), anyExpr, expr)
			})

			It("no any", func() {
				expr, err := ConditionsTranslate([]condition.Condition{
					condition.NewBoolCondition("test", false),
					condition.NewBoolCondition("test2", true),
				})
				want := map[string]interface{}{
					"op": "OR",
					"content": []ExprCell{
						{"field": "test", "op": "eq", "value": false},
						{"field": "test2", "op": "eq", "value": true},
					},
				}
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), "OR", expr["op"])
				assert.ElementsMatch(GinkgoT(), want["content"], expr["content"])
				//assert.Equal(GinkgoT(), want, expr)
				//assert.True(GinkgoT(), assert.ObjectsAreEqualValues(want, expr))

			})

		})

	})

	Describe("oldExprToCondition", func() {
		It("ok, single", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"StringEquals":{"id":["2"]}}}]`

			want := ExprCell{
				"op":    "eq",
				"field": "bk_cmdb.biz.id",
				"value": "2",
			}
			cond, err := oldExprToCondition(resourceExpression)
			assert.NoError(GinkgoT(), err)
			got, err := cond.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, got)
		})

		It("fail, singleTranslate fail", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"NotExists":{"id":["2"]}}}]`

			_, err := oldExprToCondition(resourceExpression)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "can not support operator NotExist")
		})

		It("ok, two expression", func() {
			resourceExpression := `[{"system":"bk_sops","type":"common_flow","expression":{"Any":{"id":[]}}},
		{"system":"bk_sops","type":"project","expression":{"Any":{"id":[]}}}]`

			want := map[string]interface{}{
				"op": "AND",
				"content": []map[string]interface{}{
					{"field": "bk_sops.common_flow.id", "op": "any", "value": []interface{}{}},
					{"field": "bk_sops.project.id", "op": "any", "value": []interface{}{}},
				},
			}

			cond, err := oldExprToCondition(resourceExpression)
			assert.NoError(GinkgoT(), err)
			got, err := cond.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, got)

		})
	})

	Describe("newExprToCondition", func() {
		It("ok, single", func() {
			resourceExpression := `{"StringEquals":{"bk_cmdb.biz.id":["2"]}}`

			want := ExprCell{
				"op":    "eq",
				"field": "bk_cmdb.biz.id",
				"value": "2",
			}
			cond, err := newExprToCondition(resourceExpression)
			assert.NoError(GinkgoT(), err)
			got, err := cond.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, got)
		})

		It("fail, singleTranslate fail", func() {
			resourceExpression := `{"NotExists":{"bk_cmdb.biz.id":["2"]}}`

			_, err := newExprToCondition(resourceExpression)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "can not support operator NotExist")
		})

		It("ok, two expression", func() {
			resourceExpression := `{
		    "AND": {
				"content": [
					{"Any":{"bk_sops.common_flow.id":[]}}, 
					{"Any":{"bk_sops.project.id":[]}}
				]
            }}`

			want := map[string]interface{}{
				"op": "AND",
				"content": []map[string]interface{}{
					{"field": "bk_sops.common_flow.id", "op": "any", "value": []interface{}{}},
					{"field": "bk_sops.project.id", "op": "any", "value": []interface{}{}},
				},
			}

			cond, err := newExprToCondition(resourceExpression)
			assert.NoError(GinkgoT(), err)
			got, err := cond.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, got)
		})

	})

	Describe("expressionToCondition", func() {
		var anyCondition condition.Condition

		BeforeEach(func() {
			anyCondition = condition.NewAnyCondition()
		})

		It("ok, any, expression=``", func() {
			cond, err := expressionToCondition("")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), anyCondition, cond)
		})

		It("ok, any, expression=`[]`", func() {
			cond, err := expressionToCondition("[]")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), anyCondition, cond)
		})

		It("fail, wrong new expression", func() {
			expr := `{123}`
			_, err := expressionToCondition(expr)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "unmarshalFromString new expr fail")
		})

		It("fail, wrong old expression", func() {
			_, err := expressionToCondition("123")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "unmarshalFromString old expr fail")
		})
	})

	Describe("PolicyExpressionToCondition", func() {
		anyCondition := condition.NewAnyCondition()

		It("ok, any, expression=``", func() {
			cond, err := PolicyExpressionToCondition("")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), anyCondition, cond)
		})

		It("ok, any, expression=`[]`", func() {
			cond, err := PolicyExpressionToCondition("[]")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), anyCondition, cond)
		})

		It("fail, wrong expression", func() {
			_, err := PolicyExpressionToCondition("123")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "unmarshalFromString old expr fail")
		})

		It("ok, single", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"StringEquals":{"id":["2"]}}}]`

			want := ExprCell{
				"op":    "eq",
				"field": "bk_cmdb.biz.id",
				"value": "2",
			}
			cond, err := PolicyExpressionToCondition(resourceExpression)
			assert.NoError(GinkgoT(), err)
			got, err := cond.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, got)
		})

		It("fail, singleTranslate fail", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"NotExists":{"id":["2"]}}}]`

			_, err := PolicyExpressionToCondition(resourceExpression)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "can not support operator NotExist")
		})

		It("ok, two expression", func() {
			resourceExpression := `[{"system":"bk_sops","type":"common_flow","expression":{"Any":{"id":[]}}},
		{"system":"bk_sops","type":"project","expression":{"Any":{"id":[]}}}]`

			want := ExprCell{
				"op": "AND",
				"content": []map[string]interface{}{
					{
						"field": "bk_sops.common_flow.id", "op": "any", "value": []interface{}{},
					},
					{
						"field": "bk_sops.project.id", "op": "any", "value": []interface{}{},
					},
				},
			}

			cond, err := PolicyExpressionToCondition(resourceExpression)
			assert.NoError(GinkgoT(), err)
			got, err := cond.Translate(true)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, got)
		})

	})

	Describe("PolicyExpressionTranslate", func() {
		anyExpr := ExprCell{
			"op":    "any",
			"field": "",
			"value": []interface{}{},
		}

		It("ok, any, expression=``", func() {
			expr, err := PolicyExpressionTranslate("")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), anyExpr, expr)
		})

		It("ok, any, expression=`[]`", func() {
			expr, err := PolicyExpressionTranslate("[]")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), anyExpr, expr)
		})

		It("fail, wrong expression", func() {
			_, err := PolicyExpressionTranslate("123")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "unmarshalFromString old expr fail")
		})

		It("ok, single", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"StringEquals":{"id":["2"]}}}]`

			want := ExprCell{
				"op":    "eq",
				"field": "biz.id",
				"value": "2",
			}
			got, err := PolicyExpressionTranslate(resourceExpression)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, got)
		})

		It("fail, singleTranslate fail", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"NotExists":{"id":["2"]}}}]`

			_, err := PolicyExpressionTranslate(resourceExpression)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "can not support operator NotExist")
		})

		It("ok, two expression", func() {
			resourceExpression := `[{"system":"bk_sops","type":"common_flow","expression":{"Any":{"id":[]}}},
		{"system":"bk_sops","type":"project","expression":{"Any":{"id":[]}}}]`

			want := ExprCell{
				"op": "AND",
				"content": []map[string]interface{}{
					{
						"field": "common_flow.id", "op": "any", "value": []interface{}{},
					},
					{
						"field": "project.id", "op": "any", "value": []interface{}{},
					},
				},
			}

			expr, err := PolicyExpressionTranslate(resourceExpression)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, expr)
		})

	})

	Describe("mergeContentField", func() {
		It("ok, empty content", func() {
			content := []ExprCell{}
			newC := mergeContentField(content)
			assert.Empty(GinkgoT(), newC)
		})

		It("ok, not in/eq", func() {
			content := []ExprCell{
				{
					"op":    "starts_with",
					"field": "host.os",
					"value": "abc",
				},
				{
					"op":    "gte",
					"field": "host.id",
					"value": 23,
				},
			}
			content = mergeContentField(content)
			assert.Len(GinkgoT(), content, 2)
		})

		It("ok, not eq or in", func() {
			content := []ExprCell{
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"abc", "def"},
				},
				{
					"op":      "AND",
					"content": []interface{}{},
				},
			}
			content = mergeContentField(content)
			assert.Len(GinkgoT(), content, 2)
		})

		It("ok, single in", func() {
			content := []ExprCell{
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"abc", "def"},
				},
			}
			content = mergeContentField(content)
			want := content
			assert.Len(GinkgoT(), content, 1)
			assert.EqualValues(GinkgoT(), want, content)
		})

		It("ok, single eq", func() {
			content := []ExprCell{
				{
					"op":    "eq",
					"field": "host.id",
					"value": "abc",
				},
			}
			content = mergeContentField(content)
			want := content
			assert.Len(GinkgoT(), content, 1)
			assert.EqualValues(GinkgoT(), want, content)
		})

		It("ok, in/eq, not same field", func() {
			content := []ExprCell{
				{
					"op":    "in",
					"field": "host.os",
					"value": []interface{}{"abc", "def"},
				},
				{
					"op":    "eq",
					"field": "host.id",
					"value": "abc",
				},
			}
			content = mergeContentField(content)
			assert.Len(GinkgoT(), content, 2)
		})

		It("ok, merge", func() {
			content := []ExprCell{
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"a", "b"},
				},
				{
					"op":    "eq",
					"field": "host.id",
					"value": "c",
				},
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"d", "f"},
				},
				{
					"op":    "eq",
					"field": "host.id",
					"value": "g",
				},
			}
			content = mergeContentField(content)

			want := []ExprCell{
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"a", "b", "c", "d", "f", "g"},
				},
			}
			assert.EqualValues(GinkgoT(), want, content)
		})

		It("ok, merge part", func() {
			content := []ExprCell{
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"abc", "def"},
				},
				{
					"op":      "AND",
					"content": []interface{}{},
				},
				{
					"op":    "eq",
					"field": "host.id",
					"value": "abc",
				},
			}
			content = mergeContentField(content)

			want := []ExprCell{
				{
					"op":      "AND",
					"content": []interface{}{},
				},
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"abc", "def", "abc"},
				},
			}
			assert.EqualValues(GinkgoT(), want, content)
		})
		It("ok, merge part 2", func() {
			content := []ExprCell{
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"abc", "def"},
				},
				{
					"op":    "eq",
					"field": "host.id",
					"value": "abc",
				},
				{
					"op":      "AND",
					"content": []interface{}{},
				},
			}
			content = mergeContentField(content)

			want := []ExprCell{
				{
					"op":      "AND",
					"content": []interface{}{},
				},
				{
					"op":    "in",
					"field": "host.id",
					"value": []interface{}{"abc", "def", "abc"},
				},
			}
			assert.EqualValues(GinkgoT(), want, content)
		})

	})
})

func BenchmarkMergeContentFieldNotMerge(b *testing.B) {
	content := []ExprCell{
		{
			"op":    "in",
			"field": "host.id",
			"value": []interface{}{"abc", "def"},
		},
		{
			"op":    "eq",
			"field": "host.os",
			"value": "abc",
		},
		{
			"op":      "AND",
			"content": []interface{}{},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergeContentField(content)
	}
}

func BenchmarkMergeContentFieldMerge(b *testing.B) {
	content := []ExprCell{
		{
			"op":    "in",
			"field": "host.id",
			"value": []interface{}{"abc", "def"},
		},
		{
			"op":    "eq",
			"field": "host.id",
			"value": "abc",
		},
		{
			"op":    "eq",
			"field": "host.id",
			"value": "ghi",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergeContentField(content)
	}
}
