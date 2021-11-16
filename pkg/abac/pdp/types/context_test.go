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
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
)

var _ = Describe("Context", func() {

	var req *request.Request
	var c *EvalContext
	BeforeEach(func() {
		req = &request.Request{
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
		c = NewEvalContext(req)
	})

	Describe("NewEvalContext", func() {
		It("no resources", func() {
			req := &request.Request{}
			ec := NewEvalContext(req)
			assert.NotNil(GinkgoT(), ec)

		})

		It("ok, has resource", func() {
			ec := NewEvalContext(req)
			assert.NotNil(GinkgoT(), ec)
		})

		It("ok, has resource, attribute nil", func() {
			req := &request.Request{
				Resources: []types.Resource{
					{
						ID:        "test",
						Attribute: nil,
					},
				},
			}
			ec := NewEvalContext(req)
			assert.NotNil(GinkgoT(), ec)
		})

	})

	Describe("GetAttr", func() {
		It("ok", func() {
			a, err := c.GetAttr("iam.job.id")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "job1", a)
		})

		It("miss", func() {
			a, err := c.GetAttr("bk_cmdb.job.id")
			assert.NoError(GinkgoT(), err)
			assert.Nil(GinkgoT(), a)
		})
	})

	Describe("HasResource", func() {
		It("ok", func() {
			assert.True(GinkgoT(), c.HasResource("iam.job"))
		})

		It("miss", func() {
			assert.False(GinkgoT(), c.HasResource("bk_cmdb.job"))

		})
	})

	Describe("Environment", func() {
		It("ok, has env", func() {
			assert.NotNil(GinkgoT(), c)

			// iam._bk_iam_env_:map[ts:1637044328]
			assert.True(GinkgoT(), c.HasResource(req.System+iamEnvSuffix))

			v, err := c.GetAttr(req.System + iamEnvSuffix + "." + envTimestamp)
			assert.Nil(GinkgoT(), err)
			assert.LessOrEqual(GinkgoT(), v.(int64), time.Now().Unix())
		})

		It("GetEnv", func() {
			env := c.GetEnv()
			assert.NotNil(GinkgoT(), env)
			assert.Contains(GinkgoT(), env, envTimestamp)
		})
	})

})
