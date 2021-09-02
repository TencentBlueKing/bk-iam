/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package evaluation_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/evaluation"
	pdptypes "iam/pkg/abac/pdp/types"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/cache/impls"
	"iam/pkg/cache/memory"
)

var _ = Describe("Evaluation", func() {

	var c *pdptypes.ExprContext
	var policy types.AuthPolicy
	willPassPolicy := types.AuthPolicy{
		Expression: `[
						{
							"system": "iam",
							"type": "job",
							"expression": {
								"AND": {
									"content": [	
										{
											"StringEquals": {
												"system": ["linux"]
											}
										},
										{
											"StringPrefix": {
												"path": ["/biz,1/"]
											}
										}
									]
								}
							}
						}
					]`,
		ExpressionSignature: "3cd483f956ec8cdee6a70880c328130e",
	}
	willNotPassPolicy := types.AuthPolicy{
		Expression: `[
						{
							"system": "iam",
							"type": "job",
							"expression": {
								"AND": {
									"content": [	
										{
											"StringEquals": {
												"system": ["windows"]
											}
										},
										{
											"StringPrefix": {
												"path": ["/biz,1/"]
											}
										}
									]
								}
							}
						}
					]`,
		ExpressionSignature: "7dc6d19025f790d4509e6b732ed624a9",
		ExpiredAt:           0,
	}

	BeforeEach(func() {
		request := &request.Request{
			System: "iam",
			Subject: types.Subject{
				Type: "user",
				ID:   "admin",
			},
			Action: types.Action{
				ID:        "execute_job",
				Attribute: types.NewActionAttribute(),
			},
			Resources: []types.Resource{
				{
					System: "iam",
					Type:   "job",
					ID:     "job1",
					Attribute: map[string]interface{}{
						"system": "linux",
						"path":   []interface{}{"/biz,1/set,2/", "/biz,1/set,3/"},
					},
				},
			},
		}
		request.Action.Attribute.SetResourceTypes([]types.ActionResourceType{
			{
				System: "iam",
				Type:   "job",
			},
		})
		c = pdptypes.NewExprContext(request)
		policy = types.AuthPolicy{
			Expression: "",
		}

		impls.LocalUnmarshaledExpressionCache = memory.NewMockCache(impls.UnmarshalExpression)
	})
	Describe("EvalPolicies", func() {
		It("no policies", func() {
			allowed, id, err := evaluation.EvalPolicies(c, []types.AuthPolicy{})
			assert.False(GinkgoT(), allowed)
			assert.Equal(GinkgoT(), int64(-1), id)
			assert.NoError(GinkgoT(), err)
		})

		It("ok, one policy pass", func() {
			policies := []types.AuthPolicy{
				willPassPolicy,
			}

			allowed, _, err := evaluation.EvalPolicies(c, policies)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), allowed)
		})

		It("ok, one policy not pass", func() {
			policies := []types.AuthPolicy{
				willNotPassPolicy,
			}

			allowed, _, err := evaluation.EvalPolicies(c, policies)
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), allowed)
		})

		It("ok, one pass, one fail", func() {
			policies := []types.AuthPolicy{
				willPassPolicy,
				willNotPassPolicy,
			}

			allowed, _, err := evaluation.EvalPolicies(c, policies)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), allowed)
		})

		It("ok, one fail, one pass", func() {
			policies := []types.AuthPolicy{
				willNotPassPolicy,
				willPassPolicy,
			}

			allowed, _, err := evaluation.EvalPolicies(c, policies)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), allowed)
		})

		//It("fail, EvalPolicy err", func() {
		//	policies := []types.AuthPolicy{
		//		willPassPolicy,
		//	}
		//	//c.Resource = nil
		//	allowed, _, err := evaluation.EvalPolicies(c, policies)
		//	assert.Error(GinkgoT(), err)
		//	assert.False(GinkgoT(), allowed)
		//})

	})

	Describe("FilterPolicies", func() {
		It("ok, one policy pass", func() {
			policies := []types.AuthPolicy{
				willPassPolicy,
			}

			ps, err := evaluation.FilterPolicies(c, policies)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), ps, 1)
		})

		It("ok, one policy not pass", func() {
			policies := []types.AuthPolicy{
				willNotPassPolicy,
			}

			ps, err := evaluation.FilterPolicies(c, policies)
			assert.NoError(GinkgoT(), err)
			assert.Empty(GinkgoT(), ps)
		})

		It("ok, one pass, one fail", func() {
			policies := []types.AuthPolicy{
				willPassPolicy,
				willNotPassPolicy,
			}

			ps, err := evaluation.FilterPolicies(c, policies)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), ps, 1)
		})

		//It("fail, EvalPolicy err", func() {
		//	policies := []types.AuthPolicy{
		//		willPassPolicy,
		//	}
		//	//c.Resource = nil
		//	ps, err := evaluation.FilterPolicies(c, policies)
		//	assert.Error(GinkgoT(), err)
		//	assert.Empty(GinkgoT(), ps)
		//})
	})

	Describe("EvalPolicy", func() {

		It("ctx.Action.WithoutResourceType", func() {
			c.Action.FillAttributes(1, []types.ActionResourceType{})
			allowed, err := evaluation.EvalPolicy(c, policy)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), allowed)
		})

		//It("ctx.Resource == nil", func() {
		//	//c.Resource = nil
		//	allowed, err := evaluation.EvalPolicy(c, policy)
		//	assert.Error(GinkgoT(), err)
		//	assert.Contains(GinkgoT(), err.Error(), "get resource nil")
		//	assert.False(GinkgoT(), allowed)
		//})

		It("ParseResourceConditionFromExpression fail", func() {
			policy = types.AuthPolicy{
				Expression: "123",
			}
			allowed, err := evaluation.EvalPolicy(c, policy)
			assert.Error(GinkgoT(), err)
			assert.False(GinkgoT(), allowed)
		})

		It("ok, allowed=True", func() {
			policy = types.AuthPolicy{
				Expression: `[
						{
							"system": "iam",
							"type": "job",
							"expression": {
								"AND": {
									"content": [	
										{
											"StringEquals": {
												"system": ["linux"]
											}
										},
										{
											"StringPrefix": {
												"path": ["/biz,1/"]
											}
										}
									]
								}
							}
						}
					]`,
				ExpressionSignature: "33268b97074629d05fda196e2f7e59d2",
			}

			allowed, err := evaluation.EvalPolicy(c, policy)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), allowed)
		})

		It("ok, allowed=False", func() {
			policy = types.AuthPolicy{
				Expression: `[
						{
							"system": "iam",
							"type": "job",
							"expression": {
								"AND": {
									"content": [	
										{
											"StringEquals": {
												"system": ["windows"]
											}
										},
										{
											"StringPrefix": {
												"path": ["/biz,1/"]
											}
										}
									]
								}
							}
						}
					]`,
				ExpressionSignature: "cfeeb810bf45de623f8007d25d25293a",
			}

			allowed, err := evaluation.EvalPolicy(c, policy)
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), allowed)
		})

	})
})
