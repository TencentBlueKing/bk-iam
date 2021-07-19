/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

var _ = Describe("Set", func() {

	Describe("StringSet", func() {

		Describe("New", func() {
			It("NewStringSet", func() {
				s := util.NewStringSet()

				assert.Len(GinkgoT(), s.Data, 0)
				assert.Equal(GinkgoT(), 0, s.Size())

				assert.False(GinkgoT(), s.Has("hello"))
			})

			It("NewStringSetWithValues", func() {
				s := util.NewStringSetWithValues([]string{"hello", "world"})

				assert.Len(GinkgoT(), s.Data, 2)
				assert.Equal(GinkgoT(), 2, s.Size())

				assert.True(GinkgoT(), s.Has("hello"))
			})

			It("NewFixedLengthStringSet", func() {
				s := util.NewFixedLengthStringSet(2)

				assert.Len(GinkgoT(), s.Data, 0)
				assert.Equal(GinkgoT(), 0, s.Size())
			})
		})

		Describe("Functions", func() {
			var s *util.StringSet

			BeforeEach(func() {
				s = util.NewStringSet()
				s.Add("hello")
			})

			It("Has", func() {
				assert.True(GinkgoT(), s.Has("hello"))
				assert.False(GinkgoT(), s.Has("world"))
			})

			It("Add", func() {
				s.Add("world")
				assert.True(GinkgoT(), s.Has("world"))
			})

			It("Append", func() {
				s.Append([]string{"abc", "def"}...)
				s.Append([]string{"def", "opq"}...)

				assert.Len(GinkgoT(), s.Data, 4)
				assert.Equal(GinkgoT(), 4, s.Size())

				assert.True(GinkgoT(), s.Has("abc"))
				assert.True(GinkgoT(), s.Has("def"))
				assert.True(GinkgoT(), s.Has("opq"))
			})

			It("Size", func() {
				assert.Equal(GinkgoT(), 1, s.Size())
			})

			It("ToSlice", func() {
				sli1 := s.ToSlice()
				assert.Len(GinkgoT(), sli1, 1)

				s.Add("world")
				sli2 := s.ToSlice()
				assert.Len(GinkgoT(), sli2, 2)
			})

			It("ToString", func() {
				s1 := s.ToString(",")
				assert.Equal(GinkgoT(), "hello", s1)

				s.Add("world")
				s2 := s.ToString(",")

				isEqual := s2 == "hello,world" || s2 == "world,hello"
				//assert.Equal(GinkgoT(), "hello,world", s2)
				assert.True(GinkgoT(), isEqual)

			})

			It("Diff", func() {
				// s = [hello, world]
				s.Add("world")

				// s1 = [world, foo]
				s1 := util.NewStringSetWithValues([]string{"world", "foo"})

				// the diff result
				s2 := s.Diff(s1)

				// the result = [hello]
				assert.Equal(GinkgoT(), 1, s2.Size())
				assert.True(GinkgoT(), s2.Has("hello"))
			})

		})

	})

	Describe("Int64Set", func() {

		var s *util.Int64Set

		BeforeEach(func() {
			s = util.NewInt64Set()
		})

		It("NewInt64Set", func() {
			//s := util.NewInt64Set()
			assert.Len(GinkgoT(), s.Data, 0)
			assert.Equal(GinkgoT(), 0, s.Size())
		})

		It("NewInt64SetWithValues", func() {
			s1 := util.NewInt64SetWithValues([]int64{123, 456})

			assert.Len(GinkgoT(), s1.Data, 2)
			assert.Equal(GinkgoT(), 2, s1.Size())

			assert.True(GinkgoT(), s1.Has(123))
		})

		It("NewFixedLengthInt64Set", func() {
			s1 := util.NewFixedLengthInt64Set(2)

			assert.Len(GinkgoT(), s1.Data, 0)
			assert.Equal(GinkgoT(), 0, s1.Size())
		})

		It("Add one, check size", func() {
			s.Add(123)

			assert.Len(GinkgoT(), s.Data, 1)
			assert.Equal(GinkgoT(), 1, s.Size())
		})

		It("Append", func() {
			s.Append([]int64{123, 456}...)
			s.Append([]int64{456, 789}...)

			assert.Len(GinkgoT(), s.Data, 3)
			assert.Equal(GinkgoT(), 3, s.Size())

			assert.True(GinkgoT(), s.Has(int64(123)))
			assert.True(GinkgoT(), s.Has(int64(456)))
			assert.True(GinkgoT(), s.Has(int64(789)))
		})

		It("Has 123", func() {
			assert.False(GinkgoT(), s.Has(123))
			s.Add(123)
			assert.True(GinkgoT(), s.Has(123))
		})

		It("ToSlice", func() {
			s.Add(123)
			sli1 := s.ToSlice()
			assert.Len(GinkgoT(), sli1, 1)

			s.Add(456)

			sli2 := s.ToSlice()
			assert.Len(GinkgoT(), sli2, 2)
		})

	})

	Describe("SplitStringToSet", func() {

		It("Empty string", func() {
			s := util.SplitStringToSet("", ",")
			assert.Equal(GinkgoT(), 0, s.Size())
		})

		It("Normal string a,b,c", func() {
			s := util.SplitStringToSet("a,b,c", ",")
			assert.Equal(GinkgoT(), 3, s.Size())
			assert.True(GinkgoT(), s.Has("b"))
		})
	})

})
