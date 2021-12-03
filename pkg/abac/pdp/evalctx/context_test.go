/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package evalctx

import (
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	gocache "github.com/patrickmn/go-cache"
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

})

func BenchmarkGenEnvsInReal(b *testing.B) {
	tz := "Asia/Shanghai"
	currentTime := time.Now()

	for i := 0; i < b.N; i++ {
		genTimeEnvs(tz, currentTime)
	}
}
func BenchmarkGenEnvsFromSyncMap(b *testing.B) {
	tz := "Asia/Shanghai"
	currentTime := time.Now()

	m := gocache.New(10*time.Second, 20*time.Second)
	// m := sync.Map{}
	// for _, x := range a {
	// 	m.Store(x, strconv.FormatInt(x, 10))
	// }

	for i := 0; i < b.N; i++ {
		key := tz + strconv.FormatInt(currentTime.Unix(), 10)
		// key := fmt.Sprintf("%s%d", tz, currentTime.Unix())

		_, ok := m.Get(key)
		if !ok {
			envs, err := genTimeEnvs(tz, currentTime)
			if err == nil {
				m.Set(key, envs, 0)
				// m.Store(key, envs)
			}
		}
	}
}
