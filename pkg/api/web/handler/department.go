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

	"iam/pkg/abac/pap"
	"iam/pkg/service"
	"iam/pkg/util"
)

// BatchCreateSubjectDepartments ...
func BatchCreateSubjectDepartments(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchCreateSubjectDepartments")

	var subjectDepartments []subjectDepartment
	if err := c.ShouldBindJSON(&subjectDepartments); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	papSubjectDepartments := make([]pap.SubjectDepartment, 0, len(subjectDepartments))
	for _, sd := range subjectDepartments {
		papSubjectDepartments = append(papSubjectDepartments, pap.SubjectDepartment{
			SubjectID:     sd.SubjectID,
			DepartmentIDs: sd.DepartmentIDs,
		})
	}

	ctl := pap.NewDepartmentController()
	err := ctl.BulkCreateSubjectDepartments(papSubjectDepartments)
	if err != nil {
		err = errorWrapf(err, "ctl.BulkCreateSubjectDepartments subjectDepartments=`%+v`", papSubjectDepartments)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	// TODO: 确认这里没有清subject detail?

	util.SuccessJSONResponse(c, "ok", nil)
}

// BatchDeleteSubjectDepartments ...
func BatchDeleteSubjectDepartments(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchDeleteSubjectDepartments")

	var subjectIDs []string
	if err := c.ShouldBindJSON(&subjectIDs); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if len(subjectIDs) == 0 {
		util.BadRequestErrorJSONResponse(c, "subject id can not be empty")
		return
	}

	ctl := pap.NewDepartmentController()
	err := ctl.BulkDeleteSubjectDepartments(subjectIDs)
	if err != nil {
		err = errorWrapf(err, "svc.BulkUpdateSubjectDepartments BulkDeleteSubjectDepartments=`%+v`", subjectIDs)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// BatchUpdateSubjectDepartments ...
func BatchUpdateSubjectDepartments(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchUpdateSubjectDepartments")

	var subjectDepartments []subjectDepartment
	if err := c.ShouldBindJSON(&subjectDepartments); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	papSubjectDepartments := make([]pap.SubjectDepartment, 0, len(subjectDepartments))
	for _, sd := range subjectDepartments {
		papSubjectDepartments = append(papSubjectDepartments, pap.SubjectDepartment{
			SubjectID:     sd.SubjectID,
			DepartmentIDs: sd.DepartmentIDs,
		})
	}

	ctl := pap.NewDepartmentController()
	err := ctl.BulkUpdateSubjectDepartments(papSubjectDepartments)
	if err != nil {
		err = errorWrapf(err, "svc.BulkUpdateSubjectDepartments subjectDepartments=`%+v`", papSubjectDepartments)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// ListSubjectDepartments ...
func ListSubjectDepartments(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListSubjectDepartments")

	var page pageSerializer
	if err := c.ShouldBindQuery(&page); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	page.Default()

	svc := service.NewDepartmentService()
	count, err := svc.GetSubjectDepartmentCount()
	if err != nil {
		util.SystemErrorJSONResponse(c, errorWrapf(err, "svc.GetSubjectDepartmentCount"))
		return
	}

	ctl := pap.NewDepartmentController()
	subjectDepartments, err := ctl.ListPagingSubjectDepartment(page.Limit, page.Offset)
	if err != nil {
		err = errorWrapf(err, "ctl.ListPagingSubjectDepartment limit=`%d`, offset=`%d`", page.Limit, page.Offset)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{
		"count":   count,
		"results": subjectDepartments,
	})
}
