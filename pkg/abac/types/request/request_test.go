/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package request_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
)

var _ = Describe("request", func() {
	Describe("NewRequest", func() {
		It("ok", func() {
			r := request.NewRequest()
			expected := &request.Request{
				Subject:   types.NewSubject(),
				Action:    types.NewAction(),
				Resources: []types.Resource{},
			}
			assert.Equal(GinkgoT(), expected, r)
		})
	})

	Describe("HasSingleLocalResource", func() {
		var expectedSystem = "bk_test"
		var r *request.Request
		BeforeEach(func() {
			r = request.NewRequest()
			r.System = expectedSystem
		})

		It("true", func() {
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{
					"resource_type": []types.ActionResourceType{{
						System: expectedSystem,
						Type:   "job",
					}},
				},
			}

			assert.True(GinkgoT(), r.HasSingleLocalResource())
		})

		It("false, two resourceTypes", func() {
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{
					"resource_type": []types.ActionResourceType{{
						System: expectedSystem,
						Type:   "type1",
					}, {
						System: expectedSystem,
						Type:   "type2",
					}},
				},
			}
			assert.False(GinkgoT(), r.HasSingleLocalResource())
		})

		It("false, not my system", func() {
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{
					"resource_type": []types.ActionResourceType{{
						System: "bk_iam",
						Type:   "job",
					}},
				},
			}
			assert.False(GinkgoT(), r.HasSingleLocalResource())
		})

	})

	Describe("HasRemoteResources", func() {
		var expectedSystem = "bk_test"
		var r *request.Request
		BeforeEach(func() {
			r = request.NewRequest()
			r.System = expectedSystem
		})
		It("empty, false", func() {
			assert.False(GinkgoT(), r.HasRemoteResources())
		})

		It("false", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_test",
					Type:   "job",
				},
			}
			assert.False(GinkgoT(), r.HasRemoteResources())
		})

		It("single true", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job3",
				},
			}
			assert.True(GinkgoT(), r.HasRemoteResources())
		})

		It("two true", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job3",
				},
				{
					System: "bk_test",
					Type:   "job",
				},
			}
			assert.True(GinkgoT(), r.HasRemoteResources())
		})
	})

	Describe("GetRemoteResources", func() {
		var expectedSystem = "bk_test"
		var r *request.Request
		BeforeEach(func() {
			r = request.NewRequest()
			r.System = expectedSystem
		})

		It("no remote resource", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_test",
					Type:   "job",
				},
			}

			d := r.GetRemoteResources()
			assert.Len(GinkgoT(), d, 0)
		})

		It("single remote resource", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job3",
				},
			}

			d := r.GetRemoteResources()
			assert.Len(GinkgoT(), d, 1)
			assert.Equal(GinkgoT(), "bk_iam", d[0].System)
		})

		It("both local and remote", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_test",
					Type:   "job",
				},
				{
					System: "bk_iam",
					Type:   "job3",
				},
			}

			d := r.GetRemoteResources()
			assert.Len(GinkgoT(), d, 1)
			assert.Equal(GinkgoT(), "bk_iam", d[0].System)
		})

		It("all remote resources", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_abc",
					Type:   "job",
				},
				{
					System: "bk_iam",
					Type:   "job3",
				},
			}

			d := r.GetRemoteResources()
			assert.Len(GinkgoT(), d, 2)
		})

	})
	Describe("GetSortedResources", func() {
		var expectedSystem = "bk_test"
		var r *request.Request
		BeforeEach(func() {
			r = request.NewRequest()
			r.System = expectedSystem
		})

		It("single hit", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_test",
					Type:   "job",
				},
			}

			d := r.GetSortedResources()
			assert.Len(GinkgoT(), d, 1)
			assert.Equal(GinkgoT(), expectedSystem, d[0].System)
		})

		It("single not hit", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job3",
				},
			}

			d := r.GetSortedResources()
			assert.Len(GinkgoT(), d, 1)
			assert.Equal(GinkgoT(), "bk_iam", d[0].System)
		})

		It("two hit", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job3",
				},
				{
					System: "bk_test",
					Type:   "job",
				},
			}

			d := r.GetSortedResources()
			assert.Len(GinkgoT(), d, 2)
			assert.Equal(GinkgoT(), expectedSystem, d[0].System)
		})
	})

	Describe("ValidateActionResource", func() {
		var expectedSystem = "bk_test"
		var r *request.Request
		BeforeEach(func() {
			r = request.NewRequest()
			r.System = expectedSystem
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{
					"resource_type": []types.ActionResourceType{{
						System: "bk_iam",
						Type:   "job",
					}},
				},
			}
		})

		It("ok", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job",
				},
			}
			assert.True(GinkgoT(), r.ValidateActionResource())
		})

		It("false, length not match", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job",
				},
				{
					System: "bk_iam",
					Type:   "job",
				},
			}
			assert.False(GinkgoT(), r.ValidateActionResource())
		})

		It("false, type not match", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_iam",
					Type:   "job2",
				},
			}
			assert.False(GinkgoT(), r.ValidateActionResource())
		})
	})
	Describe("ValidateActionRemoteResource", func() {
		var expectedSystem = "bk_test"
		var r *request.Request
		BeforeEach(func() {
			r = request.NewRequest()
			r.System = expectedSystem
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{
					"resource_type": []types.ActionResourceType{{
						System: "bk_test",
						Type:   "host",
					}, {
						System: "bk_job",
						Type:   "job",
					},
					},
				},
			}
		})

		It("true", func() {

			r.Resources = []types.Resource{
				{
					System: "bk_job",
					Type:   "job",
				},
			}
			assert.True(GinkgoT(), r.ValidateActionRemoteResource())

		})

		It("false, wrong local resources", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_test",
					Type:   "cluster",
				},
				{
					System: "bk_job",
					Type:   "job",
				},
			}
			assert.False(GinkgoT(), r.ValidateActionRemoteResource())
		})

		It("false, wrong remote resources count", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_job",
					Type:   "job",
				},
				{
					System: "bk_job",
					Type:   "job",
				},
			}
			assert.False(GinkgoT(), r.ValidateActionRemoteResource())
		})

		It("false, wrong remote resources type", func() {
			r.Resources = []types.Resource{
				{
					System: "bk_job",
					Type:   "job2",
				},
			}
			assert.False(GinkgoT(), r.ValidateActionRemoteResource())
		})
	})
	Describe("GetQueryResourceTypes", func() {
		var expectedSystem = "bk_test"
		var r *request.Request
		BeforeEach(func() {
			r = request.NewRequest()
			r.System = expectedSystem
			r.Resources = []types.Resource{
				{
					System: "bk_test",
					Type:   "job",
				},
			}
		})

		It("empty", func() {
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{
					"resource_type": []types.ActionResourceType{{
						System: "bk_test",
						Type:   "job",
					}},
				},
			}

			d, err := r.GetQueryResourceTypes()
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), d, 0)
		})

		It("ok", func() {
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{
					"resource_type": []types.ActionResourceType{
						{
							System: "bk_test",
							Type:   "job",
						},
						{
							System: "bk_iam",
							Type:   "obj",
						},
					},
				},
			}

			d, err := r.GetQueryResourceTypes()
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), d, 1)

			expectedQueryResourceTypes := []types.ActionResourceType{
				{
					System: "bk_iam",
					Type:   "obj",
				},
			}
			assert.Equal(GinkgoT(), expectedQueryResourceTypes, d)
		})

		It("error", func() {
			r.Action.Attribute = &types.ActionAttribute{
				Attribute: map[string]interface{}{},
			}

			_, err := r.GetQueryResourceTypes()
			assert.Error(GinkgoT(), err)
		})

	})
	//Describe("genResourceTypeKey", func() {
	//	var r *request.Request
	//	BeforeEach(func() {
	//		r = request.NewRequest()
	//	})
	//
	//	It("ok", func() {
	//
	//	})
	//})
	//Describe("getActionResourceTypeIDSet", func() {
	//	It("", func() {
	//	})
	//})
})
