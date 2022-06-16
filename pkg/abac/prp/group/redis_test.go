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
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/prp/group/mock"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("Redis", func() {
	It("newRedisRetriever", func() {
		r := NewGroupAuthTypeRedisRetriever("test", nil)
		assert.NotNil(GinkgoT(), r)
	})

	It("genKey", func() {
		r := &groupAuthTypeRedisRetriever{
			keyPrefix: "test:",
		}
		k := r.genKey(456)

		assert.Equal(GinkgoT(), "test:456", k.Key())
	})

	Describe("parseKey", func() {
		var r *groupAuthTypeRedisRetriever
		BeforeEach(func() {
			r = &groupAuthTypeRedisRetriever{
				keyPrefix: "test:",
			}
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

	Describe("batchSetGroupAuthTypeCache", func() {
		var r *groupAuthTypeRedisRetriever
		BeforeEach(func() {
			cacheimpls.GroupSystemAuthTypeCache = redis.NewMockCache("test", 5*time.Minute)
			r = &groupAuthTypeRedisRetriever{
				keyPrefix: "test:",
			}
		})

		It("ok", func() {
			err := r.batchSetGroupAuthTypeCache([]types.GroupAuthType{{
				GroupPK:  1,
				AuthType: 1,
			}, {
				GroupPK:  2,
				AuthType: 2,
			}})

			assert.NoError(GinkgoT(), err)

			var value string
			err = cacheimpls.GroupSystemAuthTypeCache.Get(cache.NewStringKey("test:1"), &value)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "1", value)

			cacheimpls.GroupSystemAuthTypeCache.Get(cache.NewStringKey("test:2"), &value)
			assert.Equal(GinkgoT(), "2", value)
		})
	})

	Describe("batchGetGroupAuthType", func() {
		var r *groupAuthTypeRedisRetriever
		BeforeEach(func() {
			cacheimpls.GroupSystemAuthTypeCache = redis.NewMockCache("test", 5*time.Minute)
			r = &groupAuthTypeRedisRetriever{
				keyPrefix: "test:",
			}
		})

		It("ok", func() {
			err := r.batchSetGroupAuthTypeCache([]types.GroupAuthType{{
				GroupPK:  1,
				AuthType: 1,
			}, {
				GroupPK:  2,
				AuthType: 2,
			}})

			assert.NoError(GinkgoT(), err)

			groupAuthTypes, missingPKs, err := r.batchGetGroupAuthType([]int64{1, 2, 3})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{GroupPK: 1, AuthType: 1}, {GroupPK: 2, AuthType: 2},
			}, groupAuthTypes)
			assert.Equal(GinkgoT(), []int64{3}, missingPKs)
		})
	})

	Describe("Retrieve", func() {
		var ctl *gomock.Controller
		var r *groupAuthTypeRedisRetriever
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			cacheimpls.GroupSystemAuthTypeCache = redis.NewMockCache("test", 5*time.Minute)
			r = &groupAuthTypeRedisRetriever{
				keyPrefix: "test:",
			}
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("no miss ok", func() {
			err := r.batchSetGroupAuthTypeCache([]types.GroupAuthType{{
				GroupPK:  1,
				AuthType: 1,
			}, {
				GroupPK:  2,
				AuthType: 2,
			}})

			assert.NoError(GinkgoT(), err)

			groupAuthTypes, err := r.Retrieve([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), groupAuthTypes, 2)
		})

		It("missingRetriever fail", func() {
			mockRetriever := mock.NewMockGroupAuthTypeRetriever(ctl)
			mockRetriever.EXPECT().Retrieve([]int64{3}).Return(nil, errors.New("missingRetriever error"))
			r.missingRetriever = mockRetriever

			err := r.batchSetGroupAuthTypeCache([]types.GroupAuthType{{
				GroupPK:  1,
				AuthType: 1,
			}, {
				GroupPK:  2,
				AuthType: 2,
			}})

			assert.NoError(GinkgoT(), err)

			_, err = r.Retrieve([]int64{1, 2, 3})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "missingRetriever")
		})

		It("missingRetriever ok", func() {
			mockRetriever := mock.NewMockGroupAuthTypeRetriever(ctl)
			mockRetriever.EXPECT().Retrieve([]int64{3}).Return([]types.GroupAuthType{{
				GroupPK:  3,
				AuthType: 3,
			}}, nil)
			r.missingRetriever = mockRetriever

			err := r.batchSetGroupAuthTypeCache([]types.GroupAuthType{{
				GroupPK:  1,
				AuthType: 1,
			}, {
				GroupPK:  2,
				AuthType: 2,
			}})

			assert.NoError(GinkgoT(), err)

			groupAuthTypes, err := r.Retrieve([]int64{1, 2, 3})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{GroupPK: 1, AuthType: 1}, {GroupPK: 2, AuthType: 2}, {GroupPK: 3, AuthType: 3},
			}, groupAuthTypes)
		})
	})

	Describe("deleteRedisGroupAuthTypeCache", func() {
		var r *groupAuthTypeRedisRetriever
		BeforeEach(func() {
			cacheimpls.GroupSystemAuthTypeCache = redis.NewMockCache("test", 5*time.Minute)
			r = &groupAuthTypeRedisRetriever{
				keyPrefix: "test:",
			}
		})

		It("ok", func() {
			err := r.batchSetGroupAuthTypeCache([]types.GroupAuthType{{
				GroupPK:  1,
				AuthType: 1,
			}, {
				GroupPK:  2,
				AuthType: 2,
			}})

			assert.NoError(GinkgoT(), err)

			err = deleteRedisGroupAuthTypeCache("test", 2)
			assert.NoError(GinkgoT(), err)

			var value string
			err = cacheimpls.GroupSystemAuthTypeCache.Get(cache.NewStringKey("test:1"), &value)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "1", value)

			err = cacheimpls.GroupSystemAuthTypeCache.Get(cache.NewStringKey("test:2"), &value)
			assert.Error(GinkgoT(), err)
		})
	})
})
