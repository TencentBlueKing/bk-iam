/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package temporary

import (
	"errors"
	"time"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"

	red "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	gocache "github.com/wklken/go-cache"
)

var _ = Describe("TemporaryPolicy", func() {
	It("genKey", func() {
		c := &temporaryPolicyRedisCache{
			keyPrefix:              "test" + ":",
			temporaryPolicyService: nil,
		}
		k := c.genKey(456)

		assert.Equal(GinkgoT(), "test:456", k.Key())
	})

	It("genHashKeyField", func() {
		c := &temporaryPolicyRedisCache{
			keyPrefix:              "test" + ":",
			temporaryPolicyService: nil,
		}
		hashKeyField := c.genHashKeyField(456, 789)

		assert.Equal(GinkgoT(), hashKeyField, redis.HashKeyField{
			Key:   "test:456",
			Field: "789",
		})
	})

	Describe("temporaryPolicyRedisCache", func() {
		var c *temporaryPolicyRedisCache
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			c = &temporaryPolicyRedisCache{
				keyPrefix:              "test" + ":",
				temporaryPolicyService: nil,
			}
			cacheimpls.TemporaryPolicyCache = redis.NewMockCache("test", 5*time.Minute)
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("get fail", func() {
			_, err := c.getThinTemporaryPoliciesFromCache(456, 789)
			assert.ErrorIs(GinkgoT(), err, red.Nil)
		})

		It("set and get ok", func() {
			ps := []types.ThinTemporaryPolicy{{}}
			c.setThinTemporaryPoliciesToCache(456, 789, ps)
			value, err := c.getThinTemporaryPoliciesFromCache(456, 789)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), ps, value)
		})

		It("ListThinBySubjectAction fail", func() {
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().ListThinBySubjectAction(
				int64(456), int64(789),
			).Return(
				nil, errors.New("list fail"),
			).AnyTimes()

			c.temporaryPolicyService = mockTemporaryPolicyService

			_, err := c.ListThinBySubjectAction(456, 789)
			assert.Contains(GinkgoT(), err.Error(), "fail")
		})

		It("ListThinBySubjectAction ok", func() {
			ps := []types.ThinTemporaryPolicy{{}}

			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(ctl)
			mockTemporaryPolicyService.EXPECT().ListThinBySubjectAction(
				int64(456), int64(789),
			).Return(
				ps, nil,
			).AnyTimes()

			c.temporaryPolicyService = mockTemporaryPolicyService

			value, err := c.ListThinBySubjectAction(456, 789)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), ps, value)
		})
	})

	Describe("temporaryPolicyLocalCache", func() {
		var c *temporaryPolicyLocalCache
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			c = &temporaryPolicyLocalCache{
				temporaryPolicyService: nil,
			}
			cacheimpls.LocalTemporaryPolicyCache = gocache.New(5*time.Minute, 5*time.Minute)
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("batchGet miss all", func() {
			_, missPKs := c.batchGet([]int64{1, 2})
			assert.Equal(GinkgoT(), missPKs, []int64{1, 2})
		})

		It("batchGet ok", func() {
			ps := []types.TemporaryPolicy{{PK: 1}, {PK: 2}}
			c.setMissing(ps)
			value, missPKs := c.batchGet([]int64{1, 2})
			assert.Equal(GinkgoT(), missPKs, []int64{})
			assert.Equal(GinkgoT(), value, ps)
		})

		It("ListByPKs fail", func() {
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(gomock.NewController(GinkgoT()))
			mockTemporaryPolicyService.EXPECT().ListByPKs(
				[]int64{1, 2},
			).Return(
				nil, errors.New("list fail"),
			).AnyTimes()

			c.temporaryPolicyService = mockTemporaryPolicyService

			_, err := c.ListByPKs([]int64{1, 2})
			assert.Contains(GinkgoT(), err.Error(), "fail")
		})

		It("ListByPKs ok", func() {
			ps := []types.TemporaryPolicy{{PK: 1}, {PK: 2}}
			mockTemporaryPolicyService := mock.NewMockTemporaryPolicyService(gomock.NewController(GinkgoT()))
			mockTemporaryPolicyService.EXPECT().ListByPKs(
				[]int64{1, 2},
			).Return(
				ps, nil,
			).AnyTimes()

			c.temporaryPolicyService = mockTemporaryPolicyService

			value, err := c.ListByPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), ps, value)
		})
	})
})
