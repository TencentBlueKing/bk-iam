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

func TestBatchCreateSubjectTemplateGroup(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/subject-template-groups", BatchCreateSubjectTemplateGroup,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"hello": "123",
			}}).BadRequest("bad request:json decode or validate fail")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("BulkCreateSubjectTemplateGroup error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().BulkCreateSubjectTemplateGroup([]pap.SubjectTemplateGroup{
			{
				Type:       "user",
				ID:         "1",
				TemplateID: 1,
				GroupID:    1,
				ExpiredAt:  10,
			},
		}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"type":        "user",
				"id":          "1",
				"template_id": 1,
				"group_id":    1,
				"expired_at":  10,
			}}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().BulkCreateSubjectTemplateGroup([]pap.SubjectTemplateGroup{
			{
				Type:       "user",
				ID:         "1",
				TemplateID: 1,
				GroupID:    1,
				ExpiredAt:  10,
			},
		}).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"type":        "user",
				"id":          "1",
				"template_id": 1,
				"group_id":    1,
				"expired_at":  10,
			}}).OK()
	})
}

func TestBatchDeleteSubjectTemplateGroup(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete", "/subject-template-groups", BatchDeleteSubjectTemplateGroup,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"hello": "123",
			}}).BadRequest("bad request:json decode or validate fail")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("BulkDeleteSubjectTemplateGroup error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().BulkDeleteSubjectTemplateGroup([]pap.SubjectTemplateGroup{
			{
				Type:       "user",
				ID:         "1",
				TemplateID: 1,
				GroupID:    1,
				ExpiredAt:  10,
			},
		}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"type":        "user",
				"id":          "1",
				"template_id": 1,
				"group_id":    1,
				"expired_at":  10,
			}}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().BulkDeleteSubjectTemplateGroup([]pap.SubjectTemplateGroup{
			{
				Type:       "user",
				ID:         "1",
				TemplateID: 1,
				GroupID:    1,
				ExpiredAt:  10,
			},
		}).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"type":        "user",
				"id":          "1",
				"template_id": 1,
				"group_id":    1,
				"expired_at":  10,
			}}).OK()
	})
}
