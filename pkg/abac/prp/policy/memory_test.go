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

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/agiledragon/gomonkey/v2"
	rds "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	gocache "github.com/wklken/go-cache"

	"iam/pkg/abac/prp/common"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("Memory", func() {
	It("newMemoryRetriever", func() {
		r := newMemoryRetriever("test", 1, nil)
		assert.NotNil(GinkgoT(), r)
	})

	It("genKey", func() {
		r := newMemoryRetriever("test", 1, nil)

		key := r.genKey("123")
		assert.Equal(GinkgoT(), "test:1:123", key)
	})

	Describe("retrieve", func() {
		var r *memoryRetriever

		var patches *gomonkey.Patches
		var retrievedPolicies []types.AuthPolicy
		var cached1, cached2, cached3 *cachedPolicy
		var hitPolicies map[string]*cachedPolicy
		now := time.Now().Unix()

		BeforeEach(func() {
			patches = gomonkey.NewPatches()

			// init cache
			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)
			cacheimpls.LocalPolicyCache = gocache.New(1*time.Minute, 1*time.Minute)

			r = newMemoryRetriever("test", 1, nil)

			now = time.Now().Unix()
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
			cached1 = &cachedPolicy{
				timestamp: now,
				policies: []types.AuthPolicy{
					{
						PK:        1,
						SubjectPK: 123,
						ExpiredAt: now + 300,
					},
				},
			}
			cached2 = &cachedPolicy{
				timestamp: now,
				policies: []types.AuthPolicy{
					{
						PK:        2,
						SubjectPK: 456,
						ExpiredAt: now + 300,
					},
				},
			}
			cached3 = &cachedPolicy{
				timestamp: now,
				policies: []types.AuthPolicy{
					{
						PK:        3,
						SubjectPK: 789,
						ExpiredAt: now + 300,
					},
				},
			}

			hitPolicies = map[string]*cachedPolicy{
				"test:1:123": cached1,
				"test:1:456": cached2,
				"test:1:789": cached3,
			}
		})
		AfterEach(func() {
			// ctl.Finish()
			patches.Reset()
		})

		It("batchFetchSubjectPolicyChangedList fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return nil, errors.New("ZRevRangeByScore fail")
				})

			policies, missingSubjectPKs, err := r.retrieve([]int64{})
			assert.NoError(GinkgoT(), err)
			assert.Empty(GinkgoT(), policies)
			assert.Nil(GinkgoT(), missingSubjectPKs)
		})
		It("all missing, no changed list", func() {
			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthPolicy, missingPKs []int64, err error) {
				return retrievedPolicies, []int64{1000}, nil
			}

			policies, missingSubjectPKs, err := r.retrieve([]int64{123, 456, 789, 1000})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 3)
			assert.Equal(GinkgoT(), int64(1), policies[0].PK)
			assert.Equal(GinkgoT(), int64(2), policies[1].PK)
			assert.Equal(GinkgoT(), int64(3), policies[2].PK)
			assert.Equal(GinkgoT(), int64(789), policies[2].SubjectPK)
			assert.Len(GinkgoT(), missingSubjectPKs, 1)
			assert.Equal(GinkgoT(), int64(1000), missingSubjectPKs[0])

			// check the cache
			_, ok := cacheimpls.LocalPolicyCache.Get("test:1:123")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalPolicyCache.Get("test:1:456")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalPolicyCache.Get("test:1:789")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalPolicyCache.Get("test:1:1000")
			assert.True(GinkgoT(), ok)
		})
		It("all hit, no change list", func() {
			for key, cached := range hitPolicies {
				cacheimpls.LocalPolicyCache.Set(key, cached, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (policies []types.AuthPolicy, missingPKs []int64, err error) {
				return nil, nil, errors.New("should not be called")
			}

			policies, missingSubjectPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 3)
			assert.Empty(GinkgoT(), missingSubjectPKs)
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

			for key, cached := range hitPolicies {
				cacheimpls.LocalPolicyCache.Set(key, cached, 0)
			}

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthPolicy, missingPKs []int64, err error) {
				assert.Len(GinkgoT(), pks, 3)
				return retrievedPolicies, nil, nil
			}

			policies, missingSubjectPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 3)
			assert.Empty(GinkgoT(), missingSubjectPKs)
		})
		It("one policy cast fail", func() {
			// all hit
			for key, cached := range hitPolicies {
				cacheimpls.LocalPolicyCache.Set(key, cached, 0)
			}
			cacheimpls.LocalPolicyCache.Set("test:1:456", "abc", 0)

			r.missingRetrieveFunc = func(pks []int64) (expressions []types.AuthPolicy, missingPKs []int64, err error) {
				return nil, nil, errors.New("should be called")
			}

			policies, missingSubjectPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "should be called", err.Error())
			assert.Nil(GinkgoT(), policies)
			assert.Nil(GinkgoT(), missingSubjectPKs)
		})
		It("one policies is empty", func() {
			// all hit
			for key, cached := range hitPolicies {
				cacheimpls.LocalPolicyCache.Set(key, cached, 0)
			}
			cacheimpls.LocalPolicyCache.Set("test:1:456", &cachedPolicy{
				timestamp: now + 100,
				policies:  []types.AuthPolicy{},
			}, 0)

			policies, missingSubjectPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 2)
			assert.Empty(GinkgoT(), missingSubjectPKs)
		})
		It("retrieve fail", func() {
			r.missingRetrieveFunc = func(pks []int64) (policies []types.AuthPolicy, missingPKs []int64, err error) {
				return nil, nil, errors.New("retrieve fail")
			}
			policies, missingSubjectPKs, err := r.retrieve([]int64{123, 456, 789})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "retrieve fail", err.Error())
			assert.Nil(GinkgoT(), policies)
			assert.Nil(GinkgoT(), missingSubjectPKs)
		})
	})

	Describe("setMissing", func() {
		var r *memoryRetriever
		BeforeEach(func() {
			r = newMemoryRetriever("test", 1, nil)
			cacheimpls.LocalPolicyCache = gocache.New(1*time.Minute, 1*time.Minute)
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

			// check the cache
			_, ok := cacheimpls.LocalPolicyCache.Get("test:1:123")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalPolicyCache.Get("test:1:456")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalPolicyCache.Get("test:1:789")
			assert.True(GinkgoT(), ok)
			_, ok = cacheimpls.LocalPolicyCache.Get("111")
			assert.False(GinkgoT(), ok)
		})
	})

	It("deleteSystemSubjectPKsFromMemory", func() {
		// only check empty
		err := deleteSystemSubjectPKsFromMemory("test", []int64{})
		assert.NoError(GinkgoT(), err)
	})

	Describe("batchDeleteSystemSubjectPKsFromMemory", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			cacheimpls.LocalPolicyCache = gocache.New(1*time.Minute, 1*time.Minute)
			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)

			ctl = gomock.NewController(GinkgoT())
			patches = gomonkey.NewPatches()
		})

		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("empty", func() {
			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem("test").Return(
				[]types.ThinAction{}, nil,
			).AnyTimes()
			patches.ApplyFunc(service.NewActionService, func() service.ActionService {
				return mockActionService
			})

			err := batchDeleteSystemSubjectPKsFromMemory([]string{}, []int64{})
			assert.NoError(GinkgoT(), err)
		})

		It("ok", func() {
			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem(gomock.Any()).Return(
				[]types.ThinAction{{
					PK:     1,
					System: "test",
					ID:     "t1",
				}}, nil,
			).AnyTimes()
			patches.ApplyFunc(service.NewActionService, func() service.ActionService {
				return mockActionService
			})

			// init local cache
			cacheimpls.LocalPolicyCache.Set("test:1:123", "abc", 0)
			_, ok := cacheimpls.LocalPolicyCache.Get("test:1:123")
			assert.True(GinkgoT(), ok)

			max := time.Now().Unix()
			min := max - policyLocalCacheTTL

			// init redis cache
			err := cacheimpls.ChangeListCache.BatchZAdd([]redis.ZData{
				{
					Key: "policy:test:1",
					Zs: []*rds.Z{
						{
							Score:  float64(min + 50),
							Member: "0000",
						},
					},
				},
				{
					Key: "policy:test1:1",
					Zs: []*rds.Z{
						{
							Score:  float64(min - 60), // will be removed
							Member: "1111",
						},
					},
				},
			})
			assert.NoError(GinkgoT(), err)

			// do delete
			err = batchDeleteSystemSubjectPKsFromMemory([]string{"test", "test2"}, []int64{123, 456})
			assert.NoError(GinkgoT(), err)

			// check the local cache
			_, ok = cacheimpls.LocalPolicyCache.Get("test:1:123")
			// NOTE: Can't delete the key from the local cache, while no actionPK
			//       so here is true
			assert.True(GinkgoT(), ok)

			// check the change list
			assert.True(GinkgoT(), cacheimpls.ChangeListCache.Exists(cache.NewStringKey("policy:test:1")))

			// _type=policy, system=test, actionPK=1
			zs, err := cacheimpls.ChangeListCache.ZRevRangeByScore("policy:test:1", min, max, 0, maxChangeListCount)
			assert.NoError(GinkgoT(), err)
			// members: 0000 + 123 + 456  = 3
			assert.Len(GinkgoT(), zs, 3)

			// _type=policy, system=test2, actionPK=1
			assert.True(GinkgoT(), cacheimpls.ChangeListCache.Exists(cache.NewStringKey("policy:test2:1")))
			zs, err = cacheimpls.ChangeListCache.ZRevRangeByScore("policy:test2:1", min, max, 0, maxChangeListCount)
			assert.NoError(GinkgoT(), err)
			//  members: 1111 + 123 + 456 - 1111 => 123 + 456 = 2
			assert.Len(GinkgoT(), zs, 2)
		})

		It("addToChangeList fail", func() {
			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem(gomock.Any()).Return(
				[]types.ThinAction{{
					PK:     1,
					System: "test",
					ID:     "t1",
				}}, nil,
			).AnyTimes()
			patches.ApplyFunc(service.NewActionService, func() service.ActionService {
				return mockActionService
			})

			patches.ApplyMethod(reflect.TypeOf(changeList), "AddToChangeList",
				func(*common.ChangeList, map[string][]string) error {
					return errors.New("addToChangeList fail")
				})

			err := batchDeleteSystemSubjectPKsFromMemory([]string{"test"}, []int64{123, 456})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "addToChangeList fail", err.Error())
		})

		It("truncate fail", func() {
			mockActionService := mock.NewMockActionService(ctl)
			mockActionService.EXPECT().ListThinActionBySystem(gomock.Any()).Return(
				[]types.ThinAction{{
					PK:     1,
					System: "test",
					ID:     "t1",
				}}, nil,
			).AnyTimes()
			patches.ApplyFunc(service.NewActionService, func() service.ActionService {
				return mockActionService
			})

			patches.ApplyMethod(reflect.TypeOf(changeList), "Truncate",
				func(*common.ChangeList, []string) error {
					return errors.New("truncate fail")
				})

			err := batchDeleteSystemSubjectPKsFromMemory([]string{"test"}, []int64{123, 456})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "truncate fail", err.Error())
		})
	})
})
