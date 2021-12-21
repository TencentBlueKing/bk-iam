/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"

	"iam/pkg/cache"
)

func TestCache_genKey(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	assert.Equal(t, "iam:test:abc", c.genKey("abc"))
}

// func TestCache_Guard(t *testing.T) {
// 	c := NewMockCache("test", 5*time.Minute)
// 	key := cache.NewStringKey("gkey")
//
// 	err := c.makeGuard(key)
// 	assert.NoError(t, err)
//
// 	hasGuard := c.hasGuard(key)
// 	assert.True(t, hasGuard)
// }

func TestCache_Set_Exists_Get(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)
	key := cache.NewStringKey("abc")

	// set
	err := c.Set(key, 1, 0)
	assert.NoError(t, err)

	// exists
	exists := c.Exists(key)
	assert.True(t, exists)

	// get
	var a int
	err = c.Get(key, &a)
	assert.NoError(t, err)
	assert.Equal(t, 1, a)
}

func retrieveTest(key cache.Key) (interface{}, error) {
	return "ok", nil
}

func TestGetInto(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	key := cache.NewStringKey("akey")

	var i string
	err := c.GetInto(key, &i, retrieveTest)
	assert.NoError(t, err)
	assert.Equal(t, "ok", i)

	var i2 string
	err = c.GetInto(key, &i2, retrieveTest)
	assert.NoError(t, err)
	assert.Equal(t, "ok", i2)
}

func TestDelete(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	key := cache.NewStringKey("dkey")

	// do delete
	err := c.Delete(key)
	assert.NoError(t, err)

	// set
	err = c.Set(key, 1, 0)
	assert.NoError(t, err)

	// do it again
	err = c.Delete(key)
	assert.NoError(t, err)
}

func TestBatchDelete(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	key1 := cache.NewStringKey("d1key")
	key2 := cache.NewStringKey("d2key")

	keys := []cache.Key{
		key1,
		key2,
	}

	// do delete 0 key
	err := c.BatchDelete(keys)
	// assert.Equal(t, int64(0), count)
	assert.NoError(t, err)

	// set
	err = c.Set(key1, 1, 0)
	assert.NoError(t, err)

	// do delete 1 key
	err = c.BatchDelete(keys)
	// assert.Equal(t, int64(1), count)
	assert.NoError(t, err)
}

// func TestExpire(t *testing.T) {
// 	c := NewMockCache("test", 5*time.Minute)
//
// 	key := cache.NewStringKey("d1key")
//
// 	err := c.Expire(key, 1*time.Minute)
// 	assert.NoError(t, err)
// }

func TestBatchExpireWithTx(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	key1 := cache.NewStringKey("d1key")
	key2 := cache.NewStringKey("d2key")

	keys := []cache.Key{
		key1,
		key2,
	}

	err := c.BatchExpireWithTx(keys, 1*time.Minute)
	assert.NoError(t, err)
}

// func TestHSet_and_HGet(t *testing.T) {
// 	c := NewMockCache("test", 5*time.Minute)
//
// 	kf := HashKeyField{
// 		Key:   "hello",
// 		Field: "a",
// 	}
//
// 	h := Hash{
// 		HashKeyField: kf,
// 		Value:        "1",
// 	}
//
// 	count, err := c.HSet(h)
// 	assert.NoError(t, err)
// 	assert.Equal(t, int64(1), count)
//
// 	// get
// 	data, err := c.HGet(kf)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "1", data)
// }

// func TestBatchHSetWithTx_and_BatchHGet(t *testing.T) {
//	c := NewMockCache("test", 5*time.Minute)
//
//	hs := []Hash{
//		{
//			HashKeyField: HashKeyField{
//				"a",
//				"attr1",
//			},
//			Value: "1",
//		},
//		{
//			HashKeyField: HashKeyField{
//				"a",
//				"attr2",
//			},
//			Value: "2",
//		},
//	}
//
//	err := c.BatchHSetWithTx(hs)
//	assert.NoError(t, err)
//
//	keyField1 := HashKeyField{
//		"a",
//		"attr1",
//	}
//
//	keyFields := []HashKeyField{
//		keyField1,
//		{
//			"a",
//			"attr2",
//		},
//		{
//			"a",
//			"attr3",
//		},
//	}
//
//	data, err := c.BatchHGet(keyFields)
//	assert.NoError(t, err)
//	assert.Len(t, data, 2)
//
//	assert.Contains(t, data, keyField1)
//	assert.Equal(t, "1", data[keyField1])
// }

