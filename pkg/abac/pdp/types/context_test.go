/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
)

var _ = Describe("Context", func() {

	Describe("NewExprContext", func() {
		It("ok", func() {
			ctx := &request.Request{}
			ec := NewExprContext(ctx)
			assert.NotNil(GinkgoT(), ec)
		})

	})

	var c *ExprContext
	BeforeEach(func() {
		request := &request.Request{
			System: "iam",
			Subject: types.Subject{
				Type: "user",
				ID:   "admin",
			},
			Action: types.Action{
				ID: "execute_job",
			},
			Resources: []types.Resource{
				{

					System:    "iam",
					Type:      "job",
					ID:        "job1",
					Attribute: map[string]interface{}{"key": "value1"},
				},
			},
		}
		c = NewExprContext(request)
	})

	Describe("GetAttr", func() {
		It("ok", func() {
			a, err := c.GetAttr("iam.job.id")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "job1", a)
		})

	})
})
