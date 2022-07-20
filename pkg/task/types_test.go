/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package task

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("task", func() {
	Describe("Message", func() {
		var m Message
		BeforeEach(func() {
			m = Message{
				GroupPK:   1,
				ActionPK:  2,
				SubjectPK: 3,
			}
		})

		It("UniqueID", func() {
			assert.Equal(GinkgoT(), "1:2:3", m.UniqueID())
		})

		It("String", func() {
			s, err := m.String()
			assert.Nil(GinkgoT(), err)
			assert.Len(GinkgoT(), s, 43)
		})

		It("NewMessageFromString", func() {
			s := `{"group_pk":1,"action_pk":2,"subject_pk":3}`
			m2, err := NewMessageFromString(s)
			assert.Nil(GinkgoT(), err)
			assert.Equal(GinkgoT(), m, m2)
		})
	})
})
