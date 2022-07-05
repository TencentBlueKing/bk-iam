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
	"iam/pkg/abac/pap"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/types"
	"iam/pkg/service"
	"iam/pkg/util"
)

// ListSystemPolicy godoc
// @Summary List policy/获取策略列表
// @Description query all authorized policies: subject/template[required]
// @ID api-web-list-system-policy
// @Tags web
// @Accept json
// @Produce json
// @Param system_id path string true "system id"
// @Param subject_type query string true "subject type"
// @Param subject_id query string true "subject id"
// @Param template_id query string true "template id"
// @Success 200 {object} util.Response{data=types.SaaSPolicy}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/systems/{system_id}/policies [get]
func ListSystemPolicy(c *gin.Context) {
	var query policySerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	systemID := c.Param("system_id")

	// 查询有相关权限的policy列表
	ctl := pap.NewPolicyController()
	policies, err := ctl.ListSaaSBySubjectSystemTemplate(
		systemID, query.SubjectType, query.SubjectID, query.TemplateID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "ListSystemPolicy",
			"systemID=`%s`, subjectType=`%s`, subjectID=`%s`", systemID, query.SubjectType, query.SubjectID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", policies)
}

func convertToInternalTypesPolicy(systemID string, subject types.Subject, id, templateID int64,
	policy policy,
) types.Policy {
	return types.Policy{
		Version: service.PolicyVersion,
		ID:      id,
		System:  systemID,
		Subject: subject,
		Action: types.Action{
			ID:        policy.ActionID,
			Attribute: types.NewActionAttribute(),
		},
		Expression: policy.ResourceExpression,
		ExpiredAt:  policy.ExpiredAt,
		TemplateID: templateID,
	}
}

// AlterPolicies godoc
// @Summary Alter policies/变更用户自定义申请策略
// @Description alter policies by custom application
// @ID api-web-alter-policies
// @Tags web
// @Accept json
// @Produce json
// @Param system_id path string true "system id"
// @Param body body policiesAlterSerializer true "create and update policies"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/systems/{system_id}/policies [post]
func AlterPolicies(c *gin.Context) {
	var body policiesAlterSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	// ? 后端api没有检查policy是否有重复, action是否存在, 暂时交由SaaS侧做检查

	systemID := c.Param("system_id")

	subject := types.Subject{
		Type:      body.Subject.Type,
		ID:        body.Subject.ID,
		Attribute: types.NewSubjectAttribute(),
	}

	createPolicies := make([]types.Policy, 0, len(body.CreatePolicies))
	for _, p := range body.CreatePolicies {
		createPolicies = append(createPolicies,
			convertToInternalTypesPolicy(systemID, subject, 0, service.PolicyTemplateIDCustom, p))
	}

	updatePolicies := make([]types.Policy, 0, len(body.UpdatePolicies))
	for _, p := range body.UpdatePolicies {
		updatePolicies = append(updatePolicies,
			convertToInternalTypesPolicy(systemID, subject, p.ID, service.PolicyTemplateIDCustom, p.policy))
	}

	ctl := pap.NewPolicyController()
	err := ctl.AlterCustomPolicies(systemID, body.Subject.Type, body.Subject.ID,
		createPolicies, updatePolicies, body.DeletePolicyIDs)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "AlterPolicies",
			"systemID=`%s`, subjectType=`%s`, subjectID=`%s`, createPolicies=`%+v`, updatePolicies=`%+v`",
			systemID, body.Subject.Type, body.Subject.ID, createPolicies, updatePolicies)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// BatchDeletePolicies godoc
// @Summary Batch delete policies/删除用户策略
// @Description batch delete policies
// @ID api-web-batch-delete-policies
// @Tags web
// @Accept json
// @Produce json
// @Param body body policiesDeleteSerializer true "delete policy info"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/policies [delete]
func BatchDeletePolicies(c *gin.Context) {
	var body policiesDeleteSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	ctl := pap.NewPolicyController()
	err := ctl.DeleteByIDs(body.SystemID, body.SubjectType, body.SubjectID, body.IDs)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "BatchDeletePolicies",
			"subjectType=`%s`, subjectID=`%s`, IDs=`%+v`", body.SubjectType, body.SubjectID, body.IDs)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// GetCustomPolicy godoc
// @Summary GetCustomPolicy/获取自定义策略
// @Description get custom policy
// @ID api-web-get-custom-policy
// @Tags web
// @Accept json
// @Produce json
// @Param system_id path string true "system id"
// @Param subject_type query string true "subject type"
// @Param subject_id query string true "subject id"
// @Param action_id query string true "action id"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/custom-policy [get]
func GetCustomPolicy(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "GetCustomPolicy")

	var query queryPolicySerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	systemID := c.Param("system_id")
	ctl := pap.NewPolicyController()
	policy, err := ctl.GetByActionTemplate(
		systemID, query.SubjectType, query.SubjectID, query.ActionID, service.PolicyTemplateIDCustom,
	)
	if err != nil {
		err = errorWrapf(err, "system=`%s`, subjectType=`%s`, subjectID=`%s`, actionID=`%+v`",
			systemID, query.SubjectType, query.SubjectID, query.ActionID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{"policy_id": policy.ID})
}

// ListPolicy godoc
// @Summary List policy/获取策略列表
// @Description query all authorized policies: subject/template[required]
// @ID api-web-list-policy
// @Tags web
// @Accept json
// @Produce json
// @Param system_id path string true "system id"
// @Param subject_type query string true "subject type"
// @Param subject_id query string true "subject id"
// @Param before_expired_at query int true "until expired at"
// @Success 200 {object} util.Response{data=types.SaaSPolicy}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/policies [get]
func ListPolicy(c *gin.Context) {
	var query queryListPolicySerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// 查询过期时间筛选的policy列表
	ctl := pap.NewPolicyController()
	policies, err := ctl.ListSaaSBySubjectTemplateBeforeExpiredAt(
		query.SubjectType, query.SubjectID, service.PolicyTemplateIDCustom, query.BeforeExpiredAt)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "ListPolicy",
			"subjectType=`%s`, subjectID=`%s`, expiredAt=`%d`",
			query.SubjectType, query.SubjectID, query.BeforeExpiredAt)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", policies)
}

