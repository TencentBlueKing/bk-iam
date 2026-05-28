/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package middleware

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"

	"iam/pkg/cacheimpls"
	"iam/pkg/util"
)

var (
	ErrUnauthorized = errors.New("jwtauth: token is unauthorized")

	ErrExpired    = errors.New("jwtauth: token is expired")
	ErrNBFInvalid = errors.New("jwtauth: token nbf validation failed")
	ErrIATInvalid = errors.New("jwtauth: token iat validation failed")

	ErrAPIGatewayJWTMissingApp             = errors.New("app not in jwt claims")
	ErrAPIGatewayJWTAppInfoParseFail       = errors.New("app info parse fail")
	ErrAPIGatewayJWTAppInfoNoAppCode       = errors.New("app_code not in app info")
	ErrAPIGatewayJWTAppCodeNotString       = errors.New("app_code not string")
	ErrAPIGatewayJWTAppInfoNoVerified      = errors.New("verified not in app info")
	ErrAPIGatewayJWTAppInfoVerifiedNotBool = errors.New("verified not bool")
	ErrAPIGatewayJWTAppNotVerified         = errors.New("app not verified")

	ErrAPIGatewayJWTAppTenantModeNotString = errors.New("tenant_mode not string")
	ErrAPIGatewayJWTAppTenantIDNotString   = errors.New("tenant_id not string")
	ErrAPIGatewayJWTAppTenantModeInvalid   = errors.New("tenant_mode invalid, must be global or single")
	ErrAPIGatewayJWTAppSingleTenantNoID    = errors.New("single tenant app must have tenant_id")
)

type AppInfoDTO struct {
	AppCode    string
	TenantMode string // 租户模式：global / single
	TenantID   string // 租户 id（global 时为空）
}

// 从网关 JWT 中解析出 app_code、tenant_mode、tenant_id
func getAppInfoFromJWTToken(jwtToken string, apiGatewayPublicKey []byte) (AppInfoDTO, error) {
	// check if in cache
	cachedAppInfo, err := cacheimpls.GetJWTTokenAppInfo(jwtToken)
	if err == nil {
		// 缓存命中
		return AppInfoDTO{
			AppCode:    cachedAppInfo.AppCode,
			TenantMode: cachedAppInfo.TenantMode,
			TenantID:   cachedAppInfo.TenantID,
		}, nil
	}

	// parse in time
	appInfo, err := verifyAppInfo(jwtToken, apiGatewayPublicKey)
	if err != nil {
		return AppInfoDTO{}, err
	}

	// set into cache
	cacheimpls.SetJWTTokenAppInfo(jwtToken, cacheimpls.AppInfo{
		AppCode:    appInfo.AppCode,
		TenantMode: appInfo.TenantMode,
		TenantID:   appInfo.TenantID,
	})
	return appInfo, nil
}

func parseBKJWTToken(tokenString string, publicKey []byte) (jwt.MapClaims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
		if err != nil {
			return pubKey, fmt.Errorf("jwt parse fail, err=%w", err)
		}
		return pubKey, nil
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		if verr, ok := err.(*jwt.ValidationError); ok {
			switch {
			case verr.Errors&jwt.ValidationErrorExpired > 0:
				return nil, ErrExpired
			case verr.Errors&jwt.ValidationErrorIssuedAt > 0:
				return nil, ErrIATInvalid
			case verr.Errors&jwt.ValidationErrorNotValidYet > 0:
				return nil, ErrNBFInvalid
			}
		}
		return nil, err
	}

	if !token.Valid {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

// verifyAppInfo 从 JWT token 中提取并校验 app 信息
func verifyAppInfo(jwtToken string, publicKey []byte) (AppInfoDTO, error) {
	// 1. 解析 JWT token
	claims, err := parseBKJWTToken(jwtToken, publicKey)
	if err != nil {
		return AppInfoDTO{}, err
	}

	// 2. 从 claims 中提取 app 字段
	appInfo, ok := claims["app"]
	if !ok {
		return AppInfoDTO{}, ErrAPIGatewayJWTMissingApp
	}

	app, ok := appInfo.(map[string]interface{})
	if !ok {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppInfoParseFail
	}

	// 3. verified 必须为 true
	verifiedRaw, ok := app["verified"]
	if !ok {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppInfoNoVerified
	}
	verified, ok := verifiedRaw.(bool)
	if !ok {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppInfoVerifiedNotBool
	}
	if !verified {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppNotVerified
	}

	// 4. app_code
	appCodeRaw, ok := app["app_code"]
	if !ok {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppInfoNoAppCode
	}
	appCode, ok := appCodeRaw.(string)
	if !ok {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppCodeNotString
	}

	// 5. tenant_mode
	var tenantMode string
	if v, exists := app["tenant_mode"]; exists && v != nil {
		tenantMode, ok = v.(string)
		if !ok {
			return AppInfoDTO{}, ErrAPIGatewayJWTAppTenantModeNotString
		}
	}

	// 6. tenant_id
	var tenantID string
	if v, exists := app["tenant_id"]; exists && v != nil {
		tenantID, ok = v.(string)
		if !ok {
			return AppInfoDTO{}, ErrAPIGatewayJWTAppTenantIDNotString
		}
	}

	// 7. 校验 tenant_mode 合法性（仅当存在时）
	if tenantMode != "" && tenantMode != util.AppTenantModeGlobal && tenantMode != util.AppTenantModeSingle {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppTenantModeInvalid
	}

	// 8. single 模式必须带 tenant_id
	if tenantMode == util.AppTenantModeSingle && tenantID == "" {
		return AppInfoDTO{}, ErrAPIGatewayJWTAppSingleTenantNoID
	}

	return AppInfoDTO{
		AppCode:    appCode,
		TenantMode: tenantMode,
		TenantID:   tenantID,
	}, nil
}
