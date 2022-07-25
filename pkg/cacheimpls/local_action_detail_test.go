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
	"errors"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/service/types"
)

var _ = Describe("LocalActionDetailCache", func() {
	var (
		patches *gomonkey.Patches
		key     = ActionIDCacheKey{
			SystemID: "test_system_id",
			ActionID: "test_action_id",
		}
	)
	AfterEach(func() {
		// Note: 每个Case测试完成后，必须将缓存删除，否则下一个将会沿用原来的缓存
		LocalActionDetailCache.Delete(key)
		patches.Reset()
	})

	Context("GetLocalActionDetail", func() {
		It("retrieveActionDetailFromRedis OK", func() {
			patches = gomonkey.ApplyFunc(retrieveActionDetailFromRedis,
				func(cache.Key) (interface{}, error) {
					return types.ActionDetail{}, nil
				})

			_, err := GetLocalActionDetail(key.SystemID, key.ActionID)
			assert.NoError(GinkgoT(), err)
		})

		It("retrieveActionDetailFromRedis Error", func() {
			patches = gomonkey.ApplyFunc(retrieveActionDetailFromRedis,
				func(cache.Key) (interface{}, error) {
					return types.ActionDetail{}, errors.New("error")
				})

			_, err := GetLocalActionDetail(key.SystemID, key.ActionID)
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "LocalActionDetailCache.Get (.*) fail", err.Error())
		})

		It("retrieveActionDetailFromRedis return wrong data", func() {
			patches = gomonkey.ApplyFunc(retrieveActionDetailFromRedis,
				func(key cache.Key) (interface{}, error) {
					return key, nil
				})

			_, err := GetLocalActionDetail(key.SystemID, key.ActionID)
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "not types.ActionDetail in cache", err.Error())
		})
	})
})
