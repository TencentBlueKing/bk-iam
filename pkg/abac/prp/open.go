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
	"database/sql"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/translate"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
)

type OpenPolicy struct {
	Version    string
	ID         int64
	ActionPK   int64
	SubjectPK  int64
	Expression map[string]interface{}
	ExpiredAt  int64
}

// OpenPolicyManager for /api/v1/open/systems/:system_id/policies api
type OpenPolicyManager interface {
	Get(_type string, systemID string, pk int64) (OpenPolicy, error)
	List(_type string, actionPK int64, expiredAt int64, offset, limit int64) (int64, []OpenPolicy, error)
	ListSubjects(_type string, systemID string, pks []int64) (map[int64]int64, error)
}

type openPolicyManager struct {
	policyService service.PolicyService
}

func NewOpenPolicyManager() OpenPolicyManager {
	return &openPolicyManager{
		policyService: service.NewPolicyService(),
	}
}

var ErrPolicyNotFound = errors.New("policy not found")

func (m *openPolicyManager) Get(_type string, systemID string, pk int64) (openPolicy OpenPolicy, err error) {
	// 1. query policy
	policy, err := m.policyService.Get(pk)
	if err != nil {
		// 不存在的情况, 404
		if errors.Is(err, sql.ErrNoRows) {
			return openPolicy, ErrPolicyNotFound
		}
		return openPolicy, err
	}

	// 2. get expression
	pkExpressionMap, err := translateExpressions([]int64{policy.ExpressionPK})
	if err != nil {
		return openPolicy, err
	}
	expression, ok := pkExpressionMap[policy.ExpressionPK]
	if !ok {
		return openPolicy, fmt.Errorf("expression pk=`%d` missing", policy.ExpressionPK)
	}

	openPolicy = OpenPolicy{
		Version:    service.PolicyVersion,
		ID:         policy.PK,
		ActionPK:   policy.ActionPK,
		SubjectPK:  policy.SubjectPK,
		Expression: expression,
		ExpiredAt:  policy.ExpiredAt,
	}

	return openPolicy, nil
}

func (m *openPolicyManager) List(
	_type string,
	actionPK int64,
	expiredAt int64,
	offset, limit int64,
) (count int64, policies []OpenPolicy, err error) {
	// 3. do query: 查询某个系统, 某个action的所有policy列表  带分页
	count, err = m.policyService.GetCountByActionBeforeExpiredAt(actionPK, expiredAt)
	if err != nil {
		return 0, nil, fmt.Errorf(
			"svc.GetCountByActionBeforeExpiredAt actionPK=`%d`, expiredAt=`%d` fail. err=%w",
			actionPK,
			expiredAt,
			err,
		)
	}

	if count == 0 {
		return 0, []OpenPolicy{}, nil
	}

	var queryPolicies []svctypes.QueryPolicy
	queryPolicies, err = m.policyService.ListPagingQueryByActionBeforeExpiredAt(actionPK, expiredAt, offset, limit)
	if err != nil {
		err = fmt.Errorf(
			"svc.ListPagingQueryByActionBeforeExpiredAt actionPK=`%d`, expiredAt=`%d`, offset=`%d`, limit=`%d` fail. err=%w",
			actionPK,
			expiredAt,
			offset,
			limit,
			err,
		)
		return 0, nil, err
	}

	policies, err = convertQueryPoliciesToOpenPolicies(queryPolicies)
	if err != nil {
		err = fmt.Errorf(
			"convertQueryPoliciesToOpenPolicies queryPolicies=`%+v` fail. err=%w",
			queryPolicies, err)
		return 0, nil, err
	}

	return count, policies, nil
}

func (m *openPolicyManager) ListSubjects(
	_type string,
	systemID string,
	pks []int64,
) (map[int64]int64, error) {
	// NOTE: 防止敏感信息泄漏, 只能查询自己系统 + 自己action的
	// 1. query policy
	policies, err := m.policyService.ListQueryByPKs(pks)
	if err != nil {
		return nil, fmt.Errorf("svc.ListQueryByPKs system=`%s`, policy_ids=`%+v` fail. err=%w",
			systemID, pks, err)
	}

	if len(policies) == 0 {
		return nil, nil
	}

	data := make(map[int64]int64, len(policies))
	for _, policy := range policies {
		sa, err := cacheimpls.GetAction(policy.ActionPK)
		if err != nil {
			log.Infof("cacheimpls.GetAction action_pk=`%d` fail. err=%s", policy.ActionPK, err.Error())

			continue
		}
		// 不是本系统的策略, 过滤掉. not my system policy, continue
		if systemID != sa.System {
			continue
		}

		data[policy.PK] = policy.SubjectPK
	}

	return data, nil
}

// translateExpressions translate expression to json format
func translateExpressions(
	expressionPKs []int64,
) (expressionMap map[int64]map[string]interface{}, err error) {
	// when the pk is -1, will translate to any
	pkExpressionStrMap := map[int64]string{
		-1: "",
	}
	if len(expressionPKs) > 0 {
		manager := NewPolicyManager()

		var exprs []svctypes.AuthExpression
		exprs, err = manager.GetExpressionsFromCache(-1, expressionPKs)
		if err != nil {
			err = fmt.Errorf("policyManager.GetExpressionsFromCache pks=`%+v` fail. err=%w", expressionPKs, err)
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
			err = fmt.Errorf("translate.PolicyExpressionTranslate expr=`%s` fail. err=%w", expr, err1)
			return
		}
		expressionMap[pk] = translatedExpr
	}
	return expressionMap, nil
}

func convertQueryPoliciesToOpenPolicies(
	policies []svctypes.QueryPolicy,
) (openPolicies []OpenPolicy, err error) {
	if len(policies) == 0 {
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
	pkExpressionMap, err := translateExpressions(expressionPKs)
	if err != nil {
		err = fmt.Errorf("translateExpressions expressionPKs=`%+v` fail. err=%w", expressionPKs, err)
		return
	}

	// loop policies to build openPolicies
	for _, p := range policies {
		// if missing the expression, continue
		expression, ok := pkExpressionMap[p.ExpressionPK]
		if !ok {
			log.Errorf(
				"convertQueryPoliciesToOpenPolicies p.ExpressionPK=`%d` missing in pkExpressionMap",
				p.ExpressionPK,
			)
			continue
		}

		openPolicies = append(openPolicies, OpenPolicy{
			Version:    service.PolicyVersion,
			ID:         p.PK,
			ActionPK:   p.ActionPK,
			SubjectPK:  p.SubjectPK,
			Expression: expression,
			ExpiredAt:  p.ExpiredAt,
		})
	}
	return openPolicies, nil
}
