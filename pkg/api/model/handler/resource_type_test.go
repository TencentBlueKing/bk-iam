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
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("ResourceType", func() {

	Describe("ResourceTypeUpdateSerializer Validate", func() {
		var slz resourceTypeUpdateSerializer
		BeforeEach(func() {
			slz = resourceTypeUpdateSerializer{}
		})

		It("name empty", func() {
			a := map[string]interface{}{
				"name": "",
			}
			valid, _ := slz.validate(a)
			assert.False(GinkgoT(), valid)
		})
		It("name_en empty", func() {
			b := map[string]interface{}{
				"name_en": "",
			}
			valid, _ := slz.validate(b)
			assert.False(GinkgoT(), valid)
		})
		It("version < 1", func() {
			c := map[string]interface{}{
				"version": 0,
			}
			valid, _ := slz.validate(c)
			assert.False(GinkgoT(), valid)
		})
		// 4. TODO: validate r.Parents
		// It("", func() {})
		// 5. TODO: validate provider_config
		// It("", func() {})

	})
})
