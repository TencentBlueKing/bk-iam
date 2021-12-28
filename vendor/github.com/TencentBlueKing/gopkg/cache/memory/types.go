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

package memory

import (
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
)

// RetrieveFunc is the type of the retrieve function.
// it retrieves the value from database, redis, apis, etc.
type RetrieveFunc func(key cache.Key) (interface{}, error)

// Cache is the interface for the cache.
type Cache interface {
	Get(key cache.Key) (interface{}, error)
	Set(key cache.Key, data interface{})

	GetString(key cache.Key) (string, error)
	GetBool(key cache.Key) (bool, error)
	GetInt(key cache.Key) (int, error)
	GetInt8(key cache.Key) (int8, error)
	GetInt16(key cache.Key) (int16, error)
	GetInt32(key cache.Key) (int32, error)
	GetInt64(key cache.Key) (int64, error)
	GetUint(key cache.Key) (uint, error)
	GetUint8(key cache.Key) (uint8, error)
	GetUint16(key cache.Key) (uint16, error)
	GetUint32(key cache.Key) (uint32, error)
	GetUint64(key cache.Key) (uint64, error)
	GetFloat32(key cache.Key) (float32, error)
	GetFloat64(key cache.Key) (float64, error)
	GetTime(key cache.Key) (time.Time, error)

	Delete(key cache.Key) error
	Exists(key cache.Key) bool

	DirectGet(key cache.Key) (interface{}, bool)

	Disabled() bool
}
