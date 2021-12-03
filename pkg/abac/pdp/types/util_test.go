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
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Util", func() {
	Describe("InterfaceToPolicyCondition", func() {
		It("ok", func() {
			expected := PolicyCondition{
				"StringEqual": {
					"id": {"1", "2"},
				},
			}

			value := map[string]interface{}{
				"StringEqual": map[string]interface{}{
					"id": []interface{}{"1", "2"},
				},
			}
			c, err := InterfaceToPolicyCondition(value)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c)
		})

		It("invalid value", func() {
			_, err := InterfaceToPolicyCondition("abc")
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), ErrTypeAssertFail, err)
		})

		It("invalid attribute, should be an array", func() {
			value := map[string]interface{}{
				"StringEqual": map[string]interface{}{
					"id": "invalid",
				},
			}
			_, err := InterfaceToPolicyCondition(value)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), ErrTypeAssertFail, err)
		})

		It("invalid operatorMap", func() {
			value := map[string]interface{}{
				"StringEqual": "invalid",
			}
			_, err := InterfaceToPolicyCondition(value)
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), ErrTypeAssertFail, err)
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
