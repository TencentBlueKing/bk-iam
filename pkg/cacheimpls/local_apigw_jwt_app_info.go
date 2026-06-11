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

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/stringx"
)

// AppInfo 从网关 JWT 中解析出的应用信息
type AppInfo struct {
	AppCode    string // 应用 code
	TenantMode string // 租户模式：global / single
	TenantID   string // 租户 id（global 时为空）
}

// APIGatewayJWTAppInfoCacheKey cache key for JWTToken
type APIGatewayJWTAppInfoCacheKey struct {
	JWTToken string
}

// Key ...
func (k APIGatewayJWTAppInfoCacheKey) Key() string {
	return stringx.MD5Hash(k.JWTToken)
}

func retrieveAPIGatewayJWTAppInfo(key cache.Key) (interface{}, error) {
	// NOTE: this func not work
	return AppInfo{}, nil
}

var (
	ErrAPIGatewayJWTCacheNotFound   = errors.New("not found")
	ErrAPIGatewayJWTValueNotAppInfo = errors.New("value not AppInfo type")
)

// GetJWTTokenAppInfo will retrieve the AppInfo of a jwtToken
func GetJWTTokenAppInfo(jwtToken string) (appInfo AppInfo, err error) {
	key := APIGatewayJWTAppInfoCacheKey{
		JWTToken: jwtToken,
	}

	value, ok := LocalAPIGatewayJWTAppInfoCache.DirectGet(key)
	if !ok {
		err = ErrAPIGatewayJWTCacheNotFound
		return
	}

	appInfo, ok = value.(AppInfo)
	if !ok {
		err = ErrAPIGatewayJWTValueNotAppInfo
		return
	}
	return appInfo, nil
}

// SetJWTTokenAppInfo will set the jwtToken-AppInfo in cache
func SetJWTTokenAppInfo(jwtToken string, appInfo AppInfo) {
	key := APIGatewayJWTAppInfoCacheKey{
		JWTToken: jwtToken,
	}
	LocalAPIGatewayJWTAppInfoCache.Set(key, appInfo)
}
