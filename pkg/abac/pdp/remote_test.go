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

			//// path all func success
			//patches = gomonkey.ApplyMethod(reflect.TypeOf(req), "fillActionDetail",
			//	func(_ *request.Request) error {
			//		return nil
			//	})
			//patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
			//	func(_ *request.Request) bool {
			//		return true
			//	})
			//patches.ApplyMethod(reflect.TypeOf(req), "fillSubjectDetail",
			//	func(_ *request.Request) error {
			//		return nil
			//	})
			//patches.ApplyMethod(reflect.TypeOf(req), "HasSingleLocalResource",
			//	func(_ *request.Request) bool {
			//		return true
			//	})
			//patches.ApplyFunc(queryPolicies, func(system string,
			//	subject types.Subject,
			//	action types.Action,
			//	withoutCache bool,
			//	entry *debug.Entry,
			//) (policies []types.AuthPolicy, err error) {
			//	return []types.AuthPolicy{}, nil
			//})
			//patches.ApplyFunc(evaluation.EvalPolicies, func(
			//	ctx *pdptypes.ExprContext, policies []types.AuthPolicy,
			//) (isPass bool, policyID int64, err error) {
			//	return true, 1, nil
			//})
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
			//assert.Equal(GinkgoT(), want, req.Resources[0].Attribute.(map[string]interface{}))
		})

	})

	Describe("queryRemoteResourceAttrs", func() {

	})

	Describe("queryExtResourceAttrs", func() {

	})
})
