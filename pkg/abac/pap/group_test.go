/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pap

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("GroupController", func() {
	Describe("createOrUpdateSubjectMembers", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			patches = gomonkey.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				switch id {
				case "1":
					return int64(1), nil
				case "2":
					return int64(2), nil
				}

				return 0, nil
			})
			patches.ApplyFunc(cacheimpls.BatchDeleteSubjectCache, func(pks []int64) error {
				return nil
			})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("service.ListMember fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListMember(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupController{
				service: mockGroupService,
			}

			_, err := manager.alterSubjectMembers("group", "1", []GroupMember{
				{
					Type:            "user",
					ID:              "2",
					PolicyExpiredAt: int64(3),
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListMember")
		})

		It("service.UpdateMembersExpiredAtWithTx fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListMember(int64(1)).Return(
				[]types.GroupMember{
					{
						PK:              1,
						SubjectPK:       2,
						PolicyExpiredAt: 2,
					},
				}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateMembersExpiredAtWithTx(
					gomock.Any(), int64(1),
					[]types.SubjectRelationPKPolicyExpiredAt{{PK: 1, SubjectPK: 2, PolicyExpiredAt: 3}},
				).
				Return(
					errors.New("error"),
				).
				AnyTimes()

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})

			manager := &groupController{
				service: mockGroupService,
			}

			_, err := manager.alterSubjectMembers("group", "1", []GroupMember{
				{
					Type:            "user",
					ID:              "2",
					PolicyExpiredAt: int64(3),
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateMembersExpiredAtWithTx")
		})

		It("bulkCreateSubjectMembers fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListMember(int64(1)).Return(
				[]types.GroupMember{}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateMembersExpiredAtWithTx(gomock.Any(), int64(1), []types.SubjectRelationPKPolicyExpiredAt{{PK: 1, SubjectPK: 2, PolicyExpiredAt: 3}}).
				Return(
					nil,
				).
				AnyTimes()
			mockGroupService.EXPECT().BulkCreateSubjectMembersWithTx(gomock.Any(), int64(1), []types.SubjectRelation{{
				SubjectPK:       2,
				ParentPK:        1,
				PolicyExpiredAt: int64(3),
			}}).Return(
				errors.New("error"),
			).AnyTimes()

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})

			manager := &groupController{
				service: mockGroupService,
			}

			_, err := manager.alterSubjectMembers("group", "1", []GroupMember{
				{
					Type:            "user",
					ID:              "2",
					PolicyExpiredAt: int64(3),
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreateSubjectMembersWithTx")
		})

		It("not create ok", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListMember(int64(1)).Return(
				[]types.GroupMember{}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateMembersExpiredAtWithTx(
					gomock.Any(), int64(1),
					[]types.SubjectRelationPKPolicyExpiredAt{{PK: 1, SubjectPK: 2, PolicyExpiredAt: 3}},
				).Return(
				nil,
			).
				AnyTimes()
			mockSystemService := mock.NewMockSystemService(ctl)
			mockSystemService.EXPECT().ListAll().Return([]types.System{}, nil).AnyTimes()

			patches.ApplyFunc(service.NewSystemService, func() service.SystemService {
				return mockSystemService
			})

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})

			manager := &groupController{
				service: mockGroupService,
			}

			_, err := manager.alterSubjectMembers("group", "1", []GroupMember{
				{
					Type:            "user",
					ID:              "2",
					PolicyExpiredAt: int64(3),
				},
			}, false)
			assert.NoError(GinkgoT(), err)
		})

		It("ok", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListMember(int64(1)).Return(
				[]types.GroupMember{}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateMembersExpiredAtWithTx(
					gomock.Any(), int64(1),
					[]types.SubjectRelationPKPolicyExpiredAt{{PK: 1, SubjectPK: 2, PolicyExpiredAt: 3}},
				).
				Return(
					nil,
				).
				AnyTimes()
			mockGroupService.EXPECT().BulkCreateSubjectMembersWithTx(gomock.Any(), int64(1), []types.SubjectRelation{{
				SubjectPK:       2,
				ParentPK:        1,
				PolicyExpiredAt: int64(3),
			}}).Return(
				nil,
			).AnyTimes()
			mockSystemService := mock.NewMockSystemService(ctl)
			mockSystemService.EXPECT().ListAll().Return([]types.System{}, nil).AnyTimes()

			patches.ApplyFunc(service.NewSystemService, func() service.SystemService {
				return mockSystemService
			})

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})

			manager := &groupController{
				service: mockGroupService,
			}

			typeCount, err := manager.alterSubjectMembers("group", "1", []GroupMember{
				{
					Type:            "user",
					ID:              "2",
					PolicyExpiredAt: int64(3),
				},
			}, true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), map[string]int64{"user": 1, "department": 0}, typeCount)
		})
	})

	Describe("DeleteSubjectMembers", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			patches = gomonkey.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) {
				switch id {
				case "1":
					return int64(1), nil
				case "2":
					return int64(2), nil
				case "3":
					return int64(3), nil
				}

				return 0, nil
			})
			patches.ApplyFunc(cacheimpls.BatchDeleteSubjectCache, func(pks []int64) error {
				return nil
			})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("service.BulkDeleteSubjectMembers fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().BulkDeleteSubjectMembers(int64(1), []int64{2}, []int64{3}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupController{
				service: mockGroupService,
			}

			_, err := manager.DeleteSubjectMembers("group", "1", []Subject{
				{
					Type: "user",
					ID:   "2",
				},
				{
					Type: "department",
					ID:   "3",
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteSubjectMembers")
		})

		It("ok", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().BulkDeleteSubjectMembers(int64(1), []int64{2}, []int64{3}).Return(
				map[string]int64{"user": 1, "department": 0}, nil,
			).AnyTimes()
			mockSystemService := mock.NewMockSystemService(ctl)
			mockSystemService.EXPECT().ListAll().Return([]types.System{}, nil).AnyTimes()

			patches.ApplyFunc(service.NewSystemService, func() service.SystemService {
				return mockSystemService
			})

			manager := &groupController{
				service: mockGroupService,
			}

			typeCount, err := manager.DeleteSubjectMembers("group", "1", []Subject{
				{
					Type: "user",
					ID:   "2",
				},
				{
					Type: "department",
					ID:   "3",
				},
			})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), map[string]int64{"user": 1, "department": 0}, typeCount)
		})
	})
})
