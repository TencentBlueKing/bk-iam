/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/api/model/handler"
)

var _ = Describe("Types", func() {
	Describe("AllBaseInfo", func() {
		var a handler.AllBaseInfo
		BeforeEach(func() {
			a = handler.AllBaseInfo{
				IDSet: map[string]string{
					"cmdb": "cmdb",
					"job":  "job",
				},
				NameSet: map[string]string{
					"cmdb_n": "cmdb",
					"job_n":  "job",
				},
				NameEnSet: map[string]string{
					"cmdb_en": "cmdb",
					"job_en":  "job",
				},
			}
		})

		It("ContainsID", func() {
			assert.False(GinkgoT(), a.ContainsID("hello"))
			assert.True(GinkgoT(), a.ContainsID("cmdb"))
		})

		It("ContainsName", func() {
			assert.False(GinkgoT(), a.ContainsName("hello"))
			assert.True(GinkgoT(), a.ContainsName("cmdb_n"))
		})

		It("ContainsNameEn", func() {
			assert.False(GinkgoT(), a.ContainsNameEn("hello"))
			assert.True(GinkgoT(), a.ContainsNameEn("job_en"))
		})

		It("ContainsNameExcludeSelf", func() {
			assert.False(GinkgoT(), a.ContainsNameExcludeSelf("hello", "cmdb"))
			assert.False(GinkgoT(), a.ContainsNameExcludeSelf("cmdb_n", "cmdb"))
			assert.True(GinkgoT(), a.ContainsNameExcludeSelf("job_n", "cmdb"))
		})

		It("ContainsNameEnExcludeSelf", func() {
			assert.False(GinkgoT(), a.ContainsNameEnExcludeSelf("hello", "cmdb"))
			assert.False(GinkgoT(), a.ContainsNameEnExcludeSelf("cmdb_en", "cmdb"))
			assert.True(GinkgoT(), a.ContainsNameEnExcludeSelf("job_en", "cmdb"))
		})
	})
})
