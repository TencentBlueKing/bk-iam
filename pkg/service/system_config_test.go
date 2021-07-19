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
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/sdao"
	"iam/pkg/database/sdao/mock"
)

var _ = Describe("SystemConfigService", func() {
	Describe("getSliceConfig cases(in system_config)", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := sdao.SaaSSystemConfig{
				Type:  "json",
				Value: `[{"1": "1"}, {"2": "2"}]`,
			}
			expectedSliceMap := []map[string]interface{}{
				{"1": "1"},
				{"2": "2"},
			}
			expected := make([]interface{}, 0, len(expectedSliceMap))
			for _, e := range expectedSliceMap {
				expected = append(expected, e)
			}

			mockSaaSSystemConfigManager := mock.NewMockSaaSSystemConfigManager(ctl)
			mockSaaSSystemConfigManager.EXPECT().Get(gomock.Any(), gomock.Any()).Return(returned, nil)

			svc := &systemConfigService{
				manager: mockSaaSSystemConfigManager,
			}

			cfg, err := svc.getSliceConfig("", "")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, cfg)
		})

		It("error", func() {
			returned := sdao.SaaSSystemConfig{
				Type:  "json",
				Value: `{"1": "1", "2": "2"}`,
			}
			mockSaaSSystemConfigManager := mock.NewMockSaaSSystemConfigManager(ctl)
			mockSaaSSystemConfigManager.EXPECT().Get(gomock.Any(), gomock.Any()).Return(returned, nil)

			svc := &systemConfigService{
				manager: mockSaaSSystemConfigManager,
			}

			_, err := svc.getSliceConfig("", "")
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("getMapConfig cases(in system_config)", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := sdao.SaaSSystemConfig{
				Type:  "json",
				Value: `{"1": "1", "2": "2"}`,
			}
			expected := map[string]interface{}{"1": "1", "2": "2"}

			mockSaaSSystemConfigManager := mock.NewMockSaaSSystemConfigManager(ctl)
			mockSaaSSystemConfigManager.EXPECT().Get(gomock.Any(), gomock.Any()).Return(returned, nil)

			svc := &systemConfigService{
				manager: mockSaaSSystemConfigManager,
			}

			cfg, err := svc.getMapConfig("", "")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, cfg)
		})

		It("error", func() {
			returned := sdao.SaaSSystemConfig{
				Type:  "json",
				Value: `[{"1": "1"}, {"2": "2"}]`,
			}
			mockSaaSSystemConfigManager := mock.NewMockSaaSSystemConfigManager(ctl)
			mockSaaSSystemConfigManager.EXPECT().Get(gomock.Any(), gomock.Any()).Return(returned, nil)

			svc := &systemConfigService{
				manager: mockSaaSSystemConfigManager,
			}

			_, err := svc.getMapConfig("", "")
			assert.Error(GinkgoT(), err)
		})
	})
})
