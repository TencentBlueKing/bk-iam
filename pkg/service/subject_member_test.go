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
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("SubjectService", func() {

	Describe("ListMember", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.ListMember("group", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("relationManager.ListMember fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListMember(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListMember("group", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListMember")
		})

		It("manager.ListByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{0}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListMember(int64(1)).Return(
				[]dao.SubjectRelation{{}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListMember("group", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "convertToSubjectMembers")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{0}).Return(
				[]dao.Subject{}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListMember(int64(1)).Return(
				[]dao.SubjectRelation{{}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			subjectMembers, err := manager.ListMember("group", "test")
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectMembers, 1)
		})
	})

	Describe("GetMemberCount", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.GetMemberCount("group", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("relationManager.GetMemberCount fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().GetMemberCount(int64(1)).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.GetMemberCount("group", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetMemberCount")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().GetMemberCount(int64(1)).Return(
				int64(1), nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			count, err := manager.GetMemberCount("group", "test")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), count)
		})
	})

	Describe("ListPagingMember", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.ListPagingMember("group", "test", int64(0), int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("relationManager.ListPagingMember fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListPagingMember(int64(1), int64(0), int64(10)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListPagingMember("group", "test", int64(0), int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPagingMember")
		})

		It("manager.ListByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{0}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListPagingMember(int64(1), int64(0), int64(10)).Return(
				[]dao.SubjectRelation{{}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListPagingMember("group", "test", int64(0), int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "convertToSubjectMembers")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{0}).Return(
				[]dao.Subject{}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListPagingMember(int64(1), int64(0), int64(10)).Return(
				[]dao.SubjectRelation{{}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			subjectMembers, err := manager.ListPagingMember("group", "test", int64(0), int64(10))
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectMembers, 1)
		})
	})

	Describe("getSubjectMapByPKs", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("empty", func() {
			manager := &subjectService{}

			subjects, err := manager.getSubjectMapByPKs([]int64{})
			assert.NoError(GinkgoT(), err)
			assert.Nil(GinkgoT(), subjects)
		})

		It("manager.ListByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByPKs([]int64{1}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			subjects, err := manager.getSubjectMapByPKs([]int64{1})
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), subjects)
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByPKs([]int64{1}).Return(
				[]dao.Subject{{PK: 1, Type: "test", ID: "id", Name: "name"}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			subjectMap, err := manager.getSubjectMapByPKs([]int64{1})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectMap, 1)
			assert.Equal(GinkgoT(), map[int64]dao.Subject{
				1: {PK: 1, Type: "test", ID: "id", Name: "name"},
			}, subjectMap)
		})
	})

	Describe("convertToSubjectMembers", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("empty", func() {
			manager := &subjectService{}

			subjects, err := manager.convertToSubjectMembers([]dao.SubjectRelation{})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjects, 0)
		})

		It("manager.ListByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByPKs([]int64{0}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			subjects, err := manager.convertToSubjectMembers([]dao.SubjectRelation{{}})
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), subjects)
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByPKs([]int64{1}).Return(
				[]dao.Subject{{PK: 1, Type: "test", ID: "id", Name: "name"}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			subjects, err := manager.convertToSubjectMembers([]dao.SubjectRelation{{SubjectPK: 1}})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjects, 1)
			assert.Equal(GinkgoT(), "test", subjects[0].Type)
		})
	})

	Describe("GetMemberCountBeforeExpiredAt", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.GetMemberCountBeforeExpiredAt("group", "test", int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("relationManager.GetMemberCountBeforeExpiredAt fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().GetMemberCountBeforeExpiredAt(int64(1), int64(10)).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.GetMemberCountBeforeExpiredAt("group", "test", int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetMemberCountBeforeExpiredAt")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().GetMemberCountBeforeExpiredAt(int64(1), int64(10)).Return(
				int64(5), nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			count, err := manager.GetMemberCountBeforeExpiredAt("group", "test", int64(10))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(5), count)
		})
	})

	Describe("ListPagingMemberBeforeExpiredAt", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.ListPagingMemberBeforeExpiredAt("group", "test", int64(10), int64(0), int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("relationManager.ListMember fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListPagingMemberBeforeExpiredAt(int64(1), int64(10), int64(0), int64(10)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListPagingMemberBeforeExpiredAt("group", "test", int64(10), int64(0), int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPagingMemberBeforeExpiredAt")
		})

		It("manager.ListByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{0}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListPagingMemberBeforeExpiredAt(int64(1), int64(10), int64(0), int64(10)).Return(
				[]dao.SubjectRelation{{}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListPagingMemberBeforeExpiredAt("group", "test", int64(10), int64(0), int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "convertToSubjectMembers")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{0}).Return(
				[]dao.Subject{}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListPagingMemberBeforeExpiredAt(int64(1), int64(10), int64(0), int64(10)).Return(
				[]dao.SubjectRelation{{}}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			subjectMembers, err := manager.ListPagingMemberBeforeExpiredAt("group", "test", int64(10), int64(0), int64(10))
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectMembers, 1)
		})
	})

	Describe("UpdateMembersExpiredAt", func() {
		var ctl *gomock.Controller
		var members []types.SubjectMember
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			members = []types.SubjectMember{
				{
					PK: int64(1),
				},
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("relationManager.UpdateExpiredAt fail", func() {
			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().UpdateExpiredAt(gomock.Any()).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				relationManager: mockSubjectRelationService,
			}

			err := manager.UpdateMembersExpiredAt(members)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateExpiredAt")
		})

		It("success", func() {
			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().UpdateExpiredAt(gomock.Any()).Return(
				nil,
			).AnyTimes()

			manager := &subjectService{
				relationManager: mockSubjectRelationService,
			}

			err := manager.UpdateMembersExpiredAt(members)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkDeleteSubjectMembers", func() {
		var ctl *gomock.Controller
		var members []types.Subject
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			members = []types.Subject{
				{
					Type: types.UserType,
					ID:   "user",
				},
				{
					Type: types.DepartmentType,
					ID:   "department",
				},
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.BulkDeleteSubjectMembers("group", "test", members)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("manager.ListByIDs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			_, err := manager.BulkDeleteSubjectMembers("group", "test", members)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("relationManager.BulkDeleteByMembersWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().BulkDeleteByMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			_, err := manager.BulkDeleteSubjectMembers("group", "test", members)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteByMembersWithTx")
		})

		It("manager.ListByIDs fail 2", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().BulkDeleteByMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("department", []string{"department"}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			_, err := manager.BulkDeleteSubjectMembers("group", "test", members)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("relationManager.BulkDeleteByMembersWithTx fail 2", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().BulkDeleteByMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("department", []string{"department"}).Return(
				[]dao.Subject{{
					PK:   int64(3),
					Type: types.UserType,
					ID:   "department",
					Name: "department",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService.EXPECT().BulkDeleteByMembersWithTx(gomock.Any(), int64(1), []int64{3}).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			_, err := manager.BulkDeleteSubjectMembers("group", "test", members)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteByMembersWithTx")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().BulkDeleteByMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("department", []string{"department"}).Return(
				[]dao.Subject{{
					PK:   int64(3),
					Type: types.UserType,
					ID:   "department",
					Name: "department",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService.EXPECT().BulkDeleteByMembersWithTx(gomock.Any(), int64(1), []int64{3}).Return(
				int64(1), nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			typeCount, err := manager.BulkDeleteSubjectMembers("group", "test", members)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), map[string]int64{
				"user":       1,
				"department": 1,
			}, typeCount)
		})
	})

	Describe("BulkCreateSubjectMembers", func() {
		var ctl *gomock.Controller
		var members []types.Subject
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			members = []types.Subject{
				{
					Type: types.UserType,
					ID:   "user",
				},
				{
					Type: types.DepartmentType,
					ID:   "department",
				},
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateSubjectMembers("group", "test", members, int64(10))
			assert.Error(GinkgoT(), err)
		})

		It("manager.ListByIDs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateSubjectMembers("group", "test", members, int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("manager.ListByIDs fail 2", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("department", []string{"department"}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateSubjectMembers("group", "test", members, int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("member don't exists pk fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("department", []string{"department"}).Return(
				[]dao.Subject{}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateSubjectMembers("group", "test", members, int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "member don't exists pk")
		})

		It("relationManager.BulkCreate fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("department", []string{"department"}).Return(
				[]dao.Subject{{
					PK:   int64(3),
					Type: types.DepartmentType,
					ID:   "department",
					Name: "department",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().BulkCreate(gomock.Any()).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			err := manager.BulkCreateSubjectMembers("group", "test", members, int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreate")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("user", []string{"user"}).Return(
				[]dao.Subject{{
					PK:   int64(2),
					Type: types.UserType,
					ID:   "user",
					Name: "user",
				}}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByIDs("department", []string{"department"}).Return(
				[]dao.Subject{{
					PK:   int64(3),
					Type: types.DepartmentType,
					ID:   "department",
					Name: "department",
				}}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().BulkCreate(gomock.Any()).Return(
				nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			err := manager.BulkCreateSubjectMembers("group", "test", members, int64(10))
			assert.NoError(GinkgoT(), err)
		})
	})
})
