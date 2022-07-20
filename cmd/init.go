/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cmd

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"iam/pkg/api/common"
	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/component"
	"iam/pkg/config"
	"iam/pkg/database"
	"iam/pkg/logging"
	"iam/pkg/metric"
	"iam/pkg/task"
	"iam/pkg/util"
)

var globalConfig *config.Config

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		panic("Config file missing")
	}
	// Use config file from the flag.
	// viper.SetConfigFile(cfgFile)
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Using config file: %s, read fail: err=%v", viper.ConfigFileUsed(), err))
	}
	var err error
	globalConfig, err = config.Load(viper.GetViper())
	if err != nil {
		panic(fmt.Sprintf("Could not load configurations from file, error: %v", err))
	}
}

func initSentry() {
	if globalConfig.Sentry.Enable {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: globalConfig.Sentry.DSN,
		})
		if err != nil {
			log.Errorf("init Sentry fail: %s", err)
			return
		}
		log.Info("init Sentry success")
	} else {
		log.Info("Sentry is not enabled, will not init it")
	}

	util.InitErrorReport(globalConfig.Sentry.Enable)
}

func initMetrics() {
	metric.InitMetrics()
	log.Info("init Metrics success")
}

func initDatabase() {
	defaultDBConfig, ok := globalConfig.DatabaseMap["iam"]
	if !ok {
		panic("database bk-iam should be configured")
	}

	if globalConfig.EnableBkAuth {
		database.InitDBClients(&defaultDBConfig, nil)
		log.Info("init Database success")
		return
	}

	// TODO: 不应该成为强依赖
	bkPaaSDBConfig, ok := globalConfig.DatabaseMap["open_paas"]
	if !ok {
		panic("bkauth is not enabled, so database open_paas should be configured")
	}
	database.InitDBClients(&defaultDBConfig, &bkPaaSDBConfig)

	log.Info("init Database success")
}

func initRedis() {
	standaloneConfig, isStandalone := globalConfig.RedisMap[redis.ModeStandalone]
	sentinelConfig, isSentinel := globalConfig.RedisMap[redis.ModeSentinel]

	if !(isStandalone || isSentinel) {
		panic("redis id=standalone or id=sentinel should be configured")
	}

	if isSentinel && isStandalone {
		log.Info("redis both id=standalone and id=sentinel configured, will use sentinel")

		delete(globalConfig.RedisMap, redis.ModeStandalone)
		isStandalone = false
	}

	if isSentinel {
		if sentinelConfig.MasterName == "" {
			panic("redis id=sentinel, the `masterName` required")
		}
		log.Info("init Redis mode=`sentinel`")
		redis.InitRedisClient(globalConfig.Debug, &sentinelConfig)
	}

	if isStandalone {
		log.Info("init Redis mode=`standalone`")
		redis.InitRedisClient(globalConfig.Debug, &standaloneConfig)
	}

	log.Info("init Redis success")
}

// NOTE: 必须在Redis init 后才能初始化 rmq
func initRmqQueue() {
	log.Info("init RMQ queue")
	// NOTE: 必须在redis初始化后才能初始化rmq
	task.InitRmqQueue(globalConfig.Debug)

	log.Info("init RMQ queue success")
}

func initLogger() {
	logging.InitLogger(&globalConfig.Logger)
}

func initCaches() {
	cacheimpls.InitCaches(false)
}

func initPolicyCacheSettings() {
	cacheimpls.InitPolicyCacheSettings(globalConfig.PolicyCache.Disabled, globalConfig.PolicyCache.ExpirationDays)
}

func initVerifyAppCodeAppSecret() {
	cacheimpls.InitVerifyAppCodeAppSecret(globalConfig.EnableBkAuth)
}

func initSuperAppCode() {
	config.InitSuperAppCode(globalConfig.SuperAppCode)
}

func initSuperUser() {
	config.InitSuperUser(globalConfig.SuperUser)
}

func initSupportShieldFeatures() {
	config.InitSupportShieldFeatures(globalConfig.SupportShieldFeatures)
}

func initSecurityAuditAppCode() {
	config.InitSecurityAuditAppCode(globalConfig.SecurityAuditAppCode)
}

func initComponents() {
	component.InitBkRemoteResourceClient()

	if globalConfig.EnableBkAuth {
		bkAuthHost, ok := globalConfig.HostMap["bkauth"]
		if !ok {
			panic("bkauth is enabled, so host bkauth should be configured")
		}

		if globalConfig.BkAppCode == "" || globalConfig.BkAppSecret == "" {
			panic("bkauth is enabled, but iam's bkAppCode and bkAppSecret is not configured")
		}

		component.InitBkAuthClient(bkAuthHost.Addr, globalConfig.BkAppCode, globalConfig.BkAppSecret)
		log.Infof("init bkauth client success, host = %s", bkAuthHost.Addr)
	}
}

func initQuota() {
	common.InitQuota(globalConfig.Quota, globalConfig.CustomQuotasMap)
}

func initSwitch() {
	common.InitSwitch(globalConfig.Switch)
}
