/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

import (
	"errors"
	"reflect"

	"iam/pkg/abac/prp/policy"
	"iam/pkg/abac/types"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	svctypes "iam/pkg/service/types"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("PolicyCurd", func() {

	Describe("DeleteByIDs", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.DeleteByIDs("test", "user", "test", []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("policyService.DeleteByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().DeleteByPKs(
				int64(1), []int64{1, 2},
			).Return(
				errors.New("delete fail"),
			).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
			}

			err := manager.DeleteByIDs("test", "user", "test", []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.DeleteByPKs")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().DeleteByPKs(
				int64(1), []int64{1, 2},
			).Return(
				nil,
			).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
			}

			err := manager.DeleteByIDs("test", "user", "test", []int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("AlterCustomPolicies", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("actionService.ListThinActionBySystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListThinActionBySystem")
		})

		It("actionService.ListActionResourceTypeIDByActionSystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListActionResourceTypeIDByActionSystem")
		})

		It("ErrCreateActionNotExists fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.AlterCustomPolicies("test", "user", "test", []types.Policy{{
				Action: types.Action{
					ID: "test",
				},
			}}, []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "action not exists")
		})

		It("ErrUpdateActionNotExists fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{{
				Action: types.Action{
					ID: "test",
				},
			}}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "action not exists")
		})

		It("policyService.AlterCustomPolicies fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().AlterCustomPolicies(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			).Return(
				map[int64][]int64{}, errors.New("alter policies fail"),
			).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,
			}

			err := manager.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.AlterPolicies")
		})

		It("AlterCustomPolicies success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().AlterCustomPolicies(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			).Return(
				map[int64][]int64{}, nil,
			).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,
			}

			err := manager.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("UpdateSubjectPoliciesExpiredAt", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.UpdateSubjectPoliciesExpiredAt("user", "test", []types.PolicyPKExpiredAt{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("empty return", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.UpdateSubjectPoliciesExpiredAt("user", "test", []types.PolicyPKExpiredAt{})
			assert.NoError(GinkgoT(), err)
		})

		It("policyService.ListQueryByPKs", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().ListQueryByPKs(gomock.Any()).Return(
				nil, errors.New("list policy fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
			}

			err := manager.UpdateSubjectPoliciesExpiredAt("user", "test", []types.PolicyPKExpiredAt{{
				PK:        1,
				ExpiredAt: 1,
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.ListQueryByPKs")
		})

		It("empty update policies return", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().ListQueryByPKs(gomock.Any()).Return(
				[]svctypes.QueryPolicy{}, nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
			}

			err := manager.UpdateSubjectPoliciesExpiredAt("user", "test", []types.PolicyPKExpiredAt{{
				PK:        1,
				ExpiredAt: 1,
			}})
			assert.NoError(GinkgoT(), err)
		})

		It("actionService.ListThinActionByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().ListQueryByPKs(gomock.Any()).Return(
				[]svctypes.QueryPolicy{{
					PK:           1,
					SubjectPK:    1,
					ExpressionPK: 0,
				}}, nil,
			).AnyTimes()
			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionByPKs(gomock.Any()).Return(
				nil, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
				actionService:  mockActionService,
			}

			err := manager.UpdateSubjectPoliciesExpiredAt("user", "test", []types.PolicyPKExpiredAt{{
				PK:        1,
				ExpiredAt: 1,
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "getSystemSetFromPolicyIDs")
		})

		It("policyService.UpdateExpiredAt fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().UpdateExpiredAt(gomock.Any()).Return(
				errors.New("update expiredat fail")).AnyTimes()
			mockPolicyService.EXPECT().ListQueryByPKs(gomock.Any()).Return(
				[]svctypes.QueryPolicy{{
					PK:           1,
					SubjectPK:    1,
					ExpressionPK: 0,
				}}, nil,
			).AnyTimes()
			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionByPKs(gomock.Any()).Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.BatchDeleteSystemSubjectPKsFromCache,
				func(systems []string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
				actionService:  mockActionService,
			}

			err := manager.UpdateSubjectPoliciesExpiredAt("user", "test", []types.PolicyPKExpiredAt{{
				PK:        1,
				ExpiredAt: 1,
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.UpdateExpiredAt")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().UpdateExpiredAt(gomock.Any()).Return(nil).AnyTimes()
			mockPolicyService.EXPECT().ListQueryByPKs(gomock.Any()).Return(
				[]svctypes.QueryPolicy{{
					PK:           1,
					SubjectPK:    1,
					ExpressionPK: 0,
				}}, nil,
			).AnyTimes()
			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionByPKs(gomock.Any()).Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
				actionService:  mockActionService,
			}

			patches = gomonkey.ApplyFunc(policy.BatchDeleteSystemSubjectPKsFromCache,
				func(systems []string, pks []int64) error {
					return nil
				})

			err := manager.UpdateSubjectPoliciesExpiredAt("user", "test", []types.PolicyPKExpiredAt{{
				PK:        1,
				ExpiredAt: 1,
			}})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("CreateAndDeleteTemplatePolicies", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.CreateAndDeleteTemplatePolicies("test", "user", "test", int64(1), []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("actionService.ListThinActionBySystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.CreateAndDeleteTemplatePolicies("test", "user", "test", int64(1), []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListThinActionBySystem")
		})

		It("actionService.ListActionResourceTypeIDByActionSystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.CreateAndDeleteTemplatePolicies("test", "user", "test", int64(1), []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListActionResourceTypeIDByActionSystem")
		})

		It("ErrActionNotExists fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.CreateAndDeleteTemplatePolicies("test", "user", "test", int64(1), []types.Policy{{
				Action: types.Action{
					ID: "test",
				},
			}}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "action not exists")
		})

		It("policyService.CreateAndDeleteTemplatePolicies fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().CreateAndDeleteTemplatePolicies(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			).Return(errors.New("create policies fail")).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,
			}

			err := manager.CreateAndDeleteTemplatePolicies("test", "user", "test", int64(1), []types.Policy{}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.CreateAndDeleteTemplatePolicies")
		})

		It("CreateAndDeleteTemplatePolicies success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().CreateAndDeleteTemplatePolicies(
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
			).Return(nil).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,
			}

			err := manager.CreateAndDeleteTemplatePolicies("test", "user", "test", int64(1), []types.Policy{}, []int64{1})
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("UpdateTemplatePolicies", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.UpdateTemplatePolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("actionService.ListThinActionBySystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.UpdateTemplatePolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListThinActionBySystem")
		})

		It("actionService.ListActionResourceTypeIDByActionSystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.UpdateTemplatePolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListActionResourceTypeIDByActionSystem")
		})

		It("ErrActionNotExists fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := manager.UpdateTemplatePolicies("test", "user", "test", []types.Policy{{
				Action: types.Action{
					ID: "test",
				},
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "action not exists")
		})

		It("policyService.UpdateTemplatePolicies fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().UpdateTemplatePolicies(
				gomock.Any(), gomock.Any(), gomock.Any(),
			).Return(errors.New("update fail")).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,
			}

			err := manager.UpdateTemplatePolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.UpdateTemplatePolicies")
		})

		It("UpdateTemplatePolicies success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().UpdateTemplatePolicies(
				gomock.Any(), gomock.Any(), gomock.Any(),
			).Return(nil).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,
			}

			err := manager.UpdateTemplatePolicies("test", "user", "test", []types.Policy{})
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("DeleteTemplatePolicies", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.DeleteTemplatePolicies("test", "user", "test", int64(1))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("policyService.DeleteTemplatePolicies fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().DeleteTemplatePolicies(
				int64(1), int64(1),
			).Return(
				errors.New("delete fail"),
			).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
			}

			err := manager.DeleteTemplatePolicies("test", "user", "test", int64(1))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.DeleteTemplatePolicies")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().DeleteTemplatePolicies(
				int64(1), int64(1),
			).Return(
				nil,
			).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			manager := &policyManager{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
			}

			err := manager.DeleteTemplatePolicies("test", "user", "test", int64(1))
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("DeleteTemporaryByIDs", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			err := manager.DeleteTemporaryByIDs("test", "user", "test", []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("temporaryPolicyService.DeleteByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().DeleteByPKs(
				int64(1), []int64{1, 2},
			).Return(
				errors.New("delete fail"),
			).AnyTimes()

			mockTemporaryPolicyRedisCache := &temporaryPolicyRedisCache{}
			patches = gomonkey.ApplyMethod(
				reflect.TypeOf(mockTemporaryPolicyRedisCache), "DeleteBySubject",
				func(c *temporaryPolicyRedisCache, subjectPK int64) error {
					return nil
				})

			patches = gomonkey.ApplyFunc(newTemporaryPolicyRedisCache,
				func(systemID string, svc service.TemporaryPolicyService) *temporaryPolicyRedisCache {
					return mockTemporaryPolicyRedisCache
				})

			manager := &policyManager{
				subjectService:         mockSubjectService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := manager.DeleteTemporaryByIDs("test", "user", "test", []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "temporaryPolicyService.DeleteByPKs")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().DeleteByPKs(
				int64(1), []int64{1, 2},
			).Return(
				nil,
			).AnyTimes()

			mockTemporaryPolicyRedisCache := &temporaryPolicyRedisCache{}
			patches = gomonkey.ApplyMethod(
				reflect.TypeOf(mockTemporaryPolicyRedisCache), "DeleteBySubject",
				func(c *temporaryPolicyRedisCache, subjectPK int64) error {
					return nil
				})

			patches = gomonkey.ApplyFunc(newTemporaryPolicyRedisCache,
				func(systemID string, svc service.TemporaryPolicyService) *temporaryPolicyRedisCache {
					return mockTemporaryPolicyRedisCache
				})

			manager := &policyManager{
				subjectService:         mockSubjectService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := manager.DeleteTemporaryByIDs("test", "user", "test", []int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("CreateTemporaryPolicies", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectService.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
			}

			_, err := manager.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectService.GetPK")
		})

		It("actionService.ListThinActionBySystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			_, err := manager.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListThinActionBySystem")
		})

		It("actionService.ListActionResourceTypeIDByActionSystem fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, errors.New("list action fail"),
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			_, err := manager.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionService.ListActionResourceTypeIDByActionSystem")
		})

		It("ErrActionNotExists fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()

			manager := &policyManager{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			_, err := manager.CreateTemporaryPolicies("test", "user", "test", []types.Policy{{
				Action: types.Action{
					ID: "test",
				},
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "action not exists")
		})

		It("temporaryPolicyService.Create fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().Create(
				gomock.Any(),
			).Return(
				[]int64{}, errors.New("create fail"),
			).AnyTimes()

			mockTemporaryPolicyRedisCache := &temporaryPolicyRedisCache{}
			patches = gomonkey.ApplyMethod(
				reflect.TypeOf(mockTemporaryPolicyRedisCache), "DeleteBySubject",
				func(c *temporaryPolicyRedisCache, subjectPK int64) error {
					return nil
				})

			patches = gomonkey.ApplyFunc(newTemporaryPolicyRedisCache,
				func(systemID string, svc service.TemporaryPolicyService) *temporaryPolicyRedisCache {
					return mockTemporaryPolicyRedisCache
				})

			manager := &policyManager{
				subjectService:         mockSubjectService,
				actionService:          mockActionService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			_, err := manager.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "temporaryPolicyService.Create")
		})

		It("CreateTemporaryPolicies success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]svctypes.ThinAction{}, nil,
			).AnyTimes()
			mockActionService.EXPECT().ListActionResourceTypeIDByActionSystem("test").Return(
				[]svctypes.ActionResourceTypeID{}, nil,
			).AnyTimes()
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().Create(
				gomock.Any(),
			).Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockTemporaryPolicyRedisCache := &temporaryPolicyRedisCache{}
			patches = gomonkey.ApplyMethod(
				reflect.TypeOf(mockTemporaryPolicyRedisCache), "DeleteBySubject",
				func(c *temporaryPolicyRedisCache, subjectPK int64) error {
					return nil
				})

			patches = gomonkey.ApplyFunc(newTemporaryPolicyRedisCache,
				func(systemID string, svc service.TemporaryPolicyService) *temporaryPolicyRedisCache {
					return mockTemporaryPolicyRedisCache
				})

			manager := &policyManager{
				subjectService:         mockSubjectService,
				actionService:          mockActionService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			pks, err := manager.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []int64{1}, pks)
		})
	})

	Describe("DeleteTemporaryBeforeExpiredAt", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("temporaryPolicyService.DeleteBeforeExpireAt fail", func() {
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().DeleteBeforeExpireAt(
				int64(1),
			).Return(
				errors.New("delete fail"),
			).AnyTimes()

			manager := &policyManager{
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := manager.DeleteTemporaryBeforeExpiredAt(int64(1))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "temporaryPolicyService.DeleteBeforeExpireAt")
		})

		It("success", func() {
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().DeleteBeforeExpireAt(
				int64(1),
			).Return(
				nil,
			).AnyTimes()

			manager := &policyManager{
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := manager.DeleteTemporaryBeforeExpiredAt(int64(1))
			assert.NoError(GinkgoT(), err)
		})

	})
})
