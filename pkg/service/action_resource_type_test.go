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

	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("ActionService", func() {
	// Describe("ListActionResourceTypes cases(in action_resource_type)", func() {
	// 	var ctl *gomock.Controller
	//
	// 	BeforeEach(func() {
	// 		ctl = gomock.NewController(GinkgoT())
	// 	})
	//
	// 	AfterEach(func() {
	// 		ctl.Finish()
	// 	})
	//
	// 	It("ok", func() {
	// 		returned := []dao.ActionResourceType{
	// 			{
	// 				ResourceTypeSystem: "bk_cmdb",
	// 				ResourceTypeID:     "host",
	// 			},
	// 			{
	// 				ResourceTypeSystem: "bk_job",
	// 				ResourceTypeID:     "job",
	// 			},
	// 		}
	//
	// 		expected := []types.ThinActionResourceType{
	// 			{
	// 				System: "bk_cmdb",
	// 				ID:     "host",
	// 			},
	// 			{
	// 				System: "bk_job",
	// 				ID:     "job",
	// 			},
	// 		}
	//
	// 		mockActionResourceTypeManager := mock.NewMockActionResourceTypeManager(ctl)
	// 		mockActionResourceTypeManager.EXPECT().ListResourceTypeByAction(
	// 			gomock.Any(), gomock.Any()).Return(returned, nil)
	//
	// 		svc := &actionService{
	// 			actionResourceTypeManager: mockActionResourceTypeManager,
	// 		}
	//
	// 		sgs, err := svc.ListThinActionResourceTypes("bk_job", "job_execute")
	// 		assert.NoError(GinkgoT(), err)
	// 		assert.Equal(GinkgoT(), expected, sgs)
	// 	})
	//
	// 	It("error", func() {
	// 		mockActionResourceTypeManager := mock.NewMockActionResourceTypeManager(ctl)
	// 		mockActionResourceTypeManager.EXPECT().ListResourceTypeByAction(
	// 			gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
	//
	// 		svc := &actionService{
	// 			actionResourceTypeManager: mockActionResourceTypeManager,
	// 		}
	//
	// 		_, err := svc.ListThinActionResourceTypes("bk_job", "job_execute")
	// 		assert.Error(GinkgoT(), err)
	// 	})
	// })

	Describe("ListActionResourceTypeIDByActionSystem", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.ActionResourceType{
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_cmdb",
					ResourceTypeID:     "host",
				},
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_job",
					ResourceTypeID:     "job",
				},
			}

			expected := []types.ActionResourceTypeID{
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_cmdb",
					ResourceTypeID:     "host",
				},
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_job",
					ResourceTypeID:     "job",
				},
			}

			mockActionResourceTypeManager := mock.NewMockActionResourceTypeManager(ctl)
			mockActionResourceTypeManager.EXPECT().ListByActionSystem(
				gomock.Any()).Return(returned, nil)

			svc := &actionService{
				actionResourceTypeManager: mockActionResourceTypeManager,
			}

			sgs, err := svc.ListActionResourceTypeIDByActionSystem("bk_job")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, sgs)
		})

		It("fail", func() {
			mockActionResourceTypeManager := mock.NewMockActionResourceTypeManager(ctl)
			mockActionResourceTypeManager.EXPECT().ListByActionSystem(
				gomock.Any()).Return(nil, errors.New("list fail"))

			svc := &actionService{
				actionResourceTypeManager: mockActionResourceTypeManager,
			}

			_, err := svc.ListActionResourceTypeIDByActionSystem("bk_job")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionResourceTypeManager.ListByActionSystem")
		})
	})
})
