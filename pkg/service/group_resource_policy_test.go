/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"errors"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/golang/mock/gomock"
	jsoniter "github.com/json-iterator/go"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

func assertJsonStringOfInt64Slice(t assert.TestingT, expected string, input string) {
	var (
		e []int64
		i []int64
	)

	err := jsoniter.UnmarshalFromString(expected, &e)
	assert.NoError(t, err)

	err = jsoniter.UnmarshalFromString(input, &i)
	assert.NoError(t, err)

	assert.ElementsMatch(t, e, i)
}

var _ = Describe("GroupResourcePolicyService", func() {
	Describe("inter func", func() {
		var (
			ctl      *gomock.Controller
			interSvc groupResourcePolicyService
		)
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			interSvc = groupResourcePolicyService{
				manager: mock.NewMockGroupResourcePolicyManager(ctl),
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})
		Context("calculateSignature", func() {
			It("same params and diff params", func() {
				s1 := interSvc.calculateSignature(
					int64(1), int64(2), "test",
					types.ResourceChangedContent{
						ActionRelatedResourceTypePK: int64(3),
						ResourceTypePK:              int64(4),
						ResourceID:                  "resource_id",
					},
				)
				s2 := interSvc.calculateSignature(
					int64(1), int64(2), "test",
					types.ResourceChangedContent{
						ActionRelatedResourceTypePK: int64(3),
						ResourceTypePK:              int64(4),
						ResourceID:                  "resource_id",
					},
				)
				s3 := interSvc.calculateSignature(
					int64(2), int64(2), "test",
					types.ResourceChangedContent{
						ActionRelatedResourceTypePK: int64(3),
						ResourceTypePK:              int64(4),
						ResourceID:                  "resource_id",
					},
				)
				assert.Equal(GinkgoT(), s1, s2)
				assert.NotEqual(GinkgoT(), s1, s3)
				assert.NotEqual(GinkgoT(), s2, s3)
			})
		})

		Context("calculateChangedActionPKs", func() {
			It("json loads error", func() {
				aks, err := interSvc.calculateChangedActionPKs("[x]", types.ResourceChangedContent{}, set.NewInt64Set())
				assert.Regexp(GinkgoT(), "jsoniter.UnmarshalFromString (.*) fail", err.Error())
				assert.Equal(GinkgoT(), "", aks)
			})
			It("old_action_pks empty", func() {
				aks, err := interSvc.calculateChangedActionPKs("", types.ResourceChangedContent{
					CreatedActionPKs: []int64{1, 2, 3},
				}, set.NewInt64SetWithValues([]int64{1, 2, 3}))
				assert.NoError(GinkgoT(), err)
				assertJsonStringOfInt64Slice(GinkgoT(), `[1,2,3]`, aks)
				assertJsonStringOfInt64Slice(GinkgoT(), `[2,1,3]`, aks)
			})
			It("old_action_pks not empty and new action_pks empty", func() {
				aks, err := interSvc.calculateChangedActionPKs("[1, 2]", types.ResourceChangedContent{
					DeletedActionPKs: []int64{1, 2, 3},
				}, set.NewInt64SetWithValues([]int64{1, 2, 3}))
				assert.NoError(GinkgoT(), err)
				assert.Empty(GinkgoT(), aks)
			})
			It("old_action_pks not empty and new action_pks not empty", func() {
				aks, err := interSvc.calculateChangedActionPKs("[1, 2]", types.ResourceChangedContent{
					CreatedActionPKs: []int64{4, 5},
					DeletedActionPKs: []int64{1},
				}, set.NewInt64SetWithValues([]int64{1, 2, 3}))
				assert.NoError(GinkgoT(), err)
				assertJsonStringOfInt64Slice(GinkgoT(), `[5, 4, 2]`, aks)
			})
		})
	})

	Context("Alter Policy", func() {
		var (
			ctl               *gomock.Controller
			mockManager       *mock.MockGroupResourcePolicyManager
			mockActionManager *mock.MockActionManager
			svc               GroupResourcePolicyService

			rcc       types.ResourceChangedContent
			signature string
			daoPolicy dao.GroupResourcePolicy
		)
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			mockManager = mock.NewMockGroupResourcePolicyManager(ctl)
			mockActionManager = mock.NewMockActionManager(ctl)
			svc = &groupResourcePolicyService{
				manager:       mockManager,
				actionManager: mockActionManager,
			}

			rcc = types.ResourceChangedContent{
				ActionRelatedResourceTypePK: int64(3),
				ResourceTypePK:              int64(4),
				ResourceID:                  "resource_id",
			}
			interSvc := groupResourcePolicyService{
				manager: mock.NewMockGroupResourcePolicyManager(ctl),
			}
			signature = interSvc.calculateSignature(int64(1), int64(2), "test", rcc)
			daoPolicy = dao.GroupResourcePolicy{
				PK:                          int64(1),
				Signature:                   signature,
				GroupPK:                     int64(1),
				TemplateID:                  int64(2),
				SystemID:                    "test",
				ActionRelatedResourceTypePK: int64(3),
				ResourceTypePK:              int64(4),
				ResourceID:                  "resource_id",
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("ListBySignatures error", func() {
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return(nil, errors.New("error"))

			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "manager.ListBySignatures fail", err.Error())
		})

		It("calculateChangedActionPKs error", func() {
			daoPolicy.ActionPKs = "json error"
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return([]dao.GroupResourcePolicy{daoPolicy}, nil)
			mockActionManager.EXPECT().ListPKBySystem(gomock.Any()).Return([]int64{1}, nil)

			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "calculateChangedActionPKs fail", err.Error())
		})

		It("BulkCreateWithTx error", func() {
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return([]dao.GroupResourcePolicy{}, nil)
			mockManager.EXPECT().BulkCreateWithTx(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			mockActionManager.EXPECT().ListPKBySystem(gomock.Any()).Return([]int64{1}, nil)

			rcc.CreatedActionPKs = []int64{1}
			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "manager.BulkCreateWithTx fail", err.Error())
		})

		It("BulkCreateWithTx ok", func() {
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return([]dao.GroupResourcePolicy{}, nil)
			mockManager.EXPECT().BulkCreateWithTx(gomock.Any(), gomock.Any()).Return(nil)
			mockActionManager.EXPECT().ListPKBySystem(gomock.Any()).Return([]int64{1}, nil)

			rcc.CreatedActionPKs = []int64{1}
			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.NoError(GinkgoT(), err)
		})

		It("BulkDeleteByPKsWithTx error", func() {
			daoPolicy.ActionPKs = "[1]"
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return([]dao.GroupResourcePolicy{daoPolicy}, nil)
			mockManager.EXPECT().BulkDeleteByPKsWithTx(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			mockActionManager.EXPECT().ListPKBySystem(gomock.Any()).Return([]int64{1}, nil)

			rcc.DeletedActionPKs = []int64{1}
			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "manager.BulkDeleteByPKsWithTx fail", err.Error())
		})

		It("BulkDeleteByPKsWithTx ok", func() {
			daoPolicy.ActionPKs = "[1]"
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return([]dao.GroupResourcePolicy{daoPolicy}, nil)
			mockManager.EXPECT().BulkDeleteByPKsWithTx(gomock.Any(), gomock.Any()).Return(nil)
			mockActionManager.EXPECT().ListPKBySystem(gomock.Any()).Return([]int64{1}, nil)

			rcc.DeletedActionPKs = []int64{1}
			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.NoError(GinkgoT(), err)
		})

		It("BulkUpdateActionPKsWithTx error", func() {
			daoPolicy.ActionPKs = "[1,2]"
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return([]dao.GroupResourcePolicy{daoPolicy}, nil)
			mockManager.EXPECT().BulkUpdateActionPKsWithTx(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			mockActionManager.EXPECT().ListPKBySystem(gomock.Any()).Return([]int64{1, 3, 4}, nil)

			rcc.CreatedActionPKs = []int64{3, 4}
			rcc.DeletedActionPKs = []int64{1}
			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "manager.BulkUpdateActionPKsWithTx fail", err.Error())
		})

		It("BulkUpdateActionPKsWithTx ok", func() {
			daoPolicy.ActionPKs = "[1,2]"
			mockManager.EXPECT().ListBySignatures([]string{signature}).Return([]dao.GroupResourcePolicy{daoPolicy}, nil)
			mockManager.EXPECT().BulkUpdateActionPKsWithTx(gomock.Any(), gomock.Any()).Return(nil)
			mockActionManager.EXPECT().ListPKBySystem(gomock.Any()).Return([]int64{1, 3, 4}, nil)

			rcc.CreatedActionPKs = []int64{3, 4}
			rcc.DeletedActionPKs = []int64{1}
			err := svc.Alter(nil, int64(1), int64(2), "test", []types.ResourceChangedContent{rcc})
			assert.NoError(GinkgoT(), err)
		})
	})

	Context("GetAuthorizedActionGroupMap", func() {
		var (
			ctl         *gomock.Controller
			mockManager *mock.MockGroupResourcePolicyManager
			svc         GroupResourcePolicyService
		)
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			mockManager = mock.NewMockGroupResourcePolicyManager(ctl)
			svc = &groupResourcePolicyService{
				manager: mockManager,
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListThinByResource error", func() {
			mockManager.EXPECT().
				ListThinByResource("test", int64(1), int64(2), "resource_test").
				Return(nil, errors.New("error"))

			_, err := svc.GetAuthorizedActionGroupMap("test", int64(1), int64(2), "resource_test")
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "manager.ListThinByResource fail", err.Error())
		})

		It("jsoniter.UnmarshalFromString error", func() {
			mockManager.EXPECT().
				ListThinByResource("test", int64(1), int64(2), "resource_test").
				Return([]dao.ThinGroupResourcePolicy{{
					GroupPK:   int64(1),
					ActionPKs: "xxx",
				}}, nil)

			_, err := svc.GetAuthorizedActionGroupMap("test", int64(1), int64(2), "resource_test")
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "jsoniter.UnmarshalFromString fail", err.Error())
		})

		It("ok", func() {
			mockManager.EXPECT().
				ListThinByResource("test", int64(1), int64(2), "resource_test").
				Return([]dao.ThinGroupResourcePolicy{{
					GroupPK:   int64(1),
					ActionPKs: "[1,2,3]",
				}, {
					GroupPK:   int64(2),
					ActionPKs: "[3]",
				}, {
					GroupPK:   int64(1),
					ActionPKs: "[3,4]",
				}}, nil)

			actionGroupPKs, err := svc.GetAuthorizedActionGroupMap("test", int64(1), int64(2), "resource_test")
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), actionGroupPKs, 4)
			for actionPK, groupPKs := range actionGroupPKs {
				switch actionPK {
				case int64(1):
					assert.Equal(GinkgoT(), []int64{1}, groupPKs)
				case int64(2):
					assert.Equal(GinkgoT(), []int64{1}, groupPKs)
				case int64(3):
					assert.Len(GinkgoT(), groupPKs, 2)
				case int64(4):
					assert.Equal(GinkgoT(), []int64{1}, groupPKs)
				}
			}
		})
	})

	Context("ListResourceByGroupAction", func() {
		var (
			ctl         *gomock.Controller
			mockManager *mock.MockGroupResourcePolicyManager
			svc         GroupResourcePolicyService
		)
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			mockManager = mock.NewMockGroupResourcePolicyManager(ctl)
			svc = &groupResourcePolicyService{
				manager: mockManager,
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListByGroupSystemActionRelatedResourceType error", func() {
			mockManager.EXPECT().
				ListByGroupSystemActionRelatedResourceType(int64(1), "test", int64(1)).
				Return(nil, errors.New("error"))

			_, err := svc.ListResourceByGroupAction(int64(1), "test", int64(1), int64(1))
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "manager.ListByGroupSystemActionRelatedResourceType fail", err.Error())
		})

		It("jsoniter.UnmarshalFromString error", func() {
			mockManager.EXPECT().
				ListByGroupSystemActionRelatedResourceType(int64(1), "test", int64(1)).
				Return([]dao.GroupResourcePolicy{{
					GroupPK:   int64(1),
					ActionPKs: "xxx",
				}}, nil)

			_, err := svc.ListResourceByGroupAction(int64(1), "test", int64(1), int64(1))
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "jsoniter.UnmarshalFromString fail", err.Error())
		})

		It("ok", func() {
			mockManager.EXPECT().
				ListByGroupSystemActionRelatedResourceType(int64(1), "test", int64(1)).
				Return([]dao.GroupResourcePolicy{{
					GroupPK:                     int64(1),
					ActionPKs:                   "[1,2]",
					ActionRelatedResourceTypePK: int64(1),
					ResourceTypePK:              int64(1),
					ResourceID:                  "test1",
				}, {
					GroupPK:                     int64(1),
					ActionPKs:                   "[1,2]",
					ActionRelatedResourceTypePK: int64(1),
					ResourceTypePK:              int64(1),
					ResourceID:                  "test2",
				}, {
					GroupPK:                     int64(1),
					ActionPKs:                   "[1,2]",
					ActionRelatedResourceTypePK: int64(1),
					ResourceTypePK:              int64(2),
					ResourceID:                  "test3",
				}, {
					GroupPK:                     int64(1),
					ActionPKs:                   "[2]",
					ActionRelatedResourceTypePK: int64(1),
					ResourceTypePK:              int64(2),
					ResourceID:                  "test4",
				}}, nil)

			resources, err := svc.ListResourceByGroupAction(int64(1), "test", int64(1), int64(1))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.Resource{
				{ResourceTypePK: int64(1), ResourceID: "test1"},
				{ResourceTypePK: int64(1), ResourceID: "test2"},
				{ResourceTypePK: int64(2), ResourceID: "test3"},
			}, resources)
		})
	})
})
