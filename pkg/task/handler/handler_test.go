/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"database/sql"
	"errors"
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	rds "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/prp/rbac/convert"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/locker"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

var _ = Describe("Handler", func() {
	Describe("handlerEvent", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			patches = gomonkey.ApplyFunc(redis.GetDefaultRedisClient, func() *rds.Client {
				return util.NewTestRedisClient()
			})

			tx := &sql.Tx{}
			patches.ApplyMethod(reflect.TypeOf(tx), "Commit", func(tx *sql.Tx) error {
				return nil
			})

			patches.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return &sqlx.Tx{Tx: tx}, nil
			})
			patches.ApplyFunc(database.RollBackWithLog, func(tx *sqlx.Tx) {})
		})

		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subjectActionGroupResourceService.Get error", func() {
			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				Get(int64(1), int64(3)).
				Return(types.SubjectActionGroupResource{}, errors.New("error"))

			handler := &groupAlterMessageHandler{
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				locker:                            locker.NewDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, []int64{2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "Get")
		})

		It("groupService.GetExpiredAtBySubjectGroup error", func() {
			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				Get(int64(1), int64(3)).
				Return(types.SubjectActionGroupResource{}, sql.ErrNoRows)
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(0), errors.New("error"))

			handler := &groupAlterMessageHandler{
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				groupService:                      mockGroupService,
				locker:                            locker.NewDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, []int64{2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetMaxExpiredAtBySubjectGroup")
		})

		It("cacheimpls.GetGroupActionAuthorizedResource error", func() {
			patches.ApplyFunc(
				cacheimpls.GetGroupActionAuthorizedResource,
				func(_, _ int64) (map[int64][]string, error) {
					return nil, errors.New("error")
				},
			)

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				Get(int64(1), int64(3)).
				Return(types.SubjectActionGroupResource{}, sql.ErrNoRows)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(0), service.ErrGroupMemberNotFound)

			handler := &groupAlterMessageHandler{
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				groupService:                      mockGroupService,
				locker:                            locker.NewDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, []int64{2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetGroupActionAuthorizedResource")
		})

		It("subjectActionGroupResourceService.CreateOrUpdateWithTx error", func() {
			patches.ApplyFunc(
				cacheimpls.GetGroupActionAuthorizedResource,
				func(_, _ int64) (map[int64][]string, error) {
					return map[int64][]string{
						1: {"1", "2"},
					}, nil
				},
			)

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				Get(int64(1), int64(3)).
				Return(types.SubjectActionGroupResource{}, sql.ErrNoRows)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(10), nil)

			mockSubjectActionGroupResourceService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), types.SubjectActionGroupResource{
					SubjectPK: 1,
					ActionPK:  3,
					GroupResource: map[int64]types.ResourceExpiredAt{
						2: {
							ExpiredAt: int64(10),
							Resources: map[int64][]string{
								1: {"1", "2"},
							},
						},
					},
				}).
				Return(errors.New("error"))

			handler := &groupAlterMessageHandler{
				groupService:                      mockGroupService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				locker:                            locker.NewDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, []int64{2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "CreateOrUpdateWithTx")
		})

		It("subjectActionExpressionService.CreateOrUpdateWithTx error", func() {
			patches.ApplyFunc(
				cacheimpls.GetGroupActionAuthorizedResource,
				func(_, _ int64) (map[int64][]string, error) {
					return map[int64][]string{
						1: {"1", "2"},
					}, nil
				},
			)
			patches.ApplyFunc(
				convert.SubjectActionGroupResourceToExpression,
				func(obj types.SubjectActionGroupResource) (expression types.SubjectActionExpression, err error) {
					return types.SubjectActionExpression{}, nil
				},
			)

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				Get(int64(1), int64(3)).
				Return(types.SubjectActionGroupResource{}, sql.ErrNoRows)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(10), nil)

			mockSubjectActionGroupResourceService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), types.SubjectActionGroupResource{
					SubjectPK: 1,
					ActionPK:  3,
					GroupResource: map[int64]types.ResourceExpiredAt{
						2: {
							ExpiredAt: int64(10),
							Resources: map[int64][]string{
								1: {"1", "2"},
							},
						},
					},
				}).
				Return(nil)

			mockSubjectActionExpressionService := mock.NewMockSubjectActionExpressionService(ctl)
			mockSubjectActionExpressionService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), gomock.Any()).
				Return(errors.New("test"))

			handler := &groupAlterMessageHandler{
				groupService:                      mockGroupService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				subjectActionExpressionService:    mockSubjectActionExpressionService,
				locker:                            locker.NewDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, []int64{2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectActionExpressionService")
		})

		It("ok", func() {
			patches.ApplyFunc(
				cacheimpls.GetGroupActionAuthorizedResource,
				func(_, _ int64) (map[int64][]string, error) {
					return map[int64][]string{
						1: {"1", "2"},
					}, nil
				},
			)
			patches.ApplyFunc(
				convert.SubjectActionGroupResourceToExpression,
				func(obj types.SubjectActionGroupResource) (expression types.SubjectActionExpression, err error) {
					return types.SubjectActionExpression{}, nil
				},
			)
			patches.ApplyFunc(cacheimpls.DeleteSubjectActionExpressionCache, func(_, _ int64) error {
				return nil
			})

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				Get(int64(1), int64(3)).
				Return(types.SubjectActionGroupResource{}, sql.ErrNoRows)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(10), nil)

			mockSubjectActionGroupResourceService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), types.SubjectActionGroupResource{
					SubjectPK: 1,
					ActionPK:  3,
					GroupResource: map[int64]types.ResourceExpiredAt{
						2: {
							ExpiredAt: int64(10),
							Resources: map[int64][]string{
								1: {"1", "2"},
							},
						},
					},
				}).
				Return(nil)

			mockSubjectActionExpressionService := mock.NewMockSubjectActionExpressionService(ctl)
			mockSubjectActionExpressionService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), gomock.Any()).
				Return(nil)

			handler := &groupAlterMessageHandler{
				groupService:                      mockGroupService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				subjectActionExpressionService:    mockSubjectActionExpressionService,
				locker:                            locker.NewDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, []int64{2})
			assert.NoError(GinkgoT(), err)
		})
	})
})
