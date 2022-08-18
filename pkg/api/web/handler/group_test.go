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

func TestBatchAddGroupMembers(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/group-members", BatchAddGroupMembers,
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

	t.Run("CreateOrUpdateGroupMembers error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().CreateOrUpdateGroupMembers("group", "1", []pap.GroupMember{
			{
				Type:      "user",
				ID:        "admin",
				ExpiredAt: 10,
			},
		}).Return(nil, errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
		})
		patches.ApplyFunc(checkSubjectGroupsQuota, func(_, _ string, _ []pap.GroupMember) error { return nil })
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
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().CreateOrUpdateGroupMembers("group", "1", []pap.GroupMember{
			{
				Type:      "user",
				ID:        "admin",
				ExpiredAt: 10,
			},
		}).Return(map[string]int64{"user": 1, "department": 0}, nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
		})
		patches.ApplyFunc(checkSubjectGroupsQuota, func(_, _ string, _ []pap.GroupMember) error { return nil })
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

func TestDeleteGroupMembers(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete", "/api/v1/group-members", BatchDeleteGroupMembers,
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
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().DeleteGroupMembers("group", "1", []pap.Subject{
			{
				Type: "user",
				ID:   "admin",
			},
		}).Return(nil, errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
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
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().DeleteGroupMembers("group", "1", []pap.Subject{
			{
				Type: "user",
				ID:   "admin",
			},
		}).Return(map[string]int64{"user": 1, "department": 0}, nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
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
			}).OK()
	})
}

func TestUpdateGroupMembersExpiredAt(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put", "/api/v1/group-members/expired_at", BatchUpdateGroupMembersExpiredAt,
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

	t.Run("UpdateGroupMembersExpiredAt error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().UpdateGroupMembersExpiredAt("group", "1", []pap.GroupMember{
			{
				Type:      "user",
				ID:        "admin",
				ExpiredAt: 10,
			},
		}).Return(errors.New("error")).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
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
		mockCtl := mock.NewMockGroupController(ctl)
		mockCtl.EXPECT().UpdateGroupMembersExpiredAt("group", "1", []pap.GroupMember{
			{
				Type:      "user",
				ID:        "admin",
				ExpiredAt: 10,
			},
		}).Return(nil).AnyTimes()
		patches = gomonkey.ApplyFunc(pap.NewGroupController, func() pap.GroupController {
			return mockCtl
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
}
