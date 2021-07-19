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
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/sync/singleflight"

	iamcache "iam/pkg/cache"
	"iam/pkg/util"
)

const (
	// while the go-redis/cache upgrade maybe not compatible with the previous version.
	// e.g. the object set by v7 can't read by v8
	// https://github.com/go-redis/cache/issues/52
	// NOTE: important!!! if upgrade the go-redis/cache version, should change the version

	// CacheVersion is loop in 00->99->00 => make sure will not conflict with previous version
	CacheVersion = "00"

	PipelineSizeThreshold = 100
)

// RetrieveFunc ...
type RetrieveFunc func(key iamcache.Key) (interface{}, error)

// Cache is a cache implements
type Cache struct {
	name              string
	keyPrefix         string
	codec             *cache.Cache
	cli               *redis.Client
	defaultExpiration time.Duration
	G                 singleflight.Group
}

// NewCache create a cache instance
func NewCache(name string, expiration time.Duration) *Cache {
	cli := GetDefaultRedisClient()

	// key format = iam:{version}:{cache_name}:{real_key}
	keyPrefix := fmt.Sprintf("iam:%s:%s", CacheVersion, name)

	codec := cache.New(&cache.Options{
		Redis: cli,
	})

	return &Cache{
		name:              name,
		keyPrefix:         keyPrefix,
		codec:             codec,
		cli:               cli,
		defaultExpiration: expiration,
	}
}

// NewMockCache will create a cache for mock
func NewMockCache(name string, expiration time.Duration) *Cache {
	cli := util.NewTestRedisClient()

	// key format = iam:{cache_name}:{real_key}
	keyPrefix := fmt.Sprintf("iam:%s", name)

	codec := cache.New(&cache.Options{
		Redis: cli,
	})

	return &Cache{
		name:              name,
		keyPrefix:         keyPrefix,
		codec:             codec,
		cli:               cli,
		defaultExpiration: expiration,
	}
}

func (c *Cache) genKey(key string) string {
	return c.keyPrefix + ":" + key
}

func (c *Cache) copyTo(source interface{}, dest interface{}) error {
	b, err := msgpack.Marshal(source)
	if err != nil {
		return err
	}

	err = msgpack.Unmarshal(b, dest)
	return err
}

// Set execute `set`
func (c *Cache) Set(key iamcache.Key, value interface{}, duration time.Duration) error {
	if duration == time.Duration(0) {
		duration = c.defaultExpiration
	}

	k := c.genKey(key.Key())
	return c.codec.Set(&cache.Item{
		Key:   k,
		Value: value,
		TTL:   duration,
	})
}

// Get execute `get`
func (c *Cache) Get(key iamcache.Key, value interface{}) error {
	k := c.genKey(key.Key())
	return c.codec.Get(context.TODO(), k, value)
}

// Exists execute `exists`
func (c *Cache) Exists(key iamcache.Key) bool {
	k := c.genKey(key.Key())

	count, err := c.cli.Exists(context.TODO(), k).Result()

	return err == nil && count == 1
}

// GetInto will retrieve the data from cache and unmarshal into the obj
func (c *Cache) GetInto(key iamcache.Key, obj interface{}, retrieveFunc RetrieveFunc) (err error) {
	// 1. get from cache, hit, return
	err = c.Get(key, obj)
	if err == nil {
		return
	}

	// 2. if missing
	// 2.1 check the guard
	// 2.2 do retrieve
	data, err, _ := c.G.Do(key.Key(), func() (interface{}, error) {
		return retrieveFunc(key)
	})
	// 2.3 do retrieve fail, make guard and return
	if err != nil {
		// if retrieve fail, should wait for few seconds for the missing-retrieve
		//c.makeGuard(key)
		return
	}

	// 3. set to cache
	errNotImportant := c.Set(key, data, 0)
	if errNotImportant != nil {
		log.Errorf("set to redis fail, key=%s, err=%s", key.Key(), errNotImportant)
	}

	// 注意, 这里基础类型无法通过 *obj = value 来赋值
	// 所以利用从缓存再次反序列化给对应指针赋值(相当于底层msgpack.unmarshal帮做了转换再次反序列化给对应指针赋值
	return c.copyTo(data, obj)
}

// Delete execute `del`
func (c *Cache) Delete(key iamcache.Key) (err error) {
	k := c.genKey(key.Key())

	ctx := context.TODO()

	_, err = c.cli.Del(ctx, k).Result()
	return err
}

// BatchDelete execute `del` with pipeline
func (c *Cache) BatchDelete(keys []iamcache.Key) error {
	newKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		newKeys = append(newKeys, c.genKey(key.Key()))
	}
	ctx := context.TODO()

	var err error
	if len(newKeys) < PipelineSizeThreshold {
		_, err = c.cli.Del(ctx, newKeys...).Result()
	} else {
		pipe := c.cli.Pipeline()

		for _, key := range newKeys {
			pipe.Del(ctx, key)
		}

		_, err = pipe.Exec(ctx)
	}
	return err
}

