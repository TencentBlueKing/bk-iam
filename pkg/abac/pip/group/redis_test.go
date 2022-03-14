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

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

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

		var retrievedSubjectGroups map[int64][]types.ThinSubjectGroup
		var hitSubjectGroups map[int64]string
		var str1, str2, str3 []byte
		// var emptyExprStr []byte
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()

			cacheimpls.SubjectGroupCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}

			retrievedSubjectGroups = map[int64][]types.ThinSubjectGroup{
				123: {
					{
						PK:              1,
						PolicyExpiredAt: 4102444800,
					},
				},
				456: {
					{
						PK:              4,
						PolicyExpiredAt: 4102444800,
					},
				},
				789: {
					{
						PK:              7,
						PolicyExpiredAt: 4102444800,
					},
				},
			}

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				return retrievedSubjectGroups, []int64{}, nil
			}

			str1, _ = cacheimpls.SubjectGroupCache.Marshal([]types.ThinSubjectGroup{
				{PK: 1, PolicyExpiredAt: 4102444800},
			})
			str2, _ = cacheimpls.SubjectGroupCache.Marshal([]types.ThinSubjectGroup{
				{PK: 4, PolicyExpiredAt: 4102444800},
			})
			str3, _ = cacheimpls.SubjectGroupCache.Marshal([]types.ThinSubjectGroup{
				{PK: 7, PolicyExpiredAt: 4102444800},
			})
			// emptyExprStr, _ = cacheimpls.SubjectGroupCache.Marshal(types.AuthExpression{})
			hitSubjectGroups = map[int64]string{
				123: string(str1),
				456: string(str2),
				789: string(str3),
			}
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("batchGet fail", func() {
			// patches.ApplyMethod(reflect.TypeOf(cacheimpls.SubjectGroupCache), "BatchGet",
			//	func(c *redis.Cache, keys []cache.Key) (map[cache.Key]string, error) {
			//		return nil, errors.New("batch get fail")
			//	})

			// empty pks, will not do retrieve
			subjectGroups, missingPKs, err := r.retrieve([]int64{})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectGroups, 0)
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
			subjectGroups, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectGroups, 3)
			assert.Equal(GinkgoT(), int64(1), subjectGroups[123][0].PK)
			assert.Equal(GinkgoT(), int64(4), subjectGroups[456][0].PK)
			assert.Equal(GinkgoT(), int64(7), subjectGroups[789][0].PK)
			assert.Empty(GinkgoT(), missingPKs)

			// check the cache
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(789)))
		})

		It("all hit", func() {
			pks := []int64{123, 456, 789}

			for pk, str := range hitSubjectGroups {
				cacheimpls.SubjectGroupCache.Set(cache.NewInt64Key(pk), str, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				return nil, nil, errors.New("should not be called")
			}

			// empty pks, will not do retrieve
			subjectGroups, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectGroups, 3)
			assert.Len(GinkgoT(), missingPKs, 0)
		})

		It("some hit, some missing", func() {
			pks := []int64{123, 456, 789}

			cacheimpls.SubjectGroupCache.Set(cache.NewInt64Key(int64(456)), string(str2), 0)

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				r := retrievedSubjectGroups
				delete(r, 456)
				return r, nil, nil
			}

			// empty pks, will not do retrieve
			subjectGroups, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectGroups, 3)
			assert.Nil(GinkgoT(), missingPKs)

			// check the cache
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(789)))
		})

		It("some hit, some missing, retrieve has missing key", func() {
			pks := []int64{123, 456, 789}

			cacheimpls.SubjectGroupCache.Set(cache.NewInt64Key(int64(456)), string(str2), 0)

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				r := retrievedSubjectGroups
				delete(r, 123)
				delete(r, 456)
				return r, []int64{123}, nil
			}

			// empty pks, will not do retrieve
			subjectGroups, missingPKs, err := r.retrieve(pks)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), subjectGroups, 3)
			assert.Len(GinkgoT(), missingPKs, 1)
			assert.Equal(GinkgoT(), int64(123), missingPKs[0])

			// check the cache
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(789)))
		})

		It("one expression unmarshal fail", func() {
			pks := []int64{123, 456, 789}

			for pk, str := range hitSubjectGroups {
				cacheimpls.SubjectGroupCache.Set(cache.NewInt64Key(pk), str, 0)
			}
			cacheimpls.SubjectGroupCache.Set(cache.NewInt64Key(int64(456)), "not a valid json", 0)

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				return nil, nil, errors.New("should do retrieve 456")
			}

			// empty pks, will not do retrieve
			expressions, missingPKs, err := r.retrieve(pks)
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), expressions)
			assert.Nil(GinkgoT(), missingPKs)
		})

		It("retrieve fail", func() {
			pks := []int64{123, 456, 789}

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
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
			cacheimpls.SubjectGroupCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}
		})

		It("ok", func() {
			err := r.setMissing(map[int64][]types.ThinSubjectGroup{
				123: {},
			}, []int64{456, 789})
			assert.NoError(GinkgoT(), err)

			// all key exists
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(456)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(789)))
		})
	})

	// helper functions
	Describe("batchGet", func() {
		var r *redisRetriever
		BeforeEach(func() {
			cacheimpls.SubjectGroupCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}
		})

		It("cache BatchGet fail", func() {
			patches := gomonkey.NewPatches()
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.SubjectGroupCache), "BatchGet",
				func(c *redis.Cache, keys []cache.Key) (map[cache.Key]string, error) {
					return nil, errors.New("batch get fail")
				})

			_, _, err := r.batchGet([]int64{123})
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			subjectGroups := map[int64][]types.ThinSubjectGroup{
				123: {
					{
						PK:              1,
						PolicyExpiredAt: 4102444800,
					},
				},
			}
			err := r.batchSet(subjectGroups)
			assert.NoError(GinkgoT(), err)

			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(123)))
			assert.False(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(456)))
		})
	})

	Describe("batchSet", func() {
		var r *redisRetriever
		var patches *gomonkey.Patches
		var noCachesubjectGroups map[int64][]types.ThinSubjectGroup
		BeforeEach(func() {
			patches = gomonkey.NewPatches()
			cacheimpls.SubjectGroupCache = redis.NewMockCache("test", 5*time.Minute)
			r = &redisRetriever{}

			noCachesubjectGroups = map[int64][]types.ThinSubjectGroup{
				123: {
					{
						PK:              1,
						PolicyExpiredAt: 4102444800,
					},
				},
				456: {
					{
						PK:              1,
						PolicyExpiredAt: 4102444800,
					},
				},
			}
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("cache Marshal fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.SubjectGroupCache), "Marshal",
				func(c *redis.Cache, value interface{}) ([]byte, error) {
					return nil, errors.New("marshal fail")
				})

			err := r.batchSet(noCachesubjectGroups)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "marshal fail", err.Error())
		})

		It("cache BatchSetWithTx fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.SubjectGroupCache), "BatchSetWithTx",
				func(c *redis.Cache, kvs []redis.KV, expiration time.Duration) error {
					return errors.New("batchSetWithTx fail")
				})
			defer patches.Reset()

			err := r.batchSet(noCachesubjectGroups)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchSetWithTx fail", err.Error())
		})

		It("ok", func() {
			err := r.batchSet(noCachesubjectGroups)
			assert.NoError(GinkgoT(), err)

			// all key exists
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(123)))
			assert.True(GinkgoT(), cacheimpls.SubjectGroupCache.Exists(cache.NewInt64Key(456)))
		})
	})
})
