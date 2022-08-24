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
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/translate"
	abactypes "iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// policy for engine will put here

var ErrUnsupportedPolicyType = errors.New("unsupported type")

type EnginePolicy struct {
	Version string
	ID      int64
	System  string
	// abac, policy with single action
	// rbac, policy with multiple actions
	ActionPKs []int64

	SubjectPK int64

	Expression map[string]interface{}
	TemplateID int64
	ExpiredAt  int64
	UpdatedAt  int64
}

type EnginePolicyManager interface {
	GetMaxPKBeforeUpdatedAt(_type string, updatedAt int64) (int64, error)
	ListPKBetweenUpdatedAt(_type string, beginUpdatedAt, endUpdatedAt int64) ([]int64, error)
	ListBetweenPK(_type string, expiredAt, minPK, maxPK int64) (policies []EnginePolicy, err error)
	ListByPKs(_type string, pks []int64) (policies []EnginePolicy, err error)
}

type enginePolicyManager struct {
	abacService service.EngineAbacPolicyService
	rbacService service.EngineRbacPolicyService
}

func NewEnginePolicyManager() EnginePolicyManager {
	return &enginePolicyManager{
		abacService: service.NewEngineAbacPolicyService(),
		rbacService: service.NewEngineRbacPolicyService(),
	}
}

func (m *enginePolicyManager) GetMaxPKBeforeUpdatedAt(_type string, updatedAt int64) (int64, error) {
	switch _type {
	case PolicyTypeAbac:
		return m.abacService.GetMaxPKBeforeUpdatedAt(updatedAt)
	case "rbac":
		maxPK, err := m.rbacService.GetMaxPKBeforeUpdatedAt(updatedAt)
		if err != nil {
			return 0, err
		}
		return realPKToOutputRbacPolicyPK(maxPK), nil
	default:
		return 0, ErrUnsupportedPolicyType
	}
}

func (m *enginePolicyManager) ListPKBetweenUpdatedAt(
	_type string,
	beginUpdatedAt, endUpdatedAt int64,
) ([]int64, error) {
	switch _type {
	case PolicyTypeAbac:
		return m.abacService.ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt)
	case PolicyTypeRbac:
		pks, err := m.rbacService.ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt)
		if err != nil {
			return nil, err
		}
		rbacPKs := make([]int64, 0, len(pks))
		for _, pk := range pks {
			rbacPKs = append(rbacPKs, realPKToOutputRbacPolicyPK(pk))
		}
		return rbacPKs, nil
	default:
		return nil, ErrUnsupportedPolicyType
	}
}

func (m *enginePolicyManager) ListBetweenPK(
	_type string,
	expiredAt, minPK, maxPK int64,
) (policies []EnginePolicy, err error) {
	switch _type {
	case PolicyTypeAbac:
		abacPolicies, err := m.abacService.ListBetweenPK(expiredAt, minPK, maxPK)
		if err != nil {
			return nil, err
		}
		return convertAbacPoliciesToEnginePolicies(abacPolicies)
	case PolicyTypeRbac:
		minPK = inputRbacPolicyPKToRealPK(minPK)
		maxPK = inputRbacPolicyPKToRealPK(maxPK)
		rbacPolicies, err := m.rbacService.ListBetweenPK(expiredAt, minPK, maxPK)
		if err != nil {
			return nil, err
		}
		return convertRbacPoliciesToEnginePolicies(rbacPolicies)
	default:
		return nil, ErrUnsupportedPolicyType
	}
}

func (m *enginePolicyManager) ListByPKs(_type string, pks []int64) (policies []EnginePolicy, err error) {
	switch _type {
	case PolicyTypeAbac:
		abacPolicies, err := m.abacService.ListByPKs(pks)
		if err != nil {
			return nil, err
		}
		return convertAbacPoliciesToEnginePolicies(abacPolicies)
	case PolicyTypeRbac:
		realPKs := inputRbacPolicyPKsToRealPKs(pks)

		rbacPolicies, err := m.rbacService.ListByPKs(realPKs)
		if err != nil {
			return nil, err
		}
		return convertRbacPoliciesToEnginePolicies(rbacPolicies)
	default:
		return nil, ErrUnsupportedPolicyType
	}
}

func convertAbacPoliciesToEnginePolicies(
	policies []types.EngineAbacPolicy,
) (enginePolicies []EnginePolicy, err error) {
	if len(policies) == 0 {
		return
	}

	// query all expression
	expressionPKs := make([]int64, 0, len(policies))
	for _, p := range policies {
		if p.ExpressionPK != AnyExpressionPK {
			expressionPKs = append(expressionPKs, p.ExpressionPK)
		}
	}
	pkExpressionMap, err := queryAndTranslateExpressions(expressionPKs)
	if err != nil {
		err = fmt.Errorf("translateExpressions expressionPKs=`%+v` fail. err=%w", expressionPKs, err)
		return
	}

	// loop policies to build enginePolicies
	for _, p := range policies {
		expression, ok := pkExpressionMap[p.ExpressionPK]
		if !ok {
			log.Errorf(
				"convertQueryPoliciesToOpenPolicies p.ExpressionPK=`%d` missing in pkExpressionMap",
				p.ExpressionPK,
			)
			continue
		}

		enginePolicies = append(enginePolicies, EnginePolicy{
			Version:    service.PolicyVersion,
			ID:         p.PK,
			ActionPKs:  []int64{p.ActionPK},
			SubjectPK:  p.SubjectPK,
			Expression: expression,
			TemplateID: p.TemplateID,
			ExpiredAt:  p.ExpiredAt,
			UpdatedAt:  p.UpdatedAt.Unix(),
		})
	}
	return enginePolicies, nil
}

func convertRbacPoliciesToEnginePolicies(policies []types.EngineRbacPolicy) ([]EnginePolicy, error) {
	queryPolicies := make([]EnginePolicy, 0, len(policies))
	for _, p := range policies {
		expr, err := constructRbacPolicyExpr(p)
		if err != nil {
			log.WithError(err).Errorf("engine rbac policy constructExpr fail, policy=`%+v", p)
			continue
		}

		queryPolicies = append(queryPolicies, EnginePolicy{
			Version:    service.PolicyVersion,
			ID:         realPKToOutputRbacPolicyPK(p.PK),
			ActionPKs:  p.ActionPKs,
			SubjectPK:  p.GroupPK,
			Expression: expr,
			TemplateID: p.TemplateID,
			ExpiredAt:  util.NeverExpiresUnixTime,
			UpdatedAt:  p.UpdatedAt.Unix(),
		})
	}
	return queryPolicies, nil
}

func constructRbacPolicyExpr(p types.EngineRbacPolicy) (exprCell map[string]interface{}, err error) {
	action_rt, err := cacheimpls.GetThinResourceType(p.ActionRelatedResourceTypePK)
	if err != nil {
		return
	}

	if p.ActionRelatedResourceTypePK == p.ResourceTypePK {
		// pipeline.id eq 123
		exprCell = translate.ExprCell{
			"op":    "eq",
			"field": action_rt.ID + ".id",
			"value": p.ResourceID,
		}
	} else {
		// pipeline._bk_iam_path_ string_contains "/project,1/"
		var rt types.ThinResourceType
		rt, err = cacheimpls.GetThinResourceType(p.ResourceTypePK)
		if err != nil {
			return
		}

		exprCell = translate.ExprCell{
			"op":    "string_contains",
			"field": action_rt.ID + abactypes.IamPathSuffix,
			"value": fmt.Sprintf("/%s,%s/", rt.ID, p.ResourceID),
		}
	}

	return exprCell, nil
}
