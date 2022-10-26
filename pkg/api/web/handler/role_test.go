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

	"iam/pkg/abac/pap"
	"iam/pkg/abac/pap/mock"
	"iam/pkg/util"
)

func TestAddRoleSubject(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/subject-roles", BatchAddRoleSubject,
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
		mockCtl := mock.NewMockRoleController(ctl)
		mockCtl.EXPECT().BulkAddSubjects("system_manager", "test", []pap.Subject{
			{
				Type: "user",
				ID:   "admin",
			},
		}).Return(errors.New("error"))

		patches = gomonkey.ApplyFunc(pap.NewRoleController, func() pap.RoleController {
			return mockCtl
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
		mockCtl := mock.NewMockRoleController(ctl)
		mockCtl.EXPECT().BulkAddSubjects("system_manager", "test", []pap.Subject{
			{
				Type: "user",
				ID:   "admin",
			},
		}).Return(nil)

		patches = gomonkey.ApplyFunc(pap.NewRoleController, func() pap.RoleController {
			return mockCtl
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
			}).OK()
	})
}

func TestBatchDeleteRoleSubject(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete", "/api/v1/subject-roles", BatchDeleteRoleSubject,
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
		mockCtl := mock.NewMockRoleController(ctl)
		mockCtl.EXPECT().BulkDeleteSubjects("system_manager", "test", []pap.Subject{
			{
				Type: "user",
				ID:   "admin",
			},
		}).Return(errors.New("test"))

		patches = gomonkey.ApplyFunc(pap.NewRoleController, func() pap.RoleController {
			return mockCtl
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
		mockCtl := mock.NewMockRoleController(ctl)
		mockCtl.EXPECT().BulkDeleteSubjects("system_manager", "test", []pap.Subject{
			{
				Type: "user",
				ID:   "admin",
			},
		}).Return(nil)

		patches = gomonkey.ApplyFunc(pap.NewRoleController, func() pap.RoleController {
			return mockCtl
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
			}).OK()
	})
}
