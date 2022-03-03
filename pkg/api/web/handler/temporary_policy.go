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
	"strconv"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/prp"
	"iam/pkg/abac/types"
	"iam/pkg/service"
	"iam/pkg/util"
)

// CreateTemporaryPolicies godoc
// @Summary Create Temporary policies/创建临时权限策略
// @Description create temporary policies by application
// @ID api-web-create-temporary-policies
// @Tags web
// @Accept json
// @Produce json
// @Param system_id path string true "system id"
// @Param body body temporaryPoliciesSerializer true "create temporary policies"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/systems/{system_id}/temporary-policies [post]
func CreateTemporaryPolicies(c *gin.Context) {
	var body temporaryPoliciesSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if ok, message := body.validate(); !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	systemID := c.Param("system_id")

	subject := types.Subject{
		Type:      body.Subject.Type,
		ID:        body.Subject.ID,
		Attribute: types.NewSubjectAttribute(),
	}

	policies := make([]types.Policy, 0, len(body.Policies))
	for _, p := range body.Policies {
		policies = append(policies,
			convertToInternalTypesPolicy(systemID, subject, 0, service.PolicyTemplateIDCustom, p))
	}

	manager := prp.NewPolicyManager()
	pks, err := manager.CreateTemporaryPolicies(
		systemID, body.Subject.Type, body.Subject.ID, policies)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "CreateTemporaryPolicies",
			"systemID=`%s`, subjectType=`%s`, subjectID=`%s`, policies=`%+v`",
			systemID, body.Subject.Type, body.Subject.ID, policies)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{"ids": pks})
}

// BatchDeleteTemporaryPolicies godoc
// @Summary Batch delete temporary policies/删除临时权限策略
// @Description batch delete temporary policies
// @ID api-web-batch-delete-temporary-policies
// @Tags web
// @Accept json
// @Produce json
// @Param body body temporaryPoliciesDeleteSerializer true "delete temporary policy info"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/temporary-policies [delete]
func BatchDeleteTemporaryPolicies(c *gin.Context) {
	var body temporaryPoliciesDeleteSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	manager := prp.NewPolicyManager()
	err := manager.DeleteTemporaryByIDs(body.SystemID, body.SubjectType, body.SubjectID, body.IDs)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "DeleteTemporaryByIDs",
			"subjectType=`%s`, subjectID=`%s`, IDs=`%+v`", body.SubjectType, body.SubjectID, body.IDs)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}

// DeleteTemporaryBeforeExpiredAt godoc
// @Summary Batch delete temporary policies before expired_at/删除指定过期时间之前临时权限策略
// @Description batch delete temporary policies before expired_at
// @ID api-web-batch-delete-temporary-policies-before-expired-at
// @Tags web
// @Accept json
// @Produce json
// @Param expired_at query int true "expired_at"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/web/temporary-policies/before_expired_at [delete]
// DeleteTemporaryBeforeExpiredAt will delete all policies by action_id
func DeleteTemporaryBeforeExpiredAt(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "DeleteTemporaryBeforeExpiredAt")

	expiredAtStr := c.Query("expired_at")
	expiredAt, err := strconv.ParseInt(expiredAtStr, 10, 64)
	if err != nil {
		err = errorWrapf(err, "strconv.ParseInt fail, expiredAt=`%s`", expiredAtStr)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	manager := prp.NewPolicyManager()
	err = manager.DeleteTemporaryBeforeExpiredAt(expiredAt)
	if err != nil {
		err = errorWrapf(err, "expiredAt=`%d`", expiredAt)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}
