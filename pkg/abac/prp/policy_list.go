/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/prp/expression"
	"iam/pkg/abac/prp/policy"
	"iam/pkg/abac/types"
	"iam/pkg/errorx"
	"iam/pkg/logging/debug"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"
)

const (
	tooLargeThreshold   = 300
	queryTypePolicy     = "ListPolicy"
	queryTypeExpression = "ListExpression"
)

var (
	emptyAuthExpression = svctypes.AuthExpression{}
)

func convertToAuthPolicy(svcPolicy svctypes.AuthPolicy, svcExpression svctypes.AuthExpression) types.AuthPolicy {
	return types.AuthPolicy{
		Version:             service.PolicyVersion,
		ID:                  svcPolicy.PK,
		Expression:          svcExpression.Expression,
		ExpressionSignature: svcExpression.Signature,
		ExpiredAt:           svcPolicy.ExpiredAt,
	}
}

func reportTooLargeQueryArguments(queryType string, count int, system, actionID, subjectType, subjectID string) {
	if count < tooLargeThreshold {
		return
	}

	log.Errorf(
		"%s too large query arguments: system=`%s`, action=`%s`, subject_type=`%s`, subject_id=`%s`, count=%d",
		queryType, system, actionID, subjectType, subjectID, count)

	// report to sentry
	util.ReportToSentry(
		fmt.Sprintf("%s: too large query arguments", queryType),
		map[string]interface{}{
			"system":       system,
			"action":       actionID,
			"subject_type": subjectType,
			"subject_id":   subjectID,
			"count":        count,
		},
	)
}

func reportTooLargeReturnedPolicies(count int, system, actionID, subjectType, subjectID string) {
	if count < tooLargeThreshold {
		return
	}

	log.Errorf(
		"too large return policies: system=`%s`, action=`%s`, subject_type=`%s`, subject_id=`%s`, count=%d",
		system, actionID, subjectType, subjectID, count)

	// report to sentry
	util.ReportToSentry(
		"too large return policies",
		map[string]interface{}{
			"system":       system,
			"action":       actionID,
			"subject_type": subjectType,
			"subject_id":   subjectID,
			"count":        count,
		},
	)
}

