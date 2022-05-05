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
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("types", func() {

	Describe("clearStructFields", func() {
		It("clearStructFields", func() {
			body := &authRequest{
				baseRequest: baseRequest{
					System: "bk_test",
					Subject: subject{
						Type: "test",
						ID:   "test",
					},
				},
				Resources: []resource{
					{
						System:    "bk_test",
						ID:        "test",
						Type:      "test",
						Attribute: map[string]interface{}{},
					},
				},
				Action: action{
					ID: "test",
				},
			}
			clearStructFields(body)
			assert.Equal(GinkgoT(), "", body.System)
			assert.Equal(GinkgoT(), "", body.Subject.ID)
			assert.Equal(GinkgoT(), "", body.Subject.Type)
			assert.Len(GinkgoT(), body.Resources, 0)
			assert.Equal(GinkgoT(), "", body.Action.ID)
		})
	})

	Describe("requestBodyPool", func() {
		var p *requestBodyPool[authRequest]
		BeforeEach(func() {
			p = newRequestBodyPool[authRequest]()
		})

		It("newRequestBodyPool", func() {
			assert.NotNil(GinkgoT(), p)
		})

		It("get", func() {
			body := p.get()
			assert.Equal(GinkgoT(), "", body.System)
			assert.Equal(GinkgoT(), "", body.Subject.ID)
			assert.Equal(GinkgoT(), "", body.Subject.Type)
			assert.Len(GinkgoT(), body.Resources, 0)
			assert.Equal(GinkgoT(), "", body.Action.ID)
		})

		It("set", func() {
			body := &authRequest{
				baseRequest: baseRequest{
					System: "bk_test",
					Subject: subject{
						Type: "test",
						ID:   "test",
					},
				},
				Resources: []resource{
					{
						System:    "bk_test",
						ID:        "test",
						Type:      "test",
						Attribute: map[string]interface{}{},
					},
				},
				Action: action{
					ID: "test",
				},
			}
			p.put(body)
			body1 := p.get()

			assert.True(GinkgoT(), body == body1)

			assert.Equal(GinkgoT(), "", body1.System)
			assert.Equal(GinkgoT(), "", body1.Subject.ID)
			assert.Equal(GinkgoT(), "", body1.Subject.Type)
			assert.Len(GinkgoT(), body1.Resources, 0)
			assert.Equal(GinkgoT(), "", body1.Action.ID)
		})
	})
})
