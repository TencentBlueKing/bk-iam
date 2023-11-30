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
	modelHandler "iam/pkg/api/model/handler"
	"iam/pkg/api/web/handler"
)

// Register ...
func Register(r *gin.RouterGroup) {
	// 系统列表
	r.GET("/systems", handler.ListSystem)

	// 资源类型列表
	r.GET("/resource-types", handler.ListResourceType)

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
		s.DELETE("/actions/:action_id", modelHandler.DeleteAction)

		s.GET("/instance-selections", handler.ListInstanceSelection)

		// system_settings
		s.GET("/system-settings/:name", handler.GetSystemSettings)

		// policy列表
		s.GET("/policies", handler.ListSystemPolicy)
		// policies 变更
		s.POST("/policies", handler.AlterPolicies)
		// 获取自定义申请的策略
		// 根据Action删除策略
		s.DELETE("/actions/:action_id/policies", handler.DeleteActionPolicies)

		// temporary policy
		// 创建临时权限
		s.POST("/temporary-policies", handler.CreateTemporaryPolicies)

		// 带分页 https://github.com/TencentBlueKing/bk-iam-saas/issues/1155
		s.GET("/subject-groups", handler.ListSystemSubjectGroups)
	}

	// policy
	{
		// 查询过期的 policy 列表
		r.GET("/policies", handler.ListPolicy)
		// 删除
		r.DELETE("/policies", handler.BatchDeletePolicies)
	}

	// temporary-policy
	{
		// temporary-policies 删除
		r.DELETE("/temporary-policies", handler.BatchDeleteTemporaryPolicies)
		r.DELETE("/temporary-policies/before_expired_at", handler.DeleteTemporaryBeforeExpiredAt)
	}

	// subject
	{
		// 查询subject列表
		r.GET("/subjects", handler.ListSubject)
		// 创建subject
		r.POST("/subjects", handler.BatchCreateSubjects)
		// 删除subject
		r.DELETE("/subjects", handler.BatchDeleteSubjects)
		// 更新subject
		r.PUT("/subjects", handler.BatchUpdateSubject)

		// TODO: change the url? here is groups
		// 筛选有过期成员的subjects
		r.POST("/subjects/before_expired_at", handler.ListExistGroupsHasMemberBeforeExpiredAt)
	}

	// group-members
	{
		// 查询subject的成员列表
		r.GET("/group-members", handler.ListGroupMember)
		// 批量添加subject成员
		r.POST("/group-members", handler.BatchAddGroupMembers)
		// 批量删除subject成员
		r.DELETE("/group-members", handler.BatchDeleteGroupMembers)
		// 批量subject成员过期时间
		r.PUT("/group-members/expired_at", handler.BatchUpdateGroupMembersExpiredAt)
		// 查询小于指定过期时间的成员列表, 批量用户组查询
		r.GET("/group-members/query", handler.ListGroupMemberBeforeExpiredAt)
	}

	// subject-template-groups
	{
		// 查询subject的成员列表
		r.GET("/template-group-members", handler.ListTemplateGroupMember)
		// 创建subject-template-group
		r.POST("/subject-template-groups", handler.BatchCreateSubjectTemplateGroup)
		// 删除subject-template-group
		r.DELETE("/subject-template-groups", handler.BatchDeleteSubjectTemplateGroup) // NEED FIX
		// 批量subject-template-group成员过期时间
		r.PUT("/subject-template-groups/expired_at", handler.BatchUpdateSubjectTemplateGroupExpiredAt)
	}

	// Resource: subject-departments
	{
		// 查询subject-department关系
		r.GET("/subject-departments", handler.ListSubjectDepartments)
		// 创建subject-department关系
		r.POST("/subject-departments", handler.BatchCreateSubjectDepartments)
		// 更新subject-department关系
		r.PUT("/subject-departments", handler.BatchUpdateSubjectDepartments)
		// 删除subject-department关系
		r.DELETE("/subject-departments", handler.BatchDeleteSubjectDepartments)
	}

	// subject-groups
	{
		// 带分页 https://github.com/TencentBlueKing/bk-iam-saas/issues/1155
		r.GET("/subject-groups", handler.ListSubjectGroups)

		// add subject-groups?type=user&id=tome&groups=1,2,3,4,5
		r.GET("/subjects-groups/belong", handler.CheckSubjectGroupsBelong)

		// subject-groups?type=user&id=tome&groups=1,2,3,4,5
		r.GET("/subjects-groups/quota", handler.CheckSubjectGroupsQuota)
	}

	// group-subject
	r.GET("/group-subject/before_expired_at", handler.ListGroupSubjectBeforeExpiredAt)

	// Resource: role-subjects
	{
		// 批量添加role subjects
		r.POST("/role-subjects", handler.BatchAddRoleSubject)
		// 批量删除role subject
		r.DELETE("/role-subjects", handler.BatchDeleteRoleSubject)
	}

	// others
	{
		// 模型变更事件
		r.GET("/model-change-event", handler.ListModelChangeEvent)
		r.PUT("/model-change-event/:event_pk", handler.UpdateModelChangeEvent)
		r.DELETE("/model-change-event", handler.BatchDeleteModelChangeEvent)

		// 清理未引用的expression
		r.DELETE("/unreferenced-expressions", handler.DeleteUnreferencedExpressions)
	}

	fg := r.Group("/freeze/subjects")
	{
		// 查询冻结用户列表
		fg.GET("", handler.ListFreezedSubjects)
		// 批量冻结
		fg.POST("", handler.BatchFreezeSubjects)
		// 批量解冻
		fg.DELETE("", handler.BatchUnfreezeSubjects)
	}
}

// RegisterV2 ...
func RegisterV2(r *gin.RouterGroup) {
	s := r.Group("/systems/:system_id")
	s.Use(common.SystemExists())
	{
		// policies 变更
		s.POST("/policies", handler.AlterPoliciesV2)

		// 查询rbac有资源实例权限的用户组
		s.POST("/rbac/resource-groups", handler.QueryRbacGroupByResource)
	}
}
