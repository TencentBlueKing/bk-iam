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
	"database/sql"
	"errors"
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("GroupAlterEventService", func() {
	Describe("CreateByGroupAction cases", func() {
		var ctl *gomock.Controller
		var svc GroupAlterEventService
		var patches *gomonkey.Patches

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			tx := &sql.Tx{}
			patches = gomonkey.ApplyMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
				return nil
			})

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return &sqlx.Tx{Tx: tx}, nil
			})
			patches.ApplyFunc(database.RollBackWithLog, func(tx *sqlx.Tx) {})
		})

		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("ok", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().ListGroupMember(int64(1)).Return([]dao.SubjectRelation{
				{SubjectPK: 11},
				{SubjectPK: 12},
			}, nil)

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().ListGroupDistinctSubjectPK(int64(1)).Return([]int64{
				11,
			}, nil)

			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(gomock.Any()).Return(nil)

			svc = &groupAlterEventService{
				manager:                     mockManager,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
				subjectGroupManager:         mockSubjectGroupManager,
			}

			err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})

		It("empty action pks", func() {
			svc = &groupAlterEventService{}
			err := svc.CreateByGroupAction(1, []int64{})
			assert.NoError(GinkgoT(), err)
		})

		It("empty subject pks", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().ListGroupMember(int64(1)).Return([]dao.SubjectRelation{}, nil)
			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().ListGroupDistinctSubjectPK(int64(1)).Return(
				[]int64{}, nil,
			)

			svc = &groupAlterEventService{
				subjectGroupManager:         mockSubjectGroupManager,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}
			err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})

		It("ListGroupMember fail", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().
				ListGroupMember(int64(1)).
				Return([]dao.SubjectRelation{}, errors.New("error"))

			svc = &groupAlterEventService{
				subjectGroupManager: mockSubjectGroupManager,
			}
			err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListGroupMember")
		})

		It("createEvent fail", func() {
			mockSubjectGroupManager := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectGroupManager.EXPECT().ListGroupMember(int64(1)).Return([]dao.SubjectRelation{
				{SubjectPK: 11},
				{SubjectPK: 12},
			}, nil)

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().ListGroupDistinctSubjectPK(int64(1)).Return([]int64{
				11,
			}, nil)

			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(gomock.Any()).Return(errors.New("error"))

			svc = &groupAlterEventService{
				manager:                     mockManager,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
				subjectGroupManager:         mockSubjectGroupManager,
			}

			err := svc.CreateByGroupAction(1, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "create")
		})
	})

	Describe("CreateByGroupSubject cases", func() {
		var ctl *gomock.Controller
		var svc GroupAlterEventService
		var patches *gomonkey.Patches

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			tx := &sql.Tx{}
			patches = gomonkey.ApplyMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
				return nil
			})

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return &sqlx.Tx{Tx: tx}, nil
			})
			patches.ApplyFunc(database.RollBackWithLog, func(tx *sqlx.Tx) {})
		})

		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
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

			err := svc.CreateByGroupSubject(1, []int64{11, 12})
			assert.NoError(GinkgoT(), err)
		})

		It("empty subject pks", func() {
			svc = &groupAlterEventService{}
			err := svc.CreateByGroupSubject(1, []int64{})
			assert.NoError(GinkgoT(), err)
		})

		It("empty action pks", func() {
			mockGroupResourcePolicyManager := mock.NewMockGroupResourcePolicyManager(ctl)
			mockGroupResourcePolicyManager.EXPECT().ListActionPKsByGroup(int64(1)).Return([]string{}, nil)

			svc = &groupAlterEventService{
				groupResourcePolicyManager: mockGroupResourcePolicyManager,
			}
			err := svc.CreateByGroupSubject(1, []int64{11, 12})
			assert.NoError(GinkgoT(), err)
		})

		It("ListActionPKsByGroup fail", func() {
			mockGroupResourcePolicyManager := mock.NewMockGroupResourcePolicyManager(ctl)
			mockGroupResourcePolicyManager.EXPECT().
				ListActionPKsByGroup(int64(1)).
				Return([]string{}, errors.New("error"))

			svc = &groupAlterEventService{
				groupResourcePolicyManager: mockGroupResourcePolicyManager,
			}
			err := svc.CreateByGroupSubject(1, []int64{11, 12})
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

			err := svc.CreateByGroupSubject(1, []int64{11, 12})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "create")
		})
	})

	Describe("ListBeforeCreateAt cases", func() {
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
			mockManager.EXPECT().ListBeforeCreateAt(int64(2), int64(3)).Return([]dao.GroupAlterEvent{
				{
					UUID:       "uuid",
					GroupPK:    1,
					ActionPKs:  "[1,2]",
					SubjectPKs: "[11,12]",
				},
			}, nil)

			svc = &groupAlterEventService{
				manager: mockManager,
			}

			events, err := svc.ListBeforeCreateAt(2, 3)
			assert.NoError(GinkgoT(), err)

			assert.Equal(GinkgoT(), []types.GroupAlterEvent{
				{
					UUID:       "uuid",
					GroupPK:    1,
					ActionPKs:  []int64{1, 2},
					SubjectPKs: []int64{11, 12},
				},
			}, events)
		})

		It("ListPKByCheckCountBeforeCreateAt fail", func() {
			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().ListBeforeCreateAt(int64(2), int64(3)).Return(nil, errors.New("error"))

			svc = &groupAlterEventService{
				manager: mockManager,
			}

			_, err := svc.ListBeforeCreateAt(2, 3)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "error")
		})
	})

	Describe("CreateBySubjectActionGroup cases", func() {
		var ctl *gomock.Controller
		var svc GroupAlterEventService
		var patches *gomonkey.Patches

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			tx := &sql.Tx{}
			patches = gomonkey.ApplyMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
				return nil
			})

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return &sqlx.Tx{Tx: tx}, nil
			})
			patches.ApplyFunc(database.RollBackWithLog, func(tx *sqlx.Tx) {})
		})

		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("ok", func() {
			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(gomock.Any()).Return(nil)

			svc = &groupAlterEventService{
				manager: mockManager,
			}

			err := svc.CreateBySubjectActionGroup(1, 1, 0)
			assert.NoError(GinkgoT(), err)
		})

		It("create fail", func() {
			mockManager := mock.NewMockGroupAlterEventManager(ctl)
			mockManager.EXPECT().Create(gomock.Any()).Return(errors.New("error"))

			svc = &groupAlterEventService{
				manager: mockManager,
			}

			err := svc.CreateBySubjectActionGroup(1, 1, 0)
			assert.Error(GinkgoT(), err)

			assert.Contains(GinkgoT(), err.Error(), "create")
		})
	})
})
