/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package abac_test

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac"
	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
)

var _ = Describe("rbac", func() {
	Describe("parseResourceNode", func() {
		var patches *gomonkey.Patches
		BeforeEach(func() {
			patches = gomonkey.ApplyFunc(
				cacheimpls.GetLocalResourceTypePK,
				func(_ string, id string) (int64, error) {
					switch id {
					case "t1":
						return 1, nil
					case "biz":
						return 2, nil
					case "set":
						return 3, nil
					case "module":
						return 4, nil
					case "func":
						return 5, nil
					default:
						return 0, errors.New("not found")
					}
				},
			)
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("_bk_iam_path_ array ok", func() {
			resourceNodes, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": []interface{}{
						"/cmdb,biz,1/test,set,2/test,module,3/test,func,4",
						"/cmdb,biz,1/test,set,2/test,module,3/test,func,5",
					},
				},
			})
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
			resourceNodes, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/cmdb,biz,1/test,set,2/test,module,3/test,func,4",
				},
			})
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

		It("no _bk_iam_path_ ok", func() {
			resourceNodes, err := abac.ParseResourceNode(types.Resource{
				System:    "test",
				Type:      "t1",
				ID:        "id1",
				Attribute: map[string]interface{}{},
			})
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
			_, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "//",
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "is not valid")
		})

		It("empty _bk_iam_path_ ok", func() {
			resourceNodes, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "",
				},
			})
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
			_, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "abc",
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "is not valid")
		})

		It("_bk_iam_path_ other error", func() {
			_, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": 123,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath is not string or array")
		})

		It("_bk_iam_path_ array other error", func() {
			_, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": []interface{}{123},
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "iamPath is not string")
		})

		It("_bk_iam_path_ not valid error", func() {
			_, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/biz,1/set,2/module,3/func,4/func,5",
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "not valid")
		})

		It("_bk_iam_path_ resource type not found error", func() {
			_, err := abac.ParseResourceNode(types.Resource{
				System: "test",
				Type:   "t1",
				ID:     "id1",
				Attribute: map[string]interface{}{
					"_bk_iam_path_": "/cmdb,biz,1/test,setx,2/test,module,3/test,func,4/test,func,5",
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "not found")
		})
	})
})
