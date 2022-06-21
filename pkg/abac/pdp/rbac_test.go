/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pdp

import (
	"errors"

	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("rbac", func() {
	Describe("parseResourceNode", func() {
		var actionResourceType types.ActionResourceType
		BeforeEach(func() {
			actionResourceType = types.ActionResourceType{
				System: "test",
				Type:   "t1",
				PK:     1,
				ResourceTypeOfInstanceSelections: []types.ThinResourceType{
					{
						System: "cmdb",
						ID:     "biz",
						PK:     2,
					},
					{
						System: "test",
						ID:     "set",
						PK:     3,
					},
					{
						System: "test",
						ID:     "module",
						PK:     4,
					},
					{
						System: "test",
						ID:     "func",
						PK:     5,
					},
				},
			}
		})

		It("_bk_iam_path_ array ok", func() {
			resourceNodes, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": []interface{}{
						"/biz,1/set,2/module,3/func,4",
						"/biz,1/set,2/module,3/func,5",
					},
				},
			}, actionResourceType)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ResourceNode{
				{
					System: "cmdb",
					Type:   "biz",
					ID:     "1",
					TypePK: 2,
				}, {
					System: "test",
					Type:   "set",
					ID:     "2",
					TypePK: 3,
				}, {
					System: "test",
					Type:   "module",
					ID:     "3",
					TypePK: 4,
				}, {
					System: "test",
					Type:   "func",
					ID:     "4",
					TypePK: 5,
				}, {
					System: "test",
					Type:   "func",
					ID:     "5",
					TypePK: 5,
				}, {
					System: "test",
					Type:   "t1",
					ID:     "id1",
					TypePK: 1,
				},
			}, resourceNodes)
		})

		It("_bk_iam_path_ string ok", func() {
			resourceNodes, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2/module,3/func,4",
				},
			}, actionResourceType)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ResourceNode{
				{
					System: "cmdb",
					Type:   "biz",
					ID:     "1",
					TypePK: 2,
				}, {
					System: "test",
					Type:   "set",
					ID:     "2",
					TypePK: 3,
				}, {
					System: "test",
					Type:   "module",
					ID:     "3",
					TypePK: 4,
				}, {
					System: "test",
					Type:   "func",
					ID:     "4",
					TypePK: 5,
				}, {
					System: "test",
					Type:   "t1",
					ID:     "id1",
					TypePK: 1,
				},
			}, resourceNodes)
		})

		It("not _bk_iam_path_ ok", func() {
			resourceNodes, err := parseResourceNode(types.Resource{
				System:    "test",
				Type:      "t1",
				ID:        "id1",
				Attribute: map[string]interface{}{},
			}, actionResourceType)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ResourceNode{
				{
					System: "test",
					Type:   "t1",
					ID:     "id1",
					TypePK: 1,
				},
			}, resourceNodes)
		})

		It("_bk_iam_path_ error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "//",
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath is not valid")
		})

		It("empty _bk_iam_path_ ok", func() {
			resourceNodes, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "",
				},
			}, actionResourceType)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ResourceNode{
				{
					System: "test",
					Type:   "t1",
					ID:     "id1",
					TypePK: 1,
				},
			}, resourceNodes)
		})

		It("_bk_iam_path_ error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "abc",
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath is not valid")
		})

		It("_bk_iam_path_ other error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": 123,
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath is not string or array")
		})

		It("_bk_iam_path_ array other error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": []interface{}{123},
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath is not string")
		})

		It("_bk_iam_path_ not valid error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2,3/module,3/func,4/func,5",
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath is not valid")
		})

		It("_bk_iam_path_ resource type not found error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/setx,2/module,3/func,4/func,5",
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath resource type not found")
		})

		It("resource type not found error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test",
				Type:   "t2",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2/module,3/func,4/func,5",
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "resource type not match")
		})

		It("resource type system not match error", func() {
			_, err := parseResourceNode(types.Resource{
				System: "test1",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2/module,3/func,4/func,5",
				},
			}, actionResourceType)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "resource type not match")
		})
	})

	Describe("rbacEval", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var action types.Action
		var resources []types.Resource
		var effectGroupPKs []int64
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			action = types.NewAction()
			action.FillAttributes(1, 1, []types.ActionResourceType{{
				System: "test",
				Type:   "t1",
				PK:     1,
				ResourceTypeOfInstanceSelections: []types.ThinResourceType{
					{
						System: "cmdb",
						ID:     "biz",
						PK:     2,
					},
					{
						System: "test",
						ID:     "set",
						PK:     3,
					},
					{
						System: "test",
						ID:     "module",
						PK:     4,
					},
					{
						System: "test",
						ID:     "func",
						PK:     5,
					},
				},
			}})

			resources = []types.Resource{{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2/module,3/func,4",
				},
			}}

			effectGroupPKs = []int64{1, 2}
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("action.Attribute.GetResourceTypes fail", func() {
			action = types.NewAction()
			_, err := rbacEval("test", action, resources, effectGroupPKs, true, nil)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "action.Attribute.GetResourceTypes")
		})

		It("resources with two types fail", func() {
			resources = []types.Resource{{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2/module,3/func,4",
				},
			}, {
				System: "test",
				Type:   "t2",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2/module,3/func,4",
				},
			}}
			_, err := rbacEval("test", action, resources, effectGroupPKs, true, nil)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "rbacEval only support action with one resource type")
		})

		It("action with two types fail", func() {
			action = types.NewAction()
			action.FillAttributes(1, 1, []types.ActionResourceType{{
				System: "test",
				Type:   "t1",
				PK:     1,
			}, {
				System: "test",
				Type:   "t2",
				PK:     1,
			}})
			_, err := rbacEval("test", action, resources, effectGroupPKs, true, nil)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "rbacEval only support action with one resource type")
		})

		It("cacheimpls.GetResourceActionAuthorizedGroupPKs fail", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetResourceActionAuthorizedGroupPKs, func(
				systemID string,
				actionPK, actionResourceTypePK, resourceTypePK int64,
				resourceID string,
			) ([]int64, error) {
				return nil, errors.New("test")
			})

			_, err := rbacEval("test", action, resources, effectGroupPKs, true, nil)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetResourceActionAuthorizedGroupPKs")
		})

		It("svc.GetAuthorizedActionGroupMap fail", func() {
			mockSvc := mock.NewMockGroupResourcePolicyService(ctl)
			mockSvc.EXPECT().
				GetAuthorizedActionGroupMap("test", int64(1), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("test")).
				AnyTimes()
			patches = gomonkey.ApplyFunc(
				service.NewGroupResourcePolicyService,
				func() service.GroupResourcePolicyService { return mockSvc },
			)

			_, err := rbacEval("test", action, resources, effectGroupPKs, false, nil)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetResourceActionAuthorizedGroupPKs")
		})

		It("not pass withCache", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetResourceActionAuthorizedGroupPKs, func(
				systemID string,
				actionPK, actionResourceTypePK, resourceTypePK int64,
				resourceID string,
			) ([]int64, error) {
				return []int64{3}, nil
			})

			isPass, err := rbacEval("test", action, resources, effectGroupPKs, true, nil)
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), isPass)
		})

		It("not pass not withCache", func() {
			mockSvc := mock.NewMockGroupResourcePolicyService(ctl)
			mockSvc.EXPECT().
				GetAuthorizedActionGroupMap("test", int64(1), gomock.Any(), gomock.Any()).
				Return(map[int64][]int64{}, nil).
				AnyTimes()
			patches = gomonkey.ApplyFunc(
				service.NewGroupResourcePolicyService,
				func() service.GroupResourcePolicyService { return mockSvc },
			)

			isPass, err := rbacEval("test", action, resources, effectGroupPKs, false, nil)
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), isPass)
		})

		It("pass withCache", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetResourceActionAuthorizedGroupPKs, func(
				systemID string,
				actionPK, actionResourceTypePK, resourceTypePK int64,
				resourceID string,
			) ([]int64, error) {
				return []int64{3, 1, 4, 5}, nil
			})

			isPass, err := rbacEval("test", action, resources, effectGroupPKs, true, nil)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), isPass)
		})

		It("pass not withCache", func() {
			mockSvc := mock.NewMockGroupResourcePolicyService(ctl)
			mockSvc.EXPECT().
				GetAuthorizedActionGroupMap("test", int64(1), gomock.Any(), gomock.Any()).
				Return(map[int64][]int64{1: {1, 4, 5}}, nil).
				AnyTimes()
			patches = gomonkey.ApplyFunc(
				service.NewGroupResourcePolicyService,
				func() service.GroupResourcePolicyService { return mockSvc },
			)

			isPass, err := rbacEval("test", action, resources, effectGroupPKs, false, nil)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), isPass)
		})
	})
})