// ListBySubjectAction ...
func (m *policyManager) ListBySubjectAction(
	system string,
	subject types.Subject,
	action types.Action,
	withoutCache bool,
	parentEntry *debug.Entry,
) (policies []types.AuthPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "ListPolicyBySubjectAction")

	entry := debug.NewSubDebug(parentEntry)
	if entry != nil {
		debug.WithValue(entry, "cacheEnabled", !withoutCache)
	}

	// 1. get effect subject pks
	debug.AddStep(entry, "Get Effect Subject PKs")
	// 通过subject对象获取PK
	effectSubjectPKs, err := getEffectSubjectPKs(subject)
	if err != nil {
		err = errorWrapf(err, "getEffectSubjectPKs subject=`%+v` fail", subject)
		return
	}
	debug.WithValue(entry, "subjectPKs", effectSubjectPKs)

	// 2. get action pk
	debug.AddStep(entry, "Get Action PK")
	actionPK, err := action.Attribute.GetPK()
	if err != nil {
		err = errorWrapf(err, "action.Attribute.GetPK action=`%+v` fail", action)
		return
	}
	debug.WithValue(entry, "actionPK", actionPK)

	// 3. get effect policies
	debug.AddStep(entry, "List Policy by Subject Action")
	reportTooLargeQueryArguments(queryTypePolicy, len(effectSubjectPKs), system, action.ID, subject.Type, subject.ID)
	var effectPolicies []svctypes.AuthPolicy
	if withoutCache {
		// 查询鉴权策略
		effectPolicies, err = m.policyService.ListAuthBySubjectAction(effectSubjectPKs, actionPK)
		if err != nil {
			err = errorWrapf(err,
				"policyService.ListBySubjectAction system=`%s`, effectSubjectPKs=`%+v`, actionPK=`%d` fail",
				system, effectSubjectPKs, actionPK)
			return
		}
	} else {
		//effectPolicies, err = m.getPoliciesFromCache(system, actionPK, effectSubjectPKs, entry)
		effectPolicies, err = policy.GetPoliciesFromCache(system, actionPK, effectSubjectPKs)
		if err != nil {
			err = errorWrapf(err,
				"getPoliciesFromCache system=`%s`, actionPK=`%d`, subjectPKs=`%+v` fail",
				system, actionPK, effectSubjectPKs)
			debug.WithError(entry, err)
			return
		}
	}
	debug.WithValue(entry, "effectPolicies", effectPolicies)

	// if no effect policies, return
	if len(effectPolicies) == 0 {
		debug.WithValue(entry, "got_no_policies", true)
		return
	}

	// if action has not resource types, will not query expression!!!!!!
	if action.WithoutResourceType() {
		debug.WithValue(entry, "without_resource_types", true)
		// only return the first policy with empty expression, will auth=True or policy=Any
		policy := effectPolicies[0]

		// NOTE: the expression will be ""
		// TODO: ? should be "" or "[]"?
		policies = append(policies, convertToAuthPolicy(policy, emptyAuthExpression))
		return
	}

	// 4. expressionPK 去重
	expressionPKs := make([]int64, 0, len(effectPolicies))
	expressionPKSet := util.NewFixedLengthInt64Set(len(effectPolicies))
	for _, p := range effectPolicies {
		if !expressionPKSet.Has(p.ExpressionPK) {
			expressionPKSet.Add(p.ExpressionPK)
			expressionPKs = append(expressionPKs, p.ExpressionPK)
		}
	}

	debug.WithValue(entry, "expressionPKs", expressionPKs)

	// 5. query expressions
	debug.AddStep(entry, "List Expression by PKs")
	reportTooLargeQueryArguments(queryTypeExpression, len(expressionPKs), system, action.ID, subject.Type, subject.ID)
	var expressions []svctypes.AuthExpression
	if withoutCache {
		expressions, err = m.policyService.ListExpressionByPKs(expressionPKs)
		if err != nil {
			err = errorWrapf(err, "policyService.ListExpressionByPKs pks=`%+v` fail", expressionPKs)
			return
		}
	} else {
		expressions, err = expression.GetExpressionsFromCache(actionPK, expressionPKs)
		if err != nil {
			err = errorWrapf(err, "GetExpressionsFromCache expressionPKs=`%+v` fail", expressionPKs)
			return
		}
	}
	debug.WithValue(entry, "expressions", expressions)

	// 6. sort and uniq the policies
	debug.AddStep(entry, "Sort and Uniq expressions")
	// 表达式去重, 一个表达式只留一个
	expressionMap := make(map[int64]svctypes.AuthExpression, len(expressions))
	for _, e := range expressions {
		expressionMap[e.PK] = e
	}

	// NOTE: any 排在前面的逻辑去掉, 应该在计算或转换的时候处理合并 remove policy with `Any` first
	signatureSet := util.NewFixedLengthStringSet(len(effectPolicies))
	for _, p := range effectPolicies {
		expression := expressionMap[p.ExpressionPK]
		if !signatureSet.Has(expression.Signature) {
			signatureSet.Add(expression.Signature)

			policies = append(policies, convertToAuthPolicy(p, expression))
		}
	}

	// 7. return
	// debug.WithValue(entry, "return policies", policies)
	reportTooLargeReturnedPolicies(len(policies), system, action.ID, subject.Type, subject.ID)
	return policies, nil
}

// GetExpressionsFromCache will retrieve expression from cache
func (m *policyManager) GetExpressionsFromCache(
	actionPK int64,
	expressionPKs []int64,
) ([]svctypes.AuthExpression, error) {
	return expression.GetExpressionsFromCache(actionPK, expressionPKs)
}
