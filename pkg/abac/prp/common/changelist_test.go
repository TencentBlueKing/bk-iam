/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common_test

import (
	"errors"
	"reflect"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	rds "github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/prp/common"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
)

var _ = Describe("Changelist", func() {
	It("NewChangeList", func() {
		a := common.NewChangeList("test", 60, 100)
		assert.NotNil(GinkgoT(), a)
	})

	Describe("FetchList", func() {
		var c *common.ChangeList
		var patches *gomonkey.Patches
		BeforeEach(func() {
			c = common.NewChangeList("test", 60, 100)

			patches = gomonkey.NewPatches()
			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("zrange fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return nil, errors.New("ZRevRangeByScore fail")
				})

			_, err := c.FetchList("abc")
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "ZRevRangeByScore fail", err.Error())
		})

		It("ok", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "ZRevRangeByScore",
				func(c *redis.Cache, k string, min int64, max int64, offset int64, count int64) ([]rds.Z, error) {
					return []rds.Z{
						{
							Score:  1000,
							Member: "123",
						},
						{
							Score:  2000,
							Member: "456",
						},
					}, nil
				})

			data, err := c.FetchList("abc")
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), data, 2)
			assert.Contains(GinkgoT(), data, "123")
			assert.Contains(GinkgoT(), data, "456")
			assert.Equal(GinkgoT(), int64(1000), data["123"])
			assert.Equal(GinkgoT(), int64(2000), data["456"])
		})
	})

	Describe("AddToChangeList", func() {
		var c *common.ChangeList
		var patches *gomonkey.Patches
		var keyMembers map[string][]string
		BeforeEach(func() {
			c = common.NewChangeList("test", 60, 100)

			patches = gomonkey.NewPatches()
			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)

			keyMembers = map[string][]string{
				"abc": {"10", "11"},
			}
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("BatchZAdd fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "BatchZAdd",
				func(c *redis.Cache, zDataList []redis.ZData) error {
					return errors.New("batchZAdd fail")
				})

			err := c.AddToChangeList(keyMembers)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "batchZAdd fail", err.Error())
		})

		It("ok", func() {
			err := c.AddToChangeList(keyMembers)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("Truncate", func() {
		var c *common.ChangeList
		var patches *gomonkey.Patches
		BeforeEach(func() {
			c = common.NewChangeList("test", 60, 100)

			patches = gomonkey.NewPatches()
			cacheimpls.ChangeListCache = redis.NewMockCache("test", 5*time.Minute)
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("ZRemove fail", func() {
			patches.ApplyMethod(reflect.TypeOf(cacheimpls.ChangeListCache), "BatchZRemove",
				func(c *redis.Cache, keys []string, min int64, max int64) error {
					return errors.New("zRemove fail")
				})

			err := c.Truncate([]string{"abc", "def"})
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), "zRemove fail", err.Error())
		})

		It("ok", func() {
			err := c.Truncate([]string{"abc", "def"})
			assert.NoError(GinkgoT(), err)
		})
	})
})
