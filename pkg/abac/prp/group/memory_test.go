/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package group

import (
	"errors"
	"reflect"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	rds "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	gocache "github.com/wklken/go-cache"

	"iam/pkg/abac/prp/group/mock"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("memory", func() {
	Describe("genKey", func() {

		It("ok", func() {
			retriever := &localCacheGroupAuthTypeRetriever{
				systemID: "test",
			}
			key := retriever.genKey(1)
			assert.Equal(GinkgoT(), "test:1", key)
		})

	})

	Describe("batchGetGroupAuthType", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)
			cacheimpls.LocalGroupSystemAuthTypeCache = gocache.New(1*time.Minute, 1*time.Minute)

			cacheimpls.LocalGroupSystemAuthTypeCache.Set("test:1", int64(1), 1*time.Minute)
			cacheimpls.LocalGroupSystemAuthTypeCache.Set("test:2", int64(1), 1*time.Minute)
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("changeList.FetchList fail", func() {
			patches = gomonkey.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return nil, errors.New("ZRevRangeByScore fail")
				})

			retriever := &localCacheGroupAuthTypeRetriever{
				systemID: "test",
			}

			_, missingPKs := retriever.batchGetGroupAuthType([]int64{1, 2})
			assert.Equal(GinkgoT(), []int64{1, 2}, missingPKs)
		})

		It("changeList.FetchList empty", func() {
			patches = gomonkey.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return []rds.Z{}, nil
				})

			retriever := &localCacheGroupAuthTypeRetriever{
				systemID: "test",
				cache:    cacheimpls.LocalGroupSystemAuthTypeCache,
			}

			groupAuthTypes, missingPKs := retriever.batchGetGroupAuthType([]int64{1, 2, 3})
			assert.Equal(GinkgoT(), []int64{3}, missingPKs)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{GroupPK: 1, AuthType: 1},
				{GroupPK: 2, AuthType: 1},
			}, groupAuthTypes)
		})

		It("changeList.FetchList miss", func() {
			now := time.Now().Unix()
			patches = gomonkey.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return []rds.Z{
						{
							Score:  float64(now + 100),
							Member: "1",
						},
					}, nil
				})

			retriever := &localCacheGroupAuthTypeRetriever{
				systemID: "test",
				cache:    cacheimpls.LocalGroupSystemAuthTypeCache,
			}

			groupAuthTypes, missingPKs := retriever.batchGetGroupAuthType([]int64{1, 2, 3})
			assert.Equal(GinkgoT(), []int64{1, 3}, missingPKs)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{GroupPK: 2, AuthType: 1},
			}, groupAuthTypes)
		})

	})

	Describe("batchSetGroupAuthTypeCache", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			cacheimpls.LocalGroupSystemAuthTypeCache = gocache.New(1*time.Minute, 1*time.Minute)
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			retriever := &localCacheGroupAuthTypeRetriever{
				systemID: "test",
				cache:    cacheimpls.LocalGroupSystemAuthTypeCache,
			}

			retriever.batchSetGroupAuthTypeCache([]types.GroupAuthType{
				{GroupPK: 1, AuthType: 1},
				{GroupPK: 2, AuthType: 2},
			})

			val, _ := cacheimpls.LocalGroupSystemAuthTypeCache.Get("test:1")
			assert.Equal(GinkgoT(), int64(1), val.(int64))

			val, _ = cacheimpls.LocalGroupSystemAuthTypeCache.Get("test:2")
			assert.Equal(GinkgoT(), int64(2), val.(int64))
		})

	})

	Describe("Retrieve", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)
			cacheimpls.LocalGroupSystemAuthTypeCache = gocache.New(1*time.Minute, 1*time.Minute)

			cacheimpls.LocalGroupSystemAuthTypeCache.Set("test:1", int64(1), 1*time.Minute)
			cacheimpls.LocalGroupSystemAuthTypeCache.Set("test:2", int64(1), 1*time.Minute)
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("no missing retrieve ok", func() {
			patches = gomonkey.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return []rds.Z{}, nil
				})

			retriever := &localCacheGroupAuthTypeRetriever{
				systemID: "test",
				cache:    cacheimpls.LocalGroupSystemAuthTypeCache,
			}

			groupAuthTypes, err := retriever.Retrieve([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{GroupPK: 1, AuthType: 1},
				{GroupPK: 2, AuthType: 1},
			}, groupAuthTypes)
		})

		It("missing retrieve err", func() {
			patches = gomonkey.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return []rds.Z{}, nil
				})

			mockRetriever := mock.NewMockGroupAuthTypeRetriever(ctl)
			mockRetriever.EXPECT().Retrieve([]int64{3}).Return(nil, errors.New("missingRetriever error"))

			retriever := &localCacheGroupAuthTypeRetriever{
				systemID:         "test",
				cache:            cacheimpls.LocalGroupSystemAuthTypeCache,
				missingRetriever: mockRetriever,
			}

			_, err := retriever.Retrieve([]int64{1, 2, 3})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "missingRetriever")
		})

		It("missing retrieve ok", func() {
			patches = gomonkey.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return []rds.Z{}, nil
				})

			mockRetriever := mock.NewMockGroupAuthTypeRetriever(ctl)
			mockRetriever.EXPECT().Retrieve([]int64{3}).Return(
				[]types.GroupAuthType{{GroupPK: 3, AuthType: 3}}, nil,
			)

			retriever := &localCacheGroupAuthTypeRetriever{
				systemID:         "test",
				cache:            cacheimpls.LocalGroupSystemAuthTypeCache,
				missingRetriever: mockRetriever,
			}

			groupAuthTypes, err := retriever.Retrieve([]int64{1, 2, 3})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{GroupPK: 1, AuthType: 1},
				{GroupPK: 2, AuthType: 1},
				{GroupPK: 3, AuthType: 3},
			}, groupAuthTypes)
		})

	})
})
