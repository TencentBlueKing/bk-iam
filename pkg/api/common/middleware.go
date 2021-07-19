/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	"fmt"

	"iam/pkg/cache/impls"
	"iam/pkg/util"

	"github.com/gin-gonic/gin"
)

// SystemExists via system_id in path
func SystemExists() gin.HandlerFunc {
	return func(c *gin.Context) {
		systemID := c.Param("system_id")
		if systemID == "" {
			util.BadRequestErrorJSONResponse(c, "system_id in url required")
			c.Abort()
			return
		}

		// use cache here
		_, err := impls.GetSystem(systemID)
		if err != nil {
			util.NotFoundJSONResponse(c, fmt.Sprintf("system(%s) not exists", systemID))
			c.Abort()
			return
		}

		c.Next()
	}
}

// SystemExistsAndClientValid ...
func SystemExistsAndClientValid() gin.HandlerFunc {
	return func(c *gin.Context) {
		systemID := c.Param("system_id")

		if systemID == "" {
			util.BadRequestErrorJSONResponse(c, "system_id in url required")
			c.Abort()
			return
		}

		// check system is exists
		// use cache here
		system, err := impls.GetSystem(systemID)
		if err != nil {
			util.NotFoundJSONResponse(c, fmt.Sprintf("system(%s) not exists", systemID))
			c.Abort()
			return
		}

		clientID := util.GetClientID(c)
		if clientID == "" {
			util.UnauthorizedJSONResponse(c, "app code and app secret required")
			c.Abort()
			return
		}

		validClients := util.SplitStringToSet(system.Clients, ",")
		if !validClients.Has(clientID) {
			util.UnauthorizedJSONResponse(c,
				fmt.Sprintf("app(%s) is not allowed to call system (%s) api", clientID, systemID))

			c.Abort()
			return
		}

		c.Next()
	}
}
