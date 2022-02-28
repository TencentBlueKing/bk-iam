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
	"iam/pkg/abac/types"
	"iam/pkg/logging/debug"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
)

// PRP ...
const PRP = "PRP"

// PolicyManager ...
type PolicyManager interface {
	// in policy_list.go

	GetByActionTemplate(
		system, subjectType, subjectID, actionID string, templateID int64) (policy types.AuthPolicy, err error)
	ListBySubjectAction(system string, subject types.Subject, action types.Action,
		withoutCache bool, entry *debug.Entry) ([]types.AuthPolicy, error) // 需要对service查询来的policy去重

	ListSaaSBySubjectSystemTemplate(system, subjectType, subjectID string, templateID int64) ([]types.SaaSPolicy,
		error)
	ListSaaSBySubjectTemplateBeforeExpiredAt(subjectType, subjectID string, templateID, expiredAt int64) (
		[]types.SaaSPolicy, error)

	// in policy_crud.go

	AlterCustomPolicies(
		systemID, subjectType, subjectID string,
		createPolicies, updatePolicies []types.Policy, deletePolicyIDs []int64) error
	UpdateSubjectPoliciesExpiredAt(subjectType, subjectID string, policies []types.PolicyPKExpiredAt) error

	DeleteByIDs(system string, subjectType, subjectID string, policyIDs []int64) error

	GetExpressionsFromCache(actionPK int64, expressionPKs []int64) ([]svctypes.AuthExpression, error)
	DeleteByActionID(systemID, actionID string) error

	// template

	CreateAndDeleteTemplatePolicies(systemID, subjectType, subjectID string, templateID int64,
		createPolicies []types.Policy, deletePolicyIDs []int64) error
	UpdateTemplatePolicies(systemID, subjectType, subjectID string, policies []types.Policy) error
	DeleteTemplatePolicies(systemID, subjectType, subjectID string, templateID int64) error
}

type policyManager struct {
	subjectService         service.SubjectService
	actionService          service.ActionService
	policyService          service.PolicyService
	temporaryPolicyService service.TemporaryPolicyService
}

// NewPolicyManager ...
func NewPolicyManager() PolicyManager {
	return &policyManager{
		subjectService:         service.NewSubjectService(),
		actionService:          service.NewActionService(),
		policyService:          service.NewPolicyService(),
		temporaryPolicyService: service.NewTemporaryPolicyService(),
	}
}
