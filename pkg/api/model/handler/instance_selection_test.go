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
	. "github.com/onsi/ginkgo/v2"

	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/util"
)

var _ = Describe("InstanceSelection", func() {
})

func TestBatchCreateInstanceSelections(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post",
		"/api/v1/model/systems/bk_test/instance-selections",
		BatchCreateInstanceSelections,
		"/api/v1/model/systems/:system_id/instance-selections",
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

	t.Run("bad request invalid id", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":         "instance_selection_Aid",
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}}).BadRequestContainsMessage("id should begin with a lowercase letter")
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

	t.Run("bad request invalid id repeat", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			validateInstanceSelectionsRepeat,
			func(instanceSelections []instanceSelectionSerializer) error {
				return errors.New("repeat")
			},
		)
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":         "instance_selection_id",
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}}).BadRequestContainsMessage("repeat")
	})

	t.Run("bad request invalid quota check", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			validateInstanceSelectionsRepeat,
			func(instanceSelections []instanceSelectionSerializer) error {
				return nil
			},
		)
		patches.ApplyFunc(
			checkAllInstanceSelectionsQuotaAndUnique,
			func(systemID string,
				instanceSelections []instanceSelectionSerializer,
			) error {
				return errors.New("quota")
			},
		)
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":         "instance_selection_id",
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}}).BadRequestContainsMessage("quota", util.ConflictError)
	})

	t.Run("manager error", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			validateInstanceSelectionsRepeat,
			func(instanceSelections []instanceSelectionSerializer) error {
				return nil
			},
		)
		patches.ApplyFunc(
			checkAllInstanceSelectionsQuotaAndUnique,
			func(systemID string,
				instanceSelections []instanceSelectionSerializer,
			) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockInstanceSelectionService(ctl)
		mockSvc.EXPECT().BulkCreate(
			"bk_test", gomock.Any(),
		).Return(
			errors.New("create error"),
		).AnyTimes()

		patches.ApplyFunc(service.NewInstanceSelectionService, func() service.InstanceSelectionService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":         "instance_selection_id",
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			validateInstanceSelectionsRepeat,
			func(instanceSelections []instanceSelectionSerializer) error {
				return nil
			},
		)
		patches.ApplyFunc(
			checkAllInstanceSelectionsQuotaAndUnique,
			func(systemID string,
				instanceSelections []instanceSelectionSerializer,
			) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockInstanceSelectionService(ctl)
		mockSvc.EXPECT().BulkCreate(
			"bk_test", gomock.Any(),
		).Return(
			nil,
		).AnyTimes()

		patches.ApplyFunc(service.NewInstanceSelectionService, func() service.InstanceSelectionService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":         "instance_selection_id",
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}}).OK()
	})
}

func TestUpdateInstanceSelection(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put",
		"/api/v1/model/systems/bk_test/instance-selections/instance_selection_id",
		UpdateInstanceSelection,
		"/api/v1/model/systems/:system_id/instance-selections/:instance_selection_id",
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequestContainsMessage("Name is required")
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

	t.Run("bad request invalid quota check", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			checkInstanceSelectionUpdateUnique,
			func(systemID string, instanceSelectionID string, name string, nameEn string) error {
				return errors.New("quota")
			},
		)
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}).BadRequestContainsMessage("quota", util.ConflictError)
	})

	t.Run("manager error", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			checkInstanceSelectionUpdateUnique,
			func(systemID string, instanceSelectionID string, name string, nameEn string) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockInstanceSelectionService(ctl)
		mockSvc.EXPECT().Update(
			"bk_test", "instance_selection_id", gomock.Any(),
		).Return(
			errors.New("update error"),
		).AnyTimes()

		patches.ApplyFunc(service.NewInstanceSelectionService, func() service.InstanceSelectionService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			checkInstanceSelectionUpdateUnique,
			func(systemID string, instanceSelectionID string, name string, nameEn string) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockInstanceSelectionService(ctl)
		mockSvc.EXPECT().Update(
			"bk_test", "instance_selection_id", gomock.Any(),
		).Return(
			nil,
		).AnyTimes()

		patches.ApplyFunc(service.NewInstanceSelectionService, func() service.InstanceSelectionService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name":       "instance_selection_name",
				"name_en":    "instance_selection_name_en",
				"is_dynamic": false,
				"resource_type_chain": []map[string]interface{}{
					{
						"system_id": "bk_test",
						"id":        "resource_type_id",
					},
				},
			}).OK()
	})
}

func TestDeleteInstanceSelection(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete",
		"/api/v1/model/systems/bk_test/instance-selections/instance_selection_id",
		DeleteInstanceSelection,
		"/api/v1/model/systems/:system_id/instance-selections/:instance_selection_id",
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
			checkInstanceSelectionIDsExist,
			func(systemID string, ids []string) error {
				return errors.New("not exists")
			},
		)
		defer restMock()

		newRequestFunc(t).BadRequestContainsMessage("not exists")
	})
}

func TestBatchDeleteInstanceSelections(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete",
		"/api/v1/model/systems/bk_test/instance-selections",
		BatchDeleteInstanceSelections,
		"/api/v1/model/systems/:system_id/instance-selections",
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
