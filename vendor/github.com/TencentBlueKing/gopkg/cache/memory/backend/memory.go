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

package backend

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const (
	DefaultCleanupInterval = 5 * time.Minute
)

// RandomExtraExpirationDurationFunc is the type of the function generate extra expiration duration
type RandomExtraExpirationDurationFunc func() time.Duration

// NewTTLCache create cache with expiration and cleanup interval,
// if cleanupInterval is 0, will use DefaultCleanupInterval
func newTTLCache(expiration time.Duration, cleanupInterval time.Duration) *gocache.Cache {
	if cleanupInterval == 0 {
		cleanupInterval = DefaultCleanupInterval
	}

	return gocache.New(expiration, cleanupInterval)
}

// NewMemoryBackend create memory backend
// - name: the name of the backend
// - expiration: the expiration duration of the cache
// - randomExtraExpirationFunc: the function generate extra expiration duration
func NewMemoryBackend(
	name string,
	expiration time.Duration,
	randomExtraExpirationFunc RandomExtraExpirationDurationFunc,
) *MemoryBackend {
	cleanupInterval := expiration + (5 * time.Minute)

	return &MemoryBackend{
		name:                      name,
		cache:                     newTTLCache(expiration, cleanupInterval),
		defaultExpiration:         expiration,
		randomExtraExpirationFunc: randomExtraExpirationFunc,
	}
}

// MemoryBackend is the backend for memory cache
type MemoryBackend struct {
	name  string
	cache *gocache.Cache

	defaultExpiration         time.Duration
	randomExtraExpirationFunc RandomExtraExpirationDurationFunc
}

// Set sets value to cache with key and expiration
func (c *MemoryBackend) Set(key string, value interface{}, duration time.Duration) {
	if duration == time.Duration(0) {
		duration = c.defaultExpiration
	}

	if c.randomExtraExpirationFunc != nil {
		duration += c.randomExtraExpirationFunc()
	}

	c.cache.Set(key, value, duration)
}

// Get gets value by key from the cache
func (c *MemoryBackend) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

// Delete deletes value by key from the cache
func (c *MemoryBackend) Delete(key string) error {
	c.cache.Delete(key)
	return nil
}
