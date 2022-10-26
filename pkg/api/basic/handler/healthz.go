/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	pkgredis "iam/pkg/cache/redis"
	"iam/pkg/config"
	"iam/pkg/database"
)

func checkDatabase(dbConfig *config.Database) error {
	c := database.NewDBClient(dbConfig)
	return c.TestConnection()
}

func checkRedis(redisConfig *config.Redis) error {
	var rds *redis.Client
	switch redisConfig.Type {
	case pkgredis.ModeStandalone:
		opt := &redis.Options{
			Addr:     redisConfig.Addr,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
			// just for live test
			PoolSize: 1,
		}

		rds = redis.NewClient(opt)
	case pkgredis.ModeSentinel:
		sentinelAddrs := strings.Split(redisConfig.SentinelAddr, ",")
		opt := &redis.FailoverOptions{
			MasterName:    redisConfig.MasterName,
			SentinelAddrs: sentinelAddrs,
			DB:            redisConfig.DB,
			Password:      redisConfig.Password,
			PoolSize:      1,
		}

		if redisConfig.SentinelPassword != "" {
			opt.SentinelPassword = redisConfig.SentinelPassword
		}

		rds = redis.NewFailoverClient(opt)
	default:
		return errors.New("invalid redis ID, should be `standalone` or `sentinel`")
	}

	defer rds.Close()
	_, err := rds.Ping(context.TODO()).Result()
	return err
}

// Healthz godoc
// @Summary healthz for server health check
// @Description /healthz to make sure the server is health
// @ID healthz
// @Tags basic
// @Accept json
// @Produce json
// @Success 200 {string} string message
// @Failure 500 {string} string message
// @Header 200 {string} X-Request-Id "the request id"
// @Router /healthz [get]
func NewHealthzHandleFunc(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. check database
		defaultDBConfig := cfg.DatabaseMap["iam"]

		dbConfigs := []config.Database{defaultDBConfig}

		if !cfg.EnableBkAuth {
			bkPaaSDBConfig := cfg.DatabaseMap["open_paas"]
			dbConfigs = append(dbConfigs, bkPaaSDBConfig)
		}

		for _, dbConfig := range dbConfigs {
			dbConfig := dbConfig
			// reset the options for check
			dbConfig.MaxIdleConns = 1
			dbConfig.MaxOpenConns = 1
			dbConfig.ConnMaxLifetimeSecond = 60

			err := checkDatabase(&dbConfig)
			if err != nil {
				message := fmt.Sprintf("db connect fail: %s [id=%s host=%s port=%d]",
					err.Error(), dbConfig.ID, dbConfig.Host, dbConfig.Port)
				c.String(http.StatusInternalServerError, message)
				return
			}
		}

		// 2. check redis
		var err error
		var addr string
		redisConfig, ok := cfg.RedisMap[pkgredis.NameCache]
		if ok {
			addr = redisConfig.Addr
			err = checkRedis(&redisConfig)
		}

		redisConfig, ok = cfg.RedisMap[pkgredis.NameMQ]
		if ok {
			addr = redisConfig.SentinelAddr
			err = checkRedis(&redisConfig)
		}

		if err != nil {
			message := fmt.Sprintf("redis(mode=%s) connect fail: %s [addr=%s]", redisConfig.ID, err.Error(), addr)
			c.String(http.StatusInternalServerError, message)
			return
		}

		// 4. return ok
		c.String(http.StatusOK, "ok")
	}
}
