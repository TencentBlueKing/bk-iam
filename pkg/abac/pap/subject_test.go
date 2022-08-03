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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("SubjectController", func() {
	Describe("BulkCreate", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("service.BulkCreate fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().BulkCreate([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &subjectController{
				service: mockSubjectService,
			}

			err := manager.BulkCreate([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreate")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().BulkCreate([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &subjectController{
				service: mockSubjectService,
			}

			err := manager.BulkCreate([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkUpdateName", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("service.BulkUpdateName fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().BulkUpdateName([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &subjectController{
				service: mockSubjectService,
			}

			err := manager.BulkUpdateName([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkUpdateName")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().BulkUpdateName([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &subjectController{
				service: mockSubjectService,
			}

			err := manager.BulkUpdateName([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkDelete", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("service.ListPKsBySubjects fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectController{
				service: mockSubjectService,
			}

			err := manager.BulkDelete([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPKsBySubjects")
		})

		It("policyService.BulkDeleteBySubjectPKsWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1, 2}, nil,
			).AnyTimes()

			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				errors.New("test"),
			).AnyTimes()

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})
			defer patches.Reset()

			manager := &subjectController{
				service:       mockSubjectService,
				policyService: mockPolicyService,
			}

			err := manager.BulkDelete([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "policyService.BulkDeleteBySubjectPKsWithTx")
		})

		It("groupService.BulkDeleteBySubjectPKsWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1, 2}, nil,
			).AnyTimes()

			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				errors.New("test"),
			).AnyTimes()
			mockGroupService.EXPECT().ListGroupMember(gomock.Any()).Return([]types.GroupMember{}, nil).AnyTimes()
			mockGroupService.EXPECT().ListGroupAuthSystemIDs(gomock.Any()).Return([]string{}, nil).AnyTimes()

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})
			defer patches.Reset()

			manager := &subjectController{
				service:       mockSubjectService,
				policyService: mockPolicyService,
				groupService:  mockGroupService,
			}

			err := manager.BulkDelete([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "groupService.BulkDeleteBySubjectPKsWithTx")
		})

		It("departmentService.BulkDeleteBySubjectPKsWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1, 2}, nil,
			).AnyTimes()

			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()
			mockGroupService.EXPECT().ListGroupMember(gomock.Any()).Return([]types.GroupMember{}, nil).AnyTimes()
			mockGroupService.EXPECT().ListGroupAuthSystemIDs(gomock.Any()).Return([]string{}, nil).AnyTimes()

			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				errors.New("test"),
			).AnyTimes()

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})
			defer patches.Reset()

			manager := &subjectController{
				service:           mockSubjectService,
				policyService:     mockPolicyService,
				groupService:      mockGroupService,
				departmentService: mockDepartmentService,
			}

			err := manager.BulkDelete([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "departmentService.BulkDeleteBySubjectPKsWithTx")
		})

		It("service.BulkDeleteByPKsWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1, 2}, nil,
			).AnyTimes()
			mockSubjectService.EXPECT().BulkDeleteByPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				errors.New("test"),
			).AnyTimes()

			mockPolicyService := mock.NewMockPolicyService(ctl)
			mockPolicyService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()
			mockGroupService.EXPECT().ListGroupMember(gomock.Any()).Return([]types.GroupMember{}, nil).AnyTimes()
			mockGroupService.EXPECT().ListGroupAuthSystemIDs(gomock.Any()).Return([]string{}, nil).AnyTimes()

			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkDeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})
			defer patches.Reset()

			manager := &subjectController{
				service:           mockSubjectService,
				policyService:     mockPolicyService,
				groupService:      mockGroupService,
				departmentService: mockDepartmentService,
			}

			err := manager.BulkDelete([]Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "service.BulkDeleteByPKsWithTx")
		})
	})
})
