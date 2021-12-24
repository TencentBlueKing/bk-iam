/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"errors"
	"reflect"
	"time"

	"github.com/agiledragon/gomonkey"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("Redis", func() {
	It("newRedisRetriever", func() {
		r := newRedisRetriever("test", 1, nil)
		assert.NotNil(GinkgoT(), r)
	})

	It("genKey", func() {
		r := newRedisRetriever("test", 1, nil)
		k := r.genKey(456)

		assert.Equal(GinkgoT(), "test:456", k.Key())
	})

	Describe("parseKey", func() {
		var r *redisRetriever
		BeforeEach(func() {
			r = newRedisRetriever("test", 1, nil)
		})

		It("fail", func() {
			subjectPK, err := r.parseKey("test:abc")
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(-1), subjectPK)
		})

		It("ok", func() {
			subjectPK, err := r.parseKey("test:123")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(123), subjectPK)
		})
	})

	Describe("retrieve", func() {
		var r *redisRetriever
		var patches *gomonkey.Patches
		var retrievedPolicies []types.AuthPolicy
		var hitPolicies map[int64]string
		var policyStr1, policyStr2, policyStr3 []byte
		var emptyPolicyStr []byte
		now := time.Now().Unix()
		BeforeEach(func() {
			patches = gomonkey.NewPatches()

			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)
			r = newRedisRetriever("test", 1, nil)

			retrievedPolicies = []types.AuthPolicy{
				{
					PK:        1,
					SubjectPK: 123,
					ExpiredAt: now + 300,
				},
				{
					PK:        2,
					SubjectPK: 456,
					ExpiredAt: now + 300,
				},
				{
					PK:        3,
					SubjectPK: 789,
					ExpiredAt: now + 300,
				},
			}

			r.missingRetrieveFunc = func(pks []int64) (policies []types.AuthPolicy, missingSubjectPKs []int64, err error) {
				return retrievedPolicies, []int64{}, nil
			}

			policyStr1, _ = cacheimpls.PolicyCache.Marshal([]types.AuthPolicy{
				{
					PK:        1,
					SubjectPK: 123,
				},
			})
			policyStr2, _ = cacheimpls.PolicyCache.Marshal([]types.AuthPolicy{
				{
					PK:        2,
					SubjectPK: 456,
				},
			})
			policyStr3, _ = cacheimpls.PolicyCache.Marshal([]types.AuthPolicy{
				{
					PK:        3,
					SubjectPK: 789,
				},
			})
			emptyPolicyStr, _ = cacheimpls.PolicyCache.Marshal([]types.AuthPolicy{})
			hitPolicies = map[int64]string{
				123: string(policyStr1),
				456: string(policyStr2),
				789: string(policyStr3),
			}

			_ = hitPolicies
			_ = emptyPolicyStr

		})
		AfterEach(func() {
			patches.Reset()
		})

		It("batchGet fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.PolicyCache), "BatchHGet",
				func(c *redis.Cache, hashKeyFields []redis.HashKeyField) (map[redis.HashKeyField]string, error) {
					return nil, errors.New("batchHGet fail")
				})

			// empty pks, will not do retrieve
			policies, missingSubjectPKs, err := r.retrieve([]int64{})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 0)
			assert.Len(GinkgoT(), missingSubjectPKs, 0)
		})

		It("all missing", func() {
			subjectPKs := []int64{123, 456, 789}

			// patches.ApplyFunc(r.batchGet,
			//	func(subjectPKs []int64) (
			//		hitPolicies map[int64]string,
			//		missSubjectPKs []int64,
			//		err error,
			//	) {
			//		return nil, subjectPKs, nil
			//	})

			policies, missingPKs, err := r.retrieve(subjectPKs)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 3)
			assert.Equal(GinkgoT(), int64(1), policies[0].PK)
			assert.Equal(GinkgoT(), int64(2), policies[1].PK)
			assert.Equal(GinkgoT(), int64(3), policies[2].PK)
			assert.Equal(GinkgoT(), int64(789), policies[2].SubjectPK)
			assert.Empty(GinkgoT(), missingPKs)
		})

		It("all hit", func() {
			subjectPKs := []int64{123, 456, 789}

			r.setMissing(retrievedPolicies, []int64{})
			r.missingRetrieveFunc = func(pks []int64) (policies []types.AuthPolicy, missingSubjectPKs []int64, err error) {
				return nil, nil, errors.New("should not be called")
			}

			policies, missingPKs, err := r.retrieve(subjectPKs)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 3)
			assert.Empty(GinkgoT(), missingPKs)

			// check the cache
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(123)))
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(456)))
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(789)))
		})

		It("one policy unmarshal fail", func() {
			r.setMissing(retrievedPolicies, []int64{})
			cacheimpls.PolicyCache.BatchHSetWithTx([]redis.Hash{
				{
					HashKeyField: redis.HashKeyField{
						Key:   r.keyPrefix + "1000",
						Field: "1",
					},
					Value: "not a valid json",
				},
			})

			subjectPKs := []int64{123, 456, 789, 1000}
			r.missingRetrieveFunc = func(pks []int64) (policies []types.AuthPolicy, missingSubjectPKs []int64, err error) {
				return nil, nil, errors.New("should do retrieve 1000")
			}
			_, _, err := r.retrieve(subjectPKs)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "should do retrieve 1000", err.Error())

		})

		It("empty policy", func() {
			r.setMissing(retrievedPolicies, []int64{})
			cacheimpls.PolicyCache.BatchHSetWithTx([]redis.Hash{
				{
					HashKeyField: redis.HashKeyField{
						Key:   r.keyPrefix + "1000",
						Field: "1",
					},
					Value: string(emptyPolicyStr),
				},
			})

			subjectPKs := []int64{123, 456, 789, 1000}
			r.missingRetrieveFunc = func(pks []int64) (policies []types.AuthPolicy, missingSubjectPKs []int64, err error) {
				return nil, nil, errors.New("should not be called")
			}
			policies, missingSubjectPKs, err := r.retrieve(subjectPKs)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 3)
			assert.Len(GinkgoT(), missingSubjectPKs, 1)
			assert.Equal(GinkgoT(), int64(1000), missingSubjectPKs[0])

		})

		It("retrieve fail", func() {
			subjectPKs := []int64{123, 456, 789}

			r.missingRetrieveFunc = func(pks []int64) (policies []types.AuthPolicy, missingSubjectPKs []int64, err error) {
				return nil, nil, errors.New("retrieve fail")
			}

			_, _, err := r.retrieve(subjectPKs)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "retrieve fail", err.Error())
		})

	})
	Describe("setMissing", func() {
		var r *redisRetriever
		BeforeEach(func() {
			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)
			r = newRedisRetriever("test", 1, nil)
		})

		It("ok", func() {
			err := r.setMissing([]types.AuthPolicy{
				{
					PK:        1,
					SubjectPK: 123,
				},
				{
					PK:        2,
					SubjectPK: 456,
				},
				{
					PK:        3,
					SubjectPK: 123,
				},
			}, []int64{789})
			assert.NoError(GinkgoT(), err)

			// all key exists
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(123)))
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(456)))
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(789)))
		})
	})

	Describe("batchGet", func() {
		var r *redisRetriever
		var patches *gomonkey.Patches
		BeforeEach(func() {
			patches = gomonkey.NewPatches()
			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)
			r = newRedisRetriever("test", 1, nil)
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("cache BatchHGet fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.PolicyCache), "BatchHGet",
				func(c *redis.Cache, hashKeyFields []redis.HashKeyField) (map[redis.HashKeyField]string, error) {
					return nil, errors.New("batchHget fail")
				})
			_, _, err := r.batchGet([]int64{123, 456})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchHget fail", err.Error())

		})

		It("all empty", func() {
			hitPolicies, missSubjectPKs, err := r.batchGet([]int64{123, 456})
			assert.NoError(GinkgoT(), err)
			assert.Empty(GinkgoT(), hitPolicies)
			assert.Len(GinkgoT(), missSubjectPKs, 2)
		})

		It("hit one", func() {
			// set one
			subjectPKPolicies := map[int64][]types.AuthPolicy{
				123: {},
			}

			err := r.batchSet(subjectPKPolicies)
			assert.NoError(GinkgoT(), err)

			// get again
			hitPolicies, missSubjectPKs, err := r.batchGet([]int64{123, 456})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), hitPolicies, 1)
			assert.Len(GinkgoT(), missSubjectPKs, 1)
			assert.Equal(GinkgoT(), int64(456), missSubjectPKs[0])
		})

		It("hit all", func() {
			// set all
			subjectPKPolicies := map[int64][]types.AuthPolicy{
				123: {},
				456: {},
			}
			err := r.batchSet(subjectPKPolicies)
			assert.NoError(GinkgoT(), err)

			// get again
			hitPolicies, missSubjectPKs, err := r.batchGet([]int64{123, 456})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), hitPolicies, 2)
			assert.Len(GinkgoT(), missSubjectPKs, 0)
		})

	})
	Describe("batchSet", func() {
		var r *redisRetriever
		var patches *gomonkey.Patches
		var subjectPKPolicies map[int64][]types.AuthPolicy
		BeforeEach(func() {
			patches = gomonkey.NewPatches()

			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)
			r = newRedisRetriever("test", 1, nil)
			subjectPKPolicies = map[int64][]types.AuthPolicy{
				123: {},
				456: {},
			}
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("ok", func() {
			err := r.batchSet(subjectPKPolicies)
			assert.NoError(GinkgoT(), err)

			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(123)))
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(r.genKey(456)))
		})

		It("marshal fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.PolicyCache), "Marshal",
				func(c *redis.Cache, value interface{}) ([]byte, error) {
					return nil, errors.New("marshal fail")
				})

			err := r.batchSet(subjectPKPolicies)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "marshal fail", err.Error())
		})

		It("cache BatchHSetWithTx fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.PolicyCache), "BatchHSetWithTx",
				func(c *redis.Cache, hashes []redis.Hash) error {
					return errors.New("batchHSetWithTx fail")
				})
			defer patches.Reset()

			err := r.batchSet(subjectPKPolicies)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchHSetWithTx fail", err.Error())
		})

		It("cache BatchExpireWithTx fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.PolicyCache), "BatchExpireWithTx",
				func(c *redis.Cache, keys []cache.Key, expiration time.Duration) error {
					return errors.New("batchExpireWithTx fail")
				})
			defer patches.Reset()

			err := r.batchSet(subjectPKPolicies)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchExpireWithTx fail", err.Error())
		})
	})

	Describe("batchDelete", func() {
		var r *redisRetriever
		var patches *gomonkey.Patches
		BeforeEach(func() {
			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)
			r = newRedisRetriever("test", 1, nil)

			patches = gomonkey.NewPatches()
		})

		AfterEach(func() {
			patches.Reset()
		})

		It("not exists, delete ok", func() {
			err := r.batchDelete([]int64{123})
			assert.NoError(GinkgoT(), err)
		})

		It("ok", func() {
			key := cache.NewStringKey("test:123")
			cacheimpls.PolicyCache.Set(key, 1, 0)
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(key))

			err := r.batchDelete([]int64{123})
			assert.NoError(GinkgoT(), err)

			assert.False(GinkgoT(), cacheimpls.PolicyCache.Exists(key))
		})

		It("batchDelete fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.PolicyCache), "BatchDelete",
				func(c *redis.Cache, keys []cache.Key) error {
					return errors.New("batchDelete fail")
				})

			err := r.batchDelete([]int64{123})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchDelete fail", err.Error())
		})

	})

	Describe("deleteSystemSubjectPKsFromRedis", func() {
		It("ok", func() {
			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)

			r := newRedisRetriever("test", 1, nil)

			key := r.genKey(123)
			cacheimpls.PolicyCache.Set(key, "", 0)
			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(cache.NewStringKey("test:123")))

			err := deleteSystemSubjectPKsFromRedis("test", []int64{123, 456})
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), cacheimpls.PolicyCache.Exists(cache.NewStringKey("test:123")))
		})
		It("empty pks, but ok", func() {
			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)

			err := deleteSystemSubjectPKsFromRedis("test", []int64{})
			assert.NoError(GinkgoT(), err)
		})

	})

	Describe("batchDeleteSystemSubjectPKsFromRedis", func() {
		var patches *gomonkey.Patches
		BeforeEach(func() {
			cacheimpls.PolicyCache = redis.NewMockCache("test", 5*time.Minute)
			patches = gomonkey.NewPatches()
		})

		AfterEach(func() {
			patches.Reset()
		})

		It("ok", func() {
			r := newRedisRetriever("test", 1, nil)

			key := r.genKey(123)
			cacheimpls.PolicyCache.Set(key, "", 0)

			assert.True(GinkgoT(), cacheimpls.PolicyCache.Exists(cache.NewStringKey("test:123")))

			err := batchDeleteSystemSubjectPKsFromRedis([]string{"test1", "test"}, []int64{123, 456})
			assert.NoError(GinkgoT(), err)
			assert.False(GinkgoT(), cacheimpls.PolicyCache.Exists(cache.NewStringKey("test:123")))
		})

		It("empty pks, but ok", func() {
			err := batchDeleteSystemSubjectPKsFromRedis([]string{}, []int64{123, 456})
			assert.NoError(GinkgoT(), err)
		})

		It("batchDelete fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.PolicyCache), "BatchDelete",
				func(c *redis.Cache, keys []cache.Key) error {
					return errors.New("batchDelete fail")
				})
			err := batchDeleteSystemSubjectPKsFromRedis([]string{"test"}, []int64{123, 456})
			assert.Error(GinkgoT(), err)

		})

	})

})
