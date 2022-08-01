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

type EngineAbacPolicyService interface {
	GetMaxPKBeforeUpdatedAt(updatedAt int64) (int64, error)
	ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) ([]int64, error)
	ListBetweenPK(expiredAt, minPK, maxPK int64) (policies []types.EngineAbacPolicy, err error)
	ListByPKs(pks []int64) (policies []types.EngineAbacPolicy, err error)
}

type engineAbacPolicyService struct {
	manager dao.EngineAbacPolicyManager
}

// NewEnginePolicyService create the EnginePolicyService
func NewEngineAbacPolicyService() EngineAbacPolicyService {
	return &engineAbacPolicyService{
		manager: dao.NewAbacEnginePolicyManager(),
	}
}

// GetMaxPKBeforeUpdatedAt ...
func (s *engineAbacPolicyService) GetMaxPKBeforeUpdatedAt(updatedAt int64) (int64, error) {
	return s.manager.GetMaxPKBeforeUpdatedAt(updatedAt)
}

// ListPKBetweenUpdatedAt ...
func (s *engineAbacPolicyService) ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) ([]int64, error) {
	return s.manager.ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt)
}

// ListBetweenPK ...
func (s *engineAbacPolicyService) ListBetweenPK(
	expiredAt, minPK, maxPK int64,
) (queryPolicies []types.EngineAbacPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(EnginePolicySVC, "ListBetweenPK")

	policies, err := s.manager.ListBetweenPK(expiredAt, minPK, maxPK)
	if err != nil {
		err = errorWrapf(err,
			"manager.ListBetweenPK expiredAt=`%d`, minPK=`%d`, maxPK=`%d` fail",
			expiredAt, minPK, maxPK,
		)
		return nil, err
	}
	return convertToEngineAbacPolicies(policies), nil
}

// ListByPKs ...
func (s *engineAbacPolicyService) ListByPKs(pks []int64) (queryPolicies []types.EngineAbacPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(EnginePolicySVC, "ListByPKs")

	policies, err := s.manager.ListByPKs(pks)
	if err != nil {
		err = errorWrapf(err, "manager.ListByPKs pks=`%+v` fail", pks)
		return nil, err
	}
	return convertToEngineAbacPolicies(policies), nil
}

func convertToEngineAbacPolicies(policies []dao.EngineAbacPolicy) (abacPolicies []types.EngineAbacPolicy) {
	if len(policies) == 0 {
		return
	}

	for _, p := range policies {
		abacPolicies = append(abacPolicies, types.EngineAbacPolicy{
			PK:           p.PK,
			SubjectPK:    p.SubjectPK,
			ActionPK:     p.ActionPK,
			ExpressionPK: p.ExpressionPK,
			ExpiredAt:    p.ExpiredAt,
			TemplateID:   p.TemplateID,
			UpdatedAt:    p.UpdatedAt,
		})
	}
	return
}
