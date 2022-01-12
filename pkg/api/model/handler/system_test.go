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
	"net/http"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cacheimpls"
	"iam/pkg/middleware"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// https://golang.org/pkg/testing/#hdr-Subtests_and_Sub_benchmarks

var _ = Describe("System", func() {

	Describe("defaultValidClients", func() {
		var c *gin.Context
		BeforeEach(func() {
			c = &gin.Context{}
			util.SetClientID(c, "abc")
		})

		It("equals", func() {
			assert.Equal(GinkgoT(), defaultValidClients(c, "abc"), "abc")
		})
		It("already contains", func() {
			assert.Equal(GinkgoT(), defaultValidClients(c, "123,abc"), "123,abc")
		})
		It("not contains", func() {
			assert.Contains(GinkgoT(), defaultValidClients(c, "123"), "123")
			assert.Contains(GinkgoT(), defaultValidClients(c, "123"), "abc")
		})

	})

})

func TestCreateSystem(t *testing.T) {
	t.Parallel()

	newRequestFunc := util.CreateNewAPIRequestFunc(
		"post", "/api/v1/systems", CreateSystem,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"hello": "123",
			}).BadRequest("bad request:ID is required")
	})

	t.Run("bad request invalid id", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"id":      "123",
				"name":    "test",
				"name_en": "test",
				"clients": "test_cli",
				"provider_config": map[string]interface{}{
					"host": "http://127.0.0.1",
					"auth": "basic",
				},
			}).BadRequestContainsMessage("bad request:invalid id")
	})

	// init the router
	r := util.SetupRouter()
	r.Use(middleware.ClientAuthMiddleware([]byte(""), false))
	url := "/api/v1/systems"
	r.POST(url, CreateSystem)

	// set the cache
	appCode := "test_app"
	appSecret := "123"

	cacheimpls.InitCaches(false)
	cacheimpls.LocalAppCodeAppSecretCache.Set(cacheimpls.AppCodeAppSecretCacheKey{
		AppCode:   appCode,
		AppSecret: appSecret,
	}, true)

	// for mock
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

	t.Run("uniq system", func(t *testing.T) {
		// path all func success
		patches = gomonkey.ApplyFunc(checkSystemCreateUnique,
			func(id, name, nameEn string) error {
				return errors.New("system info should be uniq")
			})
		defer restMock()

		apitest.New().
			Handler(r).
			Post(url).
			Header("X-Bk-App-Code", appCode).
			Header("X-Bk-App-Secret", appSecret).
			JSON(map[string]interface{}{
				"id":      "test_app",
				"name":    "test",
				"name_en": "test",
				"clients": "test_cli",
				"provider_config": map[string]interface{}{
					"host": "http://127.0.0.1",
					"auth": "basic",
				},
			}).
			Expect(t).
			Assert(util.NewResponseAssertFunc(t, func(resp util.Response) error {
				assert.Equal(t, resp.Code, util.ConflictError)
				assert.Contains(t, resp.Message, "conflict:")
				return nil
			})).
			Status(http.StatusOK).
			End()
	})

	t.Run("fail", func(t *testing.T) {
		// path all func success
		patches = gomonkey.ApplyFunc(checkSystemCreateUnique,
			func(id, name, nameEn string) error {
				return nil
			})

		ctl = gomock.NewController(t)

		mockService := mock.NewMockSystemService(ctl)
		mockService.EXPECT().Create(gomock.Any()).Return(
			errors.New("create fail"),
		).AnyTimes()
		patches.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockService
		})

		defer restMock()

		apitest.New().
			Handler(r).
			Post(url).
			Header("X-Bk-App-Code", appCode).
			Header("X-Bk-App-Secret", appSecret).
			JSON(map[string]interface{}{
				"id":      "test_app",
				"name":    "test",
				"name_en": "test",
				"clients": "test_cli",
				"provider_config": map[string]interface{}{
					"host": "http://127.0.0.1",
					"auth": "basic",
				},
			}).
			Expect(t).
			Assert(util.NewResponseAssertFunc(t, func(resp util.Response) error {
				assert.Equal(t, resp.Code, util.SystemError)
				assert.Contains(t, resp.Message, "system error")
				return nil
			})).
			Status(http.StatusOK).
			End()
	})

	t.Run("ok", func(t *testing.T) {
		// path all func success
		patches = gomonkey.ApplyFunc(checkSystemCreateUnique,
			func(id, name, nameEn string) error {
				return nil
			})

		ctl = gomock.NewController(t)

		mockService := mock.NewMockSystemService(ctl)
		mockService.EXPECT().Create(gomock.Any()).Return(
			nil,
		).AnyTimes()
		patches.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockService
		})
		defer restMock()

		apitest.New().
			Handler(r).
			Post(url).
			Header("X-Bk-App-Code", appCode).
			Header("X-Bk-App-Secret", appSecret).
			JSON(map[string]interface{}{
				"id":      "test_app",
				"name":    "test",
				"name_en": "test",
				"clients": "test_cli",
				"provider_config": map[string]interface{}{
					"host": "http://127.0.0.1",
					"auth": "basic",
				},
			}).
			Expect(t).
			Assert(util.NewResponseAssertFunc(t, func(resp util.Response) error {
				assert.Equal(t, resp.Code, util.NoError)
				return nil
			})).
			Status(http.StatusOK).
			End()
	})
}

