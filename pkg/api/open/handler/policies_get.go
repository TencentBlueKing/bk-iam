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
	"fmt"

	"github.com/gin-gonic/gin"

	"iam/pkg/abac/prp"
	"iam/pkg/cacheimpls"
	"iam/pkg/util"
)

// PolicyGet godoc
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
func PolicyGet(c *gin.Context) {
	var query policyGetSerializer
	if err := c.ShouldBindUri(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	query.initDefault()

	manager := prp.NewOpenPolicyManager()
	policy, err := manager.Get(query.Type, query.PolicyID)
	if err != nil {
		if errors.Is(err, prp.ErrPolicyNotFound) {
			util.NotFoundJSONResponse(c, err.Error())
			return
		}

		util.SystemErrorJSONResponse(c, err)
		return
	}

	data, err := convertOpenPolicyToPolicyGetResponse(policy)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	systemID := c.Param("system_id")
	if systemID != data.System {
		util.ForbiddenJSONResponse(c, fmt.Sprintf("system(%s) can't access system(%s)'s policy", systemID, data.System))
		return
	}

	util.SuccessJSONResponse(c, "ok", data)
}

func convertOpenPolicyToPolicyGetResponse(policy prp.OpenPolicy) (policyGetResponse, error) {
	subj, err := cacheimpls.GetSubjectByPK(policy.SubjectPK)
	if err != nil {
		return policyGetResponse{}, err
	}

	action, err := cacheimpls.GetAction(policy.ActionPK)
	if err != nil {
		return policyGetResponse{}, err
	}

	resp := policyGetResponse{
		Version: policy.Version,
		ID:      policy.ID,
		System:  action.System,
		Subject: policyResponseSubject{
			Type: subj.Type,
			ID:   subj.ID,
			Name: subj.Name,
		},
		Action: policyResponseAction{
			ID: action.ID,
		},
		Expression: policy.Expression,
		ExpiredAt:  policy.ExpiredAt,
	}
	return resp, nil
}
