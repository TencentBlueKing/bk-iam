/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"iam/pkg/api/common"
	"iam/pkg/service/types"
)

const superSystemID = "SUPER"

type pageSerializer struct {
	Limit  int64 `json:"limit" form:"limit" binding:"omitempty,min=0"`
	Offset int64 `json:"offset" form:"offset" binding:"omitempty,min=0"`
}

// Default ...
func (s *pageSerializer) Default() {
	if s.Limit == 0 {
		s.Limit = 20
	}
}

type listSubjectSerializer struct {
	Type string `form:"type" binding:"required,oneof=user group department"`
	pageSerializer
}

type createSubjectSerializer struct {
	Type string `json:"type" binding:"required,oneof=user group department"`
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type deleteSubjectSerializer struct {
	Type string `json:"type" binding:"required,oneof=user group department"`
	ID   string `json:"id" binding:"required"`
}

type listGroupMemberSerializer struct {
	Type string `form:"type" binding:"required,oneof=group"`
	ID   string `form:"id" binding:"required"`
	pageSerializer
}

type checkSubjectGroupsBelongSerializer struct {
	Type     string `form:"type" binding:"required,oneof=user department"`
	ID       string `form:"id" binding:"required"`
	GroupIDs string `form:"group_ids" binding:"required"`
	Inherit  bool   `form:"inherit" binding:"omitempty"`
}

type listSubjectGroupSerializer struct {
	Type            string `form:"type" binding:"required,oneof=user department"`
	ID              string `form:"id" binding:"required"`
	BeforeExpiredAt int64  `form:"before_expired_at" binding:"omitempty,min=0"`
	pageSerializer
}

type memberSerializer struct {
	Type string `json:"type" binding:"required,oneof=user department"`
	ID   string `json:"id" binding:"required"`
}

type deleteGroupMemberSerializer struct {
	Type string `json:"type" binding:"required,oneof=group"`
	ID   string `json:"id" binding:"required"`
	// 防御，避免出现一次性删除太多成员，影响性能
	Members []memberSerializer `json:"members" binding:"required,gt=0,lte=1000"`
}

type addGroupMembersSerializer struct {
	Type            string `json:"type" binding:"required,oneof=group"`
	ID              string `json:"id" binding:"required"`
	PolicyExpiredAt int64  `json:"policy_expired_at" binding:"omitempty,min=1,max=4102444800"`
	// 防御，避免出现一次性添加太多成员，影响性能
	Members []memberSerializer `json:"members" binding:"required,gt=0,lte=1000"`
}

func (s *addGroupMembersSerializer) validate() (bool, string) {
	// type为group时必须有过期时间
	if s.Type == types.GroupType && s.PolicyExpiredAt < 1 {
		return false, "policy expires time required when add group member"
	}

	if len(s.Members) > 0 {
		if valid, message := common.ValidateArray(s.Members); !valid {
			return false, message
		}
	}

	return true, "valid"
}

type subjectDepartment struct {
	SubjectID     string   `json:"id" binding:"required"`
	DepartmentIDs []string `json:"departments" binding:"required"`
}

type updateSubjectSerializer struct {
	Type string `json:"type" binding:"required,oneof=user group department"`
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type userSerializer struct {
	Type string `form:"type" binding:"required,oneof=user"`
	ID   string `form:"id" binding:"required"`
}

type baseRoleSubjectSerializer struct {
	RoleType string `form:"role_type" json:"role_type" binding:"required,oneof=super_manager system_manager"`
	SystemID string `form:"system_id" json:"system_id" binding:"required"`
}

func (s *baseRoleSubjectSerializer) validate() (bool, string) {
	if s.RoleType == types.SuperManager && s.SystemID != superSystemID {
		return false, "system_id must be SUPER if role type is super_manager"
	}
	return true, "valid"
}

type roleSubjectSerializer struct {
	baseRoleSubjectSerializer
	Subjects []userSerializer `json:"subjects" binding:"required,gt=0"`
}

func (s *roleSubjectSerializer) validate() (bool, string) {
	if valid, message := s.baseRoleSubjectSerializer.validate(); !valid {
		return valid, message
	}

	if valid, message := common.ValidateArray(s.Subjects); !valid {
		return valid, message
	}

	return true, "valid"
}

type memberExpiredAtSerializer struct {
	memberSerializer
	PolicyExpiredAt int64 `json:"policy_expired_at" binding:"omitempty,min=1,max=4102444800"`
}

type groupMemberExpiredAtSerializer struct {
	Type    string                      `json:"type" binding:"required,oneof=group"`
	ID      string                      `json:"id" binding:"required"`
	Members []memberExpiredAtSerializer `json:"members" binding:"required,gt=0,lte=1000"`
}

func (slz *groupMemberExpiredAtSerializer) validate() (bool, string) {
	if len(slz.Members) > 0 {
		if valid, message := common.ValidateArray(slz.Members); !valid {
			return false, message
		}
	}

	return true, ""
}

type listGroupMemberBeforeExpiredAtSerializer struct {
	listGroupMemberSerializer
	BeforeExpiredAt int64 `form:"before_expired_at" binding:"required,min=1,max=4102444800"`
}

type subjectSerializer struct {
	Type string `json:"type" binding:"required,oneof=group"`
	ID   string `json:"id" binding:"required"`
}

type filterSubjectsBeforeExpiredAtSerializer struct {
	Subjects        []subjectSerializer `json:"subjects" binding:"required,gt=0,lte=1000"`
	BeforeExpiredAt int64               `json:"before_expired_at" binding:"required,min=1,max=4102444800"`
}

func (slz *filterSubjectsBeforeExpiredAtSerializer) validate() (bool, string) {
	if len(slz.Subjects) > 0 {
		if valid, message := common.ValidateArray(slz.Subjects); !valid {
			return false, message
		}
	}

	return true, ""
}
