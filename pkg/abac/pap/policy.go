package pap

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"iam/pkg/abac/types"
	"iam/pkg/service"
)

// PolicyCTL ...
const PolicyCTL = "PolicyCTL"

type PolicyController interface {
	// policy list

	GetByActionTemplate(
		system, subjectType, subjectID, actionID string, templateID int64) (policy types.AuthPolicy, err error)
	ListSaaSBySubjectSystemTemplate(system, subjectType, subjectID string, templateID int64) ([]types.SaaSPolicy,
		error)
	ListSaaSBySubjectTemplateBeforeExpiredAt(subjectType, subjectID string, templateID, expiredAt int64) (
		[]types.SaaSPolicy, error)

	// policy curd

	AlterCustomPolicies(
		system, subjectType, subjectID string,
		createPolicies, updatePolicies []types.Policy, deletePolicyIDs []int64) error
	UpdateSubjectPoliciesExpiredAt(subjectType, subjectID string, policies []types.PolicyPKExpiredAt) error

	DeleteByIDs(system string, subjectType, subjectID string, policyIDs []int64) error
	DeleteByActionID(system, actionID string) error

	// template

	CreateAndDeleteTemplatePolicies(system, subjectType, subjectID string, templateID int64,
		createPolicies []types.Policy, deletePolicyIDs []int64) error
	UpdateTemplatePolicies(system, subjectType, subjectID string, policies []types.Policy) error
	DeleteTemplatePolicies(system, subjectType, subjectID string, templateID int64) error

	// temporary policy

	CreateTemporaryPolicies(
		system, subjectType, subjectID string,
		policies []types.Policy,
	) ([]int64, error)
	DeleteTemporaryByIDs(system string, subjectType, subjectID string, policyIDs []int64) error
	DeleteTemporaryBeforeExpiredAt(expiredAt int64) error
}

type policyController struct {
	subjectService         service.SubjectService
	actionService          service.ActionService
	policyService          service.PolicyService
	temporaryPolicyService service.TemporaryPolicyService

	// RBAC
	groupResourcePolicyService service.GroupResourcePolicyService
	groupService               service.GroupService
}

func NewPolicyController() PolicyController {
	return &policyController{
		subjectService:             service.NewSubjectService(),
		actionService:              service.NewActionService(),
		policyService:              service.NewPolicyService(),
		temporaryPolicyService:     service.NewTemporaryPolicyService(),
		groupResourcePolicyService: service.NewGroupResourcePolicyService(),
		groupService:               service.NewGroupService(),
	}
}
