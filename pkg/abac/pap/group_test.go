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
	"database/sql"
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
	Describe("createOrUpdateGroupMembers", func() {
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
			patches.ApplyFunc(cacheimpls.BatchDeleteSubjectGroupCache, func(pks []int64) error {
				return nil
			})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("service.ListGroupMember fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListGroupMember(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupController{
				service: mockGroupService,
			}

			_, err := manager.alterGroupMembers("group", "1", []GroupMember{
				{
					Type:      "user",
					ID:        "2",
					ExpiredAt: int64(3),
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListGroupMember")
		})

		It("service.UpdateGroupMembersExpiredAtWithTx fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListGroupMember(int64(1)).Return(
				[]types.GroupMember{
					{
						PK:        1,
						SubjectPK: 2,
						ExpiredAt: 2,
					},
				}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateGroupMembersExpiredAtWithTx(
					gomock.Any(), int64(1),
					[]types.SubjectRelationForUpdate{{PK: 1, SubjectPK: 2, ExpiredAt: 3}},
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

			_, err := manager.alterGroupMembers("group", "1", []GroupMember{
				{
					Type:      "user",
					ID:        "2",
					ExpiredAt: int64(3),
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateGroupMembersExpiredAtWithTx")
		})

		It("bulkCreateGroupMembers fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListGroupMember(int64(1)).Return(
				[]types.GroupMember{}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateGroupMembersExpiredAtWithTx(gomock.Any(), int64(1), []types.SubjectRelationForUpdate{{PK: 1, SubjectPK: 2, ExpiredAt: 3}}).
				Return(
					nil,
				).
				AnyTimes()
			mockGroupService.EXPECT().
				BulkCreateGroupMembersWithTx(gomock.Any(), int64(1), []types.SubjectRelationForCreate{{
					SubjectPK: 2,
					GroupPK:   1,
					ExpiredAt: int64(3),
				}}).
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

			_, err := manager.alterGroupMembers("group", "1", []GroupMember{
				{
					Type:      "user",
					ID:        "2",
					ExpiredAt: int64(3),
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreateGroupMembersWithTx")
		})

		It("not create ok", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListGroupMember(int64(1)).Return(
				[]types.GroupMember{}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateGroupMembersExpiredAtWithTx(
					gomock.Any(), int64(1),
					[]types.SubjectRelationForUpdate{{PK: 1, SubjectPK: 2, ExpiredAt: 3}},
				).Return(
				nil,
			).
				AnyTimes()
			mockGroupService.EXPECT().ListGroupAuthSystemIDs(int64(1)).Return([]string{}, nil).AnyTimes()
			mockGroupAlterEventService := mock.NewMockGroupAlterEventService(ctl)
			mockGroupAlterEventService.EXPECT().
				CreateByGroupSubject(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("error"))

			patches.ApplyFunc(service.NewGroupService, func() service.GroupService {
				return mockGroupService
			})

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})

			manager := &groupController{
				service:                mockGroupService,
				groupAlterEventService: mockGroupAlterEventService,
			}

			_, err := manager.alterGroupMembers("group", "1", []GroupMember{
				{
					Type:      "user",
					ID:        "2",
					ExpiredAt: int64(3),
				},
			}, false)
			assert.NoError(GinkgoT(), err)
		})

		It("ok", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListGroupMember(int64(1)).Return(
				[]types.GroupMember{}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().
				UpdateGroupMembersExpiredAtWithTx(
					gomock.Any(), int64(1),
					[]types.SubjectRelationForUpdate{{PK: 1, SubjectPK: 2, ExpiredAt: 3}},
				).
				Return(
					nil,
				).
				AnyTimes()
			mockGroupService.EXPECT().
				BulkCreateGroupMembersWithTx(gomock.Any(), int64(1), []types.SubjectRelationForCreate{{
					SubjectPK: 2,
					GroupPK:   1,
					ExpiredAt: int64(3),
				}}).
				Return(
					nil,
				).
				AnyTimes()
			mockGroupService.EXPECT().ListGroupAuthSystemIDs(int64(1)).Return([]string{}, nil).AnyTimes()
			mockGroupAlterEventService := mock.NewMockGroupAlterEventService(ctl)
			mockGroupAlterEventService.EXPECT().
				CreateByGroupSubject(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("error"))

			patches.ApplyFunc(service.NewGroupService, func() service.GroupService {
				return mockGroupService
			})

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})

			manager := &groupController{
				service:                mockGroupService,
				groupAlterEventService: mockGroupAlterEventService,
			}

			typeCount, err := manager.alterGroupMembers("group", "1", []GroupMember{
				{
					Type:      "user",
					ID:        "2",
					ExpiredAt: int64(3),
				},
			}, true)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), map[string]int64{"user": 1, "department": 0}, typeCount)
		})
	})

	Describe("DeleteGroupMembers", func() {
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
			patches.ApplyFunc(cacheimpls.BatchDeleteSubjectGroupCache, func(pks []int64) error {
				return nil
			})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("service.BulkDeleteGroupMembers fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().BulkDeleteGroupMembers(int64(1), []int64{2}, []int64{3}).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupController{
				service: mockGroupService,
			}

			_, err := manager.DeleteGroupMembers("group", "1", []Subject{
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
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteGroupMembers")
		})

		It("ok", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().BulkDeleteGroupMembers(int64(1), []int64{2}, []int64{3}).Return(
				map[string]int64{"user": 1, "department": 0}, nil,
			).AnyTimes()
			mockGroupService.EXPECT().ListGroupAuthSystemIDs(int64(1)).Return([]string{}, nil).AnyTimes()
			mockGroupAlterEventService := mock.NewMockGroupAlterEventService(ctl)
			mockGroupAlterEventService.EXPECT().
				CreateByGroupSubject(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("error"))

			patches.ApplyFunc(service.NewGroupService, func() service.GroupService {
				return mockGroupService
			})

			manager := &groupController{
				service:                mockGroupService,
				groupAlterEventService: mockGroupAlterEventService,
			}

			typeCount, err := manager.DeleteGroupMembers("group", "1", []Subject{
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

	Describe("CheckSubjectExistGroups", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalSubjectPK, func(_type, id string) (pk int64, err error) {
				if _type == "user" && id == "1" {
					return int64(1), nil
				}
				if _type == "user" && id == "2" {
					return int64(2), nil
				}
				if _type == "group" && id == "10" {
					return int64(10), nil
				}

				if _type == "group" && id == "20" {
					return int64(20), nil
				}

				return 0, sql.ErrNoRows
			})

			patches.ApplyFunc(cacheimpls.GetSubjectDepartmentPKs, func(subjectPK int64) ([]int64, error) {
				return []int64{10, 20, 30}, nil
			})
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("get user subject PK fail", func() {
			c := &groupController{
				service: mock.NewMockGroupService(ctl),
			}

			_, err := c.CheckSubjectEffectGroups("user", "notexist", true, []string{"10", "20"})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "cacheimpls.GetLocalSubjectPK")
		})

		It("get subject all group pks fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().FilterExistEffectSubjectGroupPKs(gomock.Any(), gomock.Any()).Return(
				nil, errors.New("error"),
			).AnyTimes()

			c := &groupController{
				service: mockGroupService,
			}

			_, err := c.CheckSubjectEffectGroups("user", "1", true, []string{"10", "20"})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "FilterExistEffectSubjectGroupPKs")
		})

		It("ok, all groupID valid", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().FilterExistEffectSubjectGroupPKs(gomock.Any(), gomock.Any()).Return(
				[]int64{10, 30}, nil,
			).AnyTimes()

			c := &groupController{
				service: mockGroupService,
			}

			groupIDBelong, err := c.CheckSubjectEffectGroups("user", "1", true, []string{"10", "20"})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groupIDBelong, 2)
			assert.True(GinkgoT(), groupIDBelong["10"])
			assert.False(GinkgoT(), groupIDBelong["20"])
		})

		It("ok, has invalid groupID", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().FilterExistEffectSubjectGroupPKs(gomock.Any(), gomock.Any()).Return(
				[]int64{10, 30}, nil,
			).AnyTimes()

			c := &groupController{
				service: mockGroupService,
			}

			groupIDBelong, err := c.CheckSubjectEffectGroups("user", "1", true, []string{"10", "20", "invalid"})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groupIDBelong, 3)
			assert.True(GinkgoT(), groupIDBelong["10"])
			assert.False(GinkgoT(), groupIDBelong["20"])
			assert.False(GinkgoT(), groupIDBelong["invalid"])
		})
	})
})
