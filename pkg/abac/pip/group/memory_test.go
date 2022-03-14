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
	. "github.com/onsi/ginkgo/v2"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("Memory", func() {
	It("newMemoryRetriever", func() {
		r := newMemoryRetriever(types.DepartmentType, nil)
		assert.NotNil(GinkgoT(), r)
	})

	It("genKey", func() {
		r := newMemoryRetriever(types.DepartmentType, nil)
		assert.Equal(GinkgoT(), "111", r.genKey(111))
	})

	Describe("retrieve", func() {
		var r *memoryRetriever

		// var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var retrievedSubjectGroups map[int64][]types.ThinSubjectGroup
		var cached1, cached2, cached3 *cachedSubjectGroups
		var now int64
		var hitSubjectGroups map[string]*cachedSubjectGroups

		BeforeEach(func() {
			// ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()

			// init cache
			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)
			cacheimpls.LocalSubjectGroupsCache = gocache.New(1*time.Minute, 1*time.Minute)

			r = newMemoryRetriever(types.DepartmentType, nil)

			now = time.Now().Unix()
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
			cached1 = &cachedSubjectGroups{
				timestamp: now,
				subjectGroups: []types.ThinSubjectGroup{
					{PK: 1, PolicyExpiredAt: 4102444800},
				},
			}
			cached2 = &cachedSubjectGroups{
				timestamp: now,
				subjectGroups: []types.ThinSubjectGroup{
					{PK: 1, PolicyExpiredAt: 4102444800},
				},
			}
			cached3 = &cachedSubjectGroups{
				timestamp: now,
				subjectGroups: []types.ThinSubjectGroup{
					{PK: 1, PolicyExpiredAt: 4102444800},
				},
			}

			hitSubjectGroups = map[string]*cachedSubjectGroups{
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

			subjectGroups, missingPKs, err := r.retrieve([]int64{})
			assert.NoError(GinkgoT(), err)
			assert.Empty(GinkgoT(), subjectGroups)
			assert.Nil(GinkgoT(), missingPKs)
		})
		It("all missing, no changed list", func() {
			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				return retrievedSubjectGroups, []int64{1000}, nil
			}

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789, 1000})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Contains(GinkgoT(), expressions, int64(123))
			assert.Contains(GinkgoT(), expressions, int64(456))
			assert.Contains(GinkgoT(), expressions, int64(789))
			assert.Len(GinkgoT(), missingPKs, 1)
			assert.Equal(GinkgoT(), int64(1000), missingPKs[0])

			// check the cache
			_, ok := cacheimpls.LocalSubjectGroupsCache.Get("123")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalSubjectGroupsCache.Get("456")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalSubjectGroupsCache.Get("789")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalSubjectGroupsCache.Get("1000")
			assert.True(GinkgoT(), ok)
		})
		It("all hit, no change list", func() {
			for key, cached := range hitSubjectGroups {
				cacheimpls.LocalSubjectGroupsCache.Set(key, cached, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
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

			for key, cached := range hitSubjectGroups {
				cacheimpls.LocalSubjectGroupsCache.Set(key, cached, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				assert.Len(GinkgoT(), pks, 3)
				return retrievedSubjectGroups, nil, nil
			}

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expressions, 3)
			assert.Empty(GinkgoT(), missingPKs)
		})
		It("one subjectGroup cast fail", func() {
			// all hit
			for key, cached := range hitSubjectGroups {
				cacheimpls.LocalSubjectGroupsCache.Set(key, cached, 0)
			}
			cacheimpls.LocalSubjectGroupsCache.Set("456", "abc", 0)

			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
				return nil, nil, errors.New("should be called")
			}

			expressions, missingPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "should be called", err.Error())
			assert.Nil(GinkgoT(), expressions)
			assert.Nil(GinkgoT(), missingPKs)
		})
		It("retrieve fail", func() {
			r.missingRetrieveFunc = func(pks []int64) (subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64, err error) {
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
			r = newMemoryRetriever(types.DepartmentType, nil)
			cacheimpls.LocalSubjectGroupsCache = gocache.New(1*time.Minute, 1*time.Minute)
		})

		It("ok", func() {
			err := r.setMissing(map[int64][]types.ThinSubjectGroup{
				123: {},
				456: {},
			}, []int64{789})

			assert.NoError(GinkgoT(), err)

			// check the cache
			_, ok := cacheimpls.LocalSubjectGroupsCache.Get("123")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalSubjectGroupsCache.Get("456")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalSubjectGroupsCache.Get("789")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalSubjectGroupsCache.Get("111")
			assert.False(GinkgoT(), ok)
		})
	})
})
