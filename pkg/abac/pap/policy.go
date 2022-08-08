/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pap

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"iam/pkg/abac/pap/event"
	"iam/pkg/abac/types"
	"iam/pkg/service"
)

// PolicyCTL ...
const PolicyCTL = "PolicyCTL"

type PolicyController interface {
	// TODO: 目前PolicyController集成了自定义权限、模板权限、临时权限，是否切分出TemplateController、TemporaryPolicyController ??
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

	DeleteByIDs(system string, subjectType, subjectID string, policyIDs []int64) error

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

	eventProducer event.PolicyEventProducer
}

func NewPolicyController() PolicyController {
	return &policyController{
		subjectService:         service.NewSubjectService(),
		actionService:          service.NewActionService(),
		policyService:          service.NewPolicyService(),
		temporaryPolicyService: service.NewTemporaryPolicyService(),

		eventProducer: event.NewPolicyEventProducer(),
	}
}
