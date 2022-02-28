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
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
)

// TemporaryPolicySVC ...
const TemporaryPolicySVC = "TemporaryPolicySVC"

// TemporaryPolicyService ...
type TemporaryPolicyService interface {
	// for auth
	ListThinBySubjectAction(subjectPK, actionPK int64) ([]types.ThinTemporaryPolicy, error)
	ListByPKs(pks []int64) ([]types.TemporaryPolicy, error)
}

type temporaryPolicyService struct {
	manager dao.TemporaryPolicyManager
}

// NewPolicyService ...
func NewTemporaryPolicyService() TemporaryPolicyService {
	return &temporaryPolicyService{
		manager: dao.NewTemporaryPolicyManager(),
	}
}

// ListThinBySubjectAction ...
func (s *temporaryPolicyService) ListThinBySubjectAction(
	subjectPK, actionPK int64,
) ([]types.ThinTemporaryPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(TemporaryPolicySVC, "ListThinBySubjectAction")
	nowUnix := time.Now().Unix()
	daoPolicies, err := s.manager.ListThinBySubjectAction(subjectPK, actionPK, nowUnix)
	if err != nil {
		return nil, errorWrapf(
			err, "manager.ListThinBySubjectAction subjectPK=`%d`, actionPK=`%d`, expiredAt=`%d`",
			subjectPK, actionPK, nowUnix,
		)
	}

	policies := make([]types.ThinTemporaryPolicy, 0, len(daoPolicies))
	for _, p := range daoPolicies {
		policies = append(policies, types.ThinTemporaryPolicy{
			PK:        p.PK,
			ExpiredAt: p.ExpiredAt,
		})
	}
	return policies, nil
}

// ListByPKs ...
func (s *temporaryPolicyService) ListByPKs(pks []int64) ([]types.TemporaryPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(TemporaryPolicySVC, "ListByPKs")
	daoPolicies, err := s.manager.ListByPKs(pks)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListByPKs pks=`%+v`", pks)
	}

	policies := make([]types.TemporaryPolicy, 0, len(daoPolicies))
	for _, p := range daoPolicies {
		policies = append(policies, types.TemporaryPolicy{
			PK:         p.PK,
			SubjectPK:  p.SubjectPK,
			ActionPK:   p.ActionPK,
			Expression: p.Expression,
			ExpiredAt:  p.ExpiredAt,
		})
	}
	return policies, nil
}
