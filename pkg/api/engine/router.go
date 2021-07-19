/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package engine

import (
	"github.com/gin-gonic/gin"

	"iam/pkg/api/engine/handler"
)

// in-out policy `id` equals to `pk` in iam backend, so do the convert in serializer

// Register ...
func Register(r *gin.RouterGroup) {
	// GET /api/v1/engine/policies 拉取指定条件的策略列表
	r.GET("/policies", handler.ListPolicy)

	// GET /api/v1/engine/policies/ids 查询指定条件的策略ID List
	r.GET("/policies/ids", handler.ListPolicyPKs)

	// GET /api/v1/engine/policies/ids/max 查询指定条件的策略最大ID
	r.GET("/policies/ids/max", handler.GetMaxPolicyPK)

	// GET /api/v1/engine/systems/:system_id 查询系统信息
	r.GET("/systems/:system_id", handler.GetSystem)

	// POST /api/v1/engine/credentials/verify 认证信息验证
	r.POST("/credentials/verify", handler.CredentialsVerify)
}
