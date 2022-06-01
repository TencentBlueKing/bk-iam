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
			mockSubjectService.EXPECT().ListMember("group", "test").Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListMember("group", "test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListMember")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().ListMember("group", "test").Return(
				[]dao.SubjectRelation{}, nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			subjectMembers, err := manager.ListMember("group", "test")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.SubjectMember{}, subjectMembers)
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

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByParentPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectService.EXPECT().BulkDeleteBySubjectPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
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

			err := manager.UpdateMembersExpiredAtWithTx(nil, []types.SubjectRelationPKPolicyExpiredAt{
				{
					PK:              1,
					PolicyExpiredAt: 2,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateExpiredAtWithTx")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().UpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelationPKPolicyExpiredAt{
				{
					PK:              1,
					PolicyExpiredAt: 2,
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.UpdateMembersExpiredAtWithTx(nil, []types.SubjectRelationPKPolicyExpiredAt{
				{
					PK:              1,
					PolicyExpiredAt: 2,
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkCreateSubjectMembersWithTx", func() {
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
					SubjectType:     "subject_type",
					SubjectID:       "1",
					ParentPK:        2,
					ParentType:      "parent_type",
					ParentID:        "2",
					PolicyExpiredAt: 3,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateSubjectMembersWithTx(nil, []types.SubjectRelation{
				{
					SubjectPK:       1,
					SubjectType:     "subject_type",
					SubjectID:       "1",
					ParentPK:        2,
					ParentType:      "parent_type",
					ParentID:        "2",
					PolicyExpiredAt: 3,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreateWithTx")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK:       1,
					SubjectType:     "subject_type",
					SubjectID:       "1",
					ParentPK:        2,
					ParentType:      "parent_type",
					ParentID:        "2",
					PolicyExpiredAt: 3,
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateSubjectMembersWithTx(nil, []types.SubjectRelation{
				{
					SubjectPK:       1,
					SubjectType:     "subject_type",
					SubjectID:       "1",
					ParentPK:        2,
					ParentType:      "parent_type",
					ParentID:        "2",
					PolicyExpiredAt: 3,
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})
})
