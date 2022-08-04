package prp

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

// PolicyID Rule:
// abac, table `policy` auto increment ID, 0 - 500000000
// rbac, table `rbac_group_resource_policy` auto increment ID, but scope = 500000000 - 1000000000

const rbacIDBegin = 500000000

const (
	EngineListPolicyTypeAbac = "abac"
	EngineListPolicyTypeRbac = "rbac"
)

var ErrEngineListPolicyUnsupportedType = errors.New("unsupported type")

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
	case EngineListPolicyTypeAbac:
		return m.abacService.GetMaxPKBeforeUpdatedAt(updatedAt)
	case "rbac":
		maxPK, err := m.rbacService.GetMaxPKBeforeUpdatedAt(updatedAt)
		if err != nil {
			return 0, err
		}
		return maxPK + rbacIDBegin, nil
	default:
		return 0, ErrEngineListPolicyUnsupportedType
	}
}

func (m *enginePolicyManager) ListPKBetweenUpdatedAt(
	_type string,
	beginUpdatedAt, endUpdatedAt int64,
) ([]int64, error) {
	switch _type {
	case EngineListPolicyTypeAbac:
		return m.abacService.ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt)
	case EngineListPolicyTypeRbac:
		pks, err := m.rbacService.ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt)
		if err != nil {
			return nil, err
		}
		rbacPKs := make([]int64, 0, len(pks))
		for _, pk := range pks {
			rbacPKs = append(rbacPKs, pk+rbacIDBegin)
		}
		return rbacPKs, nil
	default:
		return nil, ErrEngineListPolicyUnsupportedType
	}
}

func (m *enginePolicyManager) ListBetweenPK(
	_type string,
	expiredAt, minPK, maxPK int64,
) (policies []EnginePolicy, err error) {
	switch _type {
	case EngineListPolicyTypeAbac:
		abacPolicies, err := m.abacService.ListBetweenPK(expiredAt, minPK, maxPK)
		if err != nil {
			return nil, err
		}
		return convertAbacPolicies(abacPolicies)
	case EngineListPolicyTypeRbac:
		minPK -= rbacIDBegin
		maxPK -= rbacIDBegin
		rbacPolicies, err := m.rbacService.ListBetweenPK(expiredAt, minPK, maxPK)
		if err != nil {
			return nil, err
		}
		return convertRbacPolicies(rbacPolicies)
	default:
		return nil, ErrEngineListPolicyUnsupportedType
	}
}

func (m *enginePolicyManager) ListByPKs(_type string, pks []int64) (policies []EnginePolicy, err error) {
	switch _type {
	case EngineListPolicyTypeAbac:
		abacPolicies, err := m.abacService.ListByPKs(pks)
		if err != nil {
			return nil, err
		}
		return convertAbacPolicies(abacPolicies)
	case EngineListPolicyTypeRbac:
		realPKs := make([]int64, 0, len(pks))
		for _, pk := range pks {
			realPKs = append(realPKs, pk-rbacIDBegin)
		}

		rbacPolicies, err := m.rbacService.ListByPKs(realPKs)
		if err != nil {
			return nil, err
		}
		return convertRbacPolicies(rbacPolicies)
	default:
		return nil, ErrEngineListPolicyUnsupportedType
	}
}

func convertAbacPolicies(
	policies []types.EngineAbacPolicy,
) (enginePolicies []EnginePolicy, err error) {
	if len(policies) == 0 {
		return
	}
	// query all expression
	pkExpressionStrMap, err := queryPoliciesExpression(policies)
	if err != nil {
		// err = errorWrapf(err, "queryPolicyExpression policies length=`%d` fail", len(enginePolicies))
		return
	}

	// loop policies to build enginePolicies
	for _, p := range policies {
		expr, ok := pkExpressionStrMap[p.ExpressionPK]
		if !ok {
			log.Errorf(
				"policy.convertEngineQueryPoliciesToEnginePolicies p.ExpressionPK=`%d` missing in pkExpressionMap",
				p.ExpressionPK,
			)

			continue
		}

		expression, err1 := translate.PolicyExpressionTranslate(expr)
		if err1 != nil {
			// err = errorWrapf(err2, "translate.PolicyExpressionTranslate policy=`%+v`, expr=`%s` fail", p, p.Expression)
			return nil, err
		}

		ep := EnginePolicy{
			Version:    service.PolicyVersion,
			ID:         p.PK,
			ActionPKs:  []int64{p.ActionPK},
			SubjectPK:  p.SubjectPK,
			Expression: expression,
			TemplateID: p.TemplateID,
			ExpiredAt:  p.ExpiredAt,
			UpdatedAt:  p.UpdatedAt.Unix(),
		}

		enginePolicies = append(enginePolicies, ep)
	}
	return enginePolicies, nil
}

// AnyExpressionPK is the pk for expression=any
const AnyExpressionPK = -1

func queryPoliciesExpression(policies []types.EngineAbacPolicy) (map[int64]string, error) {
	expressionPKs := make([]int64, 0, len(policies))
	for _, p := range policies {
		if p.ExpressionPK != AnyExpressionPK {
			expressionPKs = append(expressionPKs, p.ExpressionPK)
		}
	}

	pkExpressionStrMap := map[int64]string{
		// NOTE: -1 for the `any`
		AnyExpressionPK: "",
	}
	if len(expressionPKs) > 0 {
		manager := NewPolicyManager()

		var exprs []types.AuthExpression
		var err error
		exprs, err = manager.GetExpressionsFromCache(-1, expressionPKs)
		if err != nil {
			return nil, err
		}

		for _, e := range exprs {
			pkExpressionStrMap[e.PK] = e.Expression
		}
	}
	return pkExpressionStrMap, nil
}

func convertRbacPolicies(policies []types.EngineRbacPolicy) ([]EnginePolicy, error) {
	queryPolicies := make([]EnginePolicy, 0, len(policies))
	for _, p := range policies {
		expr, err := constructRbacPolicyExpr(p)
		if err != nil {
			log.WithError(err).Errorf("engine rbac policy constructExpr fail, policy=`%+v", p)
			continue
		}

		queryPolicies = append(queryPolicies, EnginePolicy{
			Version:    service.PolicyVersion,
			ID:         p.PK + rbacIDBegin,
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
