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
)

var _ = Describe("Expression", func() {
	anyExpr := map[string]interface{}{
		"op":    "any",
		"field": "",
		"value": []string{},
	}

	// TODO: refactor it
	//	Describe("PolicyStringTranslate", func() {
	//		var policies []types.AuthPolicy
	//		var resourceTypeSet []types.ActionResourceType
	//		BeforeEach(func() {
	//			resourceTypeSet = []types.ActionResourceType{
	//				{
	//					System: "iam",
	//					Type:   "job",
	//				},
	//			}
	//		})
	//
	//		It("any", func() {
	//			policies = []types.AuthPolicy{
	//				{
	//					Expression: ``,
	//				},
	//			}
	//			ec, err := PolicyStringTranslate(policies[0].Expression)
	//			assert.NoError(GinkgoT(), err)
	//			assert.Equal(GinkgoT(), anyExpr, ec)
	//		})
	//
	//		It("any, 2", func() {
	//			policies = []types.AuthPolicy{
	//				{
	//					Expression: `[]`,
	//				},
	//			}
	//			ec, err := PolicyStringTranslate(policies[0].Expression)
	//			assert.NoError(GinkgoT(), err)
	//			assert.Equal(GinkgoT(), anyExpr, ec)
	//		})
	//
	//		It("fail, policyTranslate fail", func() {
	//			policies = []types.AuthPolicy{
	//				{
	//					Expression: `123`,
	//				},
	//			}
	//			_, err := PolicyStringTranslate(policies[0].Expression)
	//			assert.Error(GinkgoT(), err)
	//		})
	//
	//		It("ok, single policy", func() {
	//			policies = []types.AuthPolicy{
	//				{
	//					Expression: `[{"system": "iam", "type": "job",
	//"expression": {"StringEquals": {"id": ["abc"]}}}]`,
	//				},
	//			}
	//			want := map[string]interface{}{
	//				"op":    "eq",
	//				"field": "job.id",
	//				"value": "abc",
	//			}
	//			ec, err := PolicyStringTranslate(policies[0].Expression)
	//			assert.NoError(GinkgoT(), err)
	//			assert.Equal(GinkgoT(), want, ec)
	//		})
	//
	//		It("ok, multiple policy", func() {
	//			policies = []types.AuthPolicy{
	//				{
	//					Expression: `[{"system": "iam", "type": "job",
	//"expression": {"StringEquals": {"id": ["abc"]}}}]`,
	//				},
	//				{
	//					Expression: `[{"system": "iam", "type": "job",
	//"expression": {"StringEquals": {"name": ["def"]}}}]`,
	//				},
	//			}
	//			want := map[string]interface{}{
	//				"op": "OR",
	//				"content": []ExprCell{
	//					{"field": "job.id", "op": "eq", "value": "abc"},
	//					{"field": "job.name", "op": "eq", "value": "def"},
	//				},
	//			}
	//			want2 := map[string]interface{}{
	//				"op": "OR",
	//				"content": []ExprCell{
	//					{"field": "job.name", "op": "eq", "value": "def"},
	//					{"field": "job.id", "op": "eq", "value": "abc"},
	//				},
	//			}
	//			ec, err := PoliciesTranslate(policies, resourceTypeSet)
	//			assert.NoError(GinkgoT(), err)
	//			assert.True(GinkgoT(), assert.ObjectsAreEqualValues(want, ec) || assert.ObjectsAreEqualValues(want2, ec))
	//		})
	//
	//		Describe("got one any expr, merged", func() {
	//			It("ok, single any policy", func() {
	//				policies = []types.AuthPolicy{
	//					{
	//						Expression: `[{"system":"iam","type":"biz","expression":{"Any":{"id":[]}}}]`,
	//					},
	//				}
	//				ec, err := PoliciesTranslate(policies, resourceTypeSet)
	//				assert.NoError(GinkgoT(), err)
	//				assert.Equal(GinkgoT(), anyExpr, ec)
	//			})
	//
	//			It("ok, two policy, one is any", func() {
	//				policies = []types.AuthPolicy{
	//					{
	//						Expression: `[{"system": "iam", "type": "job",
	//"expression": {"StringEquals": {"id": ["abc"]}}}]`,
	//					},
	//					{
	//						Expression: `[{"system":"iam","type":"biz","expression":{"Any":{"id":[]}}}]`,
	//					},
	//				}
	//				ec, err := PoliciesTranslate(policies, resourceTypeSet)
	//				assert.NoError(GinkgoT(), err)
	//				assert.Equal(GinkgoT(), anyExpr, ec)
	//			})
	//
	//			It("ok, two policy, one is any, inverse order", func() {
	//				policies = []types.AuthPolicy{
	//					{
	//						Expression: `[{"system":"iam","type":"job","expression":{"Any":{"id":[]}}}]`,
	//					},
	//					{
	//						Expression: `[{"system": "iam", "type": "job",
	//"expression": {"StringEquals": {"id": ["abc"]}}}]`,
	//					},
	//				}
	//				ec, err := PoliciesTranslate(policies, resourceTypeSet)
	//				assert.NoError(GinkgoT(), err)
	//				assert.Equal(GinkgoT(), anyExpr, ec)
	//			})
	//
	//			It("ok, multiple policy, one is any", func() {
	//				policies = []types.AuthPolicy{
	//					{
	//						Expression: `[{"system": "iam", "type": "job",
	//		"expression": {"StringEquals": {"name": ["abc"]}}}]`,
	//					},
	//					{
	//						Expression: `[{"system": "iam", "type": "job",
	//		"expression": {"StringEquals": {"name2": ["abc"]}}}]`,
	//					},
	//					{
	//						Expression: `[{"system":"iam","type":"job","expression":{"Any":{"id":[]}}}]`,
	//					},
	//					{
	//						Expression: `[{"system": "iam", "type": "job",
	//		"expression": {"StringEquals": {"name3": ["abc"]}}}]`,
	//					},
	//					{
	//						Expression: `[{"system": "iam", "type": "job",
	//		"expression": {"StringEquals": {"id": ["def"]}}}]`,
	//					},
	//				}
	//				ec, err := PoliciesTranslate(policies, resourceTypeSet)
	//				assert.NoError(GinkgoT(), err)
	//				assert.Equal(GinkgoT(), anyExpr, ec)
	//			})
	//
	//		})
	//
	//		It("ok, two resource", func() {
	//			policies = []types.AuthPolicy{
	//				{
	//					Expression: `[{"system": "bk_job", "type": "job",
	//"expression": {"OR": {"content": [{"Any": {"id": []}}]}}},
	//{"system": "bk_cmdb", "type": "host", "expression": {"OR": {"content": [{"Any": {"id": []}}]}}}]`,
	//				},
	//			}
	//			resourceTypeSet = []types.ActionResourceType{
	//				{
	//					System: "bk_job",
	//					Type:   "job",
	//				},
	//				{
	//					System: "bk_cmdb",
	//					Type:   "host",
	//				},
	//			}
	//
	//			want := map[string]interface{}{
	//				"op": "AND",
	//				"content": []ExprCell{
	//					{
	//						"op": "OR",
	//						"content": []interface{}{
	//							ExprCell{"field": "job.id", "op": "any", "value": []interface{}{}},
	//						},
	//					},
	//					{
	//						"op": "OR",
	//						"content": []interface{}{
	//							ExprCell{"field": "host.id", "op": "any", "value": []interface{}{}},
	//						},
	//					},
	//				},
	//			}
	//			ec, err := PoliciesTranslate(policies, resourceTypeSet)
	//			assert.NoError(GinkgoT(), err)
	//			assert.EqualValues(GinkgoT(), want, ec)
	//		})
	//
	//		It("ok, merge content", func() {
	//			policies = []types.AuthPolicy{
	//				{
	//					Expression: `[{"system": "iam", "type": "job",
	//"expression": {"StringEquals": {"id": ["abc"]}}}]`,
	//				},
	//				{
	//					Expression: `[{"system": "iam", "type": "job",
	//"expression": {"StringEquals": {"id": ["def", "ghi"]}}}]`,
	//				},
	//			}
	//			want := map[string]interface{}{"field": "job.id", "op": "in", "value": []interface{}{"abc", "def", "ghi"}}
	//			ec, err := PoliciesTranslate(policies, resourceTypeSet)
	//			assert.NoError(GinkgoT(), err)
	//			assert.Equal(GinkgoT(), want, ec)
	//		})
	//
	//	})

	Describe("PolicyTranslate", func() {
		//var resourceTypeSet *util.StringSet
		//BeforeEach(func() {
		//	resourceTypeSet = util.NewStringSet()
		//})

		It("ok, any, expression=``", func() {
			expr, err := PolicyStringTranslate("")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), ExprCell(anyExpr), expr)
		})

		It("ok, any, expression=`[]`", func() {
			//expr, err := PolicyTranslate("[]", resourceTypeSet)
			expr, err := PolicyStringTranslate("[]")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), ExprCell(anyExpr), expr)
		})

		It("fail, wrong expression", func() {
			//_, err := PolicyTranslate("123", resourceTypeSet)
			_, err := PolicyStringTranslate("123")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "unmarshal resourceExpression")
		})

		It("ok, single", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"StringEquals":{"id":["2"]}}}]`
			//resourceTypeSet = util.NewStringSetWithValues([]string{"bk_cmdb:biz"})

			want := ExprCell{
				"op":    "eq",
				"field": "biz.id",
				"value": "2",
			}
			expr, err := PolicyStringTranslate(resourceExpression)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, expr)
		})

		It("fail, singleTranslate fail", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"NotExists":{"id":["2"]}}}]`

			_, err := PolicyStringTranslate(resourceExpression)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "pdp PolicyTranslate expression")
		})

		It("ok, resourceTypeSet not match, return any", func() {
			resourceExpression := `[{"system":"bk_cmdb","type":"biz","expression":{"NotExists":{"id":["2"]}}}]`
			expr, err := PolicyStringTranslate(resourceExpression)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), ExprCell(anyExpr), expr)
		})

		It("ok, two expression", func() {
			resourceExpression := `[{"system":"bk_sops","type":"common_flow","expression":{"Any":{"id":[]}}},
{"system":"bk_sops","type":"project","expression":{"Any":{"id":[]}}}]`

			want := ExprCell{
				"op": "AND",
				"content": []ExprCell{
					{
						"field": "common_flow.id", "op": "any", "value": []interface{}{},
					},
					{
						"field": "project.id", "op": "any", "value": []interface{}{},
					},
				},
			}
			expr, err := PolicyStringTranslate(resourceExpression)
			assert.NoError(GinkgoT(), err)
			assert.EqualValues(GinkgoT(), want, expr)
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
