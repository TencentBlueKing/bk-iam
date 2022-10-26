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

var _ = Describe("DepartmentService", func() {
	Describe("BulkDeleteBySubjectPKsWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.BulkDeleteWithTx fail", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkDeleteWithTx(gomock.Any(), []int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			// assert.Contains(GinkgoT(), err.Error(), "BulkDeleteWithTx")
		})

		It("success", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkDeleteWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkCreateSubjectDepartments", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("empty ok", func() {
			manager := &departmentService{}

			err := manager.BulkCreate([]types.SubjectDepartment{})
			assert.NoError(GinkgoT(), err)
		})

		It("manager.BulkCreate fail", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkCreate([]dao.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: "2,3",
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkCreate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreate")
		})

		It("ok", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkCreate([]dao.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: "2,3",
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkCreate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkDeleteSubjectDepartments", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.BulkCreate fail", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkDelete([]int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkDelete([]int64{1, 2})
			assert.Error(GinkgoT(), err)
			// assert.Contains(GinkgoT(), err.Error(), "BulkDelete")
		})

		It("ok", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkDelete([]int64{1, 2}).Return(
				nil,
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkDelete([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkUpdateSubjectDepartments", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("empty ok", func() {
			manager := &departmentService{}

			err := manager.BulkUpdate([]types.SubjectDepartment{})
			assert.NoError(GinkgoT(), err)
		})

		It("manager.BulkCreate fail", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkUpdate([]dao.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: "2,3",
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkUpdate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkUpdate")
		})

		It("ok", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().BulkUpdate([]dao.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: "2,3",
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			err := manager.BulkUpdate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("ListPagingSubjectDepartment", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListPaging fail", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().ListPaging(int64(1), int64(2)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			_, err := manager.ListPaging(int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPaging")
		})

		It("empty ok", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().ListPaging(int64(1), int64(2)).Return(
				nil, nil,
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			_, err := manager.ListPaging(int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
		})

		It("ok", func() {
			mockDepartmentService := mock.NewMockSubjectDepartmentManager(ctl)
			mockDepartmentService.EXPECT().ListPaging(int64(1), int64(2)).Return(
				[]dao.SubjectDepartment{
					{
						PK:            int64(1),
						SubjectPK:     int64(1),
						DepartmentPKs: "2,3",
					},
				}, nil,
			).AnyTimes()

			manager := &departmentService{
				manager: mockDepartmentService,
			}

			subjectDepartments, err := manager.ListPaging(int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
			assert.Equal(
				GinkgoT(),
				[]types.SubjectDepartment{{SubjectPK: int64(1), DepartmentPKs: []int64{2, 3}}},
				subjectDepartments,
			)
		})
	})
})
