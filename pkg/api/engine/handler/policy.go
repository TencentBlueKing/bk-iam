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
	"fmt"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/translate"
	"iam/pkg/abac/prp"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/util"
)

// ListPolicy godoc
// @Summary policy list
// @Description list policies of ids or between min-max ids
// @ID api-engine-policies-list
// @Tags engine
// @Accept json
// @Produce json
// @Param params query listPolicySerializer true "the list request"
// @Success 200 {object} util.Response{data=policyListResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/engine/policies [get]
func ListPolicy(c *gin.Context) {
	var query listPolicySerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	ok, message := query.validate()
	if !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}
	query.initDefault()

	var (
		policies []prp.EnginePolicy
		err      error
	)
	manager := prp.NewEnginePolicyManager()
	// 有pks, 优先查询pks的数据
	if query.hasIDs() {
		pks, _ := query.getIDs()
		policies, err = manager.ListByPKs("abac", pks)
		if err != nil {
			err = fmt.Errorf("svc.ListByPKs pks=`%s` fail. err=%w", query.IDs, err)
			util.SystemErrorJSONResponse(c, err)
			return
		}
	} else {
		policies, err = manager.ListBetweenPK("abac", query.Timestamp, query.MinID, query.MaxID)
		if err != nil {
			err = fmt.Errorf("svc.ListBetweenPK expiredAt=`%d`, minPK=`%d`, maxPK=`%d` fail. err=%w",
				query.Timestamp, query.MinID, query.MaxID, err)
			util.SystemErrorJSONResponse(c, err)
			return
		}
	}

	results, err := convertEnginePoliciesToResponse(policies)
	if err != nil {
		err = fmt.Errorf("convertEngineQueryPoliciesToEnginePolicies policies length=`%d` fail. err=%w",
			len(policies), err)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// 返回每条策略, 包含的过期时间, 接入方得二次校验
	util.SuccessJSONResponse(c, "ok", policyListResponse{
		Metadata: query,
		Results:  results,
	})
}

// ListPolicyPKs godoc
// @Summary policy ids list
// @Description list policy pks by condition
// @ID api-engine-policy-id-list
// @Tags engine
// @Accept json
// @Produce json
// @Param params query listPolicyIDsSerializer true "the list request"
// @Success 200 {object} util.Response{data=listPolicyIDsResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/engine/policies/ids [get]
func ListPolicyPKs(c *gin.Context) {
	var query listPolicyIDsSerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	ok, message := query.validate()
	if !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	manager := prp.NewEnginePolicyManager()
	pks, err := manager.ListPKBetweenUpdatedAt("abac", query.BeginUpdatedAt, query.EndUpdatedAt)
	if err != nil {
		err = fmt.Errorf("svc.ListPKBetweenUpdatedAt beginUpdatedAt=`%d`, endUpdatedAt=`%d` fail. err=%w",
			query.BeginUpdatedAt, query.EndUpdatedAt, err)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	if len(pks) == 0 {
		pks = []int64{}
	}

	util.SuccessJSONResponse(c, "ok", listPolicyIDsResponse{IDs: pks})
}

// GetMaxPolicyPK godoc
// @Summary policy max id
// @Description get max policy id by condition
// @ID api-engine-policy-id-max
// @Tags engine
// @Accept json
// @Produce json
// @Param params query getMaxPolicyIDSerializer true "the request"
// @Success 200 {object} util.Response{data=getMaxPolicyIDResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/engine/policies/ids/max [get]
func GetMaxPolicyPK(c *gin.Context) {
	var query getMaxPolicyIDSerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	manager := prp.NewEnginePolicyManager()
	pk, err := manager.GetMaxPKBeforeUpdatedAt("abac", query.UpdatedAt)
	if err != nil {
		err = fmt.Errorf("svc.GetMaxPKBeforeUpdatedAt updatedAt=`%d` fail. err=%w",
			query.UpdatedAt, err)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", getMaxPolicyIDResponse{pk})
}

// ===========================================================

func convertEnginePoliciesToResponse(
	enginePolicies []prp.EnginePolicy,
) (responses []enginePolicyResponse, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "policy.convertEnginePoliciesToResponse")

	results := make([]enginePolicyResponse, len(enginePolicies))

	for _, p := range enginePolicies {
		// 可能存在subject被删, policy还有的情况, 这时需要忽略该错误
		subj, err := cacheimpls.GetSubjectByPK(p.SubjectPK)
		if err != nil {
			err = errorWrapf(err, "cacheimpls.GetSubjectByPK get subject subject_pk=`%d` fail", p.SubjectPK)
			log.Info(err)
			continue
		}

		var systemID string
		actions := make([]policyResponseAction, 0, len(p.ActionPKs))
		for _, actionPK := range p.ActionPKs {
			act, err1 := cacheimpls.GetAction(actionPK)
			if err1 != nil {
				err = errorWrapf(err1, "cacheimpls.GetAction actionPK=`%d` fail", actionPK)
				return responses, err
			}
			systemID = act.System

			actions = append(actions, policyResponseAction{
				ID: act.ID,
			})
		}

		var action policyResponseAction
		if len(actions) == 1 {
			action = actions[0]
			actions = []policyResponseAction{}
		}

		translatedExpr, err2 := translate.PolicyExpressionTranslate(p.Expression)
		if err2 != nil {
			err = errorWrapf(err2, "translate.PolicyExpressionTranslate policy=`%+v`, expr=`%s` fail", p, p.Expression)
			return responses, err
		}

		policy := enginePolicyResponse{
			Version: service.PolicyVersion,
			ID:      p.ID,
			System:  systemID,
			Action:  action,
			Actions: actions,
			Subject: policyResponseSubject{
				Type: subj.Type,
				ID:   subj.ID,
				Name: subj.Name,
			},
			Expression: translatedExpr,
			TemplateID: p.TemplateID,
			ExpiredAt:  p.ExpiredAt,
			UpdatedAt:  p.UpdatedAt,
		}
		results = append(results, policy)
	}
	return results, nil
}
