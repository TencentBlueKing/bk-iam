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
	"iam/pkg/util"
)

// BatchAddRoleSubject ...
func BatchAddRoleSubject(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchCreateRoleSubject")

	var body subjectRoleSerializer
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

	// TODO: 校验 systemID 存在

	ctl := pap.NewRoleController()
	err := ctl.BulkAddSubjects(body.RoleType, body.SystemID, papSubjects)
	if err != nil {
		err = errorWrapf(
			err, "ctl.BulkCreateSubjectRoles roleType=`%s`, system=`%s`, subjects=`%+v`",
			body.RoleType, body.SystemID, papSubjects,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// BatchDeleteRoleSubject ...
func BatchDeleteRoleSubject(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchDeleteRoleSubject")

	var body subjectRoleSerializer
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

	ctl := pap.NewRoleController()
	err := ctl.BulkDeleteSubjects(body.RoleType, body.SystemID, papSubjects)
	if err != nil {
		err = errorWrapf(
			err, "ctl.BulkDeleteSubjectRoles roleType=`%s`, system=`%s`, subjects=`%+v`",
			body.RoleType, body.SystemID, papSubjects,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// ListRoleSubject ...
func ListRoleSubject(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BulkDeleteSubjectRole")

	var query subjectRoleQuerySerializer

	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if valid, message := query.validate(); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	ctl := pap.NewRoleController()
	subjects, err := ctl.ListSubjectByRole(query.RoleType, query.SystemID)
	if err != nil {
		err = errorWrapf(
			err, "ctl.ListSubjectByRole roleType=`%s`, system=`%s`",
			query.RoleType, query.SystemID,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", subjects)
}
