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

var _ = Describe("LocalApigwJwtClientId", func() {
	It("Key", func() {
		key := APIGatewayJWTClientIDCacheKey{
			JWTToken: "abc",
		}

		assert.Equal(GinkgoT(), "900150983cd24fb0d6963f7d28e17f72", key.Key())
	})

	It("retrieve should not work", func() {
		key := APIGatewayJWTClientIDCacheKey{
			JWTToken: "abc",
		}

		value, err := retrieveAPIGatewayJWTClientID(key)
		assert.Equal(GinkgoT(), "", value.(string))
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
			LocalAPIGatewayJWTClientIDCache = mockCache
		})

		It("not exists", func() {
			_, err := GetJWTTokenClientID("abc")
			assert.ErrorIs(GinkgoT(), err, ErrAPIGatewayJWTCacheNotFound)
		})

		It("not string", func() {
			key := APIGatewayJWTClientIDCacheKey{
				JWTToken: "abc",
			}

			LocalAPIGatewayJWTClientIDCache.Set(key, 1)

			_, err := GetJWTTokenClientID("abc")
			assert.ErrorIs(GinkgoT(), err, ErrAPIGatewayJWTClientIDNotString)
		})

		It("ok", func() {
			SetJWTTokenClientID("abc", "bk_test")

			clientID, err := GetJWTTokenClientID("abc")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "bk_test", clientID)
		})
	})
})
