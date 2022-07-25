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
	"iam/pkg/service/types"
)

var _ = Describe("ResourceTypeService", func() {
	Describe("GetByPK", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})
		It("manager.GetByPK fail", func() {
			mockManager := mock.NewMockResourceTypeManager(ctl)
			mockManager.EXPECT().
				GetByPK(int64(1)).
				Return(dao.ResourceType{}, errors.New("error"))

			svc := &resourceTypeService{
				manager: mockManager,
			}

			_, err := svc.GetByPK(1)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetByPK")
		})

		It("ok", func() {
			mockManager := mock.NewMockResourceTypeManager(ctl)
			mockManager.EXPECT().
				GetByPK(int64(1)).
				Return(dao.ResourceType{
					PK: 1, System: "system", ID: "id",
				}, nil)

			svc := &resourceTypeService{
				manager: mockManager,
			}

			resourceType, err := svc.GetByPK(int64(1))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), types.ThinResourceType{
				PK: 1, System: "system", ID: "id",
			}, resourceType)
		})
	})
})
