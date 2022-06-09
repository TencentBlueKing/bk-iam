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
)

var _ = Describe("RoleService", func() {
	Describe("ListRoleSystemIDBySubjectPK", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListSystemIDBySubjectPK fail", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().ListSystemIDBySubjectPK(int64(1)).Return(
				nil, errors.New("get pk fail"),
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			_, err := manager.ListSystemIDBySubjectPK(int64(1))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListSystemIDBySubjectPK")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().ListSystemIDBySubjectPK(int64(1)).Return(
				[]string{"test"}, nil,
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			systemIDs, err := manager.ListSystemIDBySubjectPK(int64(1))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []string{"test"}, systemIDs)
		})
	})

	Describe("ListSubjectPKByRole", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListSubjectPKByRole fail", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().ListSubjectPKByRole("super", "test").Return(
				nil, errors.New("get pk fail"),
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			_, err := manager.ListSubjectPKByRole("super", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListSubjectPKByRole")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().ListSubjectPKByRole("super", "test").Return(
				[]int64{1}, nil,
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			subjectPKs, err := manager.ListSubjectPKByRole("super", "test")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []int64{1}, subjectPKs)
		})
	})

	Describe("BulkCreateSubjectRoles", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListSubjectPKByRole fail", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().ListSubjectPKByRole("super", "test").Return(
				nil, errors.New("get pk fail"),
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			err := manager.BulkAddSubjects("super", "test", []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListSubjectPKByRole")
		})

		It("manager.BulkCreate fail", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().ListSubjectPKByRole("super", "test").Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().BulkCreate([]dao.SubjectRole{{
				RoleType:  "super",
				System:    "test",
				SubjectPK: 2,
			}, {
				RoleType:  "super",
				System:    "test",
				SubjectPK: 3,
			}}).Return(
				errors.New("test"),
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			err := manager.BulkAddSubjects("super", "test", []int64{1, 2, 3})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreate")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().ListSubjectPKByRole("super", "test").Return(
				[]int64{1}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().BulkCreate([]dao.SubjectRole{{
				RoleType:  "super",
				System:    "test",
				SubjectPK: 2,
			}, {
				RoleType:  "super",
				System:    "test",
				SubjectPK: 3,
			}}).Return(
				nil,
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			err := manager.BulkAddSubjects("super", "test", []int64{1, 2, 3})
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

		It("empty pk", func() {
			manager := &roleService{}

			err := manager.BulkDeleteSubjects("super", "test", []int64{})
			assert.NoError(GinkgoT(), err)
		})

		It("manager.BulkDelete fail", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().BulkDelete("super", "test", []int64{1}).Return(
				errors.New("test"),
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			err := manager.BulkDeleteSubjects("super", "test", []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDelete")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRoleManager(ctl)
			mockSubjectService.EXPECT().BulkDelete("super", "test", []int64{1}).Return(
				nil,
			).AnyTimes()

			manager := &roleService{
				manager: mockSubjectService,
			}

			err := manager.BulkDeleteSubjects("super", "test", []int64{1})
			assert.NoError(GinkgoT(), err)
		})
	})
})
