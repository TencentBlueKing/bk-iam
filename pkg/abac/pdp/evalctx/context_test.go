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
	"fmt"
	"strconv"
	"testing"
	"time"

	"iam/pkg/abac/pdp/condition"
	pdptypes "iam/pkg/abac/pdp/types"

	. "github.com/onsi/ginkgo/v2"
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

	Describe("SetEnv", func() {
		It("ok", func() {
			c.SetEnv(map[string]interface{}{"ts": 123})

			ts, err := c.GetAttr("iam._bk_iam_env_.ts")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), 123, ts)
		})
	})

	Describe("UnsetEnv", func() {
		It("ok", func() {
			c.SetEnv(map[string]interface{}{"ts": 123})

			ts, err := c.GetAttr("iam._bk_iam_env_.ts")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), 123, ts)

			c.UnsetEnv()

			assert.False(GinkgoT(), c.HasResource("iam._bk_iam_env_"))
		})
	})

	Describe("HasEnv", func() {
		It("ok", func() {
			assert.False(GinkgoT(), c.HasEnv())
			c.SetEnv(map[string]interface{}{"ts": 123})
			assert.True(GinkgoT(), c.HasEnv())
		})
	})

	Describe("InitEnvironments", func() {

		var noEnvCond condition.Condition
		var notEnvTimeCond condition.Condition
		var envTimeCond condition.Condition
		BeforeEach(func() {
			// init the cache
			localTimeEnvsCache = gocache.New(10*time.Second, 30*time.Second)

			c1 := pdptypes.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam.system": []interface{}{"linux"}}},
						map[string]interface{}{"StringPrefix": map[string]interface{}{"iam.path": []interface{}{"/biz,1/"}}},
					},
				},
			}
			noEnvCond, _ = condition.NewConditionFromPolicyCondition(c1)

			c2 := pdptypes.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam.host.system": []interface{}{"linux"}}},
						map[string]interface{}{"StringPrefix": map[string]interface{}{"iam.host.path": []interface{}{"/biz,1/"}}},
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam._bk_iam_env_.system": []interface{}{"iam"}}},
					},
				},
			}
			notEnvTimeCond, _ = condition.NewConditionFromPolicyCondition(c2)

			c3 := pdptypes.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam.host.system": []interface{}{"linux"}}},
						map[string]interface{}{"StringPrefix": map[string]interface{}{"iam.host.path": []interface{}{"/biz,1/"}}},
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam._bk_iam_env_.tz": []interface{}{"Asia/Shanghai"}}},
						map[string]interface{}{"NumericLt": map[string]interface{}{"iam._bk_iam_env_.hms": []interface{}{163630}}},
					},
				},
			}
			envTimeCond, _ = condition.NewConditionFromPolicyCondition(c3)
		})
		AfterEach(func() {
			c.UnsetEnv()
		})

		It("has no envs", func() {
			err := c.InitEnvironments(noEnvCond, time.Now())
			assert.NoError(GinkgoT(), err)

			assert.False(GinkgoT(), c.HasResource("iam._bk_iam_env_"))
		})

		It("has env, not time-related", func() {
			err := c.InitEnvironments(notEnvTimeCond, time.Now())
			assert.NoError(GinkgoT(), err)

			assert.False(GinkgoT(), c.HasResource("iam._bk_iam_env_"))
		})

		It("has time-related env, ok", func() {
			tz := "Asia/Shanghai"

			loc, _ := time.LoadLocation(tz)
			t, _ := time.ParseInLocation("2006-01-02 15:04:05 Z0700 MST", "2021-12-03 15:54:06 +0800 CST", loc)
			hms := int64(155406)

			fmt.Println("print the c, is nil? ", c, c == nil)

			err := c.InitEnvironments(envTimeCond, t)
			assert.NoError(GinkgoT(), err)

			assert.True(GinkgoT(), c.HasResource("iam._bk_iam_env_"))

			tzA, err := c.GetAttr("iam._bk_iam_env_.tz")
			assert.Equal(GinkgoT(), "Asia/Shanghai", tzA)
			hmsA, err := c.GetAttr("iam._bk_iam_env_.hms")
			assert.Equal(GinkgoT(), hms, hmsA)
		})

		It("has time-related env, 2 tz", func() {
			c3 := pdptypes.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam._bk_iam_env_.tz": []interface{}{"Asia/Shanghai", "America/New_York"}}},
						map[string]interface{}{"NumericLt": map[string]interface{}{"iam._bk_iam_env_.hms": []interface{}{163630}}},
					},
				},
			}
			cond, _ := condition.NewConditionFromPolicyCondition(c3)
			err := c.InitEnvironments(cond, time.Now())
			assert.Error(GinkgoT(), err)
		})

		It("has time-related env, tz wrong type", func() {
			c3 := pdptypes.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam._bk_iam_env_.tz": []interface{}{123}}},
						map[string]interface{}{"NumericLt": map[string]interface{}{"iam._bk_iam_env_.hms": []interface{}{163630}}},
					},
				},
			}
			cond, _ := condition.NewConditionFromPolicyCondition(c3)
			err := c.InitEnvironments(cond, time.Now())
			assert.Error(GinkgoT(), err)
		})

		It("has time-related env, tz wrong", func() {
			c3 := pdptypes.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{"StringEquals": map[string]interface{}{"iam._bk_iam_env_.tz": []interface{}{"wrong"}}},
						map[string]interface{}{"NumericLt": map[string]interface{}{"iam._bk_iam_env_.hms": []interface{}{163630}}},
					},
				},
			}
			cond, _ := condition.NewConditionFromPolicyCondition(c3)
			err := c.InitEnvironments(cond, time.Now())
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("envs time", func() {

		var tz string
		var t time.Time
		var hms int64
		BeforeEach(func() {
			tz = "Asia/Shanghai"

			loc, _ := time.LoadLocation(tz)
			t, _ = time.ParseInLocation("2006-01-02 15:04:05 Z0700 MST", "2021-12-03 15:54:06 +0800 CST", loc)
			hms = int64(155406)
		})

		Describe("GenTimeEnvsFromCache", func() {
			It("ok", func() {
				envs, err := GenTimeEnvsFromCache(tz, t)
				assert.NoError(GinkgoT(), err)

				assert.Len(GinkgoT(), envs, 2)
				assert.Equal(GinkgoT(), tz, envs["tz"])
				assert.Equal(GinkgoT(), hms, envs["hms"])

				envs2, err := GenTimeEnvsFromCache(tz, t)
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), envs, envs2)
			})

			It("fail", func() {
				tz := "Wrong"
				_, err := GenTimeEnvsFromCache(tz, time.Now())
				assert.Error(GinkgoT(), err)
			})
		})

		Describe("genTimeEnvs", func() {
			It("ok", func() {
				envs, err := genTimeEnvs(tz, t)
				assert.NoError(GinkgoT(), err)

				assert.Len(GinkgoT(), envs, 2)
				assert.Equal(GinkgoT(), tz, envs["tz"])
				assert.Equal(GinkgoT(), hms, envs["hms"])
			})

			It("fail", func() {
				tz := "Wrong"
				_, err := genTimeEnvs(tz, time.Now())
				assert.Error(GinkgoT(), err)
			})

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
