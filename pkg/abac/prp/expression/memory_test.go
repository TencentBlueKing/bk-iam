/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package expression

import (
	"errors"
	"reflect"
	"time"

	"github.com/agiledragon/gomonkey"
	rds "github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/prp/common"
	"iam/pkg/cache"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("Memory", func() {

	It("newMemoryRetriever", func() {
		r := newMemoryRetriever(123, nil)
		assert.NotNil(GinkgoT(), r)
	})

	It("genKey", func() {
		r := newMemoryRetriever(123, nil)
		assert.Equal(GinkgoT(), "111", r.genKey(111))
	})

	Describe("retrieve", func() {
		var r *memoryRetriever

		// var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var retrievedExpressions []types.AuthExpression
		var cached1, cached2, cached3 *cachedExpression
		var now int64
		var hitExpressions map[string]*cachedExpression

		BeforeEach(func() {
			// ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()

			// init cache
			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)
			cacheimpls.LocalExpressionCache = gocache.New(1*time.Minute, 1*time.Minute)

			r = newMemoryRetriever(1, nil)

			now = time.Now().Unix()
			retrievedExpressions = []types.AuthExpression{
				{
					PK:         123,
					Expression: "",
				},
				{
					PK:         456,
					Expression: "",
				},
				{
					PK:         789,
					Expression: "",
				},
			}
			cached1 = &cachedExpression{
				timestamp: now,
				expression: types.AuthExpression{
					PK:         123,
					Expression: "123",
					Signature:  "202cb962ac59075b964b07152d234b70",
				},
			}
			cached2 = &cachedExpression{
				timestamp: now,
				expression: types.AuthExpression{PK: 456,
					Expression: "456",
					Signature:  "250cf8b51c773f3f8dc8b4be867a9a02",
				},
			}
			cached3 = &cachedExpression{
				timestamp: now,
				expression: types.AuthExpression{
					PK:         456,
					Expression: "456",
					Signature:  "250cf8b51c773f3f8dc8b4be867a9a02",
				},
			}

			hitExpressions = map[string]*cachedExpression{
				"123": cached1,
				"456": cached2,
				"789": cached3,
			}

		})
		AfterEach(func() {
			// ctl.Finish()
			patches.Reset()
		})

		It("batchFetchActionExpressionChangedList fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return nil, errors.New("ZRevRangeByScore fail")
				})

			expressions, missingPKs, err := r.retrieve([]int64{})
			assert.NoError(GinkgoT(), err)
			assert.Empty(GinkgoT(), expressions)
			assert.Nil(GinkgoT(), missingPKs)

		})
		It("all missing, no changed list", func() {
			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return retrievedExpressions, []int64{1000}, nil
			}

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789, 1000})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Equal(GinkgoT(), int64(123), expressions[0].PK)
			assert.Equal(GinkgoT(), int64(456), expressions[1].PK)
			assert.Equal(GinkgoT(), int64(789), expressions[2].PK)
			assert.Len(GinkgoT(), missingPKs, 1)
			assert.Equal(GinkgoT(), int64(1000), missingPKs[0])

			// check the cache
			_, ok := cacheimpls.LocalExpressionCache.Get("123")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalExpressionCache.Get("456")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalExpressionCache.Get("789")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalExpressionCache.Get("1000")
			assert.True(GinkgoT(), ok)
		})
		It("all hit, no change list", func() {
			for key, cached := range hitExpressions {
				cacheimpls.LocalExpressionCache.Set(key, cached, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return nil, nil, errors.New("should not be called")
			}

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Empty(GinkgoT(), missingPKs)

		})
		It("all hit, has change list", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return []rds.Z{
						{
							Score:  float64(now + 100),
							Member: "123",
						},
						{
							Score:  float64(now + 100),
							Member: "456",
						},
						{
							Score:  float64(now + 100),
							Member: "789",
						},
					}, nil
				})

			for key, cached := range hitExpressions {
				cacheimpls.LocalExpressionCache.Set(key, cached, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				assert.Len(GinkgoT(), pks, 3)
				return retrievedExpressions, nil, nil
			}

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Empty(GinkgoT(), missingPKs)
		})
		It("one expression cast fail", func() {
			// all hit
			for key, cached := range hitExpressions {
				cacheimpls.LocalExpressionCache.Set(key, cached, 0)
			}
			cacheimpls.LocalExpressionCache.Set("456", "abc", 0)

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return nil, nil, errors.New("should be called")
			}

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "should be called", err.Error())
			assert.Nil(GinkgoT(), expressions)
			assert.Nil(GinkgoT(), missingPKs)
		})
		It("one expression is empty", func() {
			// all hit
			for key, cached := range hitExpressions {
				cacheimpls.LocalExpressionCache.Set(key, cached, 0)
			}
			cacheimpls.LocalExpressionCache.Set("456", &cachedExpression{
				timestamp:  now + 100,
				expression: types.AuthExpression{},
			}, 0)

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 2)
			assert.Empty(GinkgoT(), missingPKs)
		})
		It("retrieve fail", func() {
			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return nil, nil, errors.New("should not be called")
			}
			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.Nil(GinkgoT(), expressions)
			assert.Nil(GinkgoT(), missingPKs)
			assert.Error(GinkgoT(), err)
		})

	})

	Describe("setMissing", func() {
		var r *memoryRetriever
		BeforeEach(func() {
			r = newMemoryRetriever(123, nil)
			cacheimpls.LocalExpressionCache = gocache.New(1*time.Minute, 1*time.Minute)

		})

		It("ok", func() {
			err := r.setMissing([]types.AuthExpression{
				{
					PK: 123,
				},
				{
					PK: 456,
				},
			}, []int64{789})

			assert.NoError(GinkgoT(), err)

			// check the cache
			_, ok := cacheimpls.LocalExpressionCache.Get("123")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalExpressionCache.Get("456")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalExpressionCache.Get("789")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalExpressionCache.Get("111")
			assert.False(GinkgoT(), ok)
		})
	})

	Describe("batchDeleteExpressionsFromMemory", func() {

		var patches *gomonkey.Patches
		BeforeEach(func() {
			cacheimpls.LocalExpressionCache = gocache.New(1*time.Minute, 1*time.Minute)
			cacheimpls.ChangeListCache = redis.NewMockCache("changelist", 1*time.Minute)

			patches = gomonkey.NewPatches()
		})

		AfterEach(func() {
			patches.Reset()
		})

		It("empty", func() {
			err := batchDeleteExpressionsFromMemory(map[int64][]int64{})
			assert.NoError(GinkgoT(), err)
		})

		It("ok", func() {
			// init local cache
			cacheimpls.LocalExpressionCache.Set("123", "abc", 0)
			_, ok := cacheimpls.LocalExpressionCache.Get("123")
			assert.True(GinkgoT(), ok)

			max := time.Now().Unix()
			min := max - expressionLocalCacheTTL

			// init redis cache
			err := cacheimpls.ChangeListCache.BatchZAdd([]redis.ZData{
				{
					Key: "expression:1",
					Zs: []*rds.Z{
						{
							Score:  float64(min + 50),
							Member: "0000",
						},
					},
				},
				{
					Key: "expression:2",
					Zs: []*rds.Z{
						{
							Score:  float64(min - 300), // will be removed
							Member: "1111",
						},
					},
				},
			})
			assert.NoError(GinkgoT(), err)

			// do delete
			err = batchDeleteExpressionsFromMemory(map[int64][]int64{
				1: {123, 456},
				2: {789},
			})
			assert.NoError(GinkgoT(), err)

			// check the local cache
			_, ok = cacheimpls.LocalExpressionCache.Get("123")
			assert.False(GinkgoT(), ok)

			// check the change list
			// _type=expression,  actionPK=1
			assert.True(GinkgoT(), cacheimpls.ChangeListCache.Exists(cache.NewStringKey("expression:1")))

			zs, err := cacheimpls.ChangeListCache.ZRevRangeByScore("expression:1", min, max, 0, maxChangeListCount)
			assert.NoError(GinkgoT(), err)
			// 0000 + 123 + 456
			assert.Len(GinkgoT(), zs, 3)

			// _type=expression,  actionPK=2
			assert.True(GinkgoT(), cacheimpls.ChangeListCache.Exists(cache.NewStringKey("expression:2")))
			zs, err = cacheimpls.ChangeListCache.ZRevRangeByScore("expression:2", min, max, 0, maxChangeListCount)
			assert.NoError(GinkgoT(), err)
			//  1111 + 789 - 1111 => 789
			assert.Len(GinkgoT(), zs, 1)
		})

		It("addToChangeList fail", func() {
			patches.ApplyMethod(reflect.TypeOf(changeList), "AddToChangeList",
				func(*common.ChangeList, map[string][]string) error {
					return errors.New("addToChangeList fail")
				})

			err := batchDeleteExpressionsFromMemory(map[int64][]int64{1: {123}})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "addToChangeList fail", err.Error())
		})

		It("Truncate fail", func() {
			patches.ApplyMethod(reflect.TypeOf(changeList), "Truncate",
				func(*common.ChangeList, []string) error {
					return errors.New("truncate fail")
				})

			err := batchDeleteExpressionsFromMemory(map[int64][]int64{1: {123}})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "truncate fail", err.Error())
		})

	})

})
