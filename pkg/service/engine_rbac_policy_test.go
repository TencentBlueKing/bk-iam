/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"errors"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("PolicyEngine Rbac", func() {
	Describe("ListBetweenPK cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			updatedAt := time.Now()
			daoPolicies := []dao.EngineRbacPolicy{
				{
					PK:                          int64(1),
					GroupPK:                     int64(2),
					TemplateID:                  int64(3),
					SystemID:                    "",
					ActionPKs:                   "",
					ActionRelatedResourceTypePK: int64(4),
					ResourceTypePK:              int64(5),
					ResourceID:                  "",
					UpdatedAt:                   updatedAt,
				},
			}
			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBetweenPK(
				int64(1), int64(100),
			).Return(daoPolicies, nil)

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			policies, err := svc.ListBetweenPK(int64(1), int64(1), int64(100))
			assert.NoError(GinkgoT(), err)

			assert.Equal(GinkgoT(), []types.EngineRbacPolicy{{
				PK:                          int64(1),
				GroupPK:                     int64(2),
				TemplateID:                  int64(3),
				SystemID:                    "",
				ActionPKs:                   "",
				ActionRelatedResourceTypePK: int64(4),
				ResourceTypePK:              int64(5),
				ResourceID:                  "",
				UpdatedAt:                   updatedAt,
			}}, policies)
		})

		It("ListBetweenPK fail", func() {
			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBetweenPK(
				int64(1), int64(100),
			).Return(nil, errors.New("fail"))

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListBetweenPK(int64(1), int64(1), int64(100))
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("ListByPKs cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			updatedAt := time.Now()
			daoPolicies := []dao.EngineRbacPolicy{
				{
					PK:                          int64(1),
					GroupPK:                     int64(2),
					TemplateID:                  int64(3),
					SystemID:                    "",
					ActionPKs:                   "",
					ActionRelatedResourceTypePK: int64(4),
					ResourceTypePK:              int64(5),
					ResourceID:                  "",
					UpdatedAt:                   updatedAt,
				},
			}
			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(daoPolicies, nil)

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			policies, err := svc.ListByPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)

			assert.Equal(GinkgoT(), []types.EngineRbacPolicy{{
				PK:                          int64(1),
				GroupPK:                     int64(2),
				TemplateID:                  int64(3),
				SystemID:                    "",
				ActionPKs:                   "",
				ActionRelatedResourceTypePK: int64(4),
				ResourceTypePK:              int64(5),
				ResourceID:                  "",
				UpdatedAt:                   updatedAt,
			}}, policies)
		})

		It("ListBetweenPK fail", func() {
			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(nil, errors.New("fail"))

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListByPKs([]int64{1, 2})
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("GetMaxPKBeforeUpdatedAt cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			now := int64(1617457847)
			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().GetMaxPKBeforeUpdatedAt(now).Return(int64(1), nil)

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			cnt, err := svc.GetMaxPKBeforeUpdatedAt(now)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), cnt)
		})

		It("GetCountByActionBeforeExpiredAtBetweenUpdatedAt fail", func() {
			now := int64(1617457847)
			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().GetMaxPKBeforeUpdatedAt(now).Return(int64(0), errors.New("fail"))

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			_, err := svc.GetMaxPKBeforeUpdatedAt(now)
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("ListPKBetweenUpdatedAt cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			begin := int64(1617457847)
			end := int64(1617457850)
			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListPKBetweenUpdatedAt(begin, end).Return([]int64{1, 2}, nil)

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			pks, err := svc.ListPKBetweenUpdatedAt(begin, end)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []int64{1, 2}, pks)
		})

		It("ListPKBetweenUpdatedAt fail", func() {
			begin := int64(1617457847)
			end := int64(1617457850)

			mockPolicyManager := mock.NewMockEngineRbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListPKBetweenUpdatedAt(begin, end).Return(nil, errors.New("fail"))

			svc := engineRbacPolicyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListPKBetweenUpdatedAt(begin, end)
			assert.Error(GinkgoT(), err)
		})
	})
})