// UpdatePoliciesExpiredAt godoc
// @Summary Renew policies/权限续期
// @Description renew policies
// @ID api-web-renew-policies
// @Tags web
// @Accept json
// @Produce json
// @Param body body policiesUpdateExpiredAtSerializer true "renew policies"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/policies/expired_at [put]
func UpdatePoliciesExpiredAt(c *gin.Context) {
	var body policiesUpdateExpiredAtSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	pkExpiredAts := make([]types.PolicyPKExpiredAt, 0, len(body.Policies))
	for _, p := range body.Policies {
		pkExpiredAts = append(pkExpiredAts, types.PolicyPKExpiredAt{
			PK:        p.ID,
			ExpiredAt: p.ExpiredAt,
		})
	}

	ctl := pap.NewPolicyController()
	err := ctl.UpdateSubjectPoliciesExpiredAt(
		body.SubjectType, body.SubjectID, pkExpiredAts)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "UpdateSubjectPoliciesExpiredAt",
			"subjectType=`%s`, subjectID=`%s`, ids=`%+v`",
			body.SubjectType, body.SubjectID, pkExpiredAts)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// DeleteActionPolicies will delete all policies by action_id
func DeleteActionPolicies(c *gin.Context) {
	systemID := c.Param("system_id")
	actionID := c.Param("action_id")

	ctl := pap.NewPolicyController()
	err := ctl.DeleteByActionID(systemID, actionID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "DeleteActionPolicies",
			"systemID=`%s`, actionID=`%s`",
			systemID, actionID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// DeleteUnreferencedExpressions clean not quoted expression
func DeleteUnreferencedExpressions(c *gin.Context) {
	manager := service.NewPolicyService()

	err := manager.DeleteUnreferencedExpressions()
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "DeleteUnreferencedExpressions", "")
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}
