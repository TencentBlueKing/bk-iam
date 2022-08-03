/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package rbac

import (
	"time"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/prp/rbac/convert"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
	producermock "iam/pkg/task/producer/mock"
)

var _ = Describe("RbacPolicy", func() {
	Describe("rbacPolicyRedisRetriever", func() {
		var r *PolicyRedisRetriever
		var ctl *gomock.Controller
		var mockSubjectActionExpressionService *mock.MockSubjectActionExpressionService
		var mockSubjectActionGroupResourceService *mock.MockSubjectActionGroupResourceService
		var mockGroupAlterEventService *mock.MockGroupAlterEventService
		var mockProducer *producermock.MockProducer
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
			mockSubjectActionExpressionService = mock.NewMockSubjectActionExpressionService(ctl)
			mockSubjectActionGroupResourceService = mock.NewMockSubjectActionGroupResourceService(ctl)
			mockGroupAlterEventService = mock.NewMockGroupAlterEventService(ctl)
			mockProducer = producermock.NewMockProducer(ctl)

			r = &PolicyRedisRetriever{
				subjectActionExpressionService:    mockSubjectActionExpressionService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				groupAlterEventService:            mockGroupAlterEventService,
				alterEventProducer:                mockProducer,
			}
			cacheimpls.SubjectActionExpressionCache = redis.NewMockCache("test", 5*time.Minute)
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("cache all ok", func() {
			expressions := []types.SubjectActionExpression{
				{
					PK:         1,
					SubjectPK:  1,
					ActionPK:   1,
					Expression: "1",
					ExpiredAt:  time.Now().Add(time.Minute).Unix(),
				},
				{
					PK:         2,
					SubjectPK:  2,
					ActionPK:   1,
					Expression: "2",
					ExpiredAt:  time.Now().Add(time.Minute).Unix(),
				},
			}

			kvs := make([]redis.KV, 0, 2)
			for _, expression := range expressions {
				key := cacheimpls.SubjectActionCacheKey{
					SubjectPK: expression.SubjectPK,
					ActionPK:  expression.ActionPK,
				}

				value, _ := cacheimpls.SubjectActionExpressionCache.Marshal(expression)

				kvs = append(kvs, redis.KV{
					Key:   key.Key(),
					Value: conv.BytesToString(value),
				})
			}
			cacheimpls.SubjectActionExpressionCache.BatchSetWithTx(kvs, 0)

			es, err := r.ListBySubjectAction([]int64{1, 2}, 1)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), es, 2)
		})

		It("miss all ok", func() {
			expressions := []types.SubjectActionExpression{
				{
					PK:         1,
					SubjectPK:  1,
					ActionPK:   1,
					Expression: "1",
					ExpiredAt:  time.Now().Add(time.Minute).Unix(),
				},
				{
					PK:         2,
					SubjectPK:  2,
					ActionPK:   1,
					Expression: "2",
					ExpiredAt:  time.Now().Add(time.Minute).Unix(),
				},
			}

			mockSubjectActionExpressionService.EXPECT().
				ListBySubjectAction([]int64{1, 2}, int64(1)).
				Return(expressions, nil).
				Times(1)

			es, err := r.ListBySubjectAction([]int64{1, 2}, 1)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), es, 2)

			// with cache
			es, err = r.ListBySubjectAction([]int64{1, 2}, 1)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), es, 2)
		})

		It("part miss ok", func() {
			expressions := []types.SubjectActionExpression{
				{
					PK:         1,
					SubjectPK:  1,
					ActionPK:   1,
					Expression: "1",
					ExpiredAt:  time.Now().Add(time.Minute).Unix(),
				},
				{
					PK:         2,
					SubjectPK:  2,
					ActionPK:   1,
					Expression: "2",
					ExpiredAt:  time.Now().Add(time.Minute).Unix(),
				},
			}

			mockSubjectActionExpressionService.EXPECT().
				ListBySubjectAction([]int64{2}, int64(1)).
				Return(expressions[1:2], nil).
				Times(1)

			expression := expressions[0]
			key := cacheimpls.SubjectActionCacheKey{
				SubjectPK: expression.SubjectPK,
				ActionPK:  expression.ActionPK,
			}
			cacheimpls.SubjectActionExpressionCache.Set(key, expression, 0)

			es, err := r.ListBySubjectAction([]int64{1, 2}, 1)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), es, 2)

			// with cache
			es, err = r.ListBySubjectAction([]int64{1, 2}, 1)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), es, 2)
		})

		It("expired refresh ok", func() {
			expressions := []types.SubjectActionExpression{
				{
					PK:         1,
					SubjectPK:  1,
					ActionPK:   1,
					Expression: "1",
					ExpiredAt:  time.Now().Add(time.Minute).Unix(),
				},
				{
					PK:         2,
					SubjectPK:  2,
					ActionPK:   1,
					Expression: "2",
					ExpiredAt:  10,
				},
			}

			kvs := make([]redis.KV, 0, 2)
			for _, expression := range expressions {
				key := cacheimpls.SubjectActionCacheKey{
					SubjectPK: expression.SubjectPK,
					ActionPK:  expression.ActionPK,
				}

				value, _ := cacheimpls.SubjectActionExpressionCache.Marshal(expression)

				kvs = append(kvs, redis.KV{
					Key:   key.Key(),
					Value: conv.BytesToString(value),
				})
			}
			cacheimpls.SubjectActionExpressionCache.BatchSetWithTx(kvs, 0)

			mockSubjectActionGroupResourceService.EXPECT().
				Get(int64(2), int64(1)).
				Return(types.SubjectActionGroupResource{}, nil).
				Times(1)
			patches := gomonkey.ApplyFunc(
				convert.SubjectActionGroupResourceToExpression,
				func(r types.SubjectActionGroupResource) (types.SubjectActionExpression, error) {
					return types.SubjectActionExpression{
						PK:         2,
						SubjectPK:  2,
						ActionPK:   1,
						Expression: "test",
						ExpiredAt:  time.Now().Add(time.Minute).Unix(),
					}, nil
				},
			)
			defer patches.Reset()

			mockGroupAlterEventService.EXPECT().
				CreateBySubjectActionGroup(int64(2), int64(1), int64(0)).
				Return(int64(1), nil).
				Times(1)
			mockProducer.EXPECT().Publish("1").Return(nil).Times(1)

			es, err := r.ListBySubjectAction([]int64{1, 2}, 1)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), es, 2)
			for _, e := range es {
				if e.PK == 2 {
					assert.Equal(GinkgoT(), "test", e.Expression)
				}
			}

			// with cache
			es, err = r.ListBySubjectAction([]int64{1, 2}, 1)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), es, 2)
			for _, e := range es {
				if e.PK == 2 {
					assert.Equal(GinkgoT(), "test", e.Expression)
				}
			}
		})
	})
})
