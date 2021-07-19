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

	"github.com/dgrijalva/jwt-go"

	"iam/pkg/cache/impls"
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
)

func getClientIDFromJWTToken(jwtToken string, apiGatewayPublicKey []byte) (clientID string, err error) {
	// check if in cache
	clientID, err = impls.GetJWTTokenClientID(jwtToken)
	if err == nil {
		return
	}

	// parse in time
	clientID, err = verifyClientID(jwtToken, apiGatewayPublicKey)
	if err != nil {
		return "", err
	}
	// set into cache
	impls.SetJWTTokenClientID(jwtToken, clientID)
	return
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
