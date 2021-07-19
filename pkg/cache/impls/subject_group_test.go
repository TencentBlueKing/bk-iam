/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package impls

import (
	"errors"
	"reflect"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache"
	"iam/pkg/cache/redis"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

var _ = Describe("SubjectGroups", func() {
	BeforeEach(func() {
		var (
			expiration = 5 * time.Minute
		)
		mockCache := redis.NewMockCache("mockCache", expiration)

		SubjectGroupCache = mockCache
	})

	It("GetSubjectGroups", func() {
		ctl := gomock.NewController(GinkgoT())
		defer ctl.Finish()

		mockService := mock.NewMockSubjectService(ctl)
		mockService.EXPECT().GetThinSubjectGroups(int64(1)).Return([]types.ThinSubjectGroup{
			{
				PK:              int64(1),
				PolicyExpiredAt: int64(100000),
			},
		}, nil).AnyTimes()

		patches := gomonkey.ApplyFunc(service.NewSubjectService,
			func() service.SubjectService {
				return mockService
			})
		defer patches.Reset()

		sg, err := GetSubjectGroups(int64(1))
		assert.NoError(GinkgoT(), err)
		assert.Len(GinkgoT(), sg, 1)
	})

	Context("ListSubjectEffectGroups", func() {
		It("batchGetSubjectGroups fail", func() {
			patches := gomonkey.ApplyFunc(batchGetSubjectGroups,
				func([]int64) ([]types.ThinSubjectGroup, []int64, error) {
					return nil, nil, errors.New("error")
				})
			defer patches.Reset()

			_, err := ListSubjectEffectGroups([]int64{1, 2, 3})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "fail")
		})

		It("batchGetSubjectGroups ok, all cached", func() {
			patches := gomonkey.ApplyFunc(batchGetSubjectGroups,
				func([]int64) ([]types.ThinSubjectGroup, []int64, error) {
					return []types.ThinSubjectGroup{
						{
							PK:              2,
							PolicyExpiredAt: 21,
						},
						{
							PK:              3,
							PolicyExpiredAt: 31,
						},
					}, []int64{}, nil
				})
			defer patches.Reset()

			sgs, err := ListSubjectEffectGroups([]int64{1, 2, 3})
			assert.NoError(GinkgoT(), err)
			assert.NotEmpty(GinkgoT(), sgs)
			assert.Len(GinkgoT(), sgs, 2)
		})

		Context("batchGetSubjectGroups ok", func() {
			var ctl *gomock.Controller
			var patches *gomonkey.Patches

			BeforeEach(func() {
				ctl = gomock.NewController(GinkgoT())
				patches = gomonkey.ApplyFunc(batchGetSubjectGroups,
					func([]int64) ([]types.ThinSubjectGroup, []int64, error) {
						return []types.ThinSubjectGroup{
							{
								PK:              2,
								PolicyExpiredAt: 21,
							},
							{
								PK:              3,
								PolicyExpiredAt: 31,
							},
						}, []int64{1}, nil
					})

				SubjectGroupCache = redis.NewMockCache("mockCache", 1*time.Minute)
			})
			AfterEach(func() {
				ctl.Finish()
				patches.Reset()
			})

			It("has no cached, get from database fail", func() {
				mockService := mock.NewMockSubjectService(ctl)
				mockService.EXPECT().ListSubjectEffectGroups([]int64{1}).Return(
					nil, errors.New("error")).AnyTimes()

				patches.ApplyFunc(service.NewSubjectService,
					func() service.SubjectService {
						return mockService
					})

				// call
				_, err := ListSubjectEffectGroups([]int64{1, 2, 3})
				assert.Error(GinkgoT(), err)
				assert.Contains(GinkgoT(), err.Error(), "SubjectService.ListSubjectEffectGroups")
			})
			It("has no cached, get from database success", func() {
				mockService := mock.NewMockSubjectService(ctl)
				mockService.EXPECT().ListSubjectEffectGroups([]int64{1}).Return(
					map[int64][]types.ThinSubjectGroup{
						int64(1): {
							{
								PK:              1,
								PolicyExpiredAt: 2,
							},
						},
					}, nil).AnyTimes()

				patches.ApplyFunc(service.NewSubjectService,
					func() service.SubjectService {
						return mockService
					})

				// call
				sgs, err := ListSubjectEffectGroups([]int64{1, 2, 3})
				assert.NoError(GinkgoT(), err)
				assert.Len(GinkgoT(), sgs, 3)
			})

			It("has no cached, get from database success, has empty cached", func() {
				mockService := mock.NewMockSubjectService(ctl)
				mockService.EXPECT().ListSubjectEffectGroups([]int64{1}).Return(
					map[int64][]types.ThinSubjectGroup{}, nil).AnyTimes()

				patches.ApplyFunc(service.NewSubjectService,
					func() service.SubjectService {
						return mockService
					})

				// call
				sgs, err := ListSubjectEffectGroups([]int64{1, 2, 3})
				assert.NoError(GinkgoT(), err)
				assert.Len(GinkgoT(), sgs, 2)
			})

		})

	})

	Context("batchGetSubjectGroups", func() {
		It("SubjectGroupCache.BatchGet empty", func() {
			var (
				expiration = 5 * time.Minute
			)
			mockCache := redis.NewMockCache("mockCache", expiration)
			SubjectGroupCache = mockCache

			// call it
			sgs, noCachePKs, err := batchGetSubjectGroups([]int64{1, 2, 3})
			assert.NoError(GinkgoT(), err)
			assert.Empty(GinkgoT(), sgs)
			assert.Len(GinkgoT(), noCachePKs, 3)
		})

		It("SubjectGroupCache.BatchGet fail", func() {
			patches := gomonkey.ApplyMethod(reflect.TypeOf(SubjectGroupCache), "BatchGet",
				func(*redis.Cache, []cache.Key) (map[cache.Key]string, error) {
					return nil, errors.New("error")
				})
			defer patches.Reset()

			_, _, err := batchGetSubjectGroups([]int64{1, 2, 3})
			assert.Error(GinkgoT(), err)
		})

		Context("SubjectGroupCache.BatchGet ok", func() {
			It("no hit", func() {
				patches := gomonkey.ApplyMethod(reflect.TypeOf(SubjectGroupCache), "BatchGet",
					func(*redis.Cache, []cache.Key) (map[cache.Key]string, error) {
						return map[cache.Key]string{}, nil
					})
				defer patches.Reset()

				sgs, noCachePKs, err := batchGetSubjectGroups([]int64{1, 2, 3})
				assert.NoError(GinkgoT(), err)
				assert.Empty(GinkgoT(), sgs)
				assert.Len(GinkgoT(), noCachePKs, 3)
			})

			It("hit, unmarshal fail", func() {
				patches := gomonkey.ApplyMethod(reflect.TypeOf(SubjectGroupCache), "BatchGet",
					func(*redis.Cache, []cache.Key) (map[cache.Key]string, error) {
						return map[cache.Key]string{
							SubjectPKCacheKey{PK: 2}: "1",
						}, nil
					})
				defer patches.Reset()

				_, _, err := batchGetSubjectGroups([]int64{1, 2, 3})
				assert.Error(GinkgoT(), err)
				assert.Contains(GinkgoT(), err.Error(), "unmarshal text in cache")
			})

			It("hit, unmarshal ok", func() {
				patches := gomonkey.ApplyMethod(reflect.TypeOf(SubjectGroupCache), "BatchGet",
					func(*redis.Cache, []cache.Key) (map[cache.Key]string, error) {
						bs, _ := SubjectGroupCache.Marshal([]types.ThinSubjectGroup{
							{
								PK:              1,
								PolicyExpiredAt: 0,
							},
						})

						return map[cache.Key]string{
							SubjectPKCacheKey{PK: 2}: util.BytesToString(bs),
						}, nil
					})
				defer patches.Reset()

				sgs, noCachePKs, err := batchGetSubjectGroups([]int64{1, 2, 3})
				assert.NoError(GinkgoT(), err)
				assert.Len(GinkgoT(), sgs, 1)
				assert.Len(GinkgoT(), noCachePKs, 2)
				assert.Contains(GinkgoT(), noCachePKs, int64(1))
				assert.Contains(GinkgoT(), noCachePKs, int64(3))
			})

		})

	})

})
