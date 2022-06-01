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
	"iam/pkg/util"
)

// ListSubject 查询用户/部门/用户组列表
func ListSubject(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListSubject")

	var body listSubjectSerializer
	if err := c.ShouldBindQuery(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	body.Default()

	svc := service.NewSubjectService()
	count, err := svc.GetCount(body.Type)
	if err != nil {
		err = errorWrapf(err, "svc.GetCount type=`%s`", body.Type)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	subjects, err := svc.ListPaging(body.Type, body.Limit, body.Offset)
	if err != nil {
		err = errorWrapf(
			err, "svc.ListPaging type=`%s` limit=`%d` offset=`%d`",
			body.Type, body.Limit, body.Offset,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{
		"count":   count,
		"results": subjects,
	})
}

// BatchCreateSubjects 批量创建subject
func BatchCreateSubjects(c *gin.Context) {
	var subjects []createSubjectSerializer
	if err := c.ShouldBindJSON(&subjects); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if valid, message := common.ValidateArray(subjects); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	ctl := pap.NewSubjectController()
	papSubjects := make([]pap.Subject, 0, len(subjects))
	copier.Copy(&papSubjects, &subjects)

	err := ctl.BulkCreate(papSubjects)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "ctl.BulkCreate", "subjects=`%+v`", papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// BatchDeleteSubjects 批量删除subject
func BatchDeleteSubjects(c *gin.Context) {
	var subjects []deleteSubjectSerializer
	if err := c.ShouldBindJSON(&subjects); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if valid, message := common.ValidateArray(subjects); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}
	ctl := pap.NewSubjectController()
	papSubjects := make([]pap.Subject, 0, len(subjects))
	copier.Copy(&papSubjects, &subjects)

	err := ctl.BulkDelete(papSubjects)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "ctl.BulkDelete", "subjects=`%+v`", papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// BatchUpdateSubject ...
func BatchUpdateSubject(c *gin.Context) {
	var subjects []updateSubjectSerializer
	if err := c.ShouldBindJSON(&subjects); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if valid, message := common.ValidateArray(subjects); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	ctl := pap.NewSubjectController()
	papSubjects := make([]pap.Subject, 0, len(subjects))
	copier.Copy(&papSubjects, &subjects)

	err := ctl.BulkUpdateName(papSubjects)

	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchUpdateSubject")
	if err != nil {
		err = errorWrapf(err, "ctl.BulkUpdateName subjects=`%+v`", papSubjects)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}
