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
	"iam/pkg/service/types"
)

var _ = Describe("SubjectService", func() {

	Describe("convertToGroupExpiredAt", func() {
		It("empty", func() {
			groupExpiredAts, err := convertToGroupExpiredAt("")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), 0, len(groupExpiredAts))
		})

		It("fail", func() {
			_, err := convertToGroupExpiredAt("abc")
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			groupExpiredAts, err := convertToGroupExpiredAt(`[
				{"group_pk":1,"expired_at":1555555555},
				{"group_pk":2,"expired_at":1555555555}
			]`)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), 2, len(groupExpiredAts))
		})
	})

	Describe("convertToGroupsString", func() {
		It("empty", func() {
			groups, err := convertToGroupsString([]types.GroupExpiredAt{})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "[]", groups)
		})

		It("ok", func() {
			groups, err := convertToGroupsString([]types.GroupExpiredAt{
				{GroupPK: 1, ExpiredAt: 1555555555},
			})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), `[{"group_pk":1,"expired_at":1555555555}]`, groups)
		})
	})

	Describe("findGroupIndex", func() {
		groups := []types.GroupExpiredAt{
			{GroupPK: 1, ExpiredAt: 1555555555},
			{GroupPK: 2, ExpiredAt: 1555555555},
		}

		It("no in", func() {
			i := findGroupIndex(groups, 3)
			assert.Equal(GinkgoT(), -1, i)
		})

		It("in", func() {
			i := findGroupIndex(groups, 2)
			assert.Equal(GinkgoT(), 1, i)
		})
	})

	Describe("isMysqlDuplicateError", func() {
		It("true", func() {
			assert.True(GinkgoT(), isMysqlDuplicateError(&mysql.MySQLError{
				Number: 1062,
			}))
		})

		It("false", func() {
			assert.False(GinkgoT(), isMysqlDuplicateError(errors.New("error")))
		})
	})

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

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.createSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
			assert.Error(GinkgoT(), err)
		})

		It("groupSystemAuthTypeManager.CreateWithTx ok", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				nil,
			).AnyTimes()

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.createSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("updateSubjectSystemGroup", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("subjectSystemGroupManager.UpdateWithTx fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			rows, err := manager.updateSubjectSystemGroup(nil, "system", int64(1), []types.GroupExpiredAt{
				{
					GroupPK:   1,
					ExpiredAt: 1555555555,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(0), rows)
		})

		It("groupSystemAuthTypeManager.UpdateWithTx ok", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(1), nil,
			).AnyTimes()

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			rows, err := manager.updateSubjectSystemGroup(nil, "system", int64(1), []types.GroupExpiredAt{
				{
					GroupPK:   1,
					ExpiredAt: 1555555555,
				},
			})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), rows)
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

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
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

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "createSubjectSystemGroup")
		})

		It("updateSubjectSystemGroup fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "updateSubjectSystemGroup")
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

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "updateSubjectSystemGroup")
		})

		It("updateSubjectSystemGroup retry fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), nil,
			).AnyTimes()

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
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

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.addOrUpdateSubjectSystemGroup(nil, "system", int64(1), types.GroupExpiredAt{
				GroupPK:   1,
				ExpiredAt: 1555555555,
			})
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
					Groups: "[{\"group_pk\":2,\"expired_at\":1555555555}]",
				}, errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
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
					Groups: "[{\"group_pk\":1,\"expired_at\":1555555555}]",
				}, nil,
			).AnyTimes()

			manager := &subjectService{
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
					Groups: "[{\"group_pk\":2,\"expired_at\":1555555555}]",
				}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.removeSubjectSystemGroup(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "updateSubjectSystemGroup")
		})

		It("updateSubjectSystemGroup retry fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().GetBySystemSubject("system", int64(1)).Return(
				dao.SubjectSystemGroup{
					Groups: "[{\"group_pk\":2,\"expired_at\":1555555555}]",
				}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), nil,
			).AnyTimes()

			manager := &subjectService{
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
					Groups: "[{\"group_pk\":2,\"expired_at\":1555555555}]",
				}, nil,
			).AnyTimes()

			mockSubjectSystemGroupManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(1), nil,
			).AnyTimes()

			manager := &subjectService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			err := manager.removeSubjectSystemGroup(nil, "system", int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
		})
	})
})
