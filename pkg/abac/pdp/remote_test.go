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

	"iam/pkg/cacheimpls"

	"iam/pkg/abac/pdp/condition"
	"iam/pkg/abac/pip"

	"github.com/agiledragon/gomonkey"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
)

var _ = Describe("Remote", func() {

	Describe("fillRemoteResourceAttrs", func() {
		var req *request.Request
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			req = &request.Request{
				System: "test",
				Resources: []types.Resource{{
					System: "test",
				}},
			}
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("no remote resources", func() {
			err := fillRemoteResourceAttrs(req, []types.AuthPolicy{})
			assert.NoError(GinkgoT(), err)
		})

		It("one remote resources, queryRemoteResourceAttrs fail", func() {
			req = &request.Request{
				System: "test",
				Resources: []types.Resource{{
					System: "iam",
				}},
			}
			patches = gomonkey.ApplyFunc(queryRemoteResourceAttrs, func(
				resource *types.Resource, policies []types.AuthPolicy,
			) (attrs map[string]interface{}, err error) {
				return nil, errors.New("query remote remote resource attrs fail")
			})

			err := fillRemoteResourceAttrs(req, []types.AuthPolicy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "query remote remote resource attrs fail")
		})

		It("ok", func() {
			req = &request.Request{
				System: "test",
				Resources: []types.Resource{{
					System: "iam",
				}},
			}

			want := map[string]interface{}{
				"hello": "world",
			}
			patches = gomonkey.ApplyFunc(queryRemoteResourceAttrs, func(
				resource *types.Resource, policies []types.AuthPolicy,
			) (attrs map[string]interface{}, err error) {
				return want, nil
			})

			err := fillRemoteResourceAttrs(req, []types.AuthPolicy{})
			assert.NoError(GinkgoT(), err)

			w, e := req.Resources[0].Attribute.GetString("hello")
			assert.NoError(GinkgoT(), e)
			assert.Equal(GinkgoT(), "world", w)
			// assert.Equal(GinkgoT(), want, req.Resources[0].Attribute.(map[string]interface{}))
		})

	})

	Describe("queryRemoteResourceAttrs", func() {
		var resource *types.Resource
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			resource = &types.Resource{
				System:    "bk_cmdb",
				Type:      "host",
				ID:        "1",
				Attribute: nil,
			}
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("empty", func() {
			_, err := queryRemoteResourceAttrs(resource, []types.AuthPolicy{})
			assert.NoError(GinkgoT(), err)
		})

		It("error, cacheimpls.GetUnmarshalledResourceExpression fail", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetUnmarshalledResourceExpression,
				func(expression, signature string) (condition.Condition, error) {
					return nil, errors.New("the error")

				})

			_, err := queryRemoteResourceAttrs(resource, []types.AuthPolicy{
				{
					Version:             "1",
					ID:                  1,
					Expression:          "",
					ExpressionSignature: "",
					ExpiredAt:           0,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "the error", err.Error())
		})

		// It("error, getConditionAttrKeys fail", func() {
		// 	patches = gomonkey.ApplyFunc(getConditionAttrKeys,
		// 		func(resource *types.Resource, conditions []condition.Condition) ([]string, error) {
		// 			return nil, errors.New("the error2")
		// 		})
		//
		// 	_, err := queryRemoteResourceAttrs(resource, []types.AuthPolicy{})
		// 	assert.Error(GinkgoT(), err)
		// 	assert.Contains(GinkgoT(), err.Error(), "the error2")
		// })

		It("error, pip.QueryRemoteResourceAttribute fail", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetUnmarshalledResourceExpression,
				func(expression, signature string) (condition.Condition, error) {
					return condition.NewBoolCondition("bk_cmdb.host.isUp", true), nil

				})

			patches.ApplyFunc(getConditionAttrKeys,
				func(resource *types.Resource, conditions []condition.Condition) []string {
					return []string{"isUp"}
				})

			patches.ApplyFunc(pip.QueryRemoteResourceAttribute,
				func(system, _type, id string, keys []string) (map[string]interface{}, error) {
					return nil, errors.New("the error3")
				})

			_, err := queryRemoteResourceAttrs(resource, []types.AuthPolicy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "the error3")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetUnmarshalledResourceExpression,
				func(expression, signature string) (condition.Condition, error) {
					return condition.NewBoolCondition("bk_cmdb.host.isUp", true), nil
				})

			patches.ApplyFunc(getConditionAttrKeys,
				func(resource *types.Resource, conditions []condition.Condition) []string {
					return []string{"isUp"}
				})

			patches.ApplyFunc(pip.QueryRemoteResourceAttribute,
				func(system, _type, id string, keys []string) (map[string]interface{}, error) {
					return map[string]interface{}{"hello": "world"}, nil
				})

			attrs, err := queryRemoteResourceAttrs(resource, []types.AuthPolicy{
				{
					Version:             "1",
					ID:                  1,
					Expression:          "",
					ExpressionSignature: "",
					ExpiredAt:           0,
				},
			})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), map[string]interface{}{"hello": "world"}, attrs)
		})

	})

	Describe("queryExtResourceAttrs", func() {
		var resource *types.ExtResource
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			resource = &types.ExtResource{
				System: "bk_cmdb",
				Type:   "host",
				IDs:    []string{"1", "2"},
			}
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		// It("getConditionAttrKeys fail", func() {
		// 	patches = gomonkey.ApplyFunc(getConditionAttrKeys,
		// 		func(resource *types.Resource, conditions []condition.Condition) ([]string, error) {
		// 			return nil, errors.New("the error")
		// 		})
		//
		// 	_, err := queryExtResourceAttrs(resource, []condition.Condition{})
		// 	assert.Error(GinkgoT(), err)
		// 	assert.Contains(GinkgoT(), err.Error(), "the error")
		// })

		It("pip.BatchQueryRemoteResourcesAttribute fail", func() {
			patches = gomonkey.ApplyFunc(getConditionAttrKeys,
				func(resource *types.Resource, conditions []condition.Condition) []string {
					return []string{"id"}
				})

			patches.ApplyFunc(pip.BatchQueryRemoteResourcesAttribute,
				func(system, _type string, ids []string, keys []string) ([]map[string]interface{}, error) {
					return nil, errors.New("the error2")
				})

			_, err := queryExtResourceAttrs(resource, []condition.Condition{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "the error2")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(getConditionAttrKeys,
				func(resource *types.Resource, conditions []condition.Condition) []string {
					return []string{"id"}
				})

			patches.ApplyFunc(pip.BatchQueryRemoteResourcesAttribute,
				func(system, _type string, ids []string, keys []string) ([]map[string]interface{}, error) {
					return []map[string]interface{}{{"hello": "world"}}, nil
				})

			resources, err := queryExtResourceAttrs(resource, []condition.Condition{})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []map[string]interface{}{{"hello": "world"}}, resources)

		})

	})

	Describe("getConditionAttrKeys", func() {
		var resource *types.Resource
		BeforeEach(func() {
			resource = &types.Resource{
				System:    "bk_cmdb",
				Type:      "host",
				ID:        "1",
				Attribute: nil,
			}
		})
		It("empty", func() {
			keys := getConditionAttrKeys(resource, []condition.Condition{})
			assert.Empty(GinkgoT(), keys)
		})

		It("one condition", func() {
			keys := getConditionAttrKeys(resource, []condition.Condition{
				condition.NewBoolCondition("bk_cmdb.host.isUp", true),
			})
			assert.Len(GinkgoT(), keys, 1)
			assert.Equal(GinkgoT(), "isUp", keys[0])
		})

		It("two condition", func() {
			keys := getConditionAttrKeys(resource, []condition.Condition{
				condition.NewBoolCondition("bk_cmdb.host.isUp", true),
				condition.NewBoolCondition("bk_cmdb.host.isDown", false),
				condition.NewBoolCondition("bk_cmdb.module.isOk", false),
			})
			assert.Len(GinkgoT(), keys, 2)
			assert.ElementsMatch(GinkgoT(), []string{"isUp", "isDown"}, keys)
		})

	})

})
