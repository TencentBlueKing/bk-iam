/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util

import "strings"

// StringSet ...
type StringSet struct {
	Data map[string]struct{}
}

// NewStringSet ...
func NewStringSet() *StringSet {
	return &StringSet{
		Data: map[string]struct{}{},
	}
}

// NewStringSetWithValues ...
func NewStringSetWithValues(keys []string) *StringSet {
	set := &StringSet{
		Data: map[string]struct{}{},
	}
	for _, key := range keys {
		set.Add(key)
	}
	return set
}

// NewFixedLengthStringSet ...
func NewFixedLengthStringSet(length int) *StringSet {
	return &StringSet{
		Data: make(map[string]struct{}, length),
	}
}

// Has ...
func (s *StringSet) Has(key string) bool {
	_, ok := s.Data[key]
	return ok
}

// Add ...
func (s *StringSet) Add(key string) {
	s.Data[key] = struct{}{}
}

// Append ...
func (s *StringSet) Append(keys ...string) {
	for _, key := range keys {
		s.Data[key] = struct{}{}
	}
}

// Size ...
func (s *StringSet) Size() int {
	return len(s.Data)
}

// ToSlice ...
func (s *StringSet) ToSlice() []string {
	l := make([]string, 0, len(s.Data))
	for k := range s.Data {
		l = append(l, k)
	}
	return l
}

// ToString ...
func (s *StringSet) ToString(sep string) string {
	l := s.ToSlice()
	return strings.Join(l, sep)
}

// Diff 求差集
func (s *StringSet) Diff(b *StringSet) *StringSet {
	diffSet := NewStringSet()

	for k := range s.Data {
		if !b.Has(k) {
			diffSet.Add(k)
		}
	}
	return diffSet
}

// SplitStringToSet ...
func SplitStringToSet(s string, sep string) *StringSet {
	if s == "" {
		return &StringSet{Data: map[string]struct{}{}}
	}

	data := map[string]struct{}{}
	keys := strings.Split(s, sep)
	for _, key := range keys {
		data[key] = struct{}{}
	}
	return &StringSet{Data: data}
}

// Int64Set ...
type Int64Set struct {
	Data map[int64]struct{}
}

// NewInt64Set ...
func NewInt64Set() *Int64Set {
	return &Int64Set{
		Data: map[int64]struct{}{},
	}
}

// NewInt64SetWithValues ...
func NewInt64SetWithValues(keys []int64) *Int64Set {
	set := &Int64Set{
		Data: map[int64]struct{}{},
	}
	for _, key := range keys {
		set.Add(key)
	}
	return set
}

// NewFixedLengthInt64Set ...
func NewFixedLengthInt64Set(length int) *Int64Set {
	return &Int64Set{
		Data: make(map[int64]struct{}, length),
	}
}

// Has ...
func (s *Int64Set) Has(key int64) bool {
	_, ok := s.Data[key]
	return ok
}

// Add ...
func (s *Int64Set) Add(key int64) {
	s.Data[key] = struct{}{}
}

// Append ...
func (s *Int64Set) Append(keys ...int64) {
	for _, key := range keys {
		s.Data[key] = struct{}{}
	}
}

// Size ...
func (s *Int64Set) Size() int {
	return len(s.Data)
}

// ToSlice ...
func (s *Int64Set) ToSlice() []int64 {
	l := make([]int64, 0, len(s.Data))
	for k := range s.Data {
		l = append(l, k)
	}
	return l
}
