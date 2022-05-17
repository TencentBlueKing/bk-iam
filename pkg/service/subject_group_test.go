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
	Describe("ListSubjectGroups", func() {
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

			_, err := manager.ListSubjectGroups("group", "test", int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetPK")
		})

		It("relationManager.ListRelation fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListRelation(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListSubjectGroups("group", "test", int64(0))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListSubjectGroups")
		})

		It("relationManager.ListRelationBeforeExpiredAt fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListRelationBeforeExpiredAt(int64(1), int64(10)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListSubjectGroups("group", "test", int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListSubjectGroups")
		})

		It("manager.ListByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListRelation(int64(1)).Return(
				[]dao.SubjectRelation{
					{
						PK:       int64(1),
						ParentPK: int64(2),
					},
				}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{2}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListSubjectGroups("group", "test", int64(0))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "convertToSubjectGroup")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().GetPK("group", "test").Return(
				int64(1), nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListRelation(int64(1)).Return(
				[]dao.SubjectRelation{
					{
						PK:       int64(1),
						ParentPK: int64(2),
					},
				}, nil,
			).AnyTimes()

			mockSubjectService.EXPECT().ListByPKs([]int64{2}).Return(
				[]dao.Subject{}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			groups, err := manager.ListSubjectGroups("group", "test", int64(0))
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groups, 1)
		})
	})

	Describe("convertToSubjectGroup", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("empty", func() {
			manager := &subjectService{}

			groups, err := manager.convertToSubjectGroup([]dao.SubjectRelation{})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groups, 0)
		})

		It("manager.ListByPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByPKs([]int64{2}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.convertToSubjectGroup([]dao.SubjectRelation{
				{
					PK:       int64(1),
					ParentPK: int64(2),
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "error")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByPKs([]int64{2}).Return(
				[]dao.Subject{}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			groups, err := manager.convertToSubjectGroup([]dao.SubjectRelation{
				{
					PK:       int64(1),
					ParentPK: int64(2),
				},
			})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groups, 1)
		})
	})

	Describe("ListExistSubjectsBeforeExpiredAt", func() {
		var ctl *gomock.Controller
		var subjects []types.Subject
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			subjects = []types.Subject{
				{
					Type: "group",
					ID:   "1",
				},
				{
					Type: "group",
					ID:   "2",
				},
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListByIDs", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByIDs("group", []string{"1", "2"}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager: mockSubjectService,
			}

			_, err := manager.ListExistSubjectsBeforeExpiredAt(subjects, int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByIDs")
		})

		It("relationManager.ListParentPKsBeforeExpiredAt fail", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByIDs("group", []string{"1", "2"}).Return(
				[]dao.Subject{
					{
						PK:   int64(1),
						Type: "group",
						ID:   "1",
						Name: "1",
					},
					{
						PK:   int64(2),
						Type: "group",
						ID:   "2",
						Name: "2",
					},
				}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListParentPKsBeforeExpiredAt([]int64{1, 2}, int64(10)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			_, err := manager.ListExistSubjectsBeforeExpiredAt(subjects, int64(10))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListParentPKsBeforeExpiredAt")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectManager(ctl)
			mockSubjectService.EXPECT().ListByIDs("group", []string{"1", "2"}).Return(
				[]dao.Subject{
					{
						PK:   int64(1),
						Type: "group",
						ID:   "1",
						Name: "1",
					},
					{
						PK:   int64(2),
						Type: "group",
						ID:   "2",
						Name: "2",
					},
				}, nil,
			).AnyTimes()

			mockSubjectRelationService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectRelationService.EXPECT().ListParentPKsBeforeExpiredAt([]int64{1, 2}, int64(10)).Return(
				[]int64{1, 2}, nil,
			).AnyTimes()

			manager := &subjectService{
				manager:         mockSubjectService,
				relationManager: mockSubjectRelationService,
			}

			groups, err := manager.ListExistSubjectsBeforeExpiredAt(subjects, int64(10))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.Subject{
				{
					Type: "group",
					ID:   "1",
					Name: "1",
				},
				{
					Type: "group",
					ID:   "2",
					Name: "2",
				},
			}, groups)
		})
	})
})
