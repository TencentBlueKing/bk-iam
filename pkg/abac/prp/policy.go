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

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/TencentBlueKing/gopkg/stringx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/prp/expression"
	"iam/pkg/abac/prp/policy"
	"iam/pkg/abac/prp/rbac"
	"iam/pkg/abac/prp/temporary"
	"iam/pkg/abac/types"
	"iam/pkg/logging/debug"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"
)

// PRP ...
const PRP = "PRP"

const (
	tooLargeThreshold   = 300
	queryTypePolicy     = "ListPolicy"
	queryTypeExpression = "ListExpression"
)

var emptyAuthExpression = svctypes.AuthExpression{}

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

// PolicyManager ...
type PolicyManager interface {
	ListBySubjectAction(system string, subject types.Subject, action types.Action, effectGroupPKs []int64,
		withoutCache bool, entry *debug.Entry) ([]types.AuthPolicy, error) // 需要对service查询来的policy去重

	GetExpressionsFromCache(actionPK int64, expressionPKs []int64) ([]svctypes.AuthExpression, error)
}

type policyManager struct {
	policyService          service.PolicyService
	temporaryPolicyService service.TemporaryPolicyService
}

// NewPolicyManager ...
func NewPolicyManager() PolicyManager {
	return &policyManager{
		policyService:          service.NewPolicyService(),
		temporaryPolicyService: service.NewTemporaryPolicyService(),
	}
}

// ListBySubjectAction 查询用于鉴权的policy列表
// policy有2个来源
// 	1. 普通权限(自定义权限, 继承的用户组权限)
// 	2. 临时权限(只来自个人)
func (m *policyManager) ListBySubjectAction(
	system string,
	subject types.Subject,
	action types.Action,
	effectGroupPKs []int64,
	withoutCache bool,
	parentEntry *debug.Entry,
) (policies []types.AuthPolicy, err error) {
	entry := debug.NewSubDebug(parentEntry)
	if entry != nil {
		debug.WithValue(entry, "cacheEnabled", !withoutCache)
	}

	// 1. 查询一般权限
	debug.AddStep(entry, "query policy")
	policies, err = m.listBySubjectAction(
		system, subject, action, effectGroupPKs, withoutCache, entry,
	)
	if err != nil {
		return
	}
	debug.WithValue(entry, "policies", policies)

	// 2. 查询临时权限
	debug.AddStep(entry, "query temporary policy")
	temporaryPolicies, err := m.listTemporaryBySubjectAction(
		system, subject, action, withoutCache, entry,
	)
	if err != nil {
		return
	}
	debug.WithValue(entry, "temporaryPolicies", temporaryPolicies)

	if len(temporaryPolicies) != 0 {
		policies = append(policies, temporaryPolicies...)
	}

	debug.AddStep(entry, "get action auth type")
	actionAuthType, err := action.Attribute.GetAuthType()
	if err != nil {
		return policies, err
	}
	debug.WithValue(entry, "actionAuthType", actionAuthType)

	if actionAuthType != svctypes.AuthTypeRBAC {
		return
	}

	// 3. 查询RBAC表达式
	debug.AddStep(entry, "query rbac policy")
	rbacPolicies, err := m.listRbacBySubjectAction(
		system, subject, action, withoutCache, entry,
	)
	if err != nil {
		return
	}
	debug.WithValue(entry, "rbacPolicies", rbacPolicies)

	if len(rbacPolicies) != 0 {
		policies = append(policies, rbacPolicies...)
	}
	return policies, nil
}

// listBySubjectAction 查询普通权限
func (m *policyManager) listBySubjectAction(
	system string,
	subject types.Subject,
	action types.Action,
	effectGroupPKs []int64,
	withoutCache bool,
	parentEntry *debug.Entry,
) (policies []types.AuthPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "listBySubjectAction")

	entry := debug.NewSubDebug(parentEntry)

	// 1. get effect subject pks
	debug.AddStep(entry, "Get Effect Subject PKs")
	effectSubjectPKs, err := m.getEffectSubjectPKs(subject, effectGroupPKs)
	if err != nil {
		err = errorWrapf(err, "Get Effect Subject PKs")
		return
	}

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
	expressionPKSet := set.NewFixedLengthInt64Set(len(effectPolicies))
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
	signatureSet := set.NewFixedLengthStringSet(len(effectPolicies))
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

func (*policyManager) getEffectSubjectPKs(subject types.Subject, effectPKs []int64) ([]int64, error) {
	subjectPK, err := subject.Attribute.GetPK()
	if err != nil {
		return nil, err
	}

	effectSubjectPKs := make([]int64, 0, len(effectPKs)+1)
	effectSubjectPKs = append(effectSubjectPKs, subjectPK)
	effectSubjectPKs = append(effectSubjectPKs, effectPKs...)
	return effectSubjectPKs, nil
}

