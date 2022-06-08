/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"

	"iam/pkg/service/types"
)

var _ = Describe("SubjectDetail", func() {
	Describe("msgpack marshal/unmarshal", func() {
		It("normal data, ok", func() {
			s := types.SubjectDetail{
				DepartmentPKs: []int64{1, 3, 2},
				SubjectGroups: []types.ThinSubjectGroup{
					{
						1,
						123,
					},
					{
						2,
						456,
					},
				},
			}
			b, err := msgpack.Marshal(&s)
			assert.NoError(GinkgoT(), err)

			var us types.SubjectDetail
			err = msgpack.Unmarshal(b, &us)
			assert.NoError(GinkgoT(), err)

			assert.Len(GinkgoT(), us.DepartmentPKs, 3)
			assert.Equal(GinkgoT(), []int64{1, 3, 2}, us.DepartmentPKs)

			assert.Len(GinkgoT(), us.SubjectGroups, 2)
			assert.Equal(GinkgoT(), int64(1), us.SubjectGroups[0].PK)
			assert.Equal(GinkgoT(), int64(123), us.SubjectGroups[0].PolicyExpiredAt)
			assert.Equal(GinkgoT(), int64(2), us.SubjectGroups[1].PK)
			assert.Equal(GinkgoT(), int64(456), us.SubjectGroups[1].PolicyExpiredAt)
		})

		It("data no departments, ok", func() {
			s := types.SubjectDetail{
				DepartmentPKs: []int64{},
				SubjectGroups: []types.ThinSubjectGroup{
					{
						1,
						123,
					},
					{
						2,
						456,
					},
				},
			}
			b, err := msgpack.Marshal(&s)
			assert.NoError(GinkgoT(), err)

			var us types.SubjectDetail
			err = msgpack.Unmarshal(b, &us)
			assert.NoError(GinkgoT(), err)

			assert.Len(GinkgoT(), us.DepartmentPKs, 0)
			assert.Equal(GinkgoT(), []int64{}, us.DepartmentPKs)

			assert.Len(GinkgoT(), us.SubjectGroups, 2)
			assert.Equal(GinkgoT(), int64(1), us.SubjectGroups[0].PK)
			assert.Equal(GinkgoT(), int64(123), us.SubjectGroups[0].PolicyExpiredAt)
			assert.Equal(GinkgoT(), int64(2), us.SubjectGroups[1].PK)
			assert.Equal(GinkgoT(), int64(456), us.SubjectGroups[1].PolicyExpiredAt)
		})

		It("data not subject groups, ok", func() {
			s := types.SubjectDetail{
				DepartmentPKs: []int64{1, 2, 3},
				SubjectGroups: []types.ThinSubjectGroup{},
			}
			b, err := msgpack.Marshal(&s)
			assert.NoError(GinkgoT(), err)

			var us types.SubjectDetail
			err = msgpack.Unmarshal(b, &us)
			assert.NoError(GinkgoT(), err)

			assert.Len(GinkgoT(), us.DepartmentPKs, 3)
			assert.Equal(GinkgoT(), []int64{1, 2, 3}, us.DepartmentPKs)

			assert.Len(GinkgoT(), us.SubjectGroups, 0)
		})

		It("data both empty, ok", func() {
			s := types.SubjectDetail{
				DepartmentPKs: []int64{},
				SubjectGroups: []types.ThinSubjectGroup{},
			}
			b, err := msgpack.Marshal(&s)
			assert.NoError(GinkgoT(), err)

			var us types.SubjectDetail
			err = msgpack.Unmarshal(b, &us)
			assert.NoError(GinkgoT(), err)

			assert.Len(GinkgoT(), us.DepartmentPKs, 0)
			assert.Equal(GinkgoT(), []int64{}, us.DepartmentPKs)

			assert.Len(GinkgoT(), us.SubjectGroups, 0)
		})

		It("data both nil, ok", func() {
			s := types.SubjectDetail{}
			b, err := msgpack.Marshal(&s)
			assert.NoError(GinkgoT(), err)

			var us types.SubjectDetail
			err = msgpack.Unmarshal(b, &us)
			assert.NoError(GinkgoT(), err)

			assert.Len(GinkgoT(), us.DepartmentPKs, 0)
			assert.Equal(GinkgoT(), []int64{}, us.DepartmentPKs)

			assert.Len(GinkgoT(), us.SubjectGroups, 0)
		})
	})
})

func BenchmarkThinSubjectDetail(b *testing.B) {
	type SubjectDetail struct {
		DepartmentPKs []int64                  `json:"department_pks" msgpack:"dps"`
		SubjectGroup  []types.ThinSubjectGroup `json:"subject_group" msgpack:"sg"`
	}

	a := SubjectDetail{
		DepartmentPKs: []int64{1, 2, 3, 4, 5, 6, 7, 9, 10},
		SubjectGroup: []types.ThinSubjectGroup{
			{
				PK:              123,
				PolicyExpiredAt: 16230503601,
			},
			{
				PK:              1123,
				PolicyExpiredAt: 16230503601,
			},
			{
				PK:              11123,
				PolicyExpiredAt: 1623050389,
			},
			{
				PK:              11123,
				PolicyExpiredAt: 1623050389,
			},
			{
				PK:              11123,
				PolicyExpiredAt: 1623050389,
			},
		},
	}

	bs, _ := msgpack.Marshal(&a)
	// fmt.Println("size:", len(bs), err)

	var x SubjectDetail
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(bs, &x)
	}
}

func BenchmarkThinSubjectDetailCustomEncodeDecode(b *testing.B) {
	a := types.SubjectDetail{
		DepartmentPKs: []int64{1, 2, 3, 4, 5, 6, 7, 9, 10},
		SubjectGroups: []types.ThinSubjectGroup{
			{
				PK:              123,
				PolicyExpiredAt: 16230503601,
			},
			{
				PK:              1123,
				PolicyExpiredAt: 16230503601,
			},
			{
				PK:              11123,
				PolicyExpiredAt: 1623050389,
			},
			{
				PK:              11123,
				PolicyExpiredAt: 1623050389,
			},
			{
				PK:              11123,
				PolicyExpiredAt: 1623050389,
			},
		},
	}

	bs, _ := msgpack.Marshal(&a)
	// fmt.Println("size:", len(bs), err)

	var x types.SubjectDetail
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(bs, &x)
	}
	// fmt.Printf("+%v", x)
}
