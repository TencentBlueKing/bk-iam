/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"errors"

	"github.com/spf13/viper"
)

// Server ...
type Server struct {
	Host string
	Port int

	GraceTimeout int64

	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// Cache ...
type Cache struct {
	Disabled bool
}

// PolicyCache ...
type PolicyCache struct {
	Disabled       bool
	ExpirationDays int64
}

// Logger ...
type Logger struct {
	System    LogConfig
	API       LogConfig
	SQL       LogConfig
	Audit     LogConfig
	Web       LogConfig
	Component LogConfig
}

// LogConfig ...
type LogConfig struct {
	Level    string
	Writer   string
	Settings map[string]string
}

// Database ...
type Database struct {
	ID       string
	Host     string
	Port     int
	User     string
	Password string
	Name     string

	MaxOpenConns          int
	MaxIdleConns          int
	ConnMaxLifetimeSecond int
}

// Redis ...
type Redis struct {
	ID           string
	Addr         string
	Password     string
	DB           int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
	PoolSize     int
	MinIdleConns int
	ChannelKey   string

	// mode=sentinel required
	SentinelAddr     string
	MasterName       string
	SentinelPassword string
}

// Sentry ...
type Sentry struct {
	Enable bool
	DSN    string
}

// Quota ...
type Quota struct {
	Model map[string]int

	// NOTE: only used for rate limit middleware, will remove in the future
	API map[string]int
}

// SystemQuota store the settings for specific system
type SystemQuota struct {
	ID    string
	Quota Quota
}

type Host struct {
	ID   string
	Addr string
}

// Crypto store the keys for crypto
type Crypto struct {
	ID  string
	Key string
}

// Config ...
type Config struct {
	Debug bool

	Server Server
	Sentry Sentry

	// iam's app_code and app_secret
	BkAppCode   string
	BkAppSecret string

	SuperAppCode string
	// default superuser
	SuperUser string
	// 产品上支持接入系统配置屏蔽的功能
	SupportShieldFeatures []string
	// 内部系统app_code, 共享权限模型白名单
	ShareAppCode string

	Databases   []Database
	DatabaseMap map[string]Database

	Redis    []Redis
	RedisMap map[string]Redis

	EnableBkAuth bool
	Hosts        []Host
	HostMap      map[string]Host

	Quota Quota

	CustomQuotas    []SystemQuota
	CustomQuotasMap map[string]Quota

	Switch map[string]bool

	Cache       Cache
	PolicyCache PolicyCache
	Logger      Logger

	Cryptos map[string]*Crypto
}

// Load 从viper中读取配置文件
func Load(v *viper.Viper) (*Config, error) {
	var cfg Config
	// 将配置信息绑定到结构体上
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// parse the list to map
	// 1. database
	cfg.DatabaseMap = make(map[string]Database)
	for _, db := range cfg.Databases {
		cfg.DatabaseMap[db.ID] = db
	}

	if len(cfg.DatabaseMap) == 0 {
		return nil, errors.New("database cannot be empty")
	}

	// 2. redis
	cfg.RedisMap = make(map[string]Redis)
	for _, rds := range cfg.Redis {
		cfg.RedisMap[rds.ID] = rds
	}

	// 3. hosts
	cfg.HostMap = make(map[string]Host)
	for _, host := range cfg.Hosts {
		cfg.HostMap[host.ID] = host
	}

	// 4. init quota
	cfg.CustomQuotasMap = make(map[string]Quota)
	for _, q := range cfg.CustomQuotas {
		cfg.CustomQuotasMap[q.ID] = q.Quota
	}

	return &cfg, nil
}
