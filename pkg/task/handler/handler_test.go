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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	rds "github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

var _ = Describe("Handler", func() {
	Describe("convertToSubjectActionExpression", func() {
		It("ok", func() {
			patches := gomonkey.ApplyFunc(time.Now, func() time.Time {
				return time.Time{}
			})
			defer patches.Reset()
			patches.ApplyFunc(cacheimpls.GetAction, func(_ int64) (types.ThinAction, error) {
				return types.ThinAction{
					System: "system",
					ID:     "action",
					PK:     1,
				}, nil
			})
			patches.ApplyFunc(cacheimpls.GetLocalActionDetail, func(_, _ string) (types.ActionDetail, error) {
				return types.ActionDetail{
					ResourceTypes: []types.ThinActionResourceType{
						{System: "system", ID: "resource_type"},
					},
				}, nil
			})
			patches.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _ string) (int64, error) {
				return 1, nil
			})
			patches.ApplyFunc(cacheimpls.GetThinResourceType, func(_ int64) (types.ThinResourceType, error) {
				return types.ThinResourceType{
					System: "system",
					ID:     "resource_type2",
					PK:     2,
				}, nil
			})

			obj := types.SubjectActionGroupResource{
				SubjectPK: 1,
				ActionPK:  1,
				GroupResource: map[int64]types.ExpiredAtResource{
					1: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							1: {"resource1", "resource2"},
						},
					},
					2: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							2: {"resource3", "resource4"},
						},
					},
				},
			}

			expression, err := convertToSubjectActionExpression(obj)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expression.Expression, 209)
		})
	})

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

		It("groupService.GetExpiredAtBySubjectGroup error", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(0), errors.New("error"))

			handler := &groupAlterMessageHandler{
				groupService: mockGroupService,
				locker:       newDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, 2)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetExpiredAtBySubjectGroup")
		})

		It("cacheimpls.GetGroupActionAuthorizedResource error", func() {
			patches.ApplyFunc(
				cacheimpls.GetGroupActionAuthorizedResource,
				func(_, _ int64) (map[int64][]string, error) {
					return nil, errors.New("error")
				},
			)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(0), service.ErrGroupMemberNotFound)

			handler := &groupAlterMessageHandler{
				groupService: mockGroupService,
				locker:       newDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, 2)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetGroupActionAuthorizedResource")
		})

		It("subjectActionGroupResourceService.DeleteGroupWithTx error", func() {
			patches.ApplyFunc(
				cacheimpls.GetGroupActionAuthorizedResource,
				func(_, _ int64) (map[int64][]string, error) {
					return nil, nil
				},
			)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(0), service.ErrGroupMemberNotFound)

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				DeleteGroupResourceWithTx(gomock.Any(), int64(1), int64(3), int64(2)).
				Return(types.SubjectActionGroupResource{}, errors.New("error"))

			handler := &groupAlterMessageHandler{
				groupService:                      mockGroupService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				locker:                            newDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, 2)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "DeleteGroupWithTx")
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

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(10), nil)

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), int64(1), int64(3), int64(2), int64(10), map[int64][]string{
					1: {"1", "2"},
				}).
				Return(types.SubjectActionGroupResource{}, errors.New("error"))

			handler := &groupAlterMessageHandler{
				groupService:                      mockGroupService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				locker:                            newDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, 2)
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
				convertToSubjectActionExpression,
				func(obj types.SubjectActionGroupResource) (expression types.SubjectActionExpression, err error) {
					return types.SubjectActionExpression{}, nil
				},
			)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(10), nil)

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), int64(1), int64(3), int64(2), int64(10), map[int64][]string{
					1: {"1", "2"},
				}).
				Return(types.SubjectActionGroupResource{}, nil)

			mockSubjectActionExpressionService := mock.NewMockSubjectActionExpressionService(ctl)
			mockSubjectActionExpressionService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), gomock.Any()).
				Return(errors.New("test"))

			handler := &groupAlterMessageHandler{
				groupService:                      mockGroupService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				subjectActionExpressionService:    mockSubjectActionExpressionService,
				locker:                            newDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, 2)
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
				convertToSubjectActionExpression,
				func(obj types.SubjectActionGroupResource) (expression types.SubjectActionExpression, err error) {
					return types.SubjectActionExpression{}, nil
				},
			)

			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(int64(10), nil)

			mockSubjectActionGroupResourceService := mock.NewMockSubjectActionGroupResourceService(ctl)
			mockSubjectActionGroupResourceService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), int64(1), int64(3), int64(2), int64(10), map[int64][]string{
					1: {"1", "2"},
				}).
				Return(types.SubjectActionGroupResource{}, nil)

			mockSubjectActionExpressionService := mock.NewMockSubjectActionExpressionService(ctl)
			mockSubjectActionExpressionService.EXPECT().
				CreateOrUpdateWithTx(gomock.Any(), gomock.Any()).
				Return(nil)

			handler := &groupAlterMessageHandler{
				groupService:                      mockGroupService,
				subjectActionGroupResourceService: mockSubjectActionGroupResourceService,
				subjectActionExpressionService:    mockSubjectActionExpressionService,
				locker:                            newDistributedSubjectActionLocker(),
			}
			err := handler.alterSubjectActionGroupResource(1, 3, 2)
			assert.NoError(GinkgoT(), err)
		})
	})
})
