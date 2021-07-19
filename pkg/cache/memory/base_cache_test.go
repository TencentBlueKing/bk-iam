/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package memory

import (
	"errors"
	"testing"
	"time"

	"iam/pkg/cache"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"

	"iam/pkg/cache/memory/backend"
)

func retrieveTest(k cache.Key) (interface{}, error) {
	kStr := k.Key()
	switch kStr {
	case "a":
		return "1", nil
	case "b":
		return "2", nil
	case "error":
		return nil, errors.New("error")
	case "bool":
		return true, nil
	case "time":
		return time.Time{}, nil
	default:
		return "", nil
	}
}

func retrieveError(k cache.Key) (interface{}, error) {
	return nil, errors.New("test error")
}

func TestNewBaseCache(t *testing.T) {
	expiration := 5 * time.Minute

	be := backend.NewMemoryBackend("test", expiration)

	c := NewBaseCache(false, retrieveTest, be)

	// Disabled
	assert.False(t, c.Disabled())

	// get from cache
	aKey := cache.NewStringKey("a")
	x, err := c.Get(aKey)
	assert.NoError(t, err)
	assert.Equal(t, "1", x.(string))

	x, err = c.Get(aKey)
	assert.NoError(t, err)
	assert.Equal(t, "1", x.(string))

	assert.True(t, c.Exists(aKey))

	_, ok := c.DirectGet(aKey)
	assert.True(t, ok)

	// get string
	x, err = c.GetString(aKey)
	assert.NoError(t, err)
	assert.Equal(t, "1", x)

	// get bool
	boolKey := cache.NewStringKey("bool")
	x, err = c.GetBool(boolKey)
	assert.NoError(t, err)
	assert.Equal(t, true, x.(bool))

	// get time
	timeKey := cache.NewStringKey("time")
	x, err = c.GetTime(timeKey)
	assert.NoError(t, err)
	assert.IsType(t, time.Time{}, x)

	// get fail
	errorKey := cache.NewStringKey("error")
	x, err = c.Get(errorKey)
	assert.Error(t, err)
	assert.Nil(t, x)

	err1 := err

	// get fail twice
	x, err = c.Get(errorKey)
	assert.Error(t, err)
	assert.Nil(t, x)

	err2 := err

	// the error should be the same
	assert.Equal(t, err1, err2)

	x, err = c.GetString(errorKey)
	assert.Error(t, err)
	assert.Equal(t, "", x)

	// delete
	delKey := cache.NewStringKey("a")
	x, err = c.Get(delKey)
	assert.NoError(t, err)
	assert.Equal(t, "1", x.(string))

	err = c.Delete(delKey)
	assert.NoError(t, err)
	assert.False(t, c.Exists(delKey))

	_, ok = c.DirectGet(delKey)
	assert.False(t, ok)

	// set
	setKey := cache.NewStringKey("s")
	c.Set(setKey, "1")
	x, err = c.GetString(setKey)
	assert.NoError(t, err)
	assert.Equal(t, "1", x)

	// x, err = c.Get(delKey)
	// assert.Error(t, err)

	// disabled=true
	// c = NewCache("test", true, retrieveOK, expiration, cleanupInterval)
	c = NewBaseCache(true, retrieveTest, be)
	assert.NotNil(t, c)
	x, err = c.Get(aKey)
	assert.NoError(t, err)
	assert.Equal(t, "1", x.(string))

	_, err = c.GetString(timeKey)
	assert.Error(t, err)

	_, err = c.GetBool(aKey)
	assert.Error(t, err)

	_, err = c.GetTime(aKey)
	assert.Error(t, err)

	// retrieveError
	c = NewBaseCache(true, retrieveError, be)
	assert.NotNil(t, c)
	_, err = c.Get(aKey)
	assert.Error(t, err)

	_, err = c.GetString(timeKey)
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())

	_, err = c.GetBool(aKey)
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())

	_, err = c.GetTime(aKey)
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())

	// TODO: mock the backend first Get fail, second Get ok

	// TODO: add emptyCache here
}

func retrieveBenchmark(k cache.Key) (interface{}, error) {
	return "", nil
}

func BenchmarkRawRetrieve(b *testing.B) {
	var keys []cache.StringKey
	for i := 0; i < 100000; i++ {
		// keys = append(keys, cache.NewStringKey(util.RandString(5)))
		keys = append(keys, cache.NewStringKey("aaa"))
	}

	b.ResetTimer()
	b.ReportAllocs()

	index := 0
	for i := 0; i < b.N; i++ {
		key := keys[index]
		index++
		if index > 99999 {
			index = 0
		}
		retrieveBenchmark(key)
	}
}

func BenchmarkSingleFlightRetrieve(b *testing.B) {
	var keys []cache.StringKey
	for i := 0; i < 100000; i++ {
		// keys = append(keys, cache.NewStringKey(util.RandString(5)))
		keys = append(keys, cache.NewStringKey("aaa"))
	}

	b.ResetTimer()
	b.ReportAllocs()

	var g singleflight.Group
	index := 0
	for i := 0; i < b.N; i++ {
		key := keys[index]
		index++
		if index == 99999 {
			index = 0
		}

		g.Do(key.Key(), func() (interface{}, error) {
			return retrieveBenchmark(key)
		})
	}
}
