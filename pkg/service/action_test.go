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
	. "github.com/onsi/ginkgo"

	"iam/pkg/database/dao/mock"
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
})
