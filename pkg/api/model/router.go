/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package model

import (
	"iam/pkg/api/common"
	"iam/pkg/api/model/handler"

	"github.com/gin-gonic/gin"
)

// Register ...
func Register(r *gin.RouterGroup) {
	// system
	r.POST("/systems", handler.CreateSystem)

	// all resource in system
	s := r.Group("/systems/:system_id")
	// validate: 1) system exists 2) client_id is valid, in system.clients
	s.Use(common.SystemExistsAndClientValid())
	{
		// system
		s.PUT("", handler.UpdateSystem)
		s.GET("", handler.GetSystem)

		// system clients
		s.GET("/clients", handler.GetSystemClients)

		// resource_type
		s.POST("/resource-types", handler.BatchCreateResourceTypes)
		s.DELETE("/resource-types", handler.BatchDeleteResourceTypes)

		s.PUT("/resource-types/:resource_type_id", handler.UpdateResourceType)
		s.DELETE("/resource-types/:resource_type_id", handler.DeleteResourceType)

		// instance_selection
		s.POST("/instance-selections", handler.BatchCreateInstanceSelections)
		s.DELETE("/instance-selections", handler.BatchDeleteInstanceSelections)

		s.PUT("/instance-selections/:instance_selection_id", handler.UpdateInstanceSelection)
		s.DELETE("/instance-selections/:instance_selection_id", handler.DeleteInstanceSelection)

		// actions
		s.POST("/actions", handler.BatchCreateActions)
		s.DELETE("/actions", handler.Batchs)

		s.PUT("/actions/:action_id", handler.UpdateAction)
		s.DELETE("/actions/:action_id", handler.)

		// system config
		s.POST("/configs/:name", handler.CreateOrUpdateConfigDispatch)
		s.PUT("/configs/:name", handler.CreateOrUpdateConfigDispatch)

		// query
		s.GET("/query", handler.SystemInfoQuery)

		// token
		s.GET("/token", handler.GetToken)

		// policy
		s.DELETE("/actions/:action_id/policies", handler.DeleteActionPolicies)
	}
}
