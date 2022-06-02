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
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
)

var _ = Describe("SubjectService", func() {

	Describe("createSubjectSystemGroup", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("subjectSystemGroupManager.CreateWithTx fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.createSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.Error(GinkgoT(), err)
		})

		It("groupSystemAuthTypeManager.CreateWithTx ok", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.createSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("updateGroupsString", func() {
		var updateFunc = func(groupExpiredAtMap map[int64]int64) (map[int64]int64, error) {
			groupExpiredAtMap[2] = 1555555555
			return groupExpiredAtMap, nil
		}

		It("json.UnmarshalFromString fail", func() {
			_, err := updateGroupsString("123", updateFunc)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ReadMapCB")
		})

		It("updateFunc fail", func() {
			var newUpdateFunc = func(groupExpiredAtMap map[int64]int64) (map[int64]int64, error) {
				return nil, errors.New("update error")
			}
			_, err := updateGroupsString(`{"1": 2}`, newUpdateFunc)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "update error")
		})

		It("ok", func() {
			groups, err := updateGroupsString(`{"1": 2}`, updateFunc)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), `{"1":2,"2":1555555555}`, groups)
		})
	})

	Describe("addOrUpdateSubjectSystemGroup", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("subjectSystemGroupManager.GetBySystemSubject fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySystemSubject")
		})

		It("createSubjectSystemGroup fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, sql.ErrNoRows,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySystemSubject")
		})

		It("subjectSystemGroupManager.UpdateWithTx fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateWithTx")
		})

		It("createSubjectSystemGroup duplicate", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, sql.ErrNoRows,
			)

			mockSubjectSystemGroupManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				&mysql.MySQLError{
					Number: 1062,
				},
			)

			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, nil,
			)

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), errors.New("error"),
			)

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateWithTx")
		})

		It("updateSubjectSystemGroup retry fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "retry")
		})

		It("ok", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(1), nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), 1, 1555555555)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("removeSubjectSystemGroup", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("subjectSystemGroupManager.GetBySystemSubject fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{
					Groups: `{"2": 1555555555}`,
				}, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.removeSubjectSystemGroup(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySystemSubject")
		})

		It("no subject group fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{
					Groups: `{"1": 1555555555}`,
				}, nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.removeSubjectSystemGroup(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "no subject system group")
		})

		It("updateSubjectSystemGroup fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{
					Groups: `{"2": 1555555555}`,
				}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.removeSubjectSystemGroup(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateWithTx")
		})

		It("updateSubjectSystemGroup retry fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{
					Groups: `{"2": 1555555555}`,
				}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.removeSubjectSystemGroup(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "retry")
		})

		It("ok", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{
					Groups: `{"2": 1555555555}`,
				}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(1), nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.removeSubjectSystemGroup(nil, "system", int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
		})
	})

})
