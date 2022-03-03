/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package web

import (
	"github.com/gin-gonic/gin"

	"iam/pkg/api/common"
	modelHandle "iam/pkg/api/model/handler"
	"iam/pkg/api/web/handler"
)

// Register ...
func Register(r *gin.RouterGroup) {
	// 系统列表
	r.GET("/systems", handler.ListSystem)

	// all resource in system
	s := r.Group("/systems/:system_id")
	s.Use(common.SystemExists())
	{
		// system 信息
		s.GET("", handler.GetSystem)

		// action列表
		s.GET("/actions", handler.ListAction)
		// action detail
		s.GET("/actions/:action_id", handler.GetAction)
		s.DELETE("/actions/:action_id", modelHandle.DeleteAction)

		s.GET("/instance-selections", handler.ListInstanceSelection)

		// system_settings
		s.GET("/system-settings/:name", handler.GetSystemSettings)

		// policy列表
		s.GET("/policies", handler.ListSystemPolicy)
		// policies 变更
		s.POST("/policies", handler.AlterPolicies)
		// 获取自定义申请的策略
		s.GET("/custom-policy", handler.GetCustomPolicy)
		// 根据Action删除策略
		s.DELETE("/actions/:action_id/policies", handler.DeleteActionPolicies)

		// temporary policy
		// 创建临时权限
		s.POST("/temporary-policies", handler.CreateTemporaryPolicies)
	}

	// 资源类型列表
	r.GET("/resource-types", handler.ListResourceType)

	// 查询过期的 policy 列表
	r.GET("/policies", handler.ListPolicy)

	// 更新策略过期时间
	r.PUT("/policies/expired_at", handler.UpdatePoliciesExpiredAt)

	// Policy 删除
	r.DELETE("/policies", handler.BatchDeletePolicies)

	// temporary-policies 删除
	r.DELETE("/temporary-policies", handler.BatchDeleteTemporaryPolicies)
	r.DELETE("/temporary-policies/before_expired_at", handler.DeleteTemporaryBeforeExpiredAt)

	// 权限模板相关
	pt := r.Group("/perm-templates")
	{
		// 模板授权
		pt.POST("/policies", handler.CreateAndDeleteTemplatePolicies)
		// 模板授权更新
		pt.PUT("/policies", handler.UpdateTemplatePolicies)
		// 删除模板授权
		pt.DELETE("/policies", handler.DeleteSubjectTemplatePolicies)
	}

	// 查询subject列表
	r.GET("/subjects", handler.ListSubject)
	// 创建subject
	r.POST("/subjects", handler.BatchCreateSubjects)
	// 删除subject
	r.DELETE("/subjects", handler.BatchDeleteSubjects)
	// 更新subject
	r.PUT("/subjects", handler.BatchUpdateSubject)
	// 筛选有过期成员的subjects
	r.POST("/subjects/before_expired_at", handler.ListExistSubjectsBeforeExpiredAt)

	// 查询subject的成员列表
	r.GET("/subject-members", handler.ListSubjectMember)
	// 批量添加subject成员
	r.POST("/subject-members", handler.BatchAddSubjectMembers)
	// 批量删除subject成员
	r.DELETE("/subject-members", handler.DeleteSubjectMembers)
	// 批量subject成员过期时间
	r.PUT("/subject-members/expired_at", handler.UpdateSubjectMembersExpiredAt)

	// 查询小于指定过期时间的成员列表, 批量用户组查询
	r.GET("/subject-members/query", handler.ListSubjectMemberBeforeExpiredAt)

	// 查询subject所在的用户组/部门
	r.GET("/subject-relations", handler.GetSubjectGroup)

	// 查询subject-department关系
	r.GET("/subject-departments", handler.ListSubjectDepartments)
	// 创建subject-department关系
	r.POST("/subject-departments", handler.BatchCreateSubjectDepartments)
	// 更新subject-department关系
	r.PUT("/subject-departments", handler.BatchUpdateSubjectDepartments)
	// 删除subject-department关系
	r.DELETE("/subject-departments", handler.BatchDeleteSubjectDepartments)

	// 查询subject role
	r.GET("/subject-roles", handler.ListSubjectRole)
	// 批量添加subject role
	r.POST("/subject-roles", handler.CreateSubjectRole)
	// 批量删除subject role
	r.DELETE("/subject-roles", handler.DeleteSubjectRole)

	// 模型变更事件
	r.GET("/model-change-event", handler.ListModelChangeEvent)
	r.PUT("/model-change-event/:event_pk", handler.UpdateModelChangeEvent)

	// 清理未引用的expression
	r.DELETE("/unreferenced-expressions", handler.DeleteUnreferencedExpressions)
}
