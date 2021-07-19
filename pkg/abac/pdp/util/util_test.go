/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/types"
	"iam/pkg/abac/pdp/util"
)

var _ = Describe("Util", func() {
	Describe("InterfaceToPolicyCondition", func() {
		It("ok", func() {
			expected := types.PolicyCondition{
				"StringEqual": {
					"id": {"1", "2"},
				},
			}

			value := map[string]interface{}{
				"StringEqual": map[string]interface{}{
					"id": []interface{}{"1", "2"},
				},
			}
			c, err := util.InterfaceToPolicyCondition(value)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c)
		})

		It("invalid value", func() {
			_, err := util.InterfaceToPolicyCondition("abc")
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), util.ErrTypeAssertFail, err)
		})

		It("invalid attribute, should be an array", func() {
			value := map[string]interface{}{
				"StringEqual": map[string]interface{}{
					"id": "invalid",
				},
			}
			_, err := util.InterfaceToPolicyCondition(value)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), util.ErrTypeAssertFail, err)
		})

		It("invalid operatorMap", func() {
			value := map[string]interface{}{
				"StringEqual": "invalid",
			}
			_, err := util.InterfaceToPolicyCondition(value)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), util.ErrTypeAssertFail, err)
		})
	})
})
