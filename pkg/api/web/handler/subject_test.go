/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"

	"iam/pkg/abac/pip"
	pl "iam/pkg/abac/prp/policy"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

func TestBatchCreateSubjects(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/subjects", BatchCreateSubjects,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:json decode or validate fail, " +
			"err=json: cannot unmarshal object into Go value of type []handler.createSubjectSerializer")
	})

	t.Run("bad request subjects", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"hello": "123",
				},
			}).BadRequestContainsMessage("bad request:json decode or validate fail")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkCreate(
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"type": "user",
					"id":   "admin",
					"name": "admin",
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkCreate(
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"type": "user",
					"id":   "admin",
					"name": "admin",
				},
			}).OK()
	})
}

func TestBatchDeleteSubjects(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete", "/api/v1/subjects", BatchDeleteSubjects,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:json decode or validate fail, " +
			"err=json: cannot unmarshal object into Go value of type []handler.deleteSubjectSerializer")
	})

	t.Run("bad request subjects", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"hello": "123",
				},
			}).BadRequestContainsMessage("bad request:json decode or validate fai")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDelete(
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			nil, errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"type": "user",
					"id":   "admin",
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDelete(
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			[]int64{1}, nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})

		mockManager2 := mock.NewMockSystemService(ctl)
		mockManager2.EXPECT().ListAll().Return(
			[]types.System{}, nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockManager2
		})

		patches.ApplyFunc(pip.BatchDeleteSubjectCache, func(pks []int64) error { return nil })
		patches.ApplyFunc(cacheimpls.DeleteSubjectPK, func(_type, id string) error { return nil })
		patches.ApplyFunc(cacheimpls.DeleteLocalSubjectPK, func(_type, id string) error { return nil })
		patches.ApplyFunc(pl.BatchDeleteSystemSubjectPKsFromCache,
			func(systems []string, subjectPKs []int64) error { return nil })

		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"type": "user",
					"id":   "admin",
				},
			}).OK()
	})
}

func TestBatchUpdateSubject(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put", "/api/v1/subjects", BatchUpdateSubject,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:json decode or validate fail, " +
			"err=json: cannot unmarshal object into Go value of type []handler.updateSubjectSerializer")
	})

	t.Run("bad request subjects", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"hello": "123",
				},
			}).BadRequestContainsMessage("bad request:json decode or validate fail")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkUpdateName(
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"type": "user",
					"id":   "admin",
					"name": "admin",
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkUpdateName(
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"type": "user",
					"id":   "admin",
					"name": "admin",
				},
			}).OK()
	})
}

func TestCreateSubjectRole(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/subject-roles", CreateSubjectRole,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:RoleType is required")
	})

	t.Run("bad request subjects", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"role_type": "system_manager",
				"system_id": "test",
				"subjects": []map[string]interface{}{
					{
						"hello": "123",
					},
				},
			}).BadRequest("bad request:data in array[0], Type is required")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkCreateSubjectRoles(
			"system_manager", "test",
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"role_type": "system_manager",
				"system_id": "test",
				"subjects": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkCreateSubjectRoles(
			"system_manager", "test",
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		patches.ApplyFunc(cacheimpls.DeleteSubjectRoleSystemID, func(_type, id string) error { return nil })
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"role_type": "system_manager",
				"system_id": "test",
				"subjects": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).OK()
	})
}

func TestDeleteSubjectRole(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete", "/api/v1/subject-roles", DeleteSubjectRole,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:RoleType is required")
	})

	t.Run("bad request subjects", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"role_type": "system_manager",
				"system_id": "test",
				"subjects": []map[string]interface{}{
					{
						"hello": "123",
					},
				},
			}).BadRequest("bad request:data in array[0], Type is required")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDeleteSubjectRoles(
			"system_manager", "test",
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"role_type": "system_manager",
				"system_id": "test",
				"subjects": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDeleteSubjectRoles(
			"system_manager", "test",
			[]types.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		patches.ApplyFunc(cacheimpls.DeleteSubjectRoleSystemID, func(_type, id string) error { return nil })
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"role_type": "system_manager",
				"system_id": "test",
				"subjects": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).OK()
	})
}

