/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package debug

import (
	"github.com/gin-gonic/gin"

	"iam/pkg/api/debug/handler"
)

// Register ...
func Register(r *gin.RouterGroup) {
	q := r.Group("/query")
	{
		// 查询系统所有模型 /api/v1/debug/query/model?system=x
		q.GET("/model", handler.QueryModel)

		// 查询某个系统的所有actions [带pk]  /api/v1/debug/query/action?system=x
		q.GET("/action", handler.QueryActions)

		// 查询subject及其上级关系(部门/部门-组/组)  /api/v1/debug/query/subject?type=user&id=1
		q.GET("/subject", handler.QuerySubjects)

		// 策略查询接口  /api/v1/debug/query/policy?system=&subject_type=&subject_id=&action=    &force=1 &debug=1
		q.GET("/policy", handler.QueryPolicies)
	}

	c := r.Group("/cache")
	{
		// 获取cache中的策略 /api/v1/debug/cache/policy?system=&subject_type=&subject_id=    &action=
		c.GET("/policy", handler.QueryPolicyCache)
		// 查询缓存中的expression   /api/v1/debug/cache/expression?pks=1,2,3,4
		c.GET("/expression", handler.QueryExpressionCache)

		// TODO:  精准删除缓存 => policy / expression
	}
}
