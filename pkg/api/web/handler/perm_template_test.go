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

func TestCreateAndDeleteTemplatePolicies(t *testing.T) {
	t.Parallel()

	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/perm-templates/policies", CreateAndDeleteTemplatePolicies,
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

	t.Run("bad request invalid create policies", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"subject":     map[string]interface{}{"type": "user", "id": "test"},
				"system_id":   "test",
				"template_id": int64(1),
				"create_policies": []map[string]interface{}{{
					"hello": "123",
				}},
				"delete_policy_ids": []int64{},
			}).BadRequest("bad request:data in array[0], ActionID is required")
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
		mockPolicyCtl := mock.NewMockPolicyController(ctl)
		mockPolicyCtl.EXPECT().CreateAndDeleteTemplatePolicies(
			"test", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		).Return(
			errors.New("create policies fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewPolicyController, func() pap.PolicyController {
			return mockPolicyCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"subject":           map[string]interface{}{"type": "user", "id": "test"},
				"system_id":         "test",
				"template_id":       int64(1),
				"create_policies":   []map[string]interface{}{},
				"delete_policy_ids": []int64{1},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockPolicyCtl := mock.NewMockPolicyController(ctl)
		mockPolicyCtl.EXPECT().CreateAndDeleteTemplatePolicies(
			"test", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewPolicyController, func() pap.PolicyController {
			return mockPolicyCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"subject":           map[string]interface{}{"type": "user", "id": "test"},
				"system_id":         "test",
				"template_id":       int64(1),
				"create_policies":   []map[string]interface{}{},
				"delete_policy_ids": []int64{1},
			}).OK()
	})
}

func TestUpdateTemplatePolicies(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put", "/api/v1/perm-templates/policies", UpdateTemplatePolicies,
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

	t.Run("bad request invalid update policies", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"subject":     map[string]interface{}{"type": "user", "id": "test"},
				"system_id":   "test",
				"template_id": int64(1),
				"update_policies": []map[string]interface{}{{
					"hello": "123",
				}},
				"delete_policy_ids": []int64{},
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

	t.Run("manager error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockPolicyCtl := mock.NewMockPolicyController(ctl)
		mockPolicyCtl.EXPECT().UpdateTemplatePolicies(
			"test", gomock.Any(), gomock.Any(), gomock.Any(),
		).Return(
			errors.New("update policies fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewPolicyController, func() pap.PolicyController {
			return mockPolicyCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"subject":     map[string]interface{}{"type": "user", "id": "test"},
				"system_id":   "test",
				"template_id": int64(1),
				"update_policies": []map[string]interface{}{
					{
						"id":                  int64(1),
						"action_id":           "test",
						"resource_expression": "test",
						"expired_at":          1,
					},
				},
				"delete_policy_ids": []int64{1},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockPolicyCtl := mock.NewMockPolicyController(ctl)
		mockPolicyCtl.EXPECT().UpdateTemplatePolicies(
			"test", gomock.Any(), gomock.Any(), gomock.Any(),
		).Return(
			nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewPolicyController, func() pap.PolicyController {
			return mockPolicyCtl
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"subject":     map[string]interface{}{"type": "user", "id": "test"},
				"system_id":   "test",
				"template_id": int64(1),
				"update_policies": []map[string]interface{}{
					{
						"id":                  int64(1),
						"action_id":           "test",
						"resource_expression": "test",
						"expired_at":          1,
					},
				},
				"delete_policy_ids": []int64{1},
			}).OK()
	})
}
