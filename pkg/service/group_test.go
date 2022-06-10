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

var _ = Describe("GroupService", func() {
	Describe("ListMember", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListMember fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().ListMember(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListMember(1)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListMember")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().ListMember(int64(1)).Return(
				[]dao.SubjectRelation{}, nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			groupMembers, err := manager.ListMember(1)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupMember{}, groupMembers)
		})
	})

	Describe("BulkDeleteBySubjectPKsWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.BulkDeleteByParentPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByParentPKs(gomock.Any(), []int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteByParentPKs")
		})

		It("manager.BulkDeleteBySubjectPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByParentPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectService.EXPECT().BulkDeleteBySubjectPKs(gomock.Any(), []int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteBySubjectPKs")
		})

		It("subjectSystemGroupManager.DeleteBySubjectPKsWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByParentPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectService.EXPECT().BulkDeleteBySubjectPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectSystemGroupService := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupService.EXPECT().DeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager:                   mockSubjectService,
				subjectSystemGroupManager: mockSubjectSystemGroupService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "DeleteBySubjectPKsWithTx")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByParentPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectService.EXPECT().BulkDeleteBySubjectPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectSystemGroupService := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupService.EXPECT().DeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				manager:                   mockSubjectService,
				subjectSystemGroupManager: mockSubjectSystemGroupService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("UpdateMembersExpiredAtWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().UpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelationPKPolicyExpiredAt{
				{
					PK:              1,
					PolicyExpiredAt: 2,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.UpdateMembersExpiredAtWithTx(nil, int64(1), []types.SubjectRelationPKPolicyExpiredAt{
				{
					PK:              1,
					PolicyExpiredAt: 2,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateExpiredAtWithTx")
		})
	})

	Describe("BulkCreateGroupMembersWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK:       1,
					ParentPK:        2,
					PolicyExpiredAt: 3,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateGroupMembersWithTx(nil, int64(1), []types.SubjectRelation{
				{
					SubjectPK:       1,
					ParentPK:        2,
					PolicyExpiredAt: 3,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreateWithTx")
		})
	})

	Describe("BulkDeleteGroupMembers", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
				int64(0), errors.New("error"),
			)

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})
			defer patches.Reset()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.BulkDeleteGroupMembers(int64(1), []int64{2}, []int64{3})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteByMembersWithTx")
		})
	})

	Describe("ListThinSubjectGroupsBySubjectPKs", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListEffectRelationBySubjectPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().ListEffectRelationBySubjectPKs([]int64{1, 2}).Return(
				nil, errors.New("error"),
			)

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListEffectThinSubjectGroupsBySubjectPKs([]int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListEffectRelationBySubjectPKs")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().ListEffectRelationBySubjectPKs([]int64{1, 2}).Return(
				[]dao.EffectSubjectRelation{
					{
						ParentPK:        1,
						PolicyExpiredAt: 2,
					},
				}, nil,
			)

			manager := &groupService{
				manager: mockSubjectService,
			}

			subjectGroups, err := manager.ListEffectThinSubjectGroupsBySubjectPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ThinSubjectGroup{{GroupPK: 1, PolicyExpiredAt: 2}}, subjectGroups)
		})
	})
})
