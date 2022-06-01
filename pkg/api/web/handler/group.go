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
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"iam/pkg/abac/pap"
	"iam/pkg/api/common"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// ListSubjectMember 查询用户组的成员列表
func ListSubjectMember(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListSubjectMember")

	var subject listSubjectMemberSerializer
	if err := c.ShouldBindQuery(&subject); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	subject.Default()

	svc := service.NewGroupService()
	count, err := svc.GetMemberCount(subject.Type, subject.ID)
	if err != nil {
		err = errorWrapf(err, "type=`%s`, id=`%s`", subject.Type, subject.ID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	relations, err := svc.ListPagingMember(subject.Type, subject.ID, subject.Limit, subject.Offset)
	if err != nil {
		err = errorWrapf(
			err, "svc.ListPagingMember type=`%s`, id=`%s`, limit=`%d`, offset=`%d`",
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

// GetSubjectGroup 获取subject关联的用户组
func GetSubjectGroup(c *gin.Context) {
	var subject subjectRelationSerializer
	if err := c.ShouldBindQuery(&subject); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	svc := service.NewGroupService()
	groups, err := svc.ListSubjectGroups(subject.Type, subject.ID, subject.BeforeExpiredAt)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "svc.ListSubjectGroups",
			"type=`%s`, id=`%s`", subject.Type, subject.ID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", groups)
}

// UpdateSubjectMembersExpiredAt subject关系续期
func UpdateSubjectMembersExpiredAt(c *gin.Context) {
	var body subjectMemberExpiredAtSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "UpdateSubjectMembersExpiredAt")

	papSubjects := make([]pap.SubjectMember, 0, len(body.Members))
	for _, m := range body.Members {
		papSubjects = append(papSubjects, pap.SubjectMember{
			Type:            m.Type,
			ID:              m.ID,
			PolicyExpiredAt: m.PolicyExpiredAt,
		})
	}

	ctl := pap.NewGroupController()
	err := ctl.UpdateSubjectMembersExpiredAt(body.Type, body.ID, papSubjects)
	if err != nil {
		err = errorWrapf(err, "ctl.UpdateSubjectMembersExpiredAt",
			"type=`%s`, id=`%s`, subjects=`%+v`", body.Type, body.ID, papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// DeleteSubjectMembers 批量删除subject成员
func DeleteSubjectMembers(c *gin.Context) {
	var body deleteSubjectMemberSerializer
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
	typeCount, err := ctl.DeleteSubjectMembers(body.Type, body.ID, papSubjects)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "ctl.DeleteSubjectMembers",
			"type=`%s`, id=`%s`, subjects=`%+v`", body.Type, body.ID, papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", typeCount)
}

// BatchAddSubjectMembers 批量添加subject成员
func BatchAddSubjectMembers(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchAddSubjectMembers")

	var body addSubjectMembersSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	papSubjects := make([]pap.SubjectMember, 0, len(body.Members))
	for _, m := range body.Members {
		papSubjects = append(papSubjects, pap.SubjectMember{
			Type:            m.Type,
			ID:              m.ID,
			PolicyExpiredAt: body.PolicyExpiredAt,
		})
	}

	ctl := pap.NewGroupController()
	typeCount, err := ctl.CreateOrUpdateSubjectMembers(body.Type, body.ID, papSubjects)
	if err != nil {
		err = errorWrapf(err, "ctl.CreateOrUpdateSubjectMembers",
			"type=`%s`, id=`%s`, subjects=`%+v`", body.Type, body.ID, papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", typeCount)
}

// ListSubjectMemberBeforeExpiredAt 获取小于指定过期时间的成员列表
func ListSubjectMemberBeforeExpiredAt(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListSubjectmembersBeforeExpiredAt")

	var body listSubjectMemberBeforeExpiredAtSerializer
	if err := c.ShouldBindQuery(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	body.Default()

	svc := service.NewGroupService()
	count, err := svc.GetMemberCountBeforeExpiredAt(body.Type, body.ID, body.BeforeExpiredAt)
	if err != nil {
		err = errorWrapf(
			err, "svc.GetMemberCountBeforeExpiredAt type=`%s`, id=`%s`, beforeExpiredAt=`%d`",
			body.Type, body.ID, body.BeforeExpiredAt,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	relations, err := svc.ListPagingMemberBeforeExpiredAt(
		body.Type, body.ID, body.BeforeExpiredAt, body.Limit, body.Offset,
	)
	if err != nil {
		err = errorWrapf(
			err, "svc.ListPagingMemberBeforeExpiredAt type=`%s`, id=`%s`, beforeExpiredAt=`%d`",
			body.Type, body.ID, body.BeforeExpiredAt,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{
		"count":   count,
		"results": relations,
	})
}

// ListExistSubjectsBeforeExpiredAt 筛选出有成员过期的subjects
func ListExistSubjectsBeforeExpiredAt(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "FilterSubjectsBeforeExpiredAt")

	var body filterSubjectsBeforeExpiredAtSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	svcSubjects := make([]types.Subject, 0, len(body.Subjects))
	copier.Copy(&svcSubjects, &body.Subjects)

	svc := service.NewGroupService()
	existSubjects, err := svc.ListExistSubjectsBeforeExpiredAt(svcSubjects, body.BeforeExpiredAt)
	if err != nil {
		err = errorWrapf(
			err, "svc.ListExistSubjectsBeforeExpiredAt subjects=`%+v`, beforeExpiredAt=`%d`",
			svcSubjects, body.BeforeExpiredAt,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", existSubjects)
}
