/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-gopkg available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cache

import "fmt"

// Key is the type for the key of a cache entry.
// a struct-like object implements the key interface, so it can be used as a key in a cache.
type Key interface {
	Key() string
}

// StringKey is a string key.
type StringKey struct {
	key string
}

// NewStringKey creates a new StringKey
func NewStringKey(key string) StringKey {
	return StringKey{
		key: key,
	}
}

// Key returns the key.
func (s StringKey) Key() string {
	return s.key
}

// IntKey is an int key.
type IntKey struct {
	key int
}

// NewIntKey creates a new IntKey
func NewIntKey(key int) IntKey {
	return IntKey{
		key: key,
	}
}

// Key returns the key.
func (k IntKey) Key() string {
	return fmt.Sprintf("%d", k.key)
}

// Int64Key is an int64 key.
type Int64Key struct {
	key int64
}

// NewInt64Key creates a new Int64Key
func NewInt64Key(key int64) Int64Key {
	return Int64Key{
		key: key,
	}
}

// Key returns the key.
func (k Int64Key) Key() string {
	return fmt.Sprintf("%d", k.key)
}

// UintKey is an uint key.
type UintKey struct {
	key uint
}

// NewUintKey creates a new UintKey
func NewUintKey(key uint) UintKey {
	return UintKey{
		key: key,
	}
}

// Key returns the key.
func (k UintKey) Key() string {
	return fmt.Sprintf("%d", k.key)
}

// Uint64Key is an uint64 key.
type Uint64Key struct {
	key uint64
}

// NewUint64Key creates a new Uint64Key
func NewUint64Key(key uint64) Uint64Key {
	return Uint64Key{
		key: key,
	}
}

// Key returns the key.
func (k Uint64Key) Key() string {
	return fmt.Sprintf("%d", k.key)
}
