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
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/condition"
	"iam/pkg/abac/pdp/evalctx"
	"iam/pkg/abac/pdp/evaluation"
	"iam/pkg/abac/pip"
	"iam/pkg/abac/prp"
	"iam/pkg/abac/prp/mock"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/logging/debug"
)

var _ = Describe("Helper", func() {

	Describe("queryPolicies", func() {
		var ctl *gomock.Controller
		var mgr *mock.MockPolicyManager
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			mgr = mock.NewMockPolicyManager(ctl)

			patches = gomonkey.ApplyFunc(prp.NewPolicyManager,
				func() prp.PolicyManager {
					return mgr
				})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("error", func() {
			mgr.EXPECT().ListBySubjectAction(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				nil, errors.New("err"))

			policies, err := queryPolicies("test", types.Subject{}, types.Action{}, false, nil)
			assert.Empty(GinkgoT(), policies)
			assert.Error(GinkgoT(), err)
		})

		It("empty", func() {
			mgr.EXPECT().ListBySubjectAction(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				[]types.AuthPolicy{}, nil)
			policies, err := queryPolicies("test", types.Subject{}, types.Action{}, false, nil)
			assert.Empty(GinkgoT(), policies)
			assert.Error(GinkgoT(), err)
			assert.True(GinkgoT(), errors.Is(err, ErrNoPolicies))
		})

		It("ok", func() {
			mgr.EXPECT().ListBySubjectAction(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
				[]types.AuthPolicy{
					{
						Version:    "1",
						ID:         0,
						Expression: "",
						ExpiredAt:  0,
					},
				}, nil)

			policies, err := queryPolicies("test", types.Subject{}, types.Action{}, false, nil)
			assert.Len(GinkgoT(), policies, 1)
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("queryAndPartialEvalConditions", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var req *request.Request
		var entry *debug.Entry
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			req = &request.Request{
				System: "test",
				Action: types.Action{},
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

			policies, err := queryAndPartialEvalConditions(req, nil, false, false)
			assert.Nil(GinkgoT(), policies)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "fill action fail")
		})

		It("ValidateActionRemoteResource error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionRemoteResource",
				func(_ *request.Request) bool {
					return false
				})

			policies, err := queryAndPartialEvalConditions(req, nil, true, false)
			assert.Nil(GinkgoT(), policies)
			assert.Error(GinkgoT(), err)
		})

		It("FillSubject error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionRemoteResource",
				func(_ *request.Request) bool {
					return true
				})
			patches.ApplyFunc(fillSubjectDetail, func(req *request.Request) error {
				return errors.New("fill subject fail")
			})

			policies, err := queryAndPartialEvalConditions(req, nil, true, false)
			assert.Nil(GinkgoT(), policies)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "fill subject fail")
		})

		It("QueryPolicies error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionRemoteResource",
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
				return nil, errors.New("query policies fail")
			})
			policies, err := queryAndPartialEvalConditions(req, nil, true, false)
			assert.Nil(GinkgoT(), policies)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "query policies fail")
		})

		It("PartialEvalPolicies error", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionRemoteResource",
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
				return []types.AuthPolicy{{}}, nil
			})
			patches.ApplyFunc(evaluation.PartialEvalPolicies, func(ctx *evalctx.EvalContext,
				policie []types.AuthPolicy,
			) ([]condition.Condition, []int64, error) {
				return nil, nil, errors.New("filter error")
			})

			policies, err := queryAndPartialEvalConditions(req, nil, true, false)
			assert.Nil(GinkgoT(), policies)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "filter error")
		})

		It("PartialEvalPolicies empty", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionRemoteResource",
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
				return []types.AuthPolicy{{}}, nil
			})
			patches.ApplyFunc(evaluation.PartialEvalPolicies, func(ctx *evalctx.EvalContext,
				policies []types.AuthPolicy,
			) ([]condition.Condition, []int64, error) {
				return []condition.Condition{}, []int64{}, nil
			})

			policies, err := queryAndPartialEvalConditions(req, nil, true, false)
			assert.Len(GinkgoT(), policies, 0)
			assert.NoError(GinkgoT(), err)
		})

		It("PartialEvalPolicies ok", func() {
			patches.ApplyFunc(fillActionDetail, func(req *request.Request) error {
				return nil
			})
			patches.ApplyMethod(reflect.TypeOf(req), "ValidateActionRemoteResource",
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
				return []types.AuthPolicy{{}}, nil
			})
			patches.ApplyFunc(evaluation.PartialEvalPolicies, func(ctx *evalctx.EvalContext,
				policies []types.AuthPolicy,
			) ([]condition.Condition, []int64, error) {
				return []condition.Condition{
					condition.NewAnyCondition(),
				}, []int64{1}, nil
			})

			policies, err := queryAndPartialEvalConditions(req, entry, true, false)
			assert.Len(GinkgoT(), policies, 1)
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("fillSubjectDetail", func() {
		var r *request.Request
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			r = request.NewRequest()
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("pip.GetSubjectPK fail", func() {
			patches = gomonkey.ApplyFunc(pip.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				return -1, errors.New("get subject_pk fail")
			})
			err := fillSubjectDetail(r)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "get subject_pk fail")

		})

		It("pip.GetSubjectDetail fail", func() {
			patches = gomonkey.ApplyFunc(pip.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				return 123, nil
			})
			patches.ApplyFunc(pip.GetSubjectDetail, func(pk int64) ([]int64, []types.SubjectGroup, error) {
				return nil, nil, errors.New("get GetSubjectDetail fail")
			})

			err := fillSubjectDetail(r)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "get GetSubjectDetail fail")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(pip.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				return 123, nil
			})
			returned := []types.SubjectGroup{
				{
					PK:              1,
					PolicyExpiredAt: 123,
				},
			}

			patches.ApplyFunc(pip.GetSubjectDetail, func(pk int64) ([]int64, []types.SubjectGroup, error) {
				return []int64{1, 2, 3}, returned, nil
			})

			err := fillSubjectDetail(r)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("fillActionDetail", func() {
		var r *request.Request
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			r = request.NewRequest()
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("GetActionDetail fail", func() {
			patches = gomonkey.ApplyFunc(pip.GetActionDetail,
				func(system, id string) (pk int64,
					arts []types.ActionResourceType, err error) {
					return -1, nil, errors.New("get GetActionDetail fail")
				})

			err := fillActionDetail(r)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "get GetActionDetail fail")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(pip.GetActionDetail,
				func(system, id string) (pk int64,
					arts []types.ActionResourceType, err error) {
					return 123, []types.ActionResourceType{}, nil
				})

			err := fillActionDetail(r)
			assert.NoError(GinkgoT(), err)
		})

	})

})
