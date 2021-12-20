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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("Redis", func() {
	It("newRedisRetriever", func() {
		a := newRedisRetriever(nil)
		assert.NotNil(GinkgoT(), a)
	})

	Describe("retrieve", func() {
		var r *redisRetriever

		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var retrievedExpressions []types.AuthExpression
		var hitExpressions map[int64]string
		var exprStr1, exprStr2, exprStr3 []byte
		var emptyExprStr []byte
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()

			cacheimpls.ExpressionCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}

			retrievedExpressions =
				[]types.AuthExpression{
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

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return retrievedExpressions, []int64{}, nil
			}

			exprStr1, _ = cacheimpls.ExpressionCache.Marshal(types.AuthExpression{
				PK:         123,
				Expression: "123",
				Signature:  "202cb962ac59075b964b07152d234b70",
			})
			exprStr2, _ = cacheimpls.ExpressionCache.Marshal(types.AuthExpression{
				PK:         456,
				Expression: "456",
				Signature:  "250cf8b51c773f3f8dc8b4be867a9a02",
			})
			exprStr3, _ = cacheimpls.ExpressionCache.Marshal(types.AuthExpression{
				PK:         789,
				Expression: "789",
				Signature:  "68053af2923e00204c3ca7c6a3150cf7",
			})
			emptyExprStr, _ = cacheimpls.ExpressionCache.Marshal(types.AuthExpression{})
			hitExpressions = map[int64]string{
				123: string(exprStr1),
				456: string(exprStr2),
				789: string(exprStr3),
			}

		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("batchGet fail", func() {
			// patches.ApplyMethod(reflect.TypeOf(cacheimpls.ExpressionCache), "BatchGet",
			//	func(c *redis.Cache, keys []cache.Key) (map[cache.Key]string, error) {
			//		return nil, errors.New("batch get fail")
			//	})

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve([]int64{})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 0)
			assert.Len(GinkgoT(), missingPKs, 0)
		})

		It("all missing", func() {
			pks := []int64{123, 456, 789}

			// patches.ApplyFunc(r.batchGet,
			//	func(expressionPKs []int64) (
			//		hitExpressions map[int64]string,
			//		missExpressionPKs []int64,
			//		err error,
			//	) {
			//		return nil, pks, nil
			//	})

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Equal(GinkgoT(), int64(123), expressions[0].PK)
			assert.Equal(GinkgoT(), int64(456), expressions[1].PK)
			assert.Equal(GinkgoT(), int64(789), expressions[2].PK)
			assert.Empty(GinkgoT(), missingPKs)

			// check the cache
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(789)))
		})

		It("all hit", func() {
			pks := []int64{123, 456, 789}

			for pk, exprStr := range hitExpressions {
				cacheimpls.ExpressionCache.Set(cache.NewInt64Key(pk), exprStr, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return nil, nil, errors.New("should not be called")
			}

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Len(GinkgoT(), missingPKs, 0)

		})

		It("some hit, some missing", func() {
			pks := []int64{123, 456, 789}

			cacheimpls.ExpressionCache.Set(cache.NewInt64Key(int64(456)), string(exprStr2), 0)

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return []types.AuthExpression{
					{PK: 123},
					{PK: 789},
				}, nil, nil
			}

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Nil(GinkgoT(), missingPKs)

			// check the cache
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(789)))

		})

		It("some hit, some missing, retrieve has missing key", func() {
			pks := []int64{123, 456, 789}

			cacheimpls.ExpressionCache.Set(cache.NewInt64Key(int64(456)), string(exprStr2), 0)

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return []types.AuthExpression{
					{PK: 789},
				}, []int64{123}, nil
			}

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 2)
			assert.Len(GinkgoT(), missingPKs, 1)
			assert.Equal(GinkgoT(), int64(123), missingPKs[0])

			// check the cache
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(789)))
		})

		It("one expression unmarshal fail", func() {
			pks := []int64{123, 456, 789}

			for pk, exprStr := range hitExpressions {
				cacheimpls.ExpressionCache.Set(cache.NewInt64Key(pk), exprStr, 0)
			}
			cacheimpls.ExpressionCache.Set(cache.NewInt64Key(int64(456)), "not a valid json", 0)

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return nil, nil, errors.New("should do retrieve 456")
			}

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve(pks)
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), expressions)
			assert.Nil(GinkgoT(), missingPKs)

		})

		It("one expression is empty", func() {
			pks := []int64{123, 456, 789}

			cacheimpls.ExpressionCache.Set(cache.NewInt64Key(int64(456)), string(exprStr2), 0)
			cacheimpls.ExpressionCache.Set(cache.NewInt64Key(int64(789)), string(emptyExprStr), 0)

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return []types.AuthExpression{
					{PK: 123},
				}, nil, nil
			}

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 2)
			assert.Len(GinkgoT(), missingPKs, 1)
			assert.Equal(GinkgoT(), int64(789), missingPKs[0])

			// check the cache
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(789)))

		})

		It("retrieve fail", func() {
			pks := []int64{123, 456, 789}

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
				return nil, nil, errors.New("retrieve fail")
			}

			_, _, err := r.retrieve(pks)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "retrieve fail", err.Error())
		})

	})

	Describe("setMissing", func() {
		var r *redisRetriever
		BeforeEach(func() {
			cacheimpls.ExpressionCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}
		})

		It("ok", func() {
			err := r.setMissing([]types.AuthExpression{
				{PK: int64(123)},
			}, []int64{456, 789})
			assert.NoError(GinkgoT(), err)

			// all key exists
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(789)))
		})

	})

	// helper functions
	Describe("batchGet", func() {
		var r *redisRetriever
		BeforeEach(func() {
			cacheimpls.ExpressionCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}
		})

		It("cache BatchGet fail", func() {
			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ExpressionCache), "BatchGet",
				func(c *redis.Cache, keys []cache.Key) (map[cache.Key]string, error) {
					return nil, errors.New("batch get fail")
				})

			_, _, err := r.batchGet([]int64{123})
			assert.Error(GinkgoT(), err)

		})

		It("ok", func() {
			expression := map[int64]types.AuthExpression{
				123: {},
				456: {},
			}

			err := r.batchSet(expression)
			assert.NoError(GinkgoT(), err)

			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
		})

	})

	Describe("batchSet", func() {
		var r *redisRetriever
		var patches *gomonkey.Patches
		var noCacheExpressions map[int64]types.AuthExpression
		BeforeEach(func() {
			patches = gomonkey.NewPatches()
			cacheimpls.ExpressionCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}

			noCacheExpressions = map[int64]types.AuthExpression{
				123: {
					PK:         123,
					Expression: "",
				},
				456: {
					PK:         456,
					Expression: "",
				},
			}
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("cache Marshal fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ExpressionCache), "Marshal",
				func(c *redis.Cache, value interface{}) ([]byte, error) {
					return nil, errors.New("marshal fail")
				})

			err := r.batchSet(noCacheExpressions)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "marshal fail", err.Error())
		})

		It("cache BatchSetWithTx fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ExpressionCache), "BatchSetWithTx",
				func(c *redis.Cache, kvs []redis.KV, expiration time.Duration) error {
					return errors.New("batchSetWithTx fail")
				})
			defer patches.Reset()

			err := r.batchSet(noCacheExpressions)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchSetWithTx fail", err.Error())

		})

		It("ok", func() {

			err := r.batchSet(noCacheExpressions)
			assert.NoError(GinkgoT(), err)

			// all key exists
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
		})

	})

	Describe("batchDelete", func() {
		var r *redisRetriever
		var patches *gomonkey.Patches
		BeforeEach(func() {
			patches = gomonkey.NewPatches()

			cacheimpls.ExpressionCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("empty pks", func() {
			err := r.batchDelete([]int64{})
			assert.NoError(GinkgoT(), err)
		})

		It("not exists, delete ok", func() {
			err := r.batchDelete([]int64{123})
			assert.NoError(GinkgoT(), err)
		})

		It("cache BatchSetWithTx fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ExpressionCache), "BatchDelete",
				func(c *redis.Cache, keys []cache.Key) error {
					return errors.New("batchDelete fail")
				})
			defer patches.Reset()

			err := r.batchDelete([]int64{123})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchDelete fail", err.Error())

		})

		It("ok", func() {
			key := cache.NewInt64Key(123)
			cacheimpls.ExpressionCache.Set(key, 1, 0)
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(key))

			err := r.batchDelete([]int64{123})
			assert.NoError(GinkgoT(), err)

			assert.False(GinkgoT(), cacheimpls.ExpressionCache.Exists(key))
		})

	})

	Describe("batchDeleteExpressionsFromRedis", func() {
		var r *redisRetriever
		BeforeEach(func() {
			cacheimpls.ExpressionCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}
		})

		It("empty pks", func() {
			err := batchDeleteExpressionsFromRedis(map[int64][]int64{})
			assert.NoError(GinkgoT(), err)
		})

		It("all missing", func() {
			err := batchDeleteExpressionsFromRedis(map[int64][]int64{
				1: {123, 456},
			})
			assert.NoError(GinkgoT(), err)
		})

		It("ok", func() {
			expression := map[int64]types.AuthExpression{
				123: {},
				456: {},
			}

			err := r.batchSet(expression)
			assert.NoError(GinkgoT(), err)

			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))

			// delete
			err = batchDeleteExpressionsFromRedis(map[int64][]int64{
				1: {123},
			})
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.ExpressionCache.Exists(cache.NewInt64Key(456)))
		})

	})

})
