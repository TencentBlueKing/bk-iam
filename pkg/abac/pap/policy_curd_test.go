/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pap

import (
	"errors"

	eventmock "iam/pkg/abac/pap/event/mock"
	"iam/pkg/abac/prp/policy"
	"iam/pkg/abac/prp/temporary"
	"iam/pkg/abac/types"
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
			}

			err := policyCtl.DeleteByIDs("test", "user", "test", []int64{1, 2})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,
			}

			err := policyCtl.DeleteByIDs("test", "user", "test", []int64{1, 2})
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

			mockPolicyEventProducer := eventmock.NewMockPolicyEventProducer(ctl)
			mockPolicyEventProducer.EXPECT().PublishABACDeletePolicyEvent(gomock.Any()).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				policyService:  mockPolicyService,

				eventProducer: mockPolicyEventProducer,
			}

			err := policyCtl.DeleteByIDs("test", "user", "test", []int64{1, 2})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
			}

			err := policyCtl.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := policyCtl.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := policyCtl.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := policyCtl.AlterCustomPolicies("test", "user", "test", []types.Policy{{
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			err := policyCtl.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{{
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,
			}

			err := policyCtl.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
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

			mockPolicyEventProducer := eventmock.NewMockPolicyEventProducer(ctl)
			mockPolicyEventProducer.EXPECT().PublishABACDeletePolicyEvent(gomock.Any()).AnyTimes()

			patches = gomonkey.ApplyFunc(policy.DeleteSystemSubjectPKsFromCache,
				func(systemID string, pks []int64) error {
					return nil
				})

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
				policyService:  mockPolicyService,

				eventProducer: mockPolicyEventProducer,
			}

			err := policyCtl.AlterCustomPolicies("test", "user", "test", []types.Policy{}, []types.Policy{}, []int64{1})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
			}

			err := policyCtl.DeleteTemporaryByIDs("test", "user", "test", []int64{1, 2})
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

			patches = gomonkey.ApplyFunc(temporary.DeletePolicyBySystemSubjectFromCache,
				func(systemID string, subjectPK int64) error {
					return nil
				})

			policyCtl := &policyController{
				subjectService:         mockSubjectService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := policyCtl.DeleteTemporaryByIDs("test", "user", "test", []int64{1, 2})
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

			patches = gomonkey.ApplyFunc(temporary.DeletePolicyBySystemSubjectFromCache,
				func(systemID string, subjectPK int64) error {
					return nil
				})

			policyCtl := &policyController{
				subjectService:         mockSubjectService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := policyCtl.DeleteTemporaryByIDs("test", "user", "test", []int64{1, 2})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
			}

			_, err := policyCtl.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			_, err := policyCtl.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			_, err := policyCtl.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
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

			policyCtl := &policyController{
				subjectService: mockSubjectService,
				actionService:  mockActionService,
			}

			_, err := policyCtl.CreateTemporaryPolicies("test", "user", "test", []types.Policy{{
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

			patches = gomonkey.ApplyFunc(temporary.DeletePolicyBySystemSubjectFromCache,
				func(systemID string, subjectPK int64) error {
					return nil
				})

			policyCtl := &policyController{
				subjectService:         mockSubjectService,
				actionService:          mockActionService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			_, err := policyCtl.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
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

			patches = gomonkey.ApplyFunc(temporary.DeletePolicyBySystemSubjectFromCache,
				func(systemID string, subjectPK int64) error {
					return nil
				})

			policyCtl := &policyController{
				subjectService:         mockSubjectService,
				actionService:          mockActionService,
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			pks, err := policyCtl.CreateTemporaryPolicies("test", "user", "test", []types.Policy{})
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

		It("temporaryPolicyService.DeleteBeforeExpiredAt fail", func() {
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().DeleteBeforeExpiredAt(
				int64(1),
			).Return(
				errors.New("delete fail"),
			).AnyTimes()

			policyCtl := &policyController{
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := policyCtl.DeleteTemporaryBeforeExpiredAt(int64(1))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "temporaryPolicyService.DeleteBeforeExpiredAt")
		})

		It("success", func() {
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().DeleteBeforeExpiredAt(
				int64(1),
			).Return(
				nil,
			).AnyTimes()

			policyCtl := &policyController{
				temporaryPolicyService: mockTemporaryPolicyService,
			}

			err := policyCtl.DeleteTemporaryBeforeExpiredAt(int64(1))
			assert.NoError(GinkgoT(), err)
		})
	})
})
