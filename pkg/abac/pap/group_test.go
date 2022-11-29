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

	"iam/pkg/abac/pip"
	abacTypes "iam/pkg/abac/types"
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
				Return(errors.New("error"))

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
				Return(errors.New("error"))

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
				Return(errors.New("error"))

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

			_, err := c.CheckSubjectEffectGroups("user", "notexist", []string{"10", "20"})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "cacheimpls.GetLocalSubjectPK")
		})

		It("get subject all group pks fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListEffectThinSubjectGroupsBySubjectPKGroupPKs(gomock.Any(), gomock.Any()).Return(
				nil, errors.New("error"),
			).AnyTimes()

			c := &groupController{
				service: mockGroupService,
			}

			_, err := c.CheckSubjectEffectGroups("user", "1", []string{"10", "20"})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListEffectThinSubjectGroupsBySubjectPKGroupPKs")
		})

		It("ok, all groupID valid", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListEffectThinSubjectGroupsBySubjectPKGroupPKs(gomock.Any(), gomock.Any()).Return(
				[]types.ThinSubjectGroup{{
					GroupPK:   10,
					ExpiredAt: 1,
				}, {
					GroupPK:   30,
					ExpiredAt: 1,
				}}, nil,
			).AnyTimes()

			c := &groupController{
				service: mockGroupService,
			}

			groupIDBelong, err := c.CheckSubjectEffectGroups("user", "1", []string{"10", "20"})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groupIDBelong, 2)
			assert.Equal(GinkgoT(), map[string]interface{}{
				"belong":     true,
				"expired_at": int64(1),
			}, groupIDBelong["10"])
			assert.Equal(GinkgoT(), map[string]interface{}{
				"belong":     false,
				"expired_at": 0,
			}, groupIDBelong["20"])
		})

		It("ok, has invalid groupID", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListEffectThinSubjectGroupsBySubjectPKGroupPKs(gomock.Any(), gomock.Any()).Return(
				[]types.ThinSubjectGroup{{
					GroupPK:   10,
					ExpiredAt: 1,
				}, {
					GroupPK:   30,
					ExpiredAt: 1,
				}}, nil,
			).AnyTimes()

			c := &groupController{
				service: mockGroupService,
			}

			groupIDBelong, err := c.CheckSubjectEffectGroups("user", "1", []string{"10", "20", "invalid"})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groupIDBelong, 3)
			assert.Equal(GinkgoT(), map[string]interface{}{
				"belong":     true,
				"expired_at": int64(1),
			}, groupIDBelong["10"])
			assert.Equal(GinkgoT(), map[string]interface{}{
				"belong":     false,
				"expired_at": 0,
			}, groupIDBelong["20"])
			assert.Equal(GinkgoT(), map[string]interface{}{
				"belong":     false,
				"expired_at": 0,
			}, groupIDBelong["invalid"])
		})
	})

	Describe("ListRbacGroupByResource", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("resourceTypePK error", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				switch _type {
				case "type":
					return 1, nil
				default:
					return 0, errors.New("err")
				}
			})

			c := &groupController{}

			_, err := c.ListRbacGroupByResource("system", abacTypes.Resource{
				System:    "system",
				Type:      "type1",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "abac.ParseResourceNode")
		})

		It("GetAuthorizedActionGroupMap error", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				switch _type {
				case "type":
					return 1, nil
				default:
					return 0, errors.New("err")
				}
			})

			mockGroupResourcePolicyService := mock.NewMockGroupResourcePolicyService(ctl)
			mockGroupResourcePolicyService.EXPECT().
				GetAuthorizedActionGroupMap("system", int64(1), int64(1), "id").
				Return(
					nil, errors.New("error"),
				).
				AnyTimes()

			c := &groupController{
				groupResourcePolicyService: mockGroupResourcePolicyService,
			}

			_, err := c.ListRbacGroupByResource("system", abacTypes.Resource{
				System:    "system",
				Type:      "type",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetAuthorizedActionGroupMap")
		})

		It("groupPKsToSubjects error", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				switch _type {
				case "type":
					return 1, nil
				default:
					return 0, errors.New("err")
				}
			})

			mockGroupResourcePolicyService := mock.NewMockGroupResourcePolicyService(ctl)
			mockGroupResourcePolicyService.EXPECT().
				GetAuthorizedActionGroupMap("system", int64(1), int64(1), "id").
				Return(
					map[int64][]int64{
						1: {1, 2},
					}, nil,
				).
				AnyTimes()

			patches.ApplyFunc(cacheimpls.GetSubjectByPK, func(pk int64) (subject types.Subject, err error) {
				return types.Subject{}, errors.New("err")
			})

			c := &groupController{
				groupResourcePolicyService: mockGroupResourcePolicyService,
			}

			_, err := c.ListRbacGroupByResource("system", abacTypes.Resource{
				System:    "system",
				Type:      "type",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "groupPKsToSubjects")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				switch _type {
				case "type":
					return 1, nil
				default:
					return 0, errors.New("err")
				}
			})

			mockGroupResourcePolicyService := mock.NewMockGroupResourcePolicyService(ctl)
			mockGroupResourcePolicyService.EXPECT().
				GetAuthorizedActionGroupMap("system", int64(1), int64(1), "id").
				Return(
					map[int64][]int64{
						1: {1},
					}, nil,
				).
				AnyTimes()

			patches.ApplyFunc(cacheimpls.GetSubjectByPK, func(pk int64) (subject types.Subject, err error) {
				return types.Subject{}, nil
			})

			c := &groupController{
				groupResourcePolicyService: mockGroupResourcePolicyService,
			}

			groups, err := c.ListRbacGroupByResource("system", abacTypes.Resource{
				System:    "system",
				Type:      "type",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []Subject{{}}, groups)
		})
	})

	Describe("ListRbacGroupByActionResource", func() {
		var patches *gomonkey.Patches
		BeforeEach(func() {
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("pip.GetActionDetail fail", func() {
			patches = gomonkey.ApplyFunc(
				pip.GetActionDetail,
				func(system, id string) (pk int64, authType int64, arts []abacTypes.ActionResourceType, err error) {
					return 0, 0, nil, errors.New("err")
				},
			)

			c := &groupController{}

			_, err := c.ListRbacGroupByActionResource("system", "action", abacTypes.Resource{})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "pip.GetActionDetail")
		})

		It("authType error", func() {
			patches = gomonkey.ApplyFunc(
				pip.GetActionDetail,
				func(system, id string) (pk int64, authType int64, arts []abacTypes.ActionResourceType, err error) {
					return 0, 0, nil, nil
				},
			)

			c := &groupController{}

			_, err := c.ListRbacGroupByActionResource("system", "action", abacTypes.Resource{})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "only support rbac")
		})

		It("actionResourceTypePK error", func() {
			patches = gomonkey.ApplyFunc(
				pip.GetActionDetail,
				func(system, id string) (pk int64, authType int64, arts []abacTypes.ActionResourceType, err error) {
					return 1, 2, []abacTypes.ActionResourceType{{System: "system", Type: "type"}}, nil
				},
			)

			patches.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _ string) (int64, error) {
				return 0, errors.New("err")
			})

			c := &groupController{}

			_, err := c.ListRbacGroupByActionResource("system", "action", abacTypes.Resource{})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "cacheimpls.GetLocalResourceTypePK")
		})

		It("resourceTypePK error", func() {
			patches = gomonkey.ApplyFunc(
				pip.GetActionDetail,
				func(system, id string) (pk int64, authType int64, arts []abacTypes.ActionResourceType, err error) {
					return 1, 2, []abacTypes.ActionResourceType{{System: "system", Type: "type"}}, nil
				},
			)

			patches.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				switch _type {
				case "type":
					return 1, nil
				default:
					return 0, errors.New("err")
				}
			})

			c := &groupController{}

			_, err := c.ListRbacGroupByActionResource("system", "action", abacTypes.Resource{
				System:    "system",
				Type:      "type1",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "abac.ParseResourceNode")
		})

		It("cacheimpls.GetResourceActionAuthorizedGroupPKs error", func() {
			patches = gomonkey.ApplyFunc(
				pip.GetActionDetail,
				func(system, id string) (pk int64, authType int64, arts []abacTypes.ActionResourceType, err error) {
					return 1, 2, []abacTypes.ActionResourceType{{System: "system", Type: "type"}}, nil
				},
			)

			patches.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				return 1, nil
			})

			patches.ApplyFunc(cacheimpls.GetResourceActionAuthorizedGroupPKs, func(
				systemID string,
				actionPK, actionResourceTypePK, resourceTypePK int64,
				resourceID string,
			) ([]int64, error) {
				return nil, errors.New("err")
			})

			c := &groupController{}

			_, err := c.ListRbacGroupByActionResource("system", "action", abacTypes.Resource{
				System:    "system",
				Type:      "type",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "cacheimpls.GetResourceActionAuthorizedGroupPKs")
		})

		It("cacheimpls.GetSubjectByPK error", func() {
			patches = gomonkey.ApplyFunc(
				pip.GetActionDetail,
				func(system, id string) (pk int64, authType int64, arts []abacTypes.ActionResourceType, err error) {
					return 1, 2, []abacTypes.ActionResourceType{{System: "system", Type: "type"}}, nil
				},
			)

			patches.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				return 1, nil
			})

			patches.ApplyFunc(cacheimpls.GetResourceActionAuthorizedGroupPKs, func(
				systemID string,
				actionPK, actionResourceTypePK, resourceTypePK int64,
				resourceID string,
			) ([]int64, error) {
				return []int64{1}, nil
			})

			patches.ApplyFunc(cacheimpls.GetSubjectByPK, func(pk int64) (subject types.Subject, err error) {
				return types.Subject{}, errors.New("err")
			})

			c := &groupController{}

			_, err := c.ListRbacGroupByActionResource("system", "action", abacTypes.Resource{
				System:    "system",
				Type:      "type",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "groupPKsToSubjects")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(
				pip.GetActionDetail,
				func(system, id string) (pk int64, authType int64, arts []abacTypes.ActionResourceType, err error) {
					return 1, 2, []abacTypes.ActionResourceType{{System: "system", Type: "type"}}, nil
				},
			)

			patches.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _type string) (int64, error) {
				return 1, nil
			})

			patches.ApplyFunc(cacheimpls.GetResourceActionAuthorizedGroupPKs, func(
				systemID string,
				actionPK, actionResourceTypePK, resourceTypePK int64,
				resourceID string,
			) ([]int64, error) {
				return []int64{1}, nil
			})

			patches.ApplyFunc(cacheimpls.GetSubjectByPK, func(pk int64) (subject types.Subject, err error) {
				return types.Subject{}, nil
			})

			c := &groupController{}

			groups, err := c.ListRbacGroupByActionResource("system", "action", abacTypes.Resource{
				System:    "system",
				Type:      "type",
				ID:        "id",
				Attribute: abacTypes.Attribute{},
			})

			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []Subject{{}}, groups)
		})
	})
})
