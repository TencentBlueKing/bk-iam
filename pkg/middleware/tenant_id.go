/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云 - 权限中心 (BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package middleware

import (
	"iam/pkg/util"

	"github.com/gin-gonic/gin"
)

func BkTenantID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 tenant_id
		tenantID := c.GetHeader(util.BkTenantIDHeaderKey)
		// FIXME (nan): 需要根据是否开启多租户模式，强制要求 tenant_id 存在

		// 将 tenant_id 存储在上下文中，以便后续处理中使用
		util.SetBkTenantID(c, tenantID)

		// 继续处理请求
		c.Next()
	}
}
