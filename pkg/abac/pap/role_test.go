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

var _ = Describe("RoleController", func() {
	Describe("BulkCreateSubjectRoles", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("subjectService.ListPKsBySubjects fail", func() {
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

			manager := &roleController{
				subjectService: mockSubjectService,
			}

			err := manager.BulkAddSubjects("super", "test", []Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPKsBySubjects")
		})

		It("service.BulkCreateSubjectRoles fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockRoleService := mock.NewMockRoleService(ctl)
			mockRoleService.EXPECT().BulkAddSubjects("super", "test", []int64{1}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &roleController{
				service:        mockRoleService,
				subjectService: mockSubjectService,
			}

			err := manager.BulkAddSubjects("super", "test", []Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkAddSubjects")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockRoleService := mock.NewMockRoleService(ctl)
			mockRoleService.EXPECT().BulkAddSubjects("super", "test", []int64{1}).Return(
				nil,
			).AnyTimes()

			patches := gomonkey.ApplyFunc(
				cacheimpls.DeleteSubjectRoleSystemID,
				func(subjectType, subjectID string) error {
					return nil
				},
			)
			defer patches.Reset()

			manager := &roleController{
				service:        mockRoleService,
				subjectService: mockSubjectService,
			}

			err := manager.BulkAddSubjects("super", "test", []Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkDeleteSubjectRoles", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("subjectService.ListPKsBySubjects fail", func() {
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

			manager := &roleController{
				subjectService: mockSubjectService,
			}

			err := manager.BulkDeleteSubjects("super", "test", []Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListPKsBySubjects")
		})

		It("service.BulkDeleteSubjectRoles fail", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockRoleService := mock.NewMockRoleService(ctl)
			mockRoleService.EXPECT().BulkDeleteSubjects("super", "test", []int64{1}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &roleController{
				service:        mockRoleService,
				subjectService: mockSubjectService,
			}

			err := manager.BulkDeleteSubjects("super", "test", []Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDelete")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectService(ctl)
			mockSubjectService.EXPECT().ListPKsBySubjects([]types.Subject{
				{
					ID:   "1",
					Name: "name",
					Type: "user",
				},
			}).Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockRoleService := mock.NewMockRoleService(ctl)
			mockRoleService.EXPECT().BulkDeleteSubjects("super", "test", []int64{1}).Return(
				nil,
			).AnyTimes()

			patches := gomonkey.ApplyFunc(
				cacheimpls.DeleteSubjectRoleSystemID,
				func(subjectType, subjectID string) error {
					return nil
				},
			)
			defer patches.Reset()

			manager := &roleController{
				service:        mockRoleService,
				subjectService: mockSubjectService,
			}

			err := manager.BulkDeleteSubjects("super", "test", []Subject{{Type: "user", Name: "name", ID: "1"}})
			assert.NoError(GinkgoT(), err)
		})
	})
})
