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

// APIGatewayJWTClientIDCacheKey cache key for JWTToken
type APIGatewayJWTClientIDCacheKey struct {
	JWTToken string
}

// Key ...
func (k APIGatewayJWTClientIDCacheKey) Key() string {
	return stringx.MD5Hash(k.JWTToken)
}

func retrieveAPIGatewayJWTClientID(key cache.Key) (interface{}, error) {
	// NOTE: this func not work
	return "", nil
}

var (
	ErrAPIGatewayJWTCacheNotFound     = errors.New("not found")
	ErrAPIGatewayJWTClientIDNotString = errors.New("clientID not string")
)

// GetJWTTokenClientID will retrieve the clientID of a jwtTOken
func GetJWTTokenClientID(jwtToken string) (clientID string, err error) {
	key := APIGatewayJWTClientIDCacheKey{
		JWTToken: jwtToken,
	}

	value, ok := LocalAPIGatewayJWTClientIDCache.DirectGet(key)
	if !ok {
		err = ErrAPIGatewayJWTCacheNotFound
		return
	}

	clientID, ok = value.(string)
	if !ok {
		err = ErrAPIGatewayJWTClientIDNotString
		return
	}
	return clientID, nil
}

// SetJWTTokenClientID will set the jwtToken-clientID int cache
func SetJWTTokenClientID(jwtToken string, clientID string) {
	key := APIGatewayJWTClientIDCacheKey{
		JWTToken: jwtToken,
	}
	LocalAPIGatewayJWTClientIDCache.Set(key, clientID)
}
