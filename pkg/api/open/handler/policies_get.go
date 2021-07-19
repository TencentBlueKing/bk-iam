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
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"iam/pkg/api/common"
	"iam/pkg/cache/impls"
	"iam/pkg/service"
	"iam/pkg/util"
)

// GetPolicy godoc
// @Summary policy get
// @Description get a policy
// @ID api-open-system-policies-get
// @Tags open
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param policy_id path string true "Policy ID"
// @Success 200 {object} util.Response{data=policyGetResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/policies/{policy_id} [get]
func Get(c *gin.Context) {
	var pathParams policyGetSerializer
	if err := c.ShouldBindUri(&pathParams); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// 1. query policy
	policyService := service.NewPolicyService()
	queryPolicy, err := policyService.Get(pathParams.PolicyID)
	if err != nil {
		// 不存在的情况, 404
		if errors.Is(err, sql.ErrNoRows) {
			util.NotFoundJSONResponse(c, "policy not exist")
			return
		}

		util.SystemErrorJSONResponse(c, err)
		return
	}

	// 2. query systemAction from cache
	systemAction, err := impls.GetAction(queryPolicy.ActionPK)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	systemID := c.Param("system_id")
	if systemID != systemAction.System {
		util.ForbiddenJSONResponse(c, fmt.Sprintf("system(%s) can't access system(%s)'s policy",
			systemID, systemAction.System))
		return
	}

	// 4. query subj from cache
	subj, err := impls.GetSubjectByPK(queryPolicy.SubjectPK)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// 5. get expression
	resourceTypeSet, err := common.GetActionResourceTypeSet(systemAction.System, systemAction.ID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}
	pkExpressionMap, err := translateExpressions(resourceTypeSet, []int64{queryPolicy.ExpressionPK})
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}
	expression, ok := pkExpressionMap[queryPolicy.ExpressionPK]
	if !ok {
		util.SystemErrorJSONResponse(c, fmt.Errorf("expression pk=`%d` missing", queryPolicy.ExpressionPK))
		return
	}

	policy := policyGetResponse{
		Version: service.PolicyVersion,
		ID:      queryPolicy.PK,
		System:  systemAction.System,
		Subject: policyResponseSubject{
			Type: subj.Type,
			ID:   subj.ID,
			Name: subj.Name,
		},
		Action: policyResponseAction{
			ID: systemAction.ID,
		},
		Expression: expression,
		ExpiredAt:  queryPolicy.ExpiredAt,
	}

	util.SuccessJSONResponse(c, "ok", policy)
}
