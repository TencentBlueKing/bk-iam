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
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/translate"
	"iam/pkg/abac/prp"
	"iam/pkg/api/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/errorx"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// List godoc
// @Summary policy list
// @Description list policies of a action
// @ID api-open-system-policies-list
// @Tags open
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param params query listQuerySerializer true "the list request"
// @Success 200 {object} util.Response{data=policyListResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/policies [get]
func List(c *gin.Context) {
	// TODO: 翻页接口是否有性能问题 / 限制调用并发, 用drl
	var query listQuerySerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	ok, message := query.validate()
	if !ok {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}
	// init the default query
	query.initDefault()

	// action 必须为路径下system_id的一个合法注册的操作
	systemID := c.Param("system_id")

	// 2. action exists
	actionID := query.ActionID
	actionPK, err := cacheimpls.GetActionPK(systemID, actionID)
	if err != nil {
		// 在本系统内找不到这个action, 返回404
		if errors.Is(err, sql.ErrNoRows) {
			util.NotFoundJSONResponse(c, "action_id not exist in this system")
			return
		}

		err = fmt.Errorf("cacheimpls.GetActionPK system=`%s`, action=`%s` fail. err=%w", systemID, actionID, err)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// 3. do query: 查询某个系统, 某个action的所有policy列表  带分页
	policyService := service.NewPolicyService()
	count, err := policyService.GetCountByActionBeforeExpiredAt(actionPK, query.Timestamp)
	if err != nil {
		err = fmt.Errorf("getCountByAction actionPK=`%s`, timestamp=`%d` fail. err=%w",
			actionID, query.Timestamp, err)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	results := []thinPolicyResponse{}
	if count != 0 {
		offset := (query.Page - 1) * query.PageSize
		limit := query.PageSize
		policies, err := policyService.ListPagingQueryByActionBeforeExpiredAt(actionPK, query.Timestamp, offset, limit)
		if err != nil {
			err = fmt.Errorf(
				"listPoliciesByAction actionPK=`%s`, timestamp=`%d`, offset=`%d`, limit=`%d` fail. err=%w",
				actionID, query.Timestamp, offset, limit, err)
			util.SystemErrorJSONResponse(c, err)
			return
		}
		results, err = convertQueryPoliciesToThinPolicies(systemID, actionID, policies)
		if err != nil {
			err = fmt.Errorf(
				"convertQueryPoliciesToThinPolicies system=`%s`, action=`%s` fail. err=%w",
				systemID, actionID, err)
			util.SystemErrorJSONResponse(c, err)
			return
		}
	}

	// 返回每条策略, 包含的过期时间, 接入方得二次校验
	util.SuccessJSONResponse(c, "ok", policyListResponse{
		Metadata: policyListResponseMetadata{
			System:    systemID,
			Action:    policyResponseAction{ID: actionID},
			Timestamp: query.Timestamp,
		},
		Count:   count,
		Results: results,
	})
}

func convertQueryPoliciesToThinPolicies(
	systemID, actionID string,
	policies []types.QueryPolicy,
) (thinPolicies []thinPolicyResponse, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "policy_list.convertQueryPoliciesToThinPolicies")
	if len(policies) == 0 {
		return
	}

	// 0. get action resource type set
	resourceTypeSet, err := common.GetActionResourceTypeSet(systemID, actionID)
	if err != nil {
		err = errorWrapf(err, "getActionResourceTypeSet systemID=`%s`, actionID=`%s` fail", systemID, actionID)
		return
	}

	// 1. collect all expression pks
	expressionPKs := make([]int64, 0, len(policies))
	for _, p := range policies {
		if p.ExpressionPK != -1 {
			expressionPKs = append(expressionPKs, p.ExpressionPK)
		}
	}

	// 2. query expression from cache
	pkExpressionMap, err := translateExpressions(resourceTypeSet, expressionPKs)
	if err != nil {
		err = errorWrapf(err, "translateExpressions resourceTypeSet=`%+v`, expressionPKs=`%+v` fail",
			resourceTypeSet, expressionPKs)
		return
	}

	// loop policies to build thinPolicies
	for _, p := range policies {
		subj, err1 := cacheimpls.GetSubjectByPK(p.SubjectPK)
		// if get subject fail, continue
		if err1 != nil {
			log.Info(errorWrapf(err1,
				"policy_list.convertQueryPoliciesToThinPolicies get subject subject_pk=`%d` fail",
				p.SubjectPK))

			continue
		}

		// if missing the expression, continue
		expression, ok := pkExpressionMap[p.ExpressionPK]
		if !ok {
			log.Errorf("policy_list.convertQueryPoliciesToThinPolicies p.ExpressionPK=`%d` missing in pkExpressionMap",
				p.ExpressionPK)
			continue
		}

		thinPolicies = append(thinPolicies, thinPolicyResponse{
			Version: service.PolicyVersion,
			ID:      p.PK,
			Subject: policyResponseSubject{
				Type: subj.Type,
				ID:   subj.ID,
				Name: subj.Name,
			},
			Expression: expression,
			ExpiredAt:  p.ExpiredAt,
		})
	}
	return thinPolicies, nil
}

// translateExpressions translate expression to json formart
func translateExpressions(
	resourceTypeSet *util.StringSet,
	expressionPKs []int64,
) (expressionMap map[int64]map[string]interface{}, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "policy_list.translateExpressions")

	// when the pk is -1, will translate to any
	pkExpressionStrMap := map[int64]string{
		-1: "",
	}
	if len(expressionPKs) > 0 {
		manager := prp.NewPolicyManager()

		var exprs []types.AuthExpression
		exprs, err = manager.GetExpressionsFromCache(-1, expressionPKs)
		if err != nil {
			err = errorWrapf(err, "policyManager.GetExpressionsFromCache pks=`%+v` fail", expressionPKs)
			return
		}

		for _, e := range exprs {
			pkExpressionStrMap[e.PK] = e.Expression
		}
	}

	// translate one by one
	expressionMap = make(map[int64]map[string]interface{}, len(pkExpressionStrMap))
	for pk, expr := range pkExpressionStrMap {
		// TODO: 如何优化这里的性能?
		// TODO: 理论上, signature一样的只需要转一次
		// e.Signature
		translatedExpr, err1 := translate.PolicyExpressionTranslate(expr)
		if err1 != nil {
			err = errorWrapf(err1, "translate fail", "")
			return
		}
		expressionMap[pk] = translatedExpr
	}
	return expressionMap, nil
}