func TestBatchAddSubjectMembers(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/subject-members", BatchAddSubjectMembers,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:Type is required")
	})

	t.Run("bad request policy_expired_at", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":              "group",
				"id":                "1",
				"policy_expired_at": 0,
				"members":           []map[string]interface{}{{"type": "user", "id": "admin"}},
			}).BadRequest("bad request:policy expires time required when add group member")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager - list_member error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return(nil, errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":              "group",
				"id":                "1",
				"policy_expired_at": 10,
				"members": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).SystemError()
	})

	t.Run("manager - update_members_expired_at error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 9,
			},
		}, nil).AnyTimes()
		mockManager.EXPECT().UpdateMembersExpiredAt([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 10,
			},
		}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":              "group",
				"id":                "1",
				"policy_expired_at": 10,
				"members": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).SystemError()
	})

	t.Run("ok - not need add member", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 10,
			},
		}, nil).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":              "group",
				"id":                "1",
				"policy_expired_at": 10,
				"members": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).OK()
	})

	t.Run("manager - bulk_create subject members error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return([]types.SubjectMember{}, nil).AnyTimes()
		mockManager.EXPECT().BulkCreateSubjectMembers(
			"group",
			"1",
			[]types.Subject{{Type: "user", ID: "admin"}},
			int64(10),
		).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":              "group",
				"id":                "1",
				"policy_expired_at": 10,
				"members": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return([]types.SubjectMember{}, nil).AnyTimes()
		mockManager.EXPECT().BulkCreateSubjectMembers(
			"group",
			"1",
			[]types.Subject{{Type: "user", ID: "admin"}},
			int64(10),
		).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		patches.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) { return 1, nil })
		patches.ApplyFunc(pip.BatchDeleteSubjectCache, func(pks []int64) error { return nil })
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":              "group",
				"id":                "1",
				"policy_expired_at": 10,
				"members": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).OK()
	})
}

func TestDeleteSubjectMembers(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete", "/api/v1/subject-members", DeleteSubjectMembers,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:Type is required")
	})

	t.Run("bad request members", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":    "group",
				"id":      "1",
				"members": []map[string]interface{}{{"a": "user", "id": "admin"}},
			}).BadRequest("bad request:data in array[0], Type is required")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDeleteSubjectMembers(
			"group",
			"1",
			[]types.Subject{{Type: "user", ID: "admin"}},
		).Return(nil, errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type": "group",
				"id":   "1",
				"members": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDeleteSubjectMembers(
			"group",
			"1",
			[]types.Subject{{Type: "user", ID: "admin"}},
		).Return(map[string]int64{"user": int64(1)}, nil).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		patches.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) { return 1, nil })
		patches.ApplyFunc(pip.BatchDeleteSubjectCache, func(pks []int64) error { return nil })
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type": "group",
				"id":   "1",
				"members": []map[string]interface{}{
					{
						"type": "user",
						"id":   "admin",
					},
				},
			}).OK()
	})
}

