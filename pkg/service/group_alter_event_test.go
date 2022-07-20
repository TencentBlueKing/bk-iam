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

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("GroupAlterEventService", func() {
	Describe("CreateByGroupAction cases", func() {
		var ctl *gomock.Controller
		var svc GroupAlterEventService

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().ListGroupMember(int64(1)).Return([]dao.SubjectRelation{
				{SubjectPK: 11},
				{SubjectPK: 12},
			}, nil)

			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(dao.GroupAlterEvent{
				GroupPK:    1,
				SubjectPKs: "[11,12]",
				ActionPKs:  "[1,2]",
			}).Return(nil)

			svc = &groupAlterEventService{
				manager:             mockManager,
				subjectGroupManager: mockSubjectGroupManager,
			}

			event, err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.NoError(GinkgoT(), err)

			assert.Equal(GinkgoT(), types.GroupAlterEvent{
				GroupPK:    1,
				SubjectPKs: []int64{11, 12},
				ActionPKs:  []int64{1, 2},
			}, event)
		})

		It("empty action pks", func() {
			svc = &groupAlterEventService{}
			_, err := svc.CreateByGroupAction(1, []int64{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "empty group alter event")
		})

		It("empty subject pks", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().ListGroupMember(int64(1)).Return([]dao.SubjectRelation{}, nil)

			svc = &groupAlterEventService{
				subjectGroupManager: mockSubjectGroupManager,
			}
			_, err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "empty group alter event")
		})

		It("ListGroupMember fail", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().
				ListGroupMember(int64(1)).
				Return([]dao.SubjectRelation{}, errors.New("error"))

			svc = &groupAlterEventService{
				subjectGroupManager: mockSubjectGroupManager,
			}
			_, err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListGroupMember")
		})

		It("createEvent fail", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().ListGroupMember(int64(1)).Return([]dao.SubjectRelation{
				{SubjectPK: 11},
				{SubjectPK: 12},
			}, nil)

			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(dao.GroupAlterEvent{
				GroupPK:    1,
				SubjectPKs: "[11,12]",
				ActionPKs:  "[1,2]",
			}).Return(errors.New("error"))

			svc = &groupAlterEventService{
				manager:             mockManager,
				subjectGroupManager: mockSubjectGroupManager,
			}

			_, err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "createEvent")
		})
	})

	Describe("CreateByGroupSubject cases", func() {
		var ctl *gomock.Controller
		var svc GroupAlterEventService

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockGroupResourcePolicyManager := mock.NewMockGroupResourcePolicyManager(ctl)
			mockGroupResourcePolicyManager.EXPECT().
				ListActionPKsByGroup(int64(1)).
				Return([]string{"[1,2]", "[2,3]"}, nil)

			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(gomock.Any()).Return(nil)

			svc = &groupAlterEventService{
				manager:                    mockManager,
				groupResourcePolicyManager: mockGroupResourcePolicyManager,
			}

			event, err := svc.CreateByGroupSubject(1, []int64{11, 12})
			assert.NoError(GinkgoT(), err)

			assert.Equal(GinkgoT(), int64(1), event.GroupPK)
			assert.Equal(GinkgoT(), []int64{11, 12}, event.SubjectPKs)
			assert.Len(GinkgoT(), event.ActionPKs, 3)
		})

		It("empty subject pks", func() {
			svc = &groupAlterEventService{}
			_, err := svc.CreateByGroupSubject(1, []int64{})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "empty group alter event")
		})

		It("empty action pks", func() {
			mockGroupResourcePolicyManager := mock.NewMockGroupResourcePolicyManager(ctl)
			mockGroupResourcePolicyManager.EXPECT().ListActionPKsByGroup(int64(1)).Return([]string{}, nil)

			svc = &groupAlterEventService{
				groupResourcePolicyManager: mockGroupResourcePolicyManager,
			}
			_, err := svc.CreateByGroupSubject(1, []int64{11, 12})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "empty group alter event")
		})

		It("ListActionPKsByGroup fail", func() {
			mockGroupResourcePolicyManager := mock.NewMockGroupResourcePolicyManager(ctl)
			mockGroupResourcePolicyManager.EXPECT().
				ListActionPKsByGroup(int64(1)).
				Return([]string{}, errors.New("error"))

			svc = &groupAlterEventService{
				groupResourcePolicyManager: mockGroupResourcePolicyManager,
			}
			_, err := svc.CreateByGroupSubject(1, []int64{11, 12})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListActionPKsByGroup")
		})

		It("createEvent fail", func() {
			mockGroupResourcePolicyManager := mock.NewMockGroupResourcePolicyManager(ctl)
			mockGroupResourcePolicyManager.EXPECT().
				ListActionPKsByGroup(int64(1)).
				Return([]string{"[1,2]", "[2,3]"}, nil)

			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(gomock.Any()).Return(errors.New("error"))

			svc = &groupAlterEventService{
				manager:                    mockManager,
				groupResourcePolicyManager: mockGroupResourcePolicyManager,
			}

			_, err := svc.CreateByGroupSubject(1, []int64{11, 12})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "createEvent")
		})
	})

	Describe("ListByGroupStatus cases", func() {
		var ctl *gomock.Controller
		var svc GroupAlterEventService

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().ListByGroupStatus(int64(1), int64(0)).Return([]dao.GroupAlterEvent{
				{GroupPK: 1, ActionPKs: "[1,2]", SubjectPKs: "[11,12]"},
			}, nil)

			svc = &groupAlterEventService{
				manager: mockManager,
			}

			events, err := svc.ListUncheckedByGroup(1)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAlterEvent{
				{GroupPK: 1, ActionPKs: []int64{1, 2}, SubjectPKs: []int64{11, 12}},
			}, events)
		})

		It("ListByGroupStatus fail", func() {
			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().
				ListByGroupStatus(int64(1), int64(0)).
				Return([]dao.GroupAlterEvent{}, errors.New("error"))

			svc = &groupAlterEventService{
				manager: mockManager,
			}

			_, err := svc.ListUncheckedByGroup(1)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByGroupStatus")
		})
	})
})
