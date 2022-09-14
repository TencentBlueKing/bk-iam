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
	"errors"
	"strings"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"iam/pkg/abac/pap"
	"iam/pkg/api/common"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// ListGroupMember 查询用户组的成员列表
func ListGroupMember(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListGroupMember")

	var subject listGroupMemberSerializer
	if err := c.ShouldBindQuery(&subject); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	subject.Default()

	ctl := pap.NewGroupController()
	count, err := ctl.GetGroupMemberCount(subject.Type, subject.ID)
	if err != nil {
		err = errorWrapf(err, "type=`%s`, id=`%s`", subject.Type, subject.ID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	relations, err := ctl.ListPagingGroupMember(subject.Type, subject.ID, subject.Limit, subject.Offset)
	if err != nil {
		err = errorWrapf(
			err, "ctl.ListPagingGroupMember type=`%s`, id=`%s`, limit=`%d`, offset=`%d`",
			subject.Type, subject.ID, subject.Limit, subject.Offset,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{
		"count":   count,
		"results": relations,
	})
}

// ListSubjectGroups 获取subject关联的用户组
func ListSubjectGroups(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListSubjectGroups")

	var query listSubjectGroupSerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	query.Default()

	ctl := pap.NewGroupController()

	count, err := ctl.GetSubjectGroupCountBeforeExpiredAt(query.Type, query.ID, query.BeforeExpiredAt)
	if err != nil {
		err = errorWrapf(err, "type=`%s`, id=`%s`", query.Type, query.ID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	groups, err := ctl.ListPagingSubjectGroups(
		query.Type,
		query.ID,
		query.BeforeExpiredAt,
		query.Limit,
		query.Offset,
	)
	if err != nil {
		err = errorx.Wrapf(
			err,
			"Handler",
			"ctl.ListPagingSubjectGroups",
			"type=`%s`, id=`%s`, expiredAt=`%d`, limit=`%d`, offset=`%d`",
			query.Type,
			query.ID,
			query.BeforeExpiredAt,
			query.Limit,
			query.Offset,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{
		"count":   count,
		"results": groups,
	})
}

// CheckSubjectGroupsBelong ...
func CheckSubjectGroupsBelong(c *gin.Context) {
	var query checkSubjectGroupsBelongSerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	// input: subject.type= & subject.id= & group_ids=1,2,3,4
	// output: 个人组 + 个人-部门-组 列表中, 是否包含了这批 group_ids
	// 条件: 有效的, 即未过期的
	groupIDs := strings.Split(query.GroupIDs, ",")
	if len(groupIDs) > 100 {
		util.BadRequestErrorJSONResponse(c, "group_ids should be less than 100")
		return
	}

	ctl := pap.NewGroupController()
	groupIDBelong, err := ctl.CheckSubjectEffectGroups(query.Type, query.ID, query.Inherit, groupIDs)
	if err != nil {
		err = errorx.Wrapf(
			err,
			"Handler",
			"ctl.CheckSubjectEffectGroups type=`%s`, id=`%s` fail",
			query.Type,
			query.ID,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", groupIDBelong)
}

// BatchUpdateGroupMembersExpiredAt subject关系续期
func BatchUpdateGroupMembersExpiredAt(c *gin.Context) {
	var body groupMemberExpiredAtSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchUpdateGroupMembersExpiredAt")

	papSubjects := make([]pap.GroupMember, 0, len(body.Members))
	for _, m := range body.Members {
		papSubjects = append(papSubjects, pap.GroupMember{
			Type:      m.Type,
			ID:        m.ID,
			ExpiredAt: m.ExpiredAt,
		})
	}

	ctl := pap.NewGroupController()
	err := ctl.UpdateGroupMembersExpiredAt(body.Type, body.ID, papSubjects)
	if err != nil {
		err = errorWrapf(err, "ctl.UpdateGroupMembersExpiredAt",
			"type=`%s`, id=`%s`, subjects=`%+v`", body.Type, body.ID, papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// BatchDeleteGroupMembers 批量删除subject成员
func BatchDeleteGroupMembers(c *gin.Context) {
	var body deleteGroupMemberSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if valid, message := common.ValidateArray(body.Members); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	papSubjects := make([]pap.Subject, 0, len(body.Members))
	copier.Copy(&papSubjects, &body.Members)

	ctl := pap.NewGroupController()
	typeCount, err := ctl.DeleteGroupMembers(body.Type, body.ID, papSubjects)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "ctl.DeleteGroupMembers",
			"type=`%s`, id=`%s`, subjects=`%+v`", body.Type, body.ID, papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", typeCount)
}

// BatchAddGroupMembers 批量添加subject成员
func BatchAddGroupMembers(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchAddGroupMembers")

	var body addGroupMembersSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	papSubjects := make([]pap.GroupMember, 0, len(body.Members))
	for _, m := range body.Members {
		papSubjects = append(papSubjects, pap.GroupMember{
			Type:      m.Type,
			ID:        m.ID,
			ExpiredAt: body.ExpiredAt,
		})
	}

	// 检查subject group 数量是否超过限制
	err := checkSubjectGroupsQuota(body.Type, body.ID, papSubjects)
	if err != nil {
		if errors.Is(err, errQuota) {
			util.ConflictJSONResponse(c, err.Error())
			return
		}

		util.SystemErrorJSONResponse(c, err)
		return
	}

	ctl := pap.NewGroupController()
	typeCount, err := ctl.CreateOrUpdateGroupMembers(body.Type, body.ID, papSubjects)
	if err != nil {
		err = errorWrapf(
			err,
			"ctl.CreateOrUpdateGroupMembers",
			"type=`%s`, id=`%s`, subjects=`%+v`",
			body.Type,
			body.ID,
			papSubjects,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", typeCount)
}

// ListGroupMemberBeforeExpiredAt 获取小于指定过期时间的成员列表
func ListGroupMemberBeforeExpiredAt(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListGroupMemberBeforeExpiredAt")

	var body listGroupMemberBeforeExpiredAtSerializer
	if err := c.ShouldBindQuery(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	body.Default()

	ctl := pap.NewGroupController()
	count, err := ctl.GetGroupMemberCountBeforeExpiredAt(body.Type, body.ID, body.BeforeExpiredAt)
	if err != nil {
		err = errorWrapf(
			err, "ctl.GetGroupMemberCountBeforeExpiredAt type=`%s`, id=`%s`, beforeExpiredAt=`%d`",
			body.Type, body.ID, body.BeforeExpiredAt,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	members, err := ctl.ListPagingGroupMemberBeforeExpiredAt(
		body.Type, body.ID, body.BeforeExpiredAt, body.Limit, body.Offset,
	)
	if err != nil {
		err = errorWrapf(
			err, "ctl.ListPagingGroupMemberBeforeExpiredAt type=`%s`, id=`%s`, beforeExpiredAt=`%d`",
			body.Type, body.ID, body.BeforeExpiredAt,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{
		"count":   count,
		"results": members,
	})
}

// ListExistGroupsHasMemberBeforeExpiredAt 筛选出有成员过期的用户组
func ListExistGroupsHasMemberBeforeExpiredAt(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListExistGroupsHasMemberBeforeExpiredAt")

	var body filterSubjectsBeforeExpiredAtSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	papSubjects := make([]pap.Subject, 0, len(body.Subjects))
	copier.Copy(&papSubjects, &body.Subjects)

	ctl := pap.NewGroupController()
	existGroups, err := ctl.FilterGroupsHasMemberBeforeExpiredAt(papSubjects, body.BeforeExpiredAt)
	if err != nil {
		err = errorWrapf(
			err, "ctl.FilterGroupsHasMemberBeforeExpiredAt subjects=`%+v`, beforeExpiredAt=`%d`",
			papSubjects, body.BeforeExpiredAt,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", existGroups)
}

// CheckSubjectGroupsQuota ...
func CheckSubjectGroupsQuota(c *gin.Context) {
	var query checkSubjectGroupsQuotaSerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	// input: subject.type= & subject.id= & group_ids=1,2,3,4
	groupIDs := strings.Split(query.GroupIDs, ",")

	subjects := []pap.GroupMember{{Type: query.Type, ID: query.ID}}
	for _, groupID := range groupIDs {
		err := checkSubjectGroupsQuota(types.GroupType, groupID, subjects)
		if err != nil {
			if errors.Is(err, errQuota) {
				util.ConflictJSONResponse(c, err.Error())
				return
			}

			util.SystemErrorJSONResponse(c, err)
			return
		}
	}

	util.SuccessJSONResponse(c, "ok", nil)
}