func TestUpdateSubjectMembersExpiredAt(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put", "/api/v1/subject-members/expired_at", UpdateSubjectMembersExpiredAt,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:Type is required")
	})

	t.Run("bad request policy_expired_at", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type":    "group",
				"id":      "1",
				"members": []map[string]interface{}{{"type": "user", "policy_expired_at": 10}},
			}).BadRequest("bad request:data in array[0], ID is required")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager - list_member error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return(nil, errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type": "group",
				"id":   "1",
				"members": []map[string]interface{}{
					{
						"type":              "user",
						"id":                "admin",
						"policy_expired_at": 10,
					},
				},
			}).SystemError()
	})

	t.Run("ok - not need update member", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 10,
			},
		}, nil).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type": "group",
				"id":   "1",
				"members": []map[string]interface{}{
					{
						"type":              "user",
						"id":                "admin",
						"policy_expired_at": 10,
					},
				},
			}).OK()
	})

	t.Run("manager - update_members_expired_at error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 9,
			},
		}, nil).AnyTimes()
		mockManager.EXPECT().UpdateMembersExpiredAt([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 10,
			},
		}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type": "group",
				"id":   "1",
				"members": []map[string]interface{}{
					{
						"type":              "user",
						"id":                "admin",
						"policy_expired_at": 10,
					},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().ListMember("group", "1").Return([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 9,
			},
		}, nil).AnyTimes()
		mockManager.EXPECT().UpdateMembersExpiredAt([]types.SubjectMember{
			{
				PK:              1,
				Type:            "user",
				ID:              "admin",
				PolicyExpiredAt: 10,
			},
		}).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		patches.ApplyFunc(cacheimpls.GetSubjectPK, func(_type, id string) (pk int64, err error) { return 1, nil })
		patches.ApplyFunc(pip.BatchDeleteSubjectCache, func(pks []int64) error { return nil })
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"type": "group",
				"id":   "1",
				"members": []map[string]interface{}{
					{
						"type":              "user",
						"id":                "admin",
						"policy_expired_at": 10,
					},
				},
			}).OK()
	})
}

func TestBatchCreateSubjectDepartments(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/subject-departments", BatchCreateSubjectDepartments,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:json decode or validate fail, " +
			"err=json: cannot unmarshal object into Go value of type []handler.subjectDepartment")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkCreateSubjectDepartments(
			[]types.SubjectDepartment{{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			}},
		).Return(
			errors.New("error"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"id":          "admin",
					"departments": []string{"1", "2"},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkCreateSubjectDepartments(
			[]types.SubjectDepartment{{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"id":          "admin",
					"departments": []string{"1", "2"},
				},
			}).OK()
	})
}

func TestBatchDeleteSubjectDepartments(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete", "/api/v1/subject-departments", BatchDeleteSubjectDepartments,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:json decode or validate fail, " +
			"err=json: cannot unmarshal object into Go value of type []string")
	})

	t.Run("bad request subjects", func(t *testing.T) {
		newRequestFunc(t).JSON([]string{}).BadRequest("bad request:subject id can not be empty")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDeleteSubjectDepartments(
			[]string{"admin"},
		).Return(
			nil, errors.New("error"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).JSON([]string{"admin"}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkDeleteSubjectDepartments([]string{"admin"}).Return([]int64{1}, nil).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		patches.ApplyFunc(pip.BatchDeleteSubjectCache, func(pks []int64) error { return nil })
		defer restMock()

		newRequestFunc(t).JSON([]string{"admin"}).OK()
	})
}

func TestBatchUpdateSubjectDepartments(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put", "/api/v1/subject-departments", BatchUpdateSubjectDepartments,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:json decode or validate fail, " +
			"err=json: cannot unmarshal object into Go value of type []handler.subjectDepartment")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkUpdateSubjectDepartments(
			[]types.SubjectDepartment{{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			}},
		).Return(
			nil, errors.New("error"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"id":          "admin",
					"departments": []string{"1", "2"},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockManager := mock.NewMockSubjectService(ctl)
		mockManager.EXPECT().BulkUpdateSubjectDepartments(
			[]types.SubjectDepartment{{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			}},
		).Return(
			[]int64{1}, nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSubjectService, func() service.SubjectService {
			return mockManager
		})
		patches.ApplyFunc(pip.BatchDeleteSubjectCache, func(pks []int64) error { return nil })
		defer restMock()

		newRequestFunc(t).
			JSON([]interface{}{
				map[string]interface{}{
					"id":          "admin",
					"departments": []string{"1", "2"},
				},
			}).OK()
	})
}
