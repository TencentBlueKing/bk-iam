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

	"iam/pkg/config"
	"iam/pkg/service/types"
)

var _ = Describe("transfer", func() {
	Describe("mergeSubjectActionGroup", func() {
		var events []types.GroupAlterEvent
		BeforeEach(func() {
			events = []types.GroupAlterEvent{
				{
					GroupPK:    1,
					SubjectPKs: []int64{1, 2, 3},
					ActionPKs:  []int64{1, 2, 3},
				},
				{
					GroupPK:    2,
					SubjectPKs: []int64{1, 3},
					ActionPKs:  []int64{1, 2, 3},
				},
				{
					GroupPK:    3,
					SubjectPKs: []int64{1},
					ActionPKs:  []int64{2, 3, 4},
				},
			}
		})

		It("ok", func() {
			subjectActionGroupMap := mergeSubjectActionGroup(events)
			assert.Len(GinkgoT(), subjectActionGroupMap, 10)
			assert.Equal(GinkgoT(), subjectActionGroupMap[subjectAction{
				SubjectPK: 1,
				ActionPK:  2,
			}].Size(), 3)
			assert.Equal(GinkgoT(), subjectActionGroupMap[subjectAction{
				SubjectPK: 1,
				ActionPK:  4,
			}].Size(), 1)

			assert.Equal(GinkgoT(), subjectActionGroupMap[subjectAction{
				SubjectPK: 2,
				ActionPK:  2,
			}].Size(), 1)
		})
	})
	Describe("convertToSubjectActionAlterEvent", func() {
		var events []types.GroupAlterEvent
		BeforeEach(func() {
			config.MaxMessageGeneratedCountPreSubjectActionAlterEvent = 3
			events = []types.GroupAlterEvent{
				{
					GroupPK:    1,
					SubjectPKs: []int64{1, 2, 3},
					ActionPKs:  []int64{1, 2, 3},
				},
				{
					GroupPK:    2,
					SubjectPKs: []int64{1, 3},
					ActionPKs:  []int64{1, 2, 3},
				},
				{
					GroupPK:    3,
					SubjectPKs: []int64{1},
					ActionPKs:  []int64{2, 3, 4},
				},
			}
		})

		It("ok", func() {
			subjectActionAlterEvents := convertToSubjectActionAlterEvent(events)
			assert.Len(GinkgoT(), subjectActionAlterEvents, 4)
		})
	})
})