func TestCreateSystemClientValidation(t *testing.T) {
	t.Parallel()

	r := util.SetupRouter()
	url := "/api/v1/systems"
	r.POST(url, CreateSystem)

	// system_id not equals to app_code
	apitest.New().
		Handler(r).
		Post(url).
		Header("X-Bk-App-Code", "test_app").
		JSON(map[string]interface{}{
			"id":      "test",
			"name":    "test",
			"name_en": "test",
			"clients": "test_cli",
			"provider_config": map[string]interface{}{
				"host": "http://127.0.0.1",
				"auth": "basic",
			},
		}).
		Expect(t).
		Assert(util.NewResponseAssertFunc(t, func(resp util.Response) error {
			assert.Equal(t, resp.Code, util.BadRequestError)
			assert.Contains(t, resp.Message, "bad request:system_id should be the app_code")
			return nil
		})).
		Status(http.StatusOK).
		End()
}

func TestUpdateSystem(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"put", "/api/v1/systems/test", UpdateSystem,
	)

	t.Run("no json", func(t *testing.T) {
		newRequestFunc(t).NoJSON()
	})

	t.Run("bad request invalid json", func(t *testing.T) {
		newRequestFunc(t).
			JSON(map[string]interface{}{
				"name": 123,
			}).BadRequestContainsMessage("bad request:")
	})

	// init the router
	r := util.SetupRouter()
	r.Use(middleware.ClientAuthMiddleware([]byte(""), false))
	url := "/api/v1/systems/test"
	r.POST(url, UpdateSystem)

	// set the cache
	appCode := "test_app"
	appSecret := "123"

	cacheimpls.InitCaches(false)
	cacheimpls.LocalAppCodeAppSecretCache.Set(cacheimpls.AppCodeAppSecretCacheKey{
		AppCode:   appCode,
		AppSecret: appSecret,
	}, true)

	// for mock
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

	t.Run("uniq system", func(t *testing.T) {
		// path all func success
		patches = gomonkey.ApplyFunc(checkSystemUpdateUnique,
			func(id, name, nameEn string) error {
				return errors.New("system info should be uniq")
			})
		defer restMock()

		apitest.New().
			Handler(r).
			Post(url).
			Header("X-Bk-App-Code", appCode).
			Header("X-Bk-App-Secret", appSecret).
			JSON(map[string]interface{}{
				"id":      "test_app",
				"name":    "test",
				"name_en": "test",
				"clients": "test_cli",
				"provider_config": map[string]interface{}{
					"host": "http://127.0.0.1",
					"auth": "basic",
				},
			}).
			Expect(t).
			Assert(util.NewResponseAssertFunc(t, func(resp util.Response) error {
				assert.Equal(t, resp.Code, util.ConflictError)
				assert.Contains(t, resp.Message, "conflict:")
				return nil
			})).
			Status(http.StatusOK).
			End()
	})

	t.Run("fail", func(t *testing.T) {
		// path all func success
		patches = gomonkey.ApplyFunc(checkSystemUpdateUnique,
			func(id, name, nameEn string) error {
				return nil
			})

		ctl = gomock.NewController(t)

		mockService := mock.NewMockSystemService(ctl)
		mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(
			errors.New("update fail"),
		).AnyTimes()
		patches.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockService
		})

		defer restMock()

		apitest.New().
			Handler(r).
			Post(url).
			Header("X-Bk-App-Code", appCode).
			Header("X-Bk-App-Secret", appSecret).
			JSON(map[string]interface{}{
				"name":    "test",
				"name_en": "test",
				"clients": "test_cli",
				"provider_config": map[string]interface{}{
					"host": "http://127.0.0.1",
					"auth": "basic",
				},
			}).
			Expect(t).
			Assert(util.NewResponseAssertFunc(t, func(resp util.Response) error {
				assert.Equal(t, resp.Code, util.SystemError)
				assert.Contains(t, resp.Message, "system error")
				return nil
			})).
			Status(http.StatusOK).
			End()
	})

	t.Run("ok", func(t *testing.T) {
		// path all func success
		patches = gomonkey.ApplyFunc(checkSystemUpdateUnique,
			func(id, name, nameEn string) error {
				return nil
			})

		ctl = gomock.NewController(t)

		mockService := mock.NewMockSystemService(ctl)
		mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(
			nil,
		).AnyTimes()
		patches.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockService
		})
		patches.ApplyFunc(cacheimpls.DeleteSystemCache, func(systemID string) error {
			return nil
		})

		defer restMock()

		apitest.New().
			Handler(r).
			Post(url).
			Header("X-Bk-App-Code", appCode).
			Header("X-Bk-App-Secret", appSecret).
			JSON(map[string]interface{}{
				"name":    "test",
				"name_en": "test",
				"clients": "test_cli",
				"provider_config": map[string]interface{}{
					"host": "http://127.0.0.1",
					"auth": "basic",
				},
			}).
			Expect(t).
			Assert(util.NewResponseAssertFunc(t, func(resp util.Response) error {
				assert.Equal(t, resp.Code, util.NoError)
				return nil
			})).
			Status(http.StatusOK).
			End()
	})
}

