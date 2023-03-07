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

	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/util"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Action", func() {
})

func TestBatchCreateActions(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post",
		"/api/v1/model/systems/bk_test/actions",
		BatchCreateActions,
		"/api/v1/model/systems/:system_id/actions",
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequestContainsMessage("json decode or validate fail")
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		if ctl != nil {
			ctl.Finish()
		}
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("bad request valid fail", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateAction, func(body []actionSerializer) (bool, string) {
			return false, "valid fail"
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":                     "action_id",
				"name":                   "action_name",
				"name_en":                "action_name_en",
				"related_resource_types": []map[string]interface{}{},
				"related_actions":        []map[string]interface{}{},
			}}).BadRequestContainsMessage("valid fail")
	})

	t.Run("bad request repeat", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateAction, func(body []actionSerializer) (bool, string) {
			return true, ""
		})
		patches.ApplyFunc(validateActionsRepeat, func(actions []actionSerializer) error {
			return errors.New("repeat")
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":                     "action_id",
				"name":                   "action_name",
				"name_en":                "action_name_en",
				"related_resource_types": []map[string]interface{}{},
				"related_actions":        []map[string]interface{}{},
			}}).BadRequestContainsMessage("repeat")
	})

	t.Run("bad request quota check", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateAction, func(body []actionSerializer) (bool, string) {
			return true, ""
		})
		patches.ApplyFunc(validateActionsRepeat, func(actions []actionSerializer) error {
			return nil
		})
		patches.ApplyFunc(checkActionsQuotaAndAllUnique, func(systemID string, inActions []actionSerializer) error {
			return errors.New("quota")
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":                     "action_id",
				"name":                   "action_name",
				"name_en":                "action_name_en",
				"related_resource_types": []map[string]interface{}{},
				"related_actions":        []map[string]interface{}{},
			}}).BadRequestContainsMessage("quota", util.ConflictError)
	})

	t.Run("bad request check resource type exists error", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateAction, func(body []actionSerializer) (bool, string) {
			return true, ""
		})
		patches.ApplyFunc(validateActionsRepeat, func(actions []actionSerializer) error {
			return nil
		})
		patches.ApplyFunc(checkActionsQuotaAndAllUnique, func(systemID string, inActions []actionSerializer) error {
			return nil
		})
		patches.ApplyFunc(checkActionCreateResourceTypeAllExists, func(actions []actionSerializer) error {
			return errors.New("not exists")
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":                     "action_id",
				"name":                   "action_name",
				"name_en":                "action_name_en",
				"related_resource_types": []map[string]interface{}{},
				"related_actions":        []map[string]interface{}{},
			}}).BadRequestContainsMessage("not exists")
	})

	t.Run("manager error", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateAction, func(body []actionSerializer) (bool, string) {
			return true, ""
		})
		patches.ApplyFunc(validateActionsRepeat, func(actions []actionSerializer) error {
			return nil
		})
		patches.ApplyFunc(checkActionsQuotaAndAllUnique, func(systemID string, inActions []actionSerializer) error {
			return nil
		})
		patches.ApplyFunc(checkActionCreateResourceTypeAllExists, func(actions []actionSerializer) error {
			return nil
		})

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockActionService(ctl)
		mockSvc.EXPECT().BulkCreate(
			"bk_test", gomock.Any(),
		).Return(
			errors.New("create error"),
		).AnyTimes()

		patches.ApplyFunc(service.NewActionService, func() service.ActionService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":                     "action_id",
				"name":                   "action_name",
				"name_en":                "action_name_en",
				"related_resource_types": []map[string]interface{}{},
				"related_actions":        []map[string]interface{}{},
			}}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateAction, func(body []actionSerializer) (bool, string) {
			return true, ""
		})
		patches.ApplyFunc(validateActionsRepeat, func(actions []actionSerializer) error {
			return nil
		})
		patches.ApplyFunc(checkActionsQuotaAndAllUnique, func(systemID string, inActions []actionSerializer) error {
			return nil
		})
		patches.ApplyFunc(checkActionCreateResourceTypeAllExists, func(actions []actionSerializer) error {
			return nil
		})

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockActionService(ctl)
		mockSvc.EXPECT().BulkCreate(
			"bk_test", gomock.Any(),
		).Return(
			nil,
		).AnyTimes()

		patches.ApplyFunc(service.NewActionService, func() service.ActionService {
			return mockSvc
		})

		patches.ApplyFunc(cacheimpls.DeleteActionListCache, func(systemID string) error {
			return nil
		})

		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":                     "action_id",
				"name":                   "action_name",
				"name_en":                "action_name_en",
				"related_resource_types": []map[string]interface{}{},
				"related_actions":        []map[string]interface{}{},
			}}).OK()
	})
}

