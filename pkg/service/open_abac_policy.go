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

import (
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// OpenAbacPolicyService ...
type OpenAbacPolicyService interface {
	Get(pk int64) (types.OpenAbacPolicy, error)
	ListPagingByActionBeforeExpiredAt(
		actionPK int64, expiredAt int64, offset int64, limit int64) ([]types.OpenAbacPolicy, error)
	GetCountByActionBeforeExpiredAt(actionPK int64, expiredAt int64) (int64, error)

	ListByPKs(pks []int64) ([]types.OpenAbacPolicy, error)
}

type openAbacPolicyService struct {
	manager dao.OpenAbacPolicyManager
}

// NewPolicyService ...
func NewOpenAbacPolicyService() OpenAbacPolicyService {
	return &openAbacPolicyService{
		manager: dao.NewOpenAbacPolicyManager(),
	}
}

// Get ...
func (s *openAbacPolicyService) Get(pk int64) (daoPolicy types.OpenAbacPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "Get")
	policy, err1 := s.manager.Get(pk)
	if err1 != nil {
		err = errorWrapf(err1, "manager.Get pk=`%d` fail", pk)
		return
	}

	daoPolicy = types.OpenAbacPolicy{
		PK:           policy.PK,
		SubjectPK:    policy.SubjectPK,
		ActionPK:     policy.ActionPK,
		ExpressionPK: policy.ExpressionPK,
		ExpiredAt:    policy.ExpiredAt,
	}
	return
}

// GetCountByActionBeforeExpiredAt ...
func (s *openAbacPolicyService) GetCountByActionBeforeExpiredAt(actionPK int64, expiredAt int64) (int64, error) {
	return s.manager.GetCountByActionBeforeExpiredAt(actionPK, expiredAt)
}

// ListPagingByActionBeforeExpiredAt ...
func (s *openAbacPolicyService) ListPagingByActionBeforeExpiredAt(
	actionPK int64,
	expiredAt int64,
	offset int64,
	limit int64,
) (queryPolicies []types.OpenAbacPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "ListPagingByActionBeforeExpiredAt")

	policies, err := s.manager.ListPagingByActionPKBeforeExpiredAt(actionPK, expiredAt, offset, limit)
	if err != nil {
		err = errorWrapf(err,
			"manager.ListByActionPK actionPK=`%d`, expiredAt=`%d`, offset=`%d`, limit=`%d` fail",
			actionPK, expiredAt, offset, limit)
		return nil, err
	}

	queryPolicies = convertPoliciesToOpenAbacPolicies(policies)
	return
}

// ListByPKs ...
func (s *openAbacPolicyService) ListByPKs(pks []int64) (queryPolicies []types.OpenAbacPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "ListByPKs")

	policies, err := s.manager.ListByPKs(pks)
	if err != nil {
		err = errorWrapf(err,
			"manager.ListByPKs pks=`%+v` fail", pks)
		return nil, err
	}

	queryPolicies = convertPoliciesToOpenAbacPolicies(policies)
	return
}

func convertPoliciesToOpenAbacPolicies(policies []dao.OpenAbacPolicy) []types.OpenAbacPolicy {
	queryPolicies := make([]types.OpenAbacPolicy, 0, len(policies))
	for _, p := range policies {
		queryPolicies = append(queryPolicies, types.OpenAbacPolicy{
			PK:           p.PK,
			SubjectPK:    p.SubjectPK,
			ActionPK:     p.ActionPK,
			ExpressionPK: p.ExpressionPK,
			ExpiredAt:    p.ExpiredAt,
		})
	}
	return queryPolicies
}
