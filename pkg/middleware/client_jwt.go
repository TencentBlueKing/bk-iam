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

// 租户模式常量（与 APIGateway 签发的 JWT 保持一致）
const (
	TenantModeGlobal = "global" // 全租户应用
	TenantModeSingle = "single" // 单租户应用
)

// AppInfo 从网关 JWT 中解析出的应用信息
type AppInfo struct {
	AppCode    string // 应用 code
	TenantMode string // 租户模式：global / single
	TenantID   string // 租户 id（global 时为空）
}

func getClientIDFromJWTToken(jwtToken string, apiGatewayPublicKey []byte) (clientID string, err error) {
	// check if in cache
	clientID, err = cacheimpls.GetJWTTokenClientID(jwtToken)
	if err == nil {
		return clientID, nil
	}

	// parse in time
	clientID, err = verifyClientID(jwtToken, apiGatewayPublicKey)
	if err != nil {
		return "", err
	}
	// set into cache
	cacheimpls.SetJWTTokenClientID(jwtToken, clientID)
	return clientID, nil
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

func verifyClientID(jwtToken string, publicKey []byte) (clientID string, err error) {
	var claims jwt.MapClaims
	claims, err = parseBKJWTToken(jwtToken, publicKey)
	if err != nil {
		return
	}

	appInfo, ok := claims["app"]
	if !ok {
		err = ErrAPIGatewayJWTMissingApp
		return
	}

	app, ok := appInfo.(map[string]interface{})
	if !ok {
		err = ErrAPIGatewayJWTAppInfoParseFail
		return
	}

	verifiedRaw, ok := app["verified"]
	if !ok {
		err = ErrAPIGatewayJWTAppInfoNoVerified
		return
	}

	verified, ok := verifiedRaw.(bool)
	if !ok {
		err = ErrAPIGatewayJWTAppInfoVerifiedNotBool
		return
	}

	if !verified {
		err = ErrAPIGatewayJWTAppNotVerified
		return
	}

	appCode, ok := app["app_code"]
	if !ok {
		err = ErrAPIGatewayJWTAppInfoNoAppCode
		return
	}

	clientID, ok = appCode.(string)
	if !ok {
		err = ErrAPIGatewayJWTAppCodeNotString
		return
	}

	return clientID, nil
}

// 从网关 JWT 中解析出 app_code、tenant_mode、tenant_id
func getAppInfoFromJWTToken(jwtToken string, apiGatewayPublicKey []byte) (AppInfo, error) {
	claims, err := parseBKJWTToken(jwtToken, apiGatewayPublicKey)
	if err != nil {
		return AppInfo{}, err
	}

	return verifyAppInfo(claims)
}

// verifyAppInfo 从 jwt claims 中提取并校验 app 信息
func verifyAppInfo(claims jwt.MapClaims) (AppInfo, error) {
	appInfo, ok := claims["app"]
	if !ok {
		return AppInfo{}, ErrAPIGatewayJWTMissingApp
	}

	app, ok := appInfo.(map[string]interface{})
	if !ok {
		return AppInfo{}, ErrAPIGatewayJWTAppInfoParseFail
	}

	// 1. verified 必须为 true
	verifiedRaw, ok := app["verified"]
	if !ok {
		return AppInfo{}, ErrAPIGatewayJWTAppInfoNoVerified
	}
	verified, ok := verifiedRaw.(bool)
	if !ok {
		return AppInfo{}, ErrAPIGatewayJWTAppInfoVerifiedNotBool
	}
	if !verified {
		return AppInfo{}, ErrAPIGatewayJWTAppNotVerified
	}

	// 2. app_code
	appCodeRaw, ok := app["app_code"]
	if !ok {
		return AppInfo{}, ErrAPIGatewayJWTAppInfoNoAppCode
	}
	appCode, ok := appCodeRaw.(string)
	if !ok {
		return AppInfo{}, ErrAPIGatewayJWTAppCodeNotString
	}

	// 3. tenant_mode
	var tenantMode string
	if v, exists := app["tenant_mode"]; exists && v != nil {
		tenantMode, ok = v.(string)
		if !ok {
			return AppInfo{}, ErrAPIGatewayJWTAppTenantModeNotString
		}
	}

	// 4. tenant_id
	var tenantID string
	if v, exists := app["tenant_id"]; exists && v != nil {
		tenantID, ok = v.(string)
		if !ok {
			return AppInfo{}, ErrAPIGatewayJWTAppTenantIDNotString
		}
	}

	// 5. 校验 tenant_mode 合法性（仅当存在时）
	if tenantMode != "" && tenantMode != TenantModeGlobal && tenantMode != TenantModeSingle {
		return AppInfo{}, ErrAPIGatewayJWTAppTenantModeInvalid
	}

	// 6. single 模式必须带 tenant_id
	if tenantMode == TenantModeSingle && tenantID == "" {
		return AppInfo{}, ErrAPIGatewayJWTAppSingleTenantNoID
	}

	return AppInfo{
		AppCode:    appCode,
		TenantMode: tenantMode,
		TenantID:   tenantID,
	}, nil
}