func TestBatchSetWithTx_and_BatchGet(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	kvs := []KV{
		{
			Key:   "a",
			Value: "1",
		},
		{
			Key:   "b",
			Value: "2",
		},
	}

	err := c.BatchSetWithTx(kvs, 5*time.Minute)
	assert.NoError(t, err)

	keys := []cache.Key{
		cache.NewStringKey("a"),
		cache.NewStringKey("b"),
	}

	data, err := c.BatchGet(keys)
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	akey := cache.NewStringKey("a")
	assert.Contains(t, data, akey)
	assert.Equal(t, "1", data[akey])
}

func TestSetOneAndBatchGet(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	type Abc struct {
		X string
		Y int
		Z string
	}

	key := cache.NewStringKey("a")

	// compressionThreshold = 64
	t.Run("less than compressionThreshold", func(t *testing.T) {
		c.Set(key, Abc{
			X: "hello",
			Y: 123,
			Z: "",
		}, 5*time.Minute)

		data, err := c.BatchGet([]cache.Key{key})
		assert.NoError(t, err)
		assert.Len(t, data, 1)

		// NOTE: the string is msgpack marshal and compress, so
		value := data[key]
		fmt.Println("value", value)
		var abc Abc

		err = c.Unmarshal(conv.StringToBytes(value), &abc)
		fmt.Println("abc:", abc)
		assert.NoError(t, err)

		var def Abc
		err = msgpack.Unmarshal(conv.StringToBytes(value), &def)
		fmt.Println("def:", abc)
		assert.NoError(t, err)
	})

	t.Run("greater than compressThreshold", func(t *testing.T) {
		c.Set(key, Abc{
			X: "hello",
			Y: 123,
			Z: "123456789012345678901234567890123456789012345678901234567890",
		}, 5*time.Minute)

		data, err := c.BatchGet([]cache.Key{key})
		assert.NoError(t, err)
		assert.Len(t, data, 1)

		// NOTE: the string is msgpack marshal and compress, so
		value := data[key]
		fmt.Println("value", value)
		var abc Abc

		err = c.Unmarshal(conv.StringToBytes(value), &abc)
		fmt.Println("abc:", abc)
		assert.NoError(t, err)

		var def Abc
		err = msgpack.Unmarshal(conv.StringToBytes(value), &def)
		fmt.Println("def:", abc)
		assert.Error(t, err)
	})
}

func TestBatchSetAndGet(t *testing.T) {
	c := NewMockCache("test", 5*time.Minute)

	type Abc struct {
		X string
		Y int
		Z string
	}

	small, _ := c.Marshal(Abc{
		X: "hello",
		Y: 123,
		Z: "",
	})
	huge, _ := c.Marshal(Abc{
		X: "hello",
		Y: 123,
		Z: "123456789012345678901234567890123456789012345678901234567890",
	})

	kvs := []KV{
		{
			Key:   "a",
			Value: conv.BytesToString(small),
		},
		{
			Key:   "b",
			Value: conv.BytesToString(huge),
		},
	}

	err := c.BatchSetWithTx(kvs, 5*time.Minute)
	assert.NoError(t, err)

	// get single: small without compress
	var v1 Abc
	err = c.Get(cache.NewStringKey("a"), &v1)
	assert.NoError(t, err)
	assert.Equal(t, v1.X, "hello")
	assert.Equal(t, v1.Y, 123)
	assert.Equal(t, v1.Z, "")
	// get single: huge with compress
	var v2 Abc
	err = c.Get(cache.NewStringKey("b"), &v2)
	assert.NoError(t, err)
	assert.Equal(t, v2.X, "hello")
	assert.Equal(t, v2.Y, 123)
	assert.Equal(t, v2.Z, "123456789012345678901234567890123456789012345678901234567890")
}
