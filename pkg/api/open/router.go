/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package open

import (
	"github.com/gin-gonic/gin"

	"iam/pkg/api/common"
	"iam/pkg/api/open/handler"
)

// RegisterLegacySystemAPIs the urls: /api/v1/systems
// NOTE: should not add more apis here, move to /api/v1/open/systems/{system_id}/
func RegisterLegacySystemAPIs(r *gin.RouterGroup) {
	policies := r.Group("/:system_id/policies")
	policies.Use(common.SystemExistsAndClientValid())
	{
		// GET /api/v1/systems/:system/policies?action=x    拉取某个操作的所有策略列表
		policies.GET("", handler.PolicyList)

		// GET /api/v1/systems/:system/policies/:policy_id  查询某个策略详情(这个策略必须属于本系统)
		policies.GET("/:policy_id", handler.PolicyGet)

		// https://cloud.google.com/apis/design/design_patterns#list_sub-collections
		// GET /api/v1/systems/:system/policies/-/subjects?ids=1,2,3,4
		policies.GET("/-/subjects", handler.PoliciesSubjects)
	}
}

// Register the urls: /api/v1/open
func Register(r *gin.RouterGroup) {
	// 1. system scope /api/v1/open/systems

	// 1.1 policies
	policies := r.Group("/systems/:system_id/policies")
	policies.Use(common.SystemExistsAndClientValid())
	{
		// GET /api/v1/open/systems/:system_id/policies?action=x    拉取某个操作的所有策略列表
		policies.GET("", handler.PolicyList)

		// GET /api/v1/open/systems/:system_id/policies/:policy_id  查询某个策略详情(这个策略必须属于本系统)
		policies.GET("/:policy_id", handler.PolicyGet)

		// https://cloud.google.com/apis/design/design_patterns#list_sub-collections
		// GET /api/v1/open/systems/:system_id/policies/-/subjects?ids=1,2,3,4
		policies.GET("/-/subjects", handler.PoliciesSubjects)
	}

	// 2. subjects: users, departments, groups
	users := r.Group("/users")
	{
		// GET /user/123/groups?inherit=true
		users.GET("/:user_id/groups", handler.UserGroups)
	}

	departments := r.Group("/departments")
	{
		// GET /department/456/groups
		departments.GET("/:department_id/groups", handler.DepartmentGroups)
	}

	// 3. groups
	// groups := r.Group("/groups")
	// {
	// 	groups.GET("/:group_id/members", handler.GroupMembers)
	// 	groups.GET("/", handler.Groups)
	// 	groups.GET("/:group_id", handler.GroupGet)
	// }
}
