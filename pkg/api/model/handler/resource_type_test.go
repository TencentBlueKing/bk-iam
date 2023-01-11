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
	"github.com/stretchr/testify/assert"

	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/util"
)

var _ = Describe("ResourceType", func() {
	Describe("ResourceTypeUpdateSerializer Validate", func() {
		var slz resourceTypeUpdateSerializer
		BeforeEach(func() {
			slz = resourceTypeUpdateSerializer{}
		})

		It("name empty", func() {
			a := map[string]interface{}{
				"name": "",
			}
			valid, _ := slz.validate(a)
			assert.False(GinkgoT(), valid)
		})
		It("name_en empty", func() {
			b := map[string]interface{}{
				"name_en": "",
			}
			valid, _ := slz.validate(b)
			assert.False(GinkgoT(), valid)
		})
		It("version < 1", func() {
			c := map[string]interface{}{
				"version": 0,
			}
			valid, _ := slz.validate(c)
			assert.False(GinkgoT(), valid)
		})
		// 4. TODO: validate r.Parents
		// It("", func() {})
		// 5. TODO: validate provider_config
		// It("", func() {})
	})
})

func TestBatchCreateResourceTypes(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post",
		"/api/v1/model/systems/bk_test/resource-types",
		BatchCreateResourceTypes,
		"/api/v1/model/systems/:system_id/resource-types",
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

	t.Run("bad request invalid resource_type_id", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":      "resourceAid",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
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

	t.Run("bad request invalid resource id repeat", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateResourceTypesRepeat, func(resourceTypes []resourceTypeSerializer) error {
			return errors.New("repeat")
		})
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":      "resource_id",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
				},
			}}).BadRequestContainsMessage("repeat")
	})

	t.Run("bad request invalid resource quota check", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateResourceTypesRepeat, func(resourceTypes []resourceTypeSerializer) error {
			return nil
		})
		patches.ApplyFunc(
			checkAllResourceTypesQuotaAndUnique,
			func(systemID string, inResourceTypes []resourceTypeSerializer) error {
				return errors.New("quota")
			},
		)
		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":      "resource_id",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
				},
			}}).BadRequestContainsMessage("quota", util.ConflictError)
	})

	t.Run("manager error", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateResourceTypesRepeat, func(resourceTypes []resourceTypeSerializer) error {
			return nil
		})
		patches.ApplyFunc(
			checkAllResourceTypesQuotaAndUnique,
			func(systemID string, inResourceTypes []resourceTypeSerializer) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockResourceTypeService(ctl)
		mockSvc.EXPECT().BulkCreate(
			"bk_test", gomock.Any(),
		).Return(
			errors.New("service error"),
		).AnyTimes()

		patches.ApplyFunc(service.NewResourceTypeService, func() service.ResourceTypeService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":      "resource_id",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
				},
			}}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(validateResourceTypesRepeat, func(resourceTypes []resourceTypeSerializer) error {
			return nil
		})
		patches.ApplyFunc(
			checkAllResourceTypesQuotaAndUnique,
			func(systemID string, inResourceTypes []resourceTypeSerializer) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockResourceTypeService(ctl)
		mockSvc.EXPECT().BulkCreate(
			"bk_test", gomock.Any(),
		).Return(
			nil,
		).AnyTimes()

		patches.ApplyFunc(service.NewResourceTypeService, func() service.ResourceTypeService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON([]map[string]interface{}{{
				"id":      "resource_id",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
				},
			}}).OK()
	})
}

func TestUpdateResourceType(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put",
		"/api/v1/model/systems/bk_test/resource-types/1",
		UpdateResourceType,
		"/api/v1/model/systems/:system_id/resource-types/:resource_type_id",
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

	t.Run("bad request check unique", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			checkResourceTypeUpdateUnique,
			func(systemID string, resourceTypeID string, name string, nameEn string) error {
				return errors.New("not unique")
			},
		)
		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"id":      "resource_id",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
				},
			}).BadRequestContainsMessage("not unique", util.ConflictError)
	})

	t.Run("manager error", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			checkResourceTypeUpdateUnique,
			func(systemID string, resourceTypeID string, name string, nameEn string) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockResourceTypeService(ctl)
		mockSvc.EXPECT().Update(
			"bk_test", "1", gomock.Any(),
		).Return(
			errors.New("service error"),
		).AnyTimes()

		patches.ApplyFunc(service.NewResourceTypeService, func() service.ResourceTypeService {
			return mockSvc
		})

		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"id":      "resource_id",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
				},
			}).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(
			checkResourceTypeUpdateUnique,
			func(systemID string, resourceTypeID string, name string, nameEn string) error {
				return nil
			},
		)

		ctl = gomock.NewController(t)
		mockSvc := mock.NewMockResourceTypeService(ctl)
		mockSvc.EXPECT().Update(
			"bk_test", "1", gomock.Any(),
		).Return(
			nil,
		).AnyTimes()

		patches.ApplyFunc(service.NewResourceTypeService, func() service.ResourceTypeService {
			return mockSvc
		})

		patches.ApplyFunc(
			cacheimpls.BatchDeleteResourceTypeCache,
			func(systemID string, resourceTypeIDs []string) error {
				return nil
			},
		)

		defer restMock()

		newRequestFunc(t).
			JSON(map[string]interface{}{
				"id":      "resource_id",
				"name":    "resource_name",
				"name_en": "resource_name_en",
				"parents": []interface{}{},
				"provider_config": map[string]interface{}{
					"path": "/api/v1/resources/biz_set/query",
				},
			}).OK()
	})
}

func TestDeleteResourceType(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete",
		"/api/v1/model/systems/bk_test/resource-types/1",
		DeleteResourceType,
		"/api/v1/model/systems/:system_id/resource-types/:resource_type_id",
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

	t.Run("bad request invalid resource id repeat", func(t *testing.T) {
		patches = gomonkey.ApplyFunc(checkResourceTypeIDsExist, func(systemID string, ids []string) error {
			return errors.New("exists")
		})
		defer restMock()

		newRequestFunc(t).BadRequestContainsMessage("exists")
	})
}

func TestBatchDeleteResourceTypes(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"delete",
		"/api/v1/model/systems/bk_test/resource-types",
		BatchDeleteResourceTypes,
		"/api/v1/model/systems/:system_id/resource-types",
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

	t.Run("bad request invalid empty", func(t *testing.T) {
		newRequestFunc(t).
			JSON([]map[string]interface{}{}).BadRequestContainsMessage("the array should contain at least 1 item")
	})
}
