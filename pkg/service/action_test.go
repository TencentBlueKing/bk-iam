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

	"iam/pkg/database/dao/mock"
	"iam/pkg/database/sdao"
	smock "iam/pkg/database/sdao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("ActionService", func() {
	Describe("GetActionPK cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockActionManager := mock.NewMockActionManager(ctl)
			mockActionManager.EXPECT().GetPK("iam", "execute").Return(int64(1), nil)

			pk, err := mockActionManager.GetPK("iam", "execute")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), pk)
		})

		It("error", func() {
			mockActionManager := mock.NewMockActionManager(ctl)
			mockActionManager.EXPECT().GetPK("iam", "execute").Return(int64(1), errors.New("error"))

			_, err := mockActionManager.GetPK("iam", "execute")
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("ListBaseInfoBySystem", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListBySystem fail", func() {
			mockActionService := smock.NewMockSaaSActionManager(ctl)
			mockActionService.EXPECT().ListBySystem("test").Return(
				[]sdao.SaaSAction{}, errors.New("list by system fail"),
			).AnyTimes()

			manager := &actionService{
				saasManager: mockActionService,
			}

			_, err := manager.ListBaseInfoBySystem("test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "saasManager.ListBySystem")
		})

		It("success", func() {
			mockActionService := smock.NewMockSaaSActionManager(ctl)
			mockActionService.EXPECT().ListBySystem("test").Return(
				[]sdao.SaaSAction{}, nil,
			).AnyTimes()

			manager := &actionService{
				saasManager: mockActionService,
			}

			actions, err := manager.ListBaseInfoBySystem("test")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ActionBaseInfo{}, actions)
		})
	})

	Describe("GetAuthType", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("saasManager.GetAuthType fail", func() {
			mockActionService := smock.NewMockSaaSActionManager(ctl)
			mockActionService.EXPECT().GetAuthType("test", "action").Return(
				"", errors.New("GetAuthType"),
			).AnyTimes()

			manager := &actionService{
				saasManager: mockActionService,
			}

			_, err := manager.GetAuthType("test", "action")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetAuthType")
		})

		It("empty ok", func() {
			mockActionService := smock.NewMockSaaSActionManager(ctl)
			mockActionService.EXPECT().GetAuthType("test", "action").Return(
				"", nil,
			).AnyTimes()

			manager := &actionService{
				saasManager: mockActionService,
			}

			authType, err := manager.GetAuthType("test", "action")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), authType)
		})

		It("abac ok", func() {
			mockActionService := smock.NewMockSaaSActionManager(ctl)
			mockActionService.EXPECT().GetAuthType("test", "action").Return(
				"abac", nil,
			).AnyTimes()

			manager := &actionService{
				saasManager: mockActionService,
			}

			authType, err := manager.GetAuthType("test", "action")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(1), authType)
		})

		It("rbac ok", func() {
			mockActionService := smock.NewMockSaaSActionManager(ctl)
			mockActionService.EXPECT().GetAuthType("test", "action").Return(
				"rbac", nil,
			).AnyTimes()

			manager := &actionService{
				saasManager: mockActionService,
			}

			authType, err := manager.GetAuthType("test", "action")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(2), authType)
		})

		It("unknown fail", func() {
			mockActionService := smock.NewMockSaaSActionManager(ctl)
			mockActionService.EXPECT().GetAuthType("test", "action").Return(
				"test", nil,
			).AnyTimes()

			manager := &actionService{
				saasManager: mockActionService,
			}

			_, err := manager.GetAuthType("test", "action")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "unknown")
		})
	})
})
