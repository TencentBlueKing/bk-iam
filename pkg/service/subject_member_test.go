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
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

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

		It("manager.ListMember fail", func() {
			mockSubjectService := mock.NewMockSubjectRelationManager(ctl)
			mockSubjectService.EXPECT().ListMember("group", "test").Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				relationManager: mockSubjectService,
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

			manager := &subjectService{
				relationManager: mockSubjectService,
			}

			subjectMembers, err := manager.ListMember("group", "test")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.SubjectMember{}, subjectMembers)
		})
	})
})
