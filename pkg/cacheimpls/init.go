/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"errors"
	"math/rand"
	"time"

	"github.com/TencentBlueKing/gopkg/cache/memory"
	"github.com/TencentBlueKing/gopkg/cache/memory/backend"
	log "github.com/sirupsen/logrus"
	gocache "github.com/wklken/go-cache"

	"iam/pkg/cache/cleaner"
	"iam/pkg/cache/redis"
)

// CacheLayer ...
const CacheLayer = "Cache"

// LocalAppCodeAppSecretCache ...
var (
	LocalAppCodeAppSecretCache      *gocache.Cache
	LocalAuthAppAccessKeyCache      *gocache.Cache
	LocalSubjectCache               memory.Cache
	LocalSubjectRoleCache           memory.Cache
	LocalSystemClientsCache         memory.Cache
	LocalRemoteResourceListCache    memory.Cache
	LocalSubjectPKCache             memory.Cache
	LocalAPIGatewayJWTClientIDCache memory.Cache
	LocalActionCache                memory.Cache // for iam engine
	LocalUnmarshaledExpressionCache *gocache.Cache
	LocalGroupSystemAuthTypeCache   *gocache.Cache

	RemoteResourceCache     *redis.Cache
	ResourceTypeCache       *redis.Cache
	SubjectDepartmentCache  *redis.Cache
	SubjectPKCache          *redis.Cache
	SubjectSystemGroupCache *redis.Cache
	SystemCache             *redis.Cache
	ActionPKCache           *redis.Cache
	ActionDetailCache       *redis.Cache
	ActionListCache         *redis.Cache

	PolicyCache              *redis.Cache
	ExpressionCache          *redis.Cache
	TemporaryPolicyCache     *redis.Cache
	GroupSystemAuthTypeCache *redis.Cache

	LocalPolicyCache          *gocache.Cache
	LocalExpressionCache      *gocache.Cache
	LocalTemporaryPolicyCache *gocache.Cache
	ChangeListCache           *redis.Cache

	ActionCacheCleaner       *cleaner.CacheCleaner
	ActionListCacheCleaner   *cleaner.CacheCleaner
	ResourceTypeCacheCleaner *cleaner.CacheCleaner
	SubjectCacheCleaner      *cleaner.CacheCleaner
	SystemCacheCleaner       *cleaner.CacheCleaner
)

// ErrNotExceptedTypeFromCache ...
var ErrNotExceptedTypeFromCache = errors.New("not expected type from cache")

func newRandomDuration(seconds int) backend.RandomExtraExpirationDurationFunc {
	return func() time.Duration {
		return time.Duration(rand.Intn(seconds*1000)) * time.Millisecond
	}
}

