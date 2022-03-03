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
	"time"

	"github.com/TencentBlueKing/gopkg/stringx"
	gocache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"

	"iam/pkg/component"
	"iam/pkg/database/edao"
)

// VerifyAppCodeAppSecret ...
func VerifyAppCodeAppSecret(appCode, appSecret string) bool {
	// 1. get from cache
	key := appCode + ":" + appSecret

	value, found := LocalAppCodeAppSecretCache.Get(key)
	if found {
		return value.(bool)
	}

	// 2. get from database
	manager := edao.NewAppSecretManager()
	valid, err := manager.Exists(appCode, appSecret)
	if err != nil {
		log.Errorf("verify app_code_app_secret from bk_paas+esb_app_account fail, app_code=%s, app_secret=%s, err=%s",
			appCode,
			stringx.Truncate(appSecret, 6)+"******",
			err)
		return false
	}

	// 3. set to cache, default 12 hours, if not valid, only keep in cache for 1 minutes
	//    in case of auth server down, we can still get the valid matched accessKeys from cache
	ttl := gocache.DefaultExpiration
	if !valid {
		ttl = 1 * time.Minute
	}
	LocalAppCodeAppSecretCache.Set(key, valid, ttl)
	return valid
}

func VerifyAppCodeAppSecretFromAuth(appCode, appSecret string) bool {
	// 1. get from cache
	key := appCode + ":" + appSecret

	value, found := LocalAuthAppAccessKeyCache.Get(key)
	if found {
		return value.(bool)
	}

	// 2. get from auth
	valid, err := component.BkAuth.Verify(appCode, appSecret)
	if err != nil {
		log.Errorf("verify app_code_app_secret from auth fail, app_code=%s, app_secret=%s, err=%s",
			appCode,
			stringx.Truncate(appSecret, 6)+"******",
			err)
		return false
	}

	// 3. set to cache, default 12 hours, if not valid, only keep in cache for 1 minutes
	//    in case of auth server down, we can still get the valid matched accessKeys from cache
	ttl := gocache.DefaultExpiration
	if !valid {
		ttl = 1 * time.Minute
	}
	LocalAuthAppAccessKeyCache.Set(key, valid, ttl)
	return valid
}
