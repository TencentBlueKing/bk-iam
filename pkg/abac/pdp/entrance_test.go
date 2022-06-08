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
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/condition"
	"iam/pkg/abac/pdp/evalctx"
	"iam/pkg/abac/pdp/evaluation"
	"iam/pkg/abac/pdp/translate"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/logging/debug"
)

var _ = Describe("Entrance", func() {
	Describe("Eval", func() {
		var entry *debug.Entry
		var req *request.Request
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			// entry = debug.EntryPool.Get()
			ctl = gomock.NewController(GinkgoT())
			req = &request.Request{
				System: "test",
				Resources: []types.Resource{{
					System: "test",
				}},
			}
			entry = &debug.Entry{
				// Default is three fields, plus one optional.  Give a little extra room.
				Context:   make(debug.Fields, 6),
				Steps:     make([]debug.Step, 0, 5),
				SubDebugs: make([]*debug.Entry, 0, 5),
				Evals:     make(map[int64]string, 3),
			}

			patches = gomonkey.NewPatches()
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("FillAction error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return errors.New("fill action fail")
			})

			ok, err := Eval(req, entry, false)
			assert.False(GinkgoT(), ok)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "fill action fail")
		})

		It("ValidateAction error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return false
				})

			ok, err := Eval(req, entry, false)
			assert.False(GinkgoT(), ok)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "request resources not match action")
		})

		It("FillSubject error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return errors.New("fill subject fail")
			})

			ok, err := Eval(req, entry, false)
			assert.False(GinkgoT(), ok)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "fill subject fail")
		})

		It("QueryPolicies error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return nil, errors.New("queryPolicies fail")
			})

			ok, err := Eval(req, entry, false)
			assert.False(GinkgoT(), ok)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "queryPolicies fail")
		})
		//
		It("ok, QueryPolicies single pass", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return []types.AuthPolicy{}, nil
			})
			patches.ApplyFunc(evaluation.EvalPolicies, func(
				ctx *evalctx.EvalContext, policies []types.AuthPolicy,
			) (isPass bool, policyID int64, err error) {
				return true, 1, nil
			})

			ok, err := Eval(req, entry, false)
			assert.True(GinkgoT(), ok)
			assert.NoError(GinkgoT(), err)
		})

		It("ok, QueryPolicies single no pass", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return []types.AuthPolicy{}, nil
			})
			patches.ApplyFunc(evaluation.EvalPolicies, func(
				ctx *evalctx.EvalContext, policies []types.AuthPolicy,
			) (isPass bool, policyID int64, err error) {
				return false, -1, nil
			})

			ok, err := Eval(req, entry, false)
			assert.False(GinkgoT(), ok)
			assert.NoError(GinkgoT(), err)
		})

		It("fail, QueryPolicies single fail", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return []types.AuthPolicy{}, nil
			})
			patches.ApplyFunc(evaluation.EvalPolicies, func(
				ctx *evalctx.EvalContext, policies []types.AuthPolicy,
			) (isPass bool, policyID int64, err error) {
				return false, -1, errors.New("eval fail")
			})

			ok, err := Eval(req, entry, false)
			assert.False(GinkgoT(), ok)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "eval fail")
		})
		// TODO: add EvalPolicies multi(not single) success and fail

		It("fail, EvalPolicies error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return []types.AuthPolicy{}, nil
			})
			patches.ApplyFunc(evaluation.EvalPolicies, func(
				ctx *evalctx.EvalContext, policies []types.AuthPolicy,
			) (isPass bool, policyID int64, err error) {
				return false, -1, errors.New("test")
			})

			ok, err := Eval(req, entry, false)
			assert.False(GinkgoT(), ok)
			assert.Error(GinkgoT(), err, "test")
		})
		//
		It("ok, EvalPolicies success", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return []types.AuthPolicy{}, nil
			})
			patches.ApplyFunc(evaluation.EvalPolicies, func(
				ctx *evalctx.EvalContext, policies []types.AuthPolicy,
			) (isPass bool, policyID int64, err error) {
				return true, 1, nil
			})
			defer patches.Reset()

			ok, err := Eval(req, entry, false)
			assert.True(GinkgoT(), ok)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("Query", func() {
		var entry *debug.Entry
		var req *request.Request
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			entry = debug.EntryPool.Get()
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
			patches.Reset()
		})

		It("filter error", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return nil, errors.New("test")
			})

			expr, err := Query(req, entry, false, false)
			assert.Nil(GinkgoT(), expr)
			assert.Error(GinkgoT(), err)
		})

		It("filter empty", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return []condition.Condition{}, nil
			})

			expr, err := Query(req, entry, false, false)
			assert.Equal(GinkgoT(), expr, EmptyPolicies)
			assert.NoError(GinkgoT(), err)
		})

		It("translate error", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return []condition.Condition{
					condition.NewAnyCondition(),
				}, nil
			})
			patches.ApplyFunc(translate.ConditionsTranslate, func(policies []condition.Condition,
			) (map[string]interface{}, error) {
				return nil, errors.New("test")
			})

			expr, err := Query(req, entry, false, false)
			assert.Nil(GinkgoT(), expr)
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return []condition.Condition{
					condition.NewAnyCondition(),
				}, nil
			})
			patches.ApplyFunc(translate.ConditionsTranslate, func(policies []condition.Condition,
			) (map[string]interface{}, error) {
				return map[string]interface{}{
					"hello": "world",
				}, nil
			})

			expr, err := Query(req, entry, false, false)
			assert.Equal(GinkgoT(), expr, map[string]interface{}{
				"hello": "world",
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("QueryByExtResources", func() {
		var entry *debug.Entry
		var req *request.Request
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var extResources []types.ExtResource
		BeforeEach(func() {
			entry = debug.EntryPool.Get()
			ctl = gomock.NewController(GinkgoT())
			req = &request.Request{
				System: "test",
				Resources: []types.Resource{{
					System: "test",
				}},
			}
			extResources = []types.ExtResource{
				{
					System: "bk_cmdb",
					Type:   "host",
					IDs:    []string{"1", "2"},
				},
			}
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("filter error", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return nil, errors.New("test")
			})

			expr, resources, err := QueryByExtResources(req, []types.ExtResource{}, entry, false)
			assert.Nil(GinkgoT(), expr)
			assert.Nil(GinkgoT(), resources)
			assert.Error(GinkgoT(), err)
		})

		It("filter empty", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return []condition.Condition{}, nil
			})

			expr, resources, err := QueryByExtResources(req, extResources, entry, false)
			assert.Equal(GinkgoT(), EmptyPolicies, expr)
			assert.Equal(GinkgoT(), []types.ExtResourceWithAttribute{
				{
					System: "bk_cmdb",
					Type:   "host",
					Instances: []types.Instance{
						{
							ID:        "1",
							Attribute: map[string]interface{}{},
						},
						{
							ID:        "2",
							Attribute: map[string]interface{}{},
						},
					},
				},
			}, resources)
			assert.Nil(GinkgoT(), err)
		})

		It("query error", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return []condition.Condition{
					condition.NewAnyCondition(),
				}, nil
			})
			patches.ApplyFunc(queryExtResourceAttrs, func(
				resource *types.ExtResource,
				policies []condition.Condition,
			) (resources []map[string]interface{}, err error) {
				return nil, errors.New("test")
			})

			expr, resources, err := QueryByExtResources(req, []types.ExtResource{{}}, entry, false)
			assert.Nil(GinkgoT(), expr)
			assert.Nil(GinkgoT(), resources)
			assert.Error(GinkgoT(), err)
		})

		It("translate error", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return []condition.Condition{
					condition.NewAnyCondition(),
				}, nil
			})

			patches.ApplyFunc(queryExtResourceAttrs, func(
				resource *types.ExtResource,
				policies []condition.Condition,
			) (resources []map[string]interface{}, err error) {
				return []map[string]interface{}{}, nil
			})

			patches.ApplyFunc(translate.ConditionsTranslate, func(policies []condition.Condition,
			) (map[string]interface{}, error) {
				return nil, errors.New("test")
			})

			expr, resources, err := QueryByExtResources(req, []types.ExtResource{}, entry, false)
			assert.Nil(GinkgoT(), expr)
			assert.Nil(GinkgoT(), resources)
			assert.Error(GinkgoT(), err, "test")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(queryAndPartialEvalConditions, func(
				r *request.Request,
				entry *debug.Entry,
				willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
				withoutCache bool,
			) ([]condition.Condition, error) {
				return []condition.Condition{}, nil
			})

			patches.ApplyFunc(queryExtResourceAttrs, func(
				resource *types.ExtResource,
				policies []condition.Condition,
			) (resources []map[string]interface{}, err error) {
				return []map[string]interface{}{}, nil
			})

			patches.ApplyFunc(translate.ConditionsTranslate, func(policies []condition.Condition,
			) (map[string]interface{}, error) {
				return map[string]interface{}{}, nil
			})

			expr, resources, err := QueryByExtResources(req, extResources, entry, false)
			assert.Equal(GinkgoT(), expr, map[string]interface{}{})
			// assert.Equal(GinkgoT(), resources, []types.ExtResourceWithAttribute{})
			assert.Equal(GinkgoT(), []types.ExtResourceWithAttribute{
				{
					System: "bk_cmdb",
					Type:   "host",
					Instances: []types.Instance{
						{
							ID:        "1",
							Attribute: map[string]interface{}{},
						},
						{
							ID:        "2",
							Attribute: map[string]interface{}{},
						},
					},
				},
			}, resources)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("QueryAuthPolicies", func() {
		var entry *debug.Entry
		var req *request.Request
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			// entry = debug.EntryPool.Get()
			ctl = gomock.NewController(GinkgoT())
			req = &request.Request{
				System: "test",
				Resources: []types.Resource{{
					System: "test",
				}},
			}
			entry = &debug.Entry{
				// Default is three fields, plus one optional.  Give a little extra room.
				Context:   make(debug.Fields, 6),
				Steps:     make([]debug.Step, 0, 5),
				SubDebugs: make([]*debug.Entry, 0, 5),
				Evals:     make(map[int64]string, 3),
			}

			patches = gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionResource",
				func(_ *request.Request) bool {
					return true
				})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("FillAction error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return errors.New("fill action fail")
			})

			_, err := QueryAuthPolicies(req, entry, false)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "fill action fail")
		})

		It("FillSubject error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return errors.New("fill subject fail")
			})

			_, err := QueryAuthPolicies(req, entry, false)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "fill subject fail")
		})

		It("QueryPolicies error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return nil, errors.New("queryPolicies fail")
			})

			_, err := QueryAuthPolicies(req, entry, false)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "queryPolicies fail")
		})
		//
		It("ok, QueryPolicies single pass", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyFunc(queryPolicies, func(system string,
				subject types.Subject,
				action types.Action,
				withoutCache bool,
				entry *debug.Entry,
			) (policies []types.AuthPolicy, err error) {
				return []types.AuthPolicy{}, nil
			})

			policies, err := QueryAuthPolicies(req, entry, false)
			assert.NotNil(GinkgoT(), policies)
			assert.NoError(GinkgoT(), err)
		})
	})
})
