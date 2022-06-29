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

			_, err := manager.ListGroupAuthSystemIDs(int64(1))
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

			systems, err := manager.ListGroupAuthSystemIDs(int64(1))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []string{"1", "2"}, systems)
		})
	})

	Describe("ListGroupAuthBySystemGroupPKs", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("authTypeManger.ListAuthTypeBySystemGroups fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().ListAuthTypeBySystemGroups("test", []int64{1, 2}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			_, err := manager.ListGroupAuthBySystemGroupPKs("test", []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListAuthTypeBySystemGroups")
		})

		It("ok", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().ListAuthTypeBySystemGroups("test", []int64{1, 2}).Return(
				[]dao.GroupAuthType{{
					AuthType: int64(1),
					GroupPK:  int64(1),
				}}, nil,
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			groupAuthTypes, err := manager.ListGroupAuthBySystemGroupPKs("test", []int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{
					AuthType: int64(1),
					GroupPK:  int64(1),
				},
			}, groupAuthTypes)
		})

		It("loop ok", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().ListAuthTypeBySystemGroups("test", gomock.Any()).Return(
				[]dao.GroupAuthType{{
					AuthType: int64(1),
					GroupPK:  int64(1),
				}}, nil,
			)
			mockGroupSystemAuthTypeManager.EXPECT().ListAuthTypeBySystemGroups("test", []int64{1}).Return(
				[]dao.GroupAuthType{{
					AuthType: int64(1),
					GroupPK:  int64(2),
				}}, nil,
			)

			groupPKs := make([]int64, 0, 1001)
			for i := 0; i < 1000; i++ {
				groupPKs = append(groupPKs, 0)
			}
			groupPKs = append(groupPKs, 1)

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			groupAuthTypes, err := manager.ListGroupAuthBySystemGroupPKs("test", groupPKs)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{
					AuthType: int64(1),
					GroupPK:  int64(1),
				},
				{
					AuthType: int64(1),
					GroupPK:  int64(2),
				},
			}, groupAuthTypes)
		})
	})

	Describe("AlterGroupAuthType", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("authTypeManger.DeleteBySystemGroupWithTx fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().DeleteBySystemGroupWithTx(gomock.Any(), "test", int64(1)).Return(
				int64(0), errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			_, err := manager.AlterGroupAuthType(nil, "test", 1, 0)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "DeleteBySystemGroupWithTx")
		})

		It("no need create subject system group, ok", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().DeleteBySystemGroupWithTx(gomock.Any(), "test", int64(1)).Return(
				int64(0), nil,
			).AnyTimes()

			manager := &groupService{
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			changed, err := manager.AlterGroupAuthType(nil, "test", 1, 0)
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), changed)
		})

		It("manager.ListMember fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().DeleteBySystemGroupWithTx(gomock.Any(), "test", int64(1)).Return(
				int64(1), nil,
			).AnyTimes()
			mockSubjectRelationManger := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectRelationManger.EXPECT().ListMember(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager:        mockSubjectRelationManger,
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			_, err := manager.AlterGroupAuthType(nil, "test", 1, 0)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListMember")
		})

		It("removeSubjectSystemGroup fail", func() {
			mockGroupSystemAuthTypeManager := mock.NewMockGroupSystemAuthTypeManager(ctl)
			mockGroupSystemAuthTypeManager.EXPECT().DeleteBySystemGroupWithTx(gomock.Any(), "test", int64(1)).Return(
				int64(1), nil,
			).AnyTimes()
			mockSubjectRelationManger := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectRelationManger.EXPECT().ListMember(int64(1)).Return(
				[]dao.SubjectRelation{}, nil,
			).AnyTimes()

			manager := &groupService{
				manager:        mockSubjectRelationManger,
				authTypeManger: mockGroupSystemAuthTypeManager,
			}

			changed, err := manager.AlterGroupAuthType(nil, "test", 1, 0)
			assert.NoError(GinkgoT(), err)
			assert.True(GinkgoT(), changed)
		})
	})

	Describe("chunks", func() {
		It("ok", func() {
			assert.Equal(GinkgoT(), [][]int{}, chunks(0, 2))
			assert.Equal(GinkgoT(), [][]int{{0, 5}}, chunks(5, 6))
			assert.Equal(GinkgoT(), [][]int{{0, 2}, {2, 4}, {4, 5}}, chunks(5, 2))
		})
	})
})
