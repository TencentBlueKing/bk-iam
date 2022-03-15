/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/config"
)

var _ = Describe("Init", func() {

	Describe("quotas", func() {
		It("all default", func() {
			InitQuota(config.Quota{}, map[string]config.Quota{})

			assert.Equal(GinkgoT(), DefaultMaxActionsLimit, GetMaxActionsLimit("abc"))
			assert.Equal(GinkgoT(), DefaultMaxResourceTypesLimit, GetMaxResourceTypesLimit("abc"))
			assert.Equal(GinkgoT(), DefaultMaxInstanceSelectionsLimit, GetMaxInstanceSelectionsLimit("abc"))
		})

		It("hit config file default", func() {
			InitQuota(config.Quota{
				Model: map[string]int{
					maxActionsLimitKey:            123,
					maxResourceTypesLimitKey:      456,
					maxInstanceSelectionsLimitKey: 789,
				},
			}, map[string]config.Quota{})

			assert.Equal(GinkgoT(), 123, GetMaxActionsLimit("abc"))
			assert.Equal(GinkgoT(), 456, GetMaxResourceTypesLimit("abc"))
			assert.Equal(GinkgoT(), 789, GetMaxInstanceSelectionsLimit("abc"))
		})

		It("hit custom quotas", func() {
			InitQuota(config.Quota{
				Model: map[string]int{
					maxActionsLimitKey:            123,
					maxResourceTypesLimitKey:      456,
					maxInstanceSelectionsLimitKey: 789,
				},
			}, map[string]config.Quota{
				"abc": {
					Model: map[string]int{
						maxActionsLimitKey:            111,
						maxResourceTypesLimitKey:      222,
						maxInstanceSelectionsLimitKey: 333,
					},
				},
			})

			assert.Equal(GinkgoT(), 111, GetMaxActionsLimit("abc"))
			assert.Equal(GinkgoT(), 222, GetMaxResourceTypesLimit("abc"))
			assert.Equal(GinkgoT(), 333, GetMaxInstanceSelectionsLimit("abc"))
		})
	})

	Describe("switches", func() {
		It("hit default", func() {
			InitSwitch(map[string]bool{})

			assert.Equal(GinkgoT(), DisableCreateSystemClientValidation, GetSwitchDisableCreateSystemClientValidation())
		})
		It("set false", func() {
			InitSwitch(map[string]bool{
				triggerDisableCreateSystemClientValidationKey: false,
			})
			assert.False(GinkgoT(), GetSwitchDisableCreateSystemClientValidation())
		})
		It("set true", func() {
			InitSwitch(map[string]bool{
				triggerDisableCreateSystemClientValidationKey: true,
			})
			assert.True(GinkgoT(), GetSwitchDisableCreateSystemClientValidation())
		})
	})

})
