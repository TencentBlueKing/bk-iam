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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"iam/pkg/config"
)

// ModeStandalone ...
const (
	ModeStandalone = "standalone"
	ModeSentinel   = "sentinel"
)

var rds *redis.Client

var redisClientInitOnce sync.Once

func newStandaloneClient(redisConfig *config.Redis) *redis.Client {
	opt := &redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	}

	// set default options
	opt.DialTimeout = time.Duration(2) * time.Second
	opt.ReadTimeout = time.Duration(1) * time.Second
	opt.WriteTimeout = time.Duration(1) * time.Second
	opt.PoolSize = 20 * runtime.NumCPU()
	opt.MinIdleConns = 10 * runtime.NumCPU()
	opt.IdleTimeout = time.Duration(3) * time.Minute

	// set custom options, from config.yaml
	if redisConfig.DialTimeout > 0 {
		opt.DialTimeout = time.Duration(redisConfig.DialTimeout) * time.Second
	}
	if redisConfig.ReadTimeout > 0 {
		opt.ReadTimeout = time.Duration(redisConfig.ReadTimeout) * time.Second
	}
	if redisConfig.WriteTimeout > 0 {
		opt.WriteTimeout = time.Duration(redisConfig.WriteTimeout) * time.Second
	}

	if redisConfig.PoolSize > 0 {
		opt.PoolSize = redisConfig.PoolSize
	}
	if redisConfig.MinIdleConns > 0 {
		opt.MinIdleConns = redisConfig.MinIdleConns
	}

	log.Infof("connect to redis: "+
		"%s [db=%d, dialTimeout=%s, readTimeout=%s, writeTimeout=%s, poolSize=%d, minIdleConns=%d, idleTimeout=%s]",
		opt.Addr, opt.DB, opt.DialTimeout, opt.ReadTimeout, opt.WriteTimeout, opt.PoolSize, opt.MinIdleConns, opt.IdleTimeout)

	return redis.NewClient(opt)
}

func newSentinelClient(redisConfig *config.Redis) *redis.Client {
	sentinelAddrs := strings.Split(redisConfig.SentinelAddr, ",")
	opt := &redis.FailoverOptions{
		MasterName:    redisConfig.MasterName,
		SentinelAddrs: sentinelAddrs,
		DB:            redisConfig.DB,
		Password:      redisConfig.Password,
	}

	if redisConfig.SentinelPassword != "" {
		opt.SentinelPassword = redisConfig.SentinelPassword
	}

	// set default options
	opt.DialTimeout = 2 * time.Second
	opt.ReadTimeout = 1 * time.Second
	opt.WriteTimeout = 1 * time.Second
	opt.PoolSize = 20 * runtime.NumCPU()
	opt.MinIdleConns = 10 * runtime.NumCPU()
	opt.IdleTimeout = 3 * time.Minute

	// set custom options, from config.yaml
	if redisConfig.DialTimeout > 0 {
		opt.DialTimeout = time.Duration(redisConfig.DialTimeout) * time.Second
	}
	if redisConfig.ReadTimeout > 0 {
		opt.ReadTimeout = time.Duration(redisConfig.ReadTimeout) * time.Second
	}
	if redisConfig.WriteTimeout > 0 {
		opt.WriteTimeout = time.Duration(redisConfig.WriteTimeout) * time.Second
	}

	if redisConfig.PoolSize > 0 {
		opt.PoolSize = redisConfig.PoolSize
	}
	if redisConfig.MinIdleConns > 0 {
		opt.MinIdleConns = redisConfig.MinIdleConns
	}

	return redis.NewFailoverClient(opt)
}

// InitRedisClient ...
func InitRedisClient(debugMode bool, redisConfig *config.Redis) {
	if rds == nil {
		redisClientInitOnce.Do(func() {
			switch redisConfig.ID {
			case ModeStandalone:
				rds = newStandaloneClient(redisConfig)
			case ModeSentinel:
				rds = newSentinelClient(redisConfig)
			default:
				panic("init redis client fail, invalid redis.id, should be `standalone` or `sentinel`")
			}

			_, err := rds.Ping(context.TODO()).Result()
			if err != nil {
				log.WithError(err).Error("connect to redis fail")
				// redis is important
				if !debugMode {
					panic(err)
				}
			}
		})
	}
}

// GetDefaultRedisClient 获取默认的Redis实例
func GetDefaultRedisClient() *redis.Client {
	return rds
}
