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
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

var _ = Describe("Map", func() {
	Describe("MapValueInterfaceToString", func() {
		DescribeTable("MapValueInterfaceToString cases", func(expected string, willError bool, input map[string]interface{}) {

			data, err := util.MapValueInterfaceToString(input)
			if willError {
				assert.Error(GinkgoT(), err)
			} else {
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), expected, data["a"])
			}
		},
			Entry("will error", "", true,
				map[string]interface{}{
					"a": "1",
					"b": 2,
				}),
			Entry("ok", "1", false,
				map[string]interface{}{
					"a": "1",
					"b": "2",
				}),
		)
	})

})
