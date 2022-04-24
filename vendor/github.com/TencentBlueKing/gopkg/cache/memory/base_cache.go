/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-gopkg available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package memory

import (
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/cache/memory/backend"
)

const EmptyCacheExpiration = 5 * time.Second

// NewBaseCache creates a new baseCache instance which implements Cache interface.
func NewBaseCache(disabled bool, retrieveFunc RetrieveFunc, backend backend.Backend) Cache {
	return &BaseCache{
		backend:      backend,
		disabled:     disabled,
		retrieveFunc: retrieveFunc,
	}
}

// BaseCache is a cache which retrieves data from the backend and stores it in the cache.
type BaseCache struct {
	backend backend.Backend

	disabled     bool
	retrieveFunc RetrieveFunc
	g            singleflight.Group
}

// EmptyCache is a placeholder for the missing key
type EmptyCache struct {
	err error
}

// Exists returns true if the cache has a value for the given key.
func (c *BaseCache) Exists(key cache.Key) bool {
	k := key.Key()
	_, ok := c.backend.Get(k)
	return ok
}

// Get will get the key from cache, if missing, will call the retrieveFunc to get the data, add to cache, then return
func (c *BaseCache) Get(key cache.Key) (interface{}, error) {
	// 1. if cache is disabled, fetch and return
	if c.disabled {
		value, err := c.retrieveFunc(key)
		if err != nil {
			return nil, err
		}
		return value, nil
	}

	k := key.Key()

	// 2. get from cache
	value, ok := c.backend.Get(k)
	if ok {
		// if retrieve fail from retrieveFunc
		if emptyCache, isEmptyCache := value.(EmptyCache); isEmptyCache {
			return nil, emptyCache.err
		}
		return value, nil
	}

	// 3. if not exists in cache, retrieve it
	return c.doRetrieve(key)
}

// doRetrieve will retrieve the real data from database, redis, apis, etc.
func (c *BaseCache) doRetrieve(k cache.Key) (interface{}, error) {
	key := k.Key()

	// 3.2 fetch
	value, err, _ := c.g.Do(key, func() (interface{}, error) {
		return c.retrieveFunc(k)
	})

	if err != nil {
		// ! if error, cache it too, make it short enough(5s)
		c.backend.Set(key, EmptyCache{err: err}, EmptyCacheExpiration)
		return nil, err
	}

	// 4. set value to cache, use default expiration
	c.backend.Set(key, value, 0)

	return value, nil
}

// Set will set key-value into cache.
func (c *BaseCache) Set(key cache.Key, data interface{}) {
	k := key.Key()
	c.backend.Set(k, data, 0)
}

// Delete deletes the value from the cache for the given key.
func (c *BaseCache) Delete(key cache.Key) error {
	k := key.Key()
	return c.backend.Delete(k)
}

// DirectGet will get key from cache, without calling the retrieveFunc
func (c *BaseCache) DirectGet(key cache.Key) (interface{}, bool) {
	k := key.Key()
	return c.backend.Get(k)
}

// Disabled returns true if the cache is disabled.
func (c *BaseCache) Disabled() bool {
	return c.disabled
}

// TODO: 这里需要实现所有类型的 GetXXXX

// GetString returns a string representation of the value for the given key.
// will error if the type is not a string.
func (c *BaseCache) GetString(k cache.Key) (string, error) {
	value, err := c.Get(k)
	if err != nil {
		return "", err
	}

	v, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("not a string value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetBool returns a bool representation of the value for the given key.
// will error if the type is not a bool.
func (c *BaseCache) GetBool(k cache.Key) (bool, error) {
	value, err := c.Get(k)
	if err != nil {
		return false, err
	}

	v, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("not a string value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

var defaultZeroTime = time.Time{}

// GetTime returns a time representation of the value for the given key.
// will error if the type is not an time.Time.
func (c *BaseCache) GetTime(k cache.Key) (time.Time, error) {
	value, err := c.Get(k)
	if err != nil {
		return defaultZeroTime, err
	}

	v, ok := value.(time.Time)
	if !ok {
		return defaultZeroTime, fmt.Errorf("not a string value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetInt returns an int representation of the value for the given key.
// will error if the type is not an int.
func (c *BaseCache) GetInt(k cache.Key) (int, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("not a int value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetInt8 returns an int8 representation of the value for the given key.
// will error if the type is not an int8.
func (c *BaseCache) GetInt8(k cache.Key) (int8, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(int8)
	if !ok {
		return 0, fmt.Errorf("not a int8 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetInt16 returns an int16 representation of the value for the given key.
// will error if the type is not an int16.
func (c *BaseCache) GetInt16(k cache.Key) (int16, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(int16)
	if !ok {
		return 0, fmt.Errorf("not a int16 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetInt32 returns an int32 representation of the value for the given key.
// will error if the type is not an int32.
func (c *BaseCache) GetInt32(k cache.Key) (int32, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(int32)
	if !ok {
		return 0, fmt.Errorf("not a int32 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetInt64 returns an int64 representation of the value for the given key.
// will error if the type is not an int64.
func (c *BaseCache) GetInt64(k cache.Key) (int64, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("not a int64 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetUint returns an uint representation of the value for the given key.
// will error if the type is not an uint.
func (c *BaseCache) GetUint(k cache.Key) (uint, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(uint)
	if !ok {
		return 0, fmt.Errorf("not a uint value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetUint8 returns an uint8 representation of the value for the given key.
// will error if the type is not an uint8.
func (c *BaseCache) GetUint8(k cache.Key) (uint8, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(uint8)
	if !ok {
		return 0, fmt.Errorf("not a uint8 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetUint16 returns an uint16 representation of the value for the given key.
// will error if the type is not an uint16.
func (c *BaseCache) GetUint16(k cache.Key) (uint16, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(uint16)
	if !ok {
		return 0, fmt.Errorf("not a uint16 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetUint32 returns an uint32 representation of the value for the given key.
// will error if the type is not an uint32.
func (c *BaseCache) GetUint32(k cache.Key) (uint32, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(uint32)
	if !ok {
		return 0, fmt.Errorf("not a uint32 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetUint64 returns an uint64 representation of the value for the given key.
// will error if the type is not an uint64.
func (c *BaseCache) GetUint64(k cache.Key) (uint64, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(uint64)
	if !ok {
		return 0, fmt.Errorf("not a uint64 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetFloat32 returns a float32 representation of the value for the given key.
// will error if the type is not a float32.
func (c *BaseCache) GetFloat32(k cache.Key) (float32, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(float32)
	if !ok {
		return 0, fmt.Errorf("not a float32 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}

// GetFloat64 returns a float64 representation of the value for the given key.
// will error if the type is not a float64.
func (c *BaseCache) GetFloat64(k cache.Key) (float64, error) {
	value, err := c.Get(k)
	if err != nil {
		return 0, err
	}

	v, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("not a float64 value. key=%s, value=%v(%T)", k.Key(), value, value)
	}
	return v, nil
}
