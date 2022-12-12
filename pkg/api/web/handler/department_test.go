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
			}).BadRequest("bad request:json decode or validate fail")
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
		mockCtl := mock.NewMockDepartmentController(ctl)
		mockCtl.EXPECT().BulkCreate([]pap.SubjectDepartment{
			{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			},
		}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewDepartmentController, func() pap.DepartmentController {
			return mockCtl
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
		mockCtl := mock.NewMockDepartmentController(ctl)
		mockCtl.EXPECT().BulkCreate([]pap.SubjectDepartment{
			{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			},
		}).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewDepartmentController, func() pap.DepartmentController {
			return mockCtl
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
			}).BadRequest("bad request:json decode or validate fail")
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
		mockCtl := mock.NewMockDepartmentController(ctl)
		mockCtl.EXPECT().BulkDelete([]string{"admin"}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewDepartmentController, func() pap.DepartmentController {
			return mockCtl
		})
		defer restMock()

		newRequestFunc(t).JSON([]string{"admin"}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockCtl := mock.NewMockDepartmentController(ctl)
		mockCtl.EXPECT().BulkDelete([]string{"admin"}).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewDepartmentController, func() pap.DepartmentController {
			return mockCtl
		})
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
			}).BadRequest("bad request:json decode or validate fail")
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
		mockCtl := mock.NewMockDepartmentController(ctl)
		mockCtl.EXPECT().BulkUpdate([]pap.SubjectDepartment{
			{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			},
		}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewDepartmentController, func() pap.DepartmentController {
			return mockCtl
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
		mockCtl := mock.NewMockDepartmentController(ctl)
		mockCtl.EXPECT().BulkUpdate([]pap.SubjectDepartment{
			{
				SubjectID:     "admin",
				DepartmentIDs: []string{"1", "2"},
			},
		}).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewDepartmentController, func() pap.DepartmentController {
			return mockCtl
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
