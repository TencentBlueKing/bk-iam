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
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/prp"
	"iam/pkg/abac/types"
	"iam/pkg/errorx"
	"iam/pkg/util"
)

// CreateAndDeleteTemplatePolicies godoc
// @Summary Create and delete template policy/创建删除模板授权的策略
// @Description Create and delete template policy
// @ID api-web-create-delete-template-policies
// @Tags web
// @Accept json
// @Produce json
// @Param body body createAndDeleteTemplatePolicySerializer true "create and delete templates policies"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/perm-templates/policies [post]
func CreateAndDeleteTemplatePolicies(c *gin.Context) {
	var body createAndDeleteTemplatePolicySerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	systemID := body.SystemID
	subject := types.Subject{
		Type:      body.Subject.Type,
		ID:        body.Subject.ID,
		Attribute: types.NewSubjectAttribute(),
	}

	createPolicies := make([]types.Policy, 0, len(body.CreatePolicies))
	for _, p := range body.CreatePolicies {
		createPolicies = append(createPolicies,
			convertToInternalTypesPolicy(systemID, subject, 0, body.TemplateID, p))
	}

	manager := prp.NewPolicyManager()
	err := manager.CreateAndDeleteTemplatePolicies(systemID, body.Subject.Type, body.Subject.ID, body.TemplateID,
		createPolicies, body.DeletePolicyIDs)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "CreateAndDeleteTemplatePolicies",
			"systemID=`%s`, subjectType=`%s`, subjectID=`%s`, templateID=`%d`, "+
				"createPolicies=`%+v`, deletePolicies=`%+v`",
			systemID, body.Subject.Type, body.Subject.ID, body.TemplateID, createPolicies, body.DeletePolicyIDs)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// UpdateTemplatePolicies godoc
// @Summary update template policy/更新模板授权的策略
// @Description update template policy
// @ID api-web-update-template-policies
// @Tags web
// @Accept json
// @Produce json
// @Param body body updateTemplatePolicySerializer true "update templates policies"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/perm-templates/policies [put]
func UpdateTemplatePolicies(c *gin.Context) {
	var body updateTemplatePolicySerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	systemID := body.SystemID
	subject := types.Subject{
		Type:      body.Subject.Type,
		ID:        body.Subject.ID,
		Attribute: types.NewSubjectAttribute(),
	}

	updatePolicies := make([]types.Policy, 0, len(body.UpdatePolicies))
	for _, p := range body.UpdatePolicies {
		updatePolicies = append(updatePolicies,
			convertToInternalTypesPolicy(systemID, subject, p.ID, body.TemplateID, p.policy))
	}

	manager := prp.NewPolicyManager()
	err := manager.UpdateTemplatePolicies(systemID, body.Subject.Type, body.Subject.ID,
		updatePolicies)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "UpdateTemplatePolicies",
			"systemID=`%s`, subjectType=`%s`, subjectID=`%s`, templateID=`%d`, updatePolicies=`%+v`",
			systemID, body.Subject.Type, body.Subject.ID, body.TemplateID, updatePolicies)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// DeleteSubjectTemplatePolicies godoc
// @Summary delete template policy/删除模板授权的策略
// @Description delete template policy
// @ID api-web-delete-template-policies
// @Tags web
// @Accept json
// @Produce json
// @Param body body subjectTemplateSerializer true "delete templates policies"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/perm-templates/policies [delete]
func DeleteSubjectTemplatePolicies(c *gin.Context) {
	var body subjectTemplateSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	manager := prp.NewPolicyManager()
	err := manager.DeleteTemplatePolicies(body.SystemID, body.SubjectType, body.SubjectID, body.TemplateID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "DeleteTemplatePolicies",
			"systemID=`%s`, subjectType=`%s`, subjectID=`%s`, templateID=`%d`",
			body.SystemID, body.SubjectType, body.SubjectID, body.TemplateID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}
