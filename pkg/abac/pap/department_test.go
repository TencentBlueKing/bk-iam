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
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cacheimpls"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("DepartmentController", func() {
	Describe("ListPagingSubjectDepartment", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("service.ListPagingSubjectDepartment fail", func() {
			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().ListPaging(int64(1), int64(2)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &departmentController{
				service: mockDepartmentService,
			}

			_, err := manager.ListPaging(int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPaging")
		})

		It("ok", func() {
			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().ListPaging(int64(1), int64(2)).Return(
				[]types.SubjectDepartment{
					{
						SubjectPK:     int64(1),
						DepartmentPKs: []int64{2, 3},
					},
				}, nil,
			).AnyTimes()

			patches := gomonkey.ApplyFunc(cacheimpls.GetSubjectByPK, func(pk int64) (subject types.Subject, err error) {
				switch pk {
				case 1:
					return types.Subject{
						ID:   "1",
						Type: "user",
					}, nil
				case 2:
					return types.Subject{
						ID:   "2",
						Type: "department",
					}, nil
				case 3:
					return types.Subject{
						ID:   "3",
						Type: "department",
					}, nil
				}

				return types.Subject{}, nil
			})
			defer patches.Reset()

			manager := &departmentController{
				service: mockDepartmentService,
			}

			subjectDepartmets, err := manager.ListPaging(int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []SubjectDepartment{
				{
					SubjectID:     "1",
					DepartmentIDs: []string{"2", "3"},
				},
			}, subjectDepartmets)
		})
	})

	Describe("BulkCreate", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("service.BulkCreate fail", func() {
			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkCreate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			patches := gomonkey.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				switch id {
				case "1":
					return int64(1), nil
				case "2":
					return int64(2), nil
				case "3":
					return int64(3), nil
				}

				return 0, nil
			})
			defer patches.Reset()

			manager := &departmentController{
				service: mockDepartmentService,
			}

			err := manager.BulkCreate([]SubjectDepartment{{SubjectID: "1", DepartmentIDs: []string{"2", "3"}}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreate")
		})

		It("ok", func() {
			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkCreate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			}).Return(
				nil,
			).AnyTimes()

			patches := gomonkey.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				switch id {
				case "1":
					return int64(1), nil
				case "2":
					return int64(2), nil
				case "3":
					return int64(3), nil
				}

				return 0, nil
			})
			defer patches.Reset()

			manager := &departmentController{
				service: mockDepartmentService,
			}

			err := manager.BulkCreate([]SubjectDepartment{{SubjectID: "1", DepartmentIDs: []string{"2", "3"}}})
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

		It("service.BulkUpdateSubjectDepartments fail", func() {
			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkUpdate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			patches := gomonkey.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				switch id {
				case "1":
					return int64(1), nil
				case "2":
					return int64(2), nil
				case "3":
					return int64(3), nil
				}

				return 0, nil
			})
			defer patches.Reset()

			manager := &departmentController{
				service: mockDepartmentService,
			}

			err := manager.BulkUpdate([]SubjectDepartment{{SubjectID: "1", DepartmentIDs: []string{"2", "3"}}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkUpdate")
		})

		It("ok", func() {
			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkUpdate([]types.SubjectDepartment{
				{
					SubjectPK:     1,
					DepartmentPKs: []int64{2, 3},
				},
			}).Return(
				nil,
			).AnyTimes()

			patches := gomonkey.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				switch id {
				case "1":
					return int64(1), nil
				case "2":
					return int64(2), nil
				case "3":
					return int64(3), nil
				}

				return 0, nil
			})
			patches.ApplyFunc(cacheimpls.BatchDeleteSubjectDepartmentCache, func(pks []int64) error {
				return nil
			})
			defer patches.Reset()

			manager := &departmentController{
				service: mockDepartmentService,
			}

			err := manager.BulkUpdate([]SubjectDepartment{{SubjectID: "1", DepartmentIDs: []string{"2", "3"}}})
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

		It("subjectService.ListPKsBySubjects fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{{Type: "user", ID: "1"}}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &departmentController{
				subjectService: mockSubjectService,
			}

			err := manager.BulkDelete([]string{"1"})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPKsBySubjects")
		})

		It("service.BulkDeleteSubjectDepartments fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{{Type: "user", ID: "1"}}).Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkDelete([]int64{1}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &departmentController{
				service:        mockDepartmentService,
				subjectService: mockSubjectService,
			}

			err := manager.BulkDelete([]string{"1"})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDelete")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{{Type: "user", ID: "1"}}).Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockDepartmentService := mock.NewMockDepartmentService(ctl)
			mockDepartmentService.EXPECT().BulkDelete([]int64{1}).Return(
				nil,
			).AnyTimes()

			patches := gomonkey.ApplyFunc(cacheimpls.BatchDeleteSubjectDepartmentCache, func(pks []int64) error {
				return nil
			})
			defer patches.Reset()

			manager := &departmentController{
				service:        mockDepartmentService,
				subjectService: mockSubjectService,
			}

			err := manager.BulkDelete([]string{"1"})
			assert.NoError(GinkgoT(), err)
		})
	})
})
