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

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
)

var _ = Describe("OpenAbacPolicyService", func() {
	Describe("ListQueryByPKs cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.OpenAbacPolicy{
				{
					Policy: dao.Policy{
						PK:           1,
						ExpressionPK: 1,
						TemplateID:   0,
					},
				},
				{
					Policy: dao.Policy{
						PK:           2,
						ExpressionPK: 2,
						TemplateID:   1,
					},
				},
			}
			mockPolicyManager := mock.NewMockOpenAbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(returned, nil)

			svc := openAbacPolicyService{
				manager: mockPolicyManager,
			}

			policies, err := svc.ListByPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 2)
		})

		It("fail", func() {
			mockPolicyManager := mock.NewMockOpenAbacPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(nil, errors.New("list fail"))

			svc := openAbacPolicyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListByPKs([]int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "manager.ListByPKs")
		})
	})
})
