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

var _ = Describe("GroupService", func() {

	Describe("createOrUpdateGroupAuthType", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("groupSystemAuthTypeManager.GetBySystemGroup fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{}, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySystemGroup")
			assert.False(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(0), rows)
		})

		It("groupSystemAuthTypeManager.CreateWithTx ok", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{}, sql.ErrNoRows,
			).AnyTimes()
			mockGroupSystemAuthTypeManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(1), rows)
		})

		It("groupSystemAuthTypeManager.CreateWithTx fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{}, sql.ErrNoRows,
			).AnyTimes()
			mockGroupSystemAuthTypeManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				&mysql.MySQLError{
					Number: 1062,
				},
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "concurrency conflict")
			assert.False(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(0), rows)
		})

		It("groupSystemAuthTypeManager.CreateWithTx fail 2", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{}, sql.ErrNoRows,
			).AnyTimes()
			mockGroupSystemAuthTypeManager.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySystemGroup")
			assert.False(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(0), rows)
		})

		It("not need update", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{
					AuthType: int64(2),
				}, nil,
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(0), rows)
		})

		It("groupSystemAuthTypeManager.UpdateWithTx fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{
					AuthType: int64(1),
				}, nil,
			).AnyTimes()

			mockGroupSystemAuthTypeManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateWithTx")
			assert.False(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(0), rows)
		})

		It("groupSystemAuthTypeManager.UpdateWithTx duplicate", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{
					AuthType: int64(1),
				}, nil,
			).AnyTimes()

			mockGroupSystemAuthTypeManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(0), nil,
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "concurrency conflict")
			assert.False(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(0), rows)
		})

		It("groupSystemAuthTypeManager.UpdateWithTx ok", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().GetBySystemGroup("system", int64(1)).Return(
				dao.GroupSystemAuthType{
					AuthType: int64(1),
				}, nil,
			).AnyTimes()

			mockGroupSystemAuthTypeManager.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).Return(
				int64(1), nil,
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			created, rows, err := manager.createOrUpdateGroupAuthType(nil, "system", int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), created)
			assert.Equal(GinkgoT(), int64(1), rows)
		})
	})

	Describe("listGroupAuthSystem", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("groupSystemAuthTypeManager.ListByGroup fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().ListByGroup(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			_, err := manager.listGroupAuthSystem(int64(1))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByGroup")
		})

		It("ok", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().ListByGroup(int64(1)).Return(
				[]dao.GroupSystemAuthType{
					{SystemID: "1"},
					{SystemID: "2"},
				}, nil,
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			systems, err := manager.listGroupAuthSystem(int64(1))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []string{"1", "2"}, systems)
		})
	})
})