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
	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("SubjectService", func() {

	Describe("GetPK", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetPK fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(0), errors.New("get pk fail"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.GetPK("user", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("user", "test").Return(
				int64(1), nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			pk, err := manager.GetPK("user", "test")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), pk)
		})
	})

	Describe("ListPKsBySubjects", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("ListByIDs user fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByIDs("user", []string{"test"}).Return(
				[]dao.Subject{}, errors.New("list pk fail"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.ListPKsBySubjects([]types.Subject{{
				Type: "user",
				ID:   "test",
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("ListByIDs department fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByIDs("department", []string{"test"}).Return(
				[]dao.Subject{}, errors.New("list pk fail"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.ListPKsBySubjects([]types.Subject{{
				Type: "department",
				ID:   "test",
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("ListByIDs group fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByIDs("group", []string{"test"}).Return(
				[]dao.Subject{}, errors.New("list pk fail"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.ListPKsBySubjects([]types.Subject{{
				Type: "group",
				ID:   "test",
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByIDs(gomock.Any(), []string{"test"}).Return(
				[]dao.Subject{
					{PK: 1},
				}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			pks, err := manager.ListPKsBySubjects([]types.Subject{{
				Type: "group",
				ID:   "test",
			}, {
				Type: "user",
				ID:   "test",
			}, {
				Type: "department",
				ID:   "test",
			}})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), pks, []int64{1, 1, 1})
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

		It("manager.BulkCreate fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().BulkCreate([]dao.Subject{
				{
					Type: "user",
					ID:   "admin",
					Name: "admin",
				},
			}).Return(
				errors.New("create fail"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreate([]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "create")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().BulkCreate([]dao.Subject{
				{
					Type: "user",
					ID:   "admin",
					Name: "admin",
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreate([]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}})
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

		It("manager.BulkCreate fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().BulkUpdate([]dao.Subject{
				{
					Type: "user",
					ID:   "admin",
					Name: "admin",
				},
			}).Return(
				errors.New("update fail"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkUpdateName([]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "update")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().BulkUpdate([]dao.Subject{
				{
					Type: "user",
					ID:   "admin",
					Name: "admin",
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			err := manager.BulkUpdateName([]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}})
			assert.NoError(GinkgoT(), err)
		})
	})
})
