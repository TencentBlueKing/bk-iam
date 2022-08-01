/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

type EngineRbacPolicyService interface {
	GetMaxPKBeforeUpdatedAt(updatedAt int64) (int64, error)
	ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) ([]int64, error)
	ListBetweenPK(expiredAt, minPK, maxPK int64) (policies []types.EngineRbacPolicy, err error)
	ListByPKs(pks []int64) (policies []types.EngineRbacPolicy, err error)
}

type engineRbacPolicyService struct {
	manager dao.EngineRbacPolicyManager
}

// NewRbacEnginePolicyService create the EnginePolicyService
func NewEngineRbacPolicyService() EngineRbacPolicyService {
	return &engineRbacPolicyService{
		manager: dao.NewRbacEnginePolicyManager(),
	}
}

// GetMaxPKBeforeUpdatedAt ...
func (s *engineRbacPolicyService) GetMaxPKBeforeUpdatedAt(updatedAt int64) (int64, error) {
	return s.manager.GetMaxPKBeforeUpdatedAt(updatedAt)
}

// ListPKBetweenUpdatedAt ...
func (s *engineRbacPolicyService) ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) ([]int64, error) {
	return s.manager.ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt)
}

// ListBetweenPK ...
func (s *engineRbacPolicyService) ListBetweenPK(
	expiredAt, minPK, maxPK int64,
) (queryPolicies []types.EngineRbacPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(EnginePolicySVC, "ListBetweenPK")

	policies, err := s.manager.ListBetweenPK(minPK, maxPK)
	if err != nil {
		err = errorWrapf(err,
			"manager.ListBetweenPK expiredAt=`%d`, minPK=`%d`, maxPK=`%d` fail",
			expiredAt, minPK, maxPK,
		)
		return nil, err
	}

	return convertToEngineRbacPolicies(policies), nil
}

// ListByPKs ...
func (s *engineRbacPolicyService) ListByPKs(pks []int64) (queryPolicies []types.EngineRbacPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(EnginePolicySVC, "ListByPKs")

	policies, err := s.manager.ListByPKs(pks)
	if err != nil {
		err = errorWrapf(err, "manager.ListByPKs pks=`%+v` fail", pks)
		return nil, err
	}
	return convertToEngineRbacPolicies(policies), nil
}

func convertToEngineRbacPolicies(policies []dao.EngineRbacPolicy) (rbacPolicies []types.EngineRbacPolicy) {
	if len(policies) == 0 {
		return
	}

	for _, p := range policies {
		rbacPolicies = append(rbacPolicies, types.EngineRbacPolicy{
			PK:                          p.PK,
			GroupPK:                     p.GroupPK,
			TemplateID:                  p.TemplateID,
			SystemID:                    p.SystemID,
			ActionPKs:                   p.ActionPKs,
			ActionRelatedResourceTypePK: p.ActionRelatedResourceTypePK,
			ResourceTypePK:              p.ResourceTypePK,
			ResourceID:                  p.ResourceID,
			UpdatedAt:                   p.UpdatedAt,
		})
	}
	return
}