// listTemporaryBySubjectAction 查询临时权限
func (m *policyManager) listTemporaryBySubjectAction(
	system string,
	subject types.Subject,
	action types.Action,
	withoutCache bool,
	parentEntry *debug.Entry,
) (polices []types.AuthPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "listTemporaryBySubjectAction")

	entry := debug.NewSubDebug(parentEntry)

	// 1. get subject pk
	debug.AddStep(entry, "Get Subject PK")
	subjectPK, err := subject.Attribute.GetPK()
	if err != nil {
		err = errorWrapf(err, "subject.Attribute subject=`%+v` fail", subject)
		return
	}
	debug.WithValue(entry, "subjectPK", subjectPK)

	// 2. get action pk
	debug.AddStep(entry, "Get Action PK")
	actionPK, err := action.Attribute.GetPK()
	if err != nil {
		err = errorWrapf(err, "action.Attribute.GetPK action=`%+v` fail", action)
		return
	}
	debug.WithValue(entry, "actionPK", actionPK)

	var retriever temporary.TemporaryPolicyRetriever
	if withoutCache {
		retriever = m.temporaryPolicyService
	} else {
		retriever = temporary.NewTemporaryPolicyCacheRetriever(system, m.temporaryPolicyService)
	}

	// 3. 查询在有效期内的临时权限pks
	debug.AddStep(entry, "List ThinTemporary Policy By Subject Action")
	thinTemporaryPolices, err := retriever.ListThinBySubjectAction(
		subjectPK, actionPK,
	)
	if err != nil {
		err = errorWrapf(err,
			"retriever.ListThinBySubjectAction subjectPK=`%d`, actionPK=`%d` fail",
			subjectPK, actionPK)
		return nil, err
	}
	debug.WithValue(entry, "thinTemporaryPolicies", thinTemporaryPolices)

	if len(thinTemporaryPolices) == 0 {
		debug.AddStep(entry, "thin temporary policies is empty so return")
		return
	}

	nowTimestamp := time.Now().Unix()
	pks := make([]int64, 0, len(thinTemporaryPolices))
	for _, p := range thinTemporaryPolices {
		if p.ExpiredAt > nowTimestamp {
			pks = append(pks, p.PK)
		}
	}
	debug.WithValue(entry, "temporaryPolicyPKs", pks)

	if len(pks) == 0 {
		debug.AddStep(entry, "all temporary policy expired so return")
		return
	}

	// 4. 查询临时权限数据
	var temporaryPolicies []svctypes.TemporaryPolicy
	debug.AddStep(entry, "List Temporary Policy By pks")
	temporaryPolicies, err = retriever.ListByPKs(pks)
	if err != nil {
		err = errorWrapf(err, "retriever.ListByPKs pks=`%+v` fail", pks)
		return
	}
	debug.WithValue(entry, "temporaryPolicies", temporaryPolicies)

	// 5. 转换数据结构
	polices = make([]types.AuthPolicy, 0, len(temporaryPolicies))
	for _, p := range temporaryPolicies {
		polices = append(polices, types.AuthPolicy{
			Version:             service.PolicyVersion,
			ID:                  p.PK,
			Expression:          p.Expression,
			ExpiredAt:           p.ExpiredAt,
			ExpressionSignature: stringx.MD5Hash(p.Expression),
		})
	}

	return polices, nil
}

// listRbacBySubjectAction 查询rbac表达式
func (m *policyManager) listRbacBySubjectAction(
	system string,
	subject types.Subject,
	action types.Action,
	withoutCache bool,
	parentEntry *debug.Entry,
) (polices []types.AuthPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "listRbacBySubjectAction")

	entry := debug.NewSubDebug(parentEntry)

	// 1. get subject department pk
	debug.AddStep(entry, "Get Subject Department PK")
	departmentPKs, err := subject.Attribute.GetDepartments()
	if err != nil {
		err = errorWrapf(err, "subject.Attribute.GetDepartments subject=`%+v` fail", subject)
		return
	}
	debug.WithValue(entry, "departmentPKs", departmentPKs)

	// 2. get effect subject pks
	debug.AddStep(entry, "Get Effect Subject PKs")
	effectSubjectPKs, err := m.getEffectSubjectPKs(subject, departmentPKs)
	if err != nil {
		err = errorWrapf(err, "Get Effect Subject PKs")
		return
	}
	debug.WithValue(entry, "effectSubjectPKs", effectSubjectPKs)

	// 3. get action pk
	debug.AddStep(entry, "Get Action PK")
	actionPK, err := action.Attribute.GetPK()
	if err != nil {
		err = errorWrapf(err, "action.Attribute.GetPK action=`%+v` fail", action)
		return
	}
	debug.WithValue(entry, "actionPK", actionPK)

	var retriever rbac.RbacPolicyRetriever
	if withoutCache {
		retriever = rbac.NewRbacPolicyDatabaseRetriever()
	} else {
		retriever = rbac.NewRbacPolicyRedisRetriever()
	}

	// 4. 查询rbac表达式
	debug.AddStep(entry, "List Rbac Expression By Subject Action")
	rbacExpressions, err := retriever.ListBySubjectAction(effectSubjectPKs, actionPK)
	if err != nil {
		err = errorWrapf(
			err,
			"retriever.ListBySubjectAction subjectPKs=`%+v`, actionPK=`%d` fail",
			effectSubjectPKs,
			actionPK,
		)
		return
	}
	debug.WithValue(entry, "rbacExpressions", rbacExpressions)

	// 5. 转换数据结构
	polices = make([]types.AuthPolicy, 0, len(rbacExpressions))
	for _, p := range rbacExpressions {
		polices = append(polices, types.AuthPolicy{
			Version:             service.PolicyVersion,
			ID:                  p.PK,
			Expression:          p.Expression,
			ExpiredAt:           p.ExpiredAt,
			ExpressionSignature: stringx.MD5Hash(p.Expression),
		})
	}

	return polices, nil
}

// GetExpressionsFromCache will retrieve expression from cache
func (m *policyManager) GetExpressionsFromCache(
	actionPK int64,
	expressionPKs []int64,
) ([]svctypes.AuthExpression, error) {
	return expression.GetExpressionsFromCache(actionPK, expressionPKs)
}