func TestGetSystem(t *testing.T) {
	t.Parallel()

	newRequestFunc := util.CreateNewAPIRequestFunc(
		"get", "/api/v1/systems/test", GetSystem,
	)

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("svc error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockSVC := mock.NewMockSystemService(ctl)
		mockSVC.EXPECT().Get(
			gomock.Any(),
		).Return(
			types.System{}, errors.New("get system fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockSVC
		})
		defer restMock()

		newRequestFunc(t).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockSVC := mock.NewMockSystemService(ctl)
		mockSVC.EXPECT().Get(
			gomock.Any(),
		).Return(
			types.System{ID: "test"}, nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockSVC
		})
		defer restMock()

		newRequestFunc(t).OK()
	})
}

func TestGetSystemClients(t *testing.T) {
	newRequestFunc := util.CreateNewAPIRequestFunc(
		"get", "/api/v1/systems/test/clients", GetSystemClients,
	)

	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	restMock := func() {
		ctl.Finish()
		if patches != nil {
			patches.Reset()
		}
	}

	t.Run("svc error", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockSVC := mock.NewMockSystemService(ctl)
		mockSVC.EXPECT().Get(
			gomock.Any(),
		).Return(
			types.System{}, errors.New("get system fail"),
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockSVC
		})
		defer restMock()

		newRequestFunc(t).SystemError()
	})

	t.Run("ok", func(t *testing.T) {
		ctl = gomock.NewController(t)
		mockSVC := mock.NewMockSystemService(ctl)
		mockSVC.EXPECT().Get(
			gomock.Any(),
		).Return(
			types.System{ID: "test"}, nil,
		).AnyTimes()
		patches = gomonkey.ApplyFunc(service.NewSystemService, func() service.SystemService {
			return mockSVC
		})
		defer restMock()

		newRequestFunc(t).OK()
	})
}