// Cache should only know about get/retrieve data
// ! DO NOT CARE ABOUT WHAT THE DATA WILL BE USED FOR
func InitCaches(disabled bool) {
	LocalAppCodeAppSecretCache = gocache.New(12*time.Hour, 5*time.Minute)

	// auth app_code/app_secret cache
	LocalAuthAppAccessKeyCache = gocache.New(12*time.Hour, 5*time.Minute)

	// 影响: engine增量同步

	LocalSubjectCache = memory.NewCache(
		"local_subject",
		disabled,
		retrieveSubject,
		1*time.Minute,
		newRandomDuration(30),
	)

	// 影响: job查询cmdb的资源进行鉴权

	LocalRemoteResourceListCache = memory.NewCache(
		"local_remote_resource_list",
		disabled,
		retrieveRemoteResourceList,
		30*time.Second,
		newRandomDuration(10),
	)

	// 影响: 每次鉴权

	LocalSubjectPKCache = memory.NewCache(
		"local_subject_pk",
		disabled,
		retrieveSubjectPKFromRedis,
		30*time.Minute,
		nil,
	)

	// 影响: 每次鉴权 => 理论上, 也可以改成两级cache

	LocalSubjectRoleCache = memory.NewCache(
		"local_subject_role",
		disabled,
		retrieveSubjectRole,
		1*time.Minute,
		newRandomDuration(30),
	)

	// 影响: 每次鉴权 => system_id比较集中, singleflight可以防止大的并发落db

	LocalSystemClientsCache = memory.NewCache(
		"local_system_clients",
		disabled,
		retrieveSystemClients,
		1*time.Minute,
		newRandomDuration(30),
	)

	// 影响: engine接口/policy查询接口

	LocalActionCache = memory.NewCache(
		"local_action",
		disabled,
		retrieveAction,
		30*time.Minute,
		newRandomDuration(30),
	)

	// 无影响, 重算而已不查db

	LocalAPIGatewayJWTClientIDCache = memory.NewCache(
		"local_apigw_jwt_client_id",
		disabled,
		retrieveAPIGatewayJWTClientID,
		30*time.Second,
		nil,
	)

	// 无影响, 重算而已不查db

	LocalUnmarshaledExpressionCache = gocache.New(30*time.Minute, 5*time.Minute)

	// 影响: 每次鉴权

	LocalGroupSystemAuthTypeCache = gocache.New(10*time.Minute, 5*time.Minute)

	//  ==========================

	// NOTE: short key in 3 chars, make the redis key short enough, for better performance
	//     rem = remote
	//     res = resource
	//     typ = type
	//     sys = system
	//     act = action
	//     dtl = detail
	//     sub = subject
	//     pl  = policy
	//     ex  = expression
	//     cl = change list
	//     grp = group

	// inner system model
	SystemCache = redis.NewCache(
		"sys",
		30*time.Minute,
	)

	SubjectSystemGroupCache = redis.NewCache(
		"sys_sub_grp",
		30*time.Minute,
	)

	ResourceTypeCache = redis.NewCache(
		"res_typ",
		30*time.Minute,
	)

	RemoteResourceCache = redis.NewCache(
		"rem_res",
		5*time.Minute,
	)

	ActionPKCache = redis.NewCache(
		"act_pk",
		30*time.Minute,
	)

	ActionDetailCache = redis.NewCache(
		"act_dtl:2",
		30*time.Minute,
	)

	SubjectPKCache = redis.NewCache(
		"sub_pk",
		30*time.Minute,
	)

	SubjectDepartmentCache = redis.NewCache(
		"sub_dep",
		30*time.Minute,
	)
	ActionListCache = redis.NewCache(
		"all_act:2",
		30*time.Minute,
	)

	LocalPolicyCache = gocache.New(5*time.Minute, 5*time.Minute)
	LocalExpressionCache = gocache.New(5*time.Minute, 5*time.Minute)
	LocalTemporaryPolicyCache = gocache.New(5*time.Minute, 5*time.Minute)
	ChangeListCache = redis.NewCache("cl", 5*time.Minute)

	PolicyCache = redis.NewCache(
		"pl",
		30*time.Minute,
	)

	ExpressionCache = redis.NewCache(
		"ex",
		30*time.Minute,
	)

	TemporaryPolicyCache = redis.NewCache(
		"tpl",
		30*time.Minute,
	)

	GroupSystemAuthTypeCache = redis.NewCache(
		"gat",
		30*time.Minute,
	)

	ActionCacheCleaner = cleaner.NewCacheCleaner("ActionCacheCleaner", actionCacheDeleter{})
	go ActionCacheCleaner.Run()

	ActionListCacheCleaner = cleaner.NewCacheCleaner("ActionListCacheCleaner", actionListCacheDeleter{})
	go ActionListCacheCleaner.Run()

	ResourceTypeCacheCleaner = cleaner.NewCacheCleaner("ResourceTypeCacheCleaner", resourceTypeCacheDeleter{})
	go ResourceTypeCacheCleaner.Run()

	SubjectCacheCleaner = cleaner.NewCacheCleaner("SubjectCacheCleaner", subjectCacheDeleter{})
	go SubjectCacheCleaner.Run()

	SystemCacheCleaner = cleaner.NewCacheCleaner("SystemCacheCleaner", systemCacheDeleter{})
	go SystemCacheCleaner.Run()
}

// PolicyCacheDisabled 策略缓存默认打开
var PolicyCacheDisabled = false

// PolicyCacheExpiration 策略缓存默认保留7天
var PolicyCacheExpiration = 7 * 24 * time.Hour

// GroupSystemAuthTypeCacheExpiration 策略缓存默认保留7天
var GroupSystemAuthTypeCacheExpiration = 7 * 24 * time.Hour

// InitPolicyCacheSettings ...
func InitPolicyCacheSettings(disabled bool, expirationDays int64) {
	PolicyCacheDisabled = disabled
	if expirationDays != 0 {
		PolicyCacheExpiration = time.Duration(expirationDays) * 24 * time.Hour
	}

	log.Infof("init LocalPolicyCache disabled=%t, expiration=%s", PolicyCacheDisabled, PolicyCacheExpiration)
	if PolicyCacheDisabled {
		log.Warn("the LocalPolicyCache is disabled! Will query policy from database!")
	}
}

// InitVerifyAppCodeAppSecret ...
func InitVerifyAppCodeAppSecret(enableBkAuth bool) {
	if enableBkAuth {
		VerifyAppCodeAppSecret = VerifyAppCodeAppSecretFromAuth
		log.Infof("init VerifyAppCodeAppSecret to VerifyAppCodeAppSecretFromAuth")
	} else {
		VerifyAppCodeAppSecret = VerifyAppCodeAppSecretFromDB
		log.Infof("init VerifyAppCodeAppSecret to VerifyAppCodeAppSecretFromDB")
	}
}
