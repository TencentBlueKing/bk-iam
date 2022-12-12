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
	papMock "iam/pkg/abac/pap/mock"
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
			}).BadRequest("bad request:json decode or validate fail, err=json:")
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
		mockManager := papMock.NewMockSubjectController(ctl)
		mockManager.EXPECT().BulkCreate(
			[]pap.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewSubjectController, func() pap.SubjectController {
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
		mockManager := papMock.NewMockSubjectController(ctl)
		mockManager.EXPECT().BulkCreate(
			[]pap.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewSubjectController, func() pap.SubjectController {
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
			}).BadRequest("bad request:json decode or validate fail, err=json:")
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
		mockManager := papMock.NewMockSubjectController(ctl)
		mockManager.EXPECT().BulkDeleteUserAndDepartment(
			[]pap.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewSubjectController, func() pap.SubjectController {
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
		mockManager := papMock.NewMockSubjectController(ctl)
		mockManager.EXPECT().BulkDeleteUserAndDepartment(
			[]pap.Subject{{
				Type: "user",
				ID:   "admin",
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewSubjectController, func() pap.SubjectController {
			return mockManager
		})

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
			}).BadRequest("bad request:json decode or validate fail, err=json:")
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
		mockManager := papMock.NewMockSubjectController(ctl)
		mockManager.EXPECT().BulkUpdateName(
			[]pap.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewSubjectController, func() pap.SubjectController {
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
		mockManager := papMock.NewMockSubjectController(ctl)
		mockManager.EXPECT().BulkUpdateName(
			[]pap.Subject{{
				Type: "user",
				ID:   "admin",
				Name: "admin",
			}},
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewSubjectController, func() pap.SubjectController {
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
