/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/cache/memory"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("LocalApigwJWTAppInfo", func() {
	It("Key", func() {
		key := APIGatewayJWTAppInfoCacheKey{
			JWTToken: "abc",
		}

		assert.Equal(GinkgoT(), "900150983cd24fb0d6963f7d28e17f72", key.Key())
	})

	It("retrieve should not work", func() {
		key := APIGatewayJWTAppInfoCacheKey{
			JWTToken: "abc",
		}

		value, err := retrieveAPIGatewayJWTAppInfo(key)
		assert.Equal(GinkgoT(), AppInfo{}, value.(AppInfo))
		assert.NoError(GinkgoT(), err)
	})

	Describe("mock cache", func() {
		BeforeEach(func() {
			expiration := 5 * time.Minute

			retrieveFunc := func(key cache.Key) (interface{}, error) {
				return true, nil
			}
			mockCache := memory.NewCache(
				"mockCache", false, retrieveFunc, expiration, nil)
			LocalAPIGatewayJWTAppInfoCache = mockCache
		})

		It("not exists", func() {
			_, err := GetJWTTokenAppInfo("abc")
			assert.ErrorIs(GinkgoT(), err, ErrAPIGatewayJWTCacheNotFound)
		})

		It("not AppInfo", func() {
			key := APIGatewayJWTAppInfoCacheKey{
				JWTToken: "abc",
			}

			LocalAPIGatewayJWTAppInfoCache.Set(key, 1)

			_, err := GetJWTTokenAppInfo("abc")
			assert.ErrorIs(GinkgoT(), err, ErrAPIGatewayJWTValueNotAppInfo)
		})

		It("ok", func() {
			expectedAppInfo := AppInfo{
				AppCode:    "bk_test",
				TenantMode: "global",
				TenantID:   "",
			}
			SetJWTTokenAppInfo("abc", expectedAppInfo)

			appInfo, err := GetJWTTokenAppInfo("abc")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expectedAppInfo, appInfo)
		})
	})
})