// BatchExpireWithTx execute `expire` with tx pipeline
func (c *Cache) BatchExpireWithTx(keys []iamcache.Key, expiration time.Duration) error {
	pipe := c.cli.TxPipeline()
	ctx := context.TODO()

	for _, k := range keys {
		key := c.genKey(k.Key())
		pipe.Expire(ctx, key, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// KV is a key-value pair
type KV struct {
	Key   string
	Value string
}

// BatchGet execute `get` with pipeline
func (c *Cache) BatchGet(keys []iamcache.Key) (map[iamcache.Key]string, error) {
	pipe := c.cli.Pipeline()

	ctx := context.TODO()

	cmds := map[iamcache.Key]*redis.StringCmd{}
	for _, k := range keys {
		key := c.genKey(k.Key())
		cmd := pipe.Get(ctx, key)

		cmds[k] = cmd
	}

	_, err := pipe.Exec(ctx)
	// 当批量操作, 里面有个key不存在, err = redis.Nil; 但是不应该影响其他存在的key的获取
	// Nil reply returned by Redis when key does not exist.
	if err != nil && err != redis.Nil {
		return nil, err
	}

	values := make(map[iamcache.Key]string, len(cmds))
	for hkf, cmd := range cmds {
		// maybe err or key missing
		// only return the HashKeyField who get value success from redis
		val, err := cmd.Result()
		if err != nil {
			continue
		} else {
			values[hkf] = val
		}
	}
	return values, nil
}

// BatchSetWithTx execute `set` with tx pipeline
func (c *Cache) BatchSetWithTx(kvs []KV, expiration time.Duration) error {
	// tx, all success or all fail
	pipe := c.cli.TxPipeline()

	ctx := context.TODO()

	for _, kv := range kvs {
		key := c.genKey(kv.Key)
		pipe.Set(ctx, key, kv.Value, expiration)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// ZData is a sorted-set data for redis `key: {member: score}`
type ZData struct {
	Key string
	Zs  []*redis.Z
}

// BatchZAdd execute `zadd` with pipeline
func (c *Cache) BatchZAdd(zDataList []ZData) error {
	pipe := c.cli.TxPipeline()
	ctx := context.TODO()

	for _, zData := range zDataList {
		key := c.genKey(zData.Key)
		pipe.ZAdd(ctx, key, zData.Zs...)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// ZRevRangeByScore execute `zrevrangebyscorewithscores`
func (c *Cache) ZRevRangeByScore(k string, min int64, max int64, offset int64, count int64) ([]redis.Z, error) {
	// 时间戳, 从大到小排序
	ctx := context.TODO()

	key := c.genKey(k)
	// TODO: add limit, offset, count => to ignore the too large list size
	// LIMIT 0 -1 equals no args
	cmds := c.cli.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    strconv.FormatInt(min, 10),
		Max:    strconv.FormatInt(max, 10),
		Offset: offset,
		Count:  count,
	})

	return cmds.Result()
}

// BatchZRemove execute `zremrangebyscore` with pipeline
func (c *Cache) BatchZRemove(keys []string, min int64, max int64) error {
	pipe := c.cli.TxPipeline()
	ctx := context.TODO()

	minStr := strconv.FormatInt(min, 10)
	maxStr := strconv.FormatInt(max, 10)

	for _, k := range keys {
		key := c.genKey(k)
		pipe.ZRemRangeByScore(ctx, key, minStr, maxStr)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// HashKeyField is a hash data for redis, `Key: field -> `
type HashKeyField struct {
	Key   string
	Field string
}

// Hash is a hash data  `Key: field->value`
type Hash struct {
	HashKeyField
	Value string
}

// BatchHSetWithTx execute `hset` with tx pipeline
func (c *Cache) BatchHSetWithTx(hashes []Hash) error {
	// tx, all success or all fail
	pipe := c.cli.TxPipeline()
	ctx := context.TODO()

	for _, h := range hashes {
		key := c.genKey(h.Key)
		pipe.HSet(ctx, key, h.Field, h.Value)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// BatchHGet execute `hget` with pipeline
func (c *Cache) BatchHGet(hashKeyFields []HashKeyField) (map[HashKeyField]string, error) {
	pipe := c.cli.Pipeline()

	ctx := context.TODO()
	cmds := make(map[HashKeyField]*redis.StringCmd, len(hashKeyFields))
	for _, h := range hashKeyFields {
		key := c.genKey(h.Key)
		cmd := pipe.HGet(ctx, key, h.Field)

		cmds[h] = cmd
	}

	_, err := pipe.Exec(ctx)
	// 当批量操作, 里面有个key不存在, err = redis.Nil; 但是不应该影响其他存在的key的获取
	// Nil reply returned by Redis when key does not exist.
	if err != nil && err != redis.Nil {
		return nil, err
	}

	values := make(map[HashKeyField]string, len(cmds))
	for hkf, cmd := range cmds {
		// maybe err or key missing
		// only return the HashKeyField who get value success from redis
		val, err := cmd.Result()
		if err != nil {
			continue
		} else {
			values[hkf] = val
		}
	}
	return values, nil
}

// HKeys execute `hkeys`
func (c *Cache) HKeys(hashKey string) ([]string, error) {
	key := c.genKey(hashKey)
	return c.cli.HKeys(context.TODO(), key).Result()
}

// Unmarshal with compress, via go-redis/cache, use s2 compression
// Note: YOU SHOULD NOT USE THE RAW msgpack.Unmarshal directly! will panic with decode fail
func (c *Cache) Unmarshal(b []byte, value interface{}) error {
	return c.codec.Unmarshal(b, value)
}

// Marshal with compress, via go-redis/cache, use s2 compression
// Note: YOU SHOULD NOT USE THE RAW msgpack.Marshal directly!
func (c *Cache) Marshal(value interface{}) ([]byte, error) {
	return c.codec.Marshal(value)
}
