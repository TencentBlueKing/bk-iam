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
	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("ActionService", func() {

	Describe("ListThinActionBySystem", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListBySystem fail", func() {
			mockActionService := mock.NewMockActionManager(ctl)
			mockActionService.EXPECT().ListBySystem("test").Return(
				[]dao.Action{}, errors.New("list by system fail"),
			).AnyTimes()

			manager := &actionService{
				manager: mockActionService,
			}

			_, err := manager.ListThinActionBySystem("test")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "manager.ListBySystem")
		})

		It("success", func() {
			mockActionService := mock.NewMockActionManager(ctl)
			mockActionService.EXPECT().ListBySystem("test").Return(
				[]dao.Action{}, nil,
			).AnyTimes()

			manager := &actionService{
				manager: mockActionService,
			}

			actions, err := manager.ListThinActionBySystem("test")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ThinAction{}, actions)
		})
	})
})
