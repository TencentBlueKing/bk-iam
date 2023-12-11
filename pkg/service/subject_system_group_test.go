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
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/mock/gomock"
	jsoniter "github.com/json-iterator/go"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
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

			err := manager.createSubjectSystemGroup(nil, "system", int64(1), map[int64]int64{1: 1555555555})
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

			err := manager.createSubjectSystemGroup(nil, "system", int64(1), map[int64]int64{1: 1555555555})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("updateGroupsString", func() {
		updateFunc := func(groupExpiredAtMap map[int64]int64) (map[int64]int64, error) {
			groupExpiredAtMap[2] = 1555555555
			return groupExpiredAtMap, nil
		}

		It("json.UnmarshalFromString fail", func() {
			_, err := updateGroupsString("123", updateFunc)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ReadMapCB")
		})

		It("updateFunc fail", func() {
			newUpdateFunc := func(groupExpiredAtMap map[int64]int64) (map[int64]int64, error) {
				return nil, errors.New("update error")
			}
			_, err := updateGroupsString(`{"1": 2}`, newUpdateFunc)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "update error")
		})

		It("ok", func() {
			groups, err := updateGroupsString(`{"1": 2}`, updateFunc)
			assert.NoError(GinkgoT(), err)

			groupMap := map[int64]int64{}
			jsoniter.UnmarshalFromString(groups, &groupMap)
			assert.Equal(GinkgoT(), map[int64]int64{1: 2, 2: 1555555555}, groupMap)
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

			err := manager.addOrUpdateSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{1: 1555555555})
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

			err := manager.addOrUpdateSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{1: 1555555555})
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

			err := manager.addOrUpdateSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{1: 1555555555})
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

			err := manager.addOrUpdateSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{1: 1555555555})
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

			err := manager.addOrUpdateSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{1: 1555555555})
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

			err := manager.addOrUpdateSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{1: 1555555555})
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

			err := manager.removeSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{2: 0})
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

			err := manager.removeSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{2: 0})
			assert.NoError(GinkgoT(), err)
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

			err := manager.removeSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{2: 0})
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

			err := manager.removeSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{2: 0})
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

			err := manager.removeSubjectSystemGroup(nil, int64(1), "system", map[int64]int64{2: 0})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("convertSystemSubjectGroupsToThinSubjectGroup", func() {
		It("empty ok", func() {
			groups, err := convertSystemSubjectGroupsToThinSubjectGroup("")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), 0, len(groups))
		})

		It("UnmarshalFromString fail", func() {
			_, err := convertSystemSubjectGroupsToThinSubjectGroup("abc")
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			groups, err := convertSystemSubjectGroupsToThinSubjectGroup(`{"1": 1555555555}`)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ThinSubjectGroup{{GroupPK: 1, ExpiredAt: 1555555555}}, groups)
		})
	})

	Describe("ListEffectThinSubjectGroups", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("subjectSystemGroupManager.ListEffectSubjectGroups fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().ListSubjectGroups("system", []int64{1}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			_, err := manager.ListEffectThinSubjectGroups("system", []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListSubjectGroups")
		})

		It("UnmarshalFromString fail", func() {
			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().ListSubjectGroups("system", []int64{1}).Return(
				[]dao.SubjectGroups{{SubjectPK: int64(1), Groups: `abc`}}, nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			_, err := manager.ListEffectThinSubjectGroups("system", []int64{1})
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			ts := time.Now().Unix() + 10

			mockSubjectSystemGroupManager := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupManager.EXPECT().ListSubjectGroups("system", []int64{1}).Return(
				[]dao.SubjectGroups{{SubjectPK: int64(1), Groups: fmt.Sprintf(`{"2": %d}`, ts)}}, nil,
			).AnyTimes()

			manager := &groupService{
				subjectSystemGroupManager: mockSubjectSystemGroupManager,
			}

			groups, err := manager.ListEffectThinSubjectGroups("system", []int64{1})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), map[int64][]types.ThinSubjectGroup{1: {{
				GroupPK:   2,
				ExpiredAt: ts,
			}}}, groups)
		})
	})
})
