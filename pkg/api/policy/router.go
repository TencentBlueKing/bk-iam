/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"github.com/gin-gonic/gin"

	"iam/pkg/api/common"
	"iam/pkg/api/policy/handler"
)

// Register the urls: /api/v1/policy
func Register(r *gin.RouterGroup) {
	// in auth.go
	// 鉴权
	r.POST("/auth", handler.Auth)
	// 批量鉴权 - actions批量
	r.POST("/auth_by_actions", handler.BatchAuthByActions)
	// 批量鉴权 - resources批量
	r.POST("/auth_by_resources", handler.BatchAuthByResources)

	// in query.go
	// 查询
	r.POST("/query", handler.Query)
	// 批量查询
	r.POST("/query_by_actions", handler.BatchQueryByActions)
	// 批量第三方依赖策略查询
	r.POST("/query_by_ext_resources", handler.QueryByExtResources)
}

// RegisterV2 will register the urls: /api/v2/policy/systems/:system_id/[auth|query|...]/  (has suffix slash!)
func RegisterV2(r *gin.RouterGroup) {
	// all resource in system
	s := r.Group("/systems/:system_id")
	// validate: 1) system exists 2) client_id is valid, in system.clients
	s.Use(common.SystemExistsAndClientValid())
	{
		// in auth_v2.go
		// 鉴权
		s.POST("/auth/", handler.AuthV2)

		// in auth_v2.go
		// 批量鉴权
		r.POST("/auth_by_actions/", handler.BatchAuthV2ByActions)

		// in query_v2.go
		// 查询
		s.POST("/query/", handler.QueryV2)

		// in query_v2.go
		// 批量查询
		r.POST("/query_by_actions/", handler.BatchQueryByActions)
	}
}