func TestUpdateAction(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put",
		"/api/v1/model/systems/bk_test/actions/action_id",
		UpdateAction,
		"/api/v1/model/systems/:system_id/actions/:action_id",
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		if ctl != nil {
			ctl.Finish()
		}
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("bad request check update unique", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(checkActionUpdateUnique, func(systemID, actionID, name, nameEn string) error {
			return errors.New("unique")
		})
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name":                   "action_name",
				"name_en":                "action_name_en",
				"related_resource_types": []map[string]interface{}{},
				"related_actions":        []map[string]interface{}{},
			}).BadRequestContainsMessage("unique", util.ConflictError)
	})

	t.Run("bad request check auth type fail", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(checkActionUpdateUnique, func(systemID, actionID, name, nameEn string) error {
			return nil
		})
		patches.ApplyFunc(
			checkUpdatedActionAuthType,
			func(systemID, actionID, authType string, relatedResourceTypes []relatedResourceType) error {
				return errors.New("check auth type fail")
			},
		)
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name":            "action_name",
				"name_en":         "action_name_en",
				"related_actions": []map[string]interface{}{},
			}).BadRequestContainsMessage("check auth type fail", util.ConflictError)
	})

	t.Run("manager error", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(checkActionUpdateUnique, func(systemID, actionID, name, nameEn string) error {
			return nil
		})
		patches.ApplyFunc(
			checkUpdatedActionAuthType,
			func(systemID, actionID, authType string, relatedResourceTypes []relatedResourceType) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockActionService(ctl)
		mockSvc.EXPECT().Update(
			"bk_test", "action_id", gomock.Any(),
		).Return(
			errors.New("update error"),
		).AnyTimes()

		patches.ApplyFunc(service.NewActionService, func() service.ActionService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name":            "action_name",
				"name_en":         "action_name_en",
				"related_actions": []map[string]interface{}{},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(checkActionUpdateUnique, func(systemID, actionID, name, nameEn string) error {
			return nil
		})
		patches.ApplyFunc(
			checkUpdatedActionAuthType,
			func(systemID, actionID, authType string, relatedResourceTypes []relatedResourceType) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockActionService(ctl)
		mockSvc.EXPECT().Update(
			"bk_test", "action_id", gomock.Any(),
		).Return(
			nil,
		).AnyTimes()

		patches.ApplyFunc(service.NewActionService, func() service.ActionService {
			return mockSvc
		})

		patches.ApplyFunc(cacheimpls.BatchDeleteActionCache, func(systemID string, actionIDs []string) error {
			return nil
		})
		patches.ApplyFunc(cacheimpls.DeleteActionListCache, func(systemID string) error {
			return nil
		})

		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name":            "action_name",
				"name_en":         "action_name_en",
				"related_actions": []map[string]interface{}{},
			}).OK()
	})
}

func TestDeleteAction(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete",
		"/api/v1/model/systems/bk_test/actions/action_id",
		DeleteAction,
		"/api/v1/model/systems/:system_id/actions/:action_id",
	)

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		if ctl != nil {
			ctl.Finish()
		}
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("bad request invalid check exists", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			checkActionIDsExist,
			func(systemID string, ids []string) error {
				return errors.New("not exists")
			},
		)
		defer restMock()

		newRequestFunc(t).BadRequestContainsMessage("not exists")
	})
}

func TestBatchBatchDeleteActions(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete",
		"/api/v1/model/systems/bk_test/actions",
		BatchDeleteActions,
		"/api/v1/model/systems/:system_id/actions",
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequestContainsMessage("json decode or validate fail")
	})

	t.Run("bad request empty", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]map[string]interface{}{}).BadRequestContainsMessage("the array should contain at least 1 item")
	})
}
