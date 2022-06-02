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
	"fmt"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cacheimpls"
	"iam/pkg/config"
	"iam/pkg/util"
)

const (
	APIGatewayRequest = "apigw"
)

// NewClientAuthMiddleware create the middleware by config, support both raw app_code/app_secret and APIGateway
func NewClientAuthMiddleware(c *config.Config) gin.HandlerFunc {
	var apiGatewayPublicKey []byte
	apigwCrypto, ok := c.Cryptos["apigateway_public_key"]
	if ok {
		apiGatewayPublicKey = []byte(apigwCrypto.Key)
	}

	return ClientAuthMiddleware(apiGatewayPublicKey)
}

// ClientAuthMiddleware ...
func ClientAuthMiddleware(apiGatewayPublicKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debug("Middleware: ClientAuthMiddleware")

		requestFrom := c.GetHeader("X-Bkapi-From")

		var clientID string

		// X-Bkapi-From: apigw
		if requestFrom == APIGatewayRequest {
			jwtToken := c.GetHeader("X-Bkapi-JWT")
			// should not be empty
			if jwtToken == "" {
				util.UnauthorizedJSONResponse(c, "request from apigateway jwt token should not be empty!")
				c.Abort()
				return
			}
			if len(apiGatewayPublicKey) == 0 {
				util.UnauthorizedJSONResponse(
					c,
					"iam apigateway public key is not configured, not support request from apigateway",
				)
				c.Abort()
				return
			}

			var err error
			clientID, err = getClientIDFromJWTToken(jwtToken, apiGatewayPublicKey)
			if err != nil {
				message := fmt.Sprintf("request from apigateway jwt token invalid! err=%s", err.Error())
				util.UnauthorizedJSONResponse(c, message)
				c.Abort()
				return
			}
		} else {
			appCode := c.GetHeader("X-Bk-App-Code")
			appSecret := c.GetHeader("X-Bk-App-Secret")

			// 1. check not empty
			if appCode == "" || appSecret == "" {
				util.UnauthorizedJSONResponse(c, "app code and app secret required")
				c.Abort()
				return
			}

			// 2. validate from cache -> database
			valid := cacheimpls.VerifyAppCodeAppSecret(appCode, appSecret)
			if !valid {
				util.UnauthorizedJSONResponse(c, "app code or app secret wrong")
				c.Abort()
				return
			}

			clientID = appCode
		}

		// 3. set client_id
		util.SetClientID(c, clientID)

		c.Next()
	}
}

// SuperClientMiddleware ...
func SuperClientMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debug("Middleware: SuperClientMiddleware")

		appCode := util.GetClientID(c)
		if !config.SuperAppCodeSet.Has(appCode) {
			util.UnauthorizedJSONResponse(c, "super client app code wrong")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ShareClientMiddleware check if the client can access the APIs of the model share
func ShareClientMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debug("Middleware: ShareClientMiddleware")

		appCode := util.GetClientID(c)
		if !config.ShareAppCodeSet.Has(appCode) {
			util.UnauthorizedJSONResponse(
				c,
				"share client app code wrong, app_code is not in the share client whitelist",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}
