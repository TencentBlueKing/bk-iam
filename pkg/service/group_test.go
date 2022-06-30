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
	Describe("ListGroupMember", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListGroupMember fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListGroupMember(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListGroupMember(1)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListGroupMember")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListGroupMember(int64(1)).Return(
				[]dao.SubjectRelation{}, nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			groupMembers, err := manager.ListGroupMember(1)
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

		It("manager.BulkDeleteByGroupPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByGroupPKs(gomock.Any(), []int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteByGroupPKs")
		})

		It("manager.BulkDeleteBySubjectPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByGroupPKs(gomock.Any(), []int64{1, 2}).Return(
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
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByGroupPKs(gomock.Any(), []int64{1, 2}).Return(
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
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByGroupPKs(gomock.Any(), []int64{1, 2}).Return(
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

	Describe("UpdateGroupMembersExpiredAtWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().UpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelationForUpdateExpiredAt{
				{
					PK:        1,
					ExpiredAt: 2,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.UpdateGroupMembersExpiredAtWithTx(nil, int64(1), []types.SubjectRelationForUpdate{
				{
					PK:        1,
					ExpiredAt: 2,
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
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   2,
					ExpiredAt: 3,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateGroupMembersWithTx(nil, int64(1), []types.SubjectRelationForCreate{
				{
					SubjectPK: 1,
					GroupPK:   2,
					ExpiredAt: 3,
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
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByGroupMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
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
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteByGroupMembersWithTx")
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

		It("manager.ListThinRelationAfterExpiredAtBySubjectPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListThinRelationAfterExpiredAtBySubjectPKs([]int64{1, 2}, gomock.Any()).Return(
				nil, errors.New("error"),
			)

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListEffectThinSubjectGroupsBySubjectPKs([]int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListThinRelationAfterExpiredAtBySubjectPKs")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListThinRelationAfterExpiredAtBySubjectPKs([]int64{1, 2}, gomock.Any()).Return(
				[]dao.ThinSubjectRelation{
					{
						GroupPK:   1,
						ExpiredAt: 2,
					},
				}, nil,
			)

			manager := &groupService{
				manager: mockSubjectService,
			}

			subjectGroups, err := manager.ListEffectThinSubjectGroupsBySubjectPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ThinSubjectGroup{{GroupPK: 1, ExpiredAt: 2}}, subjectGroups)
		})
	})

	Describe("FilterExistEffectSubjectGroupPKs", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.FilterExistEffectSubjectGroupPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				FilterSubjectPKsExistGroupPKsAfterExpiredAt([]int64{123}, []int64{1}, gomock.Any()).
				Return(
					nil, errors.New("error"),
				).
				AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.FilterExistEffectSubjectGroupPKs([]int64{123}, []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "FilterSubjectPKsExistGroupPKsAfterExpiredAt")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				FilterSubjectPKsExistGroupPKsAfterExpiredAt([]int64{123}, []int64{1}, gomock.Any()).
				Return(
					[]int64{1, 2, 3}, nil,
				).
				AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			groupPKs, err := manager.FilterExistEffectSubjectGroupPKs([]int64{123}, []int64{1})
			assert.NoError(GinkgoT(), err)
			assert.ElementsMatch(GinkgoT(), []int64{1, 2, 3}, groupPKs)
		})
	})
})
