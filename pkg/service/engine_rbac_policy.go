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
	log "github.com/sirupsen/logrus"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

type engineRbacPolicyService struct {
	manager dao.EngineRbacPolicyManager
}

// NewRbacEnginePolicyService create the EnginePolicyService
func NewRbacEnginePolicyService() EnginePolicyService {
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
) (queryPolicies []types.EnginePolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(EnginePolicySVC, "ListBetweenPK")

	policies, err := s.manager.ListBetweenPK(minPK, maxPK)
	if err != nil {
		err = errorWrapf(err,
			"manager.ListBetweenPK expiredAt=`%d`, minPK=`%d`, maxPK=`%d` fail",
			expiredAt, minPK, maxPK,
		)
		return nil, err
	}

	queryPolicies = convertEngineRbacPoliciesToEnginePolicies(policies)
	return queryPolicies, nil
}

// ListByPKs ...
func (s *engineRbacPolicyService) ListByPKs(pks []int64) (queryPolicies []types.EnginePolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(EnginePolicySVC, "ListByPKs")

	policies, err := s.manager.ListByPKs(pks)
	if err != nil {
		err = errorWrapf(err, "manager.ListByPKs pks=`%+v` fail", pks)
		return nil, err
	}

	queryPolicies = convertEngineRbacPoliciesToEnginePolicies(policies)
	return queryPolicies, nil
}

func convertEngineRbacPoliciesToEnginePolicies(policies []dao.EngineRbacPolicy) []types.EnginePolicy {
	queryPolicies := make([]types.EnginePolicy, 0, len(policies))
	for _, p := range policies {
		actionPKs, err := util.StringToInt64Slice(p.ActionPKs, ",")
		if err != nil {
			log.WithError(err).
				Errorf("engine rbac policy action pks convert to int64 slice fail, actionPKs=`%+v`", p.ActionPKs)
			continue
		}
		// expr, err1 := constructExpr(p)
		// if err1 != nil {
		// 	log.WithError(err).Errorf("engine rbac policy constructExpr fail, policy=`%+v", p)
		// 	continue
		// }
		expr := ""
		queryPolicies = append(queryPolicies, types.EnginePolicy{
			Version:   PolicyVersion,
			ID:        p.PK,
			ActionPKs: actionPKs,
			SubjectPK: p.GroupPK,
			// ExpressionPK: p.ExpressionPK,
			ExpressionStr: expr,
			TemplateID:    p.TemplateID,
			ExpiredAt:     util.NeverExpiresUnixTime,
			UpdatedAt:     p.UpdatedAt.Unix(),
		})
	}
	return queryPolicies
}

// func constructExpr(p dao.EngineRbacPolicy) (string, error) {
// 	var exprCell expression.ExprCell

// 	action_rt, err := cacheimpls.GetThinResourceType(p.ActionRelatedResourceTypePK)
// 	if err != nil {
// 		return "", err
// 	}

// 	if p.ActionRelatedResourceTypePK == p.ResourceTypePK {
// 		// pipeline.id eq 123
// 		exprCell = expression.ExprCell{
// 			OP:    operator.Eq,
// 			Field: action_rt.ID + ".id",
// 			Value: p.ResourceID,
// 		}
// 	} else {
// 		// pipeline._bk_iam_path_ string_contains "/project,1/"
// 		rt, err := cacheimpls.GetThinResourceType(p.ResourceTypePK)
// 		if err != nil {
// 			return "", err
// 		}

// 		exprCell = expression.ExprCell{
// 			OP:    operator.StringContains,
// 			Field: action_rt.ID + abactypes.IamPathSuffix,
// 			Value: fmt.Sprintf("/%s,%s/", rt.ID, p.ResourceID),
// 		}
// 	}

// 	return jsoniter.MarshalToString(exprCell)
// }
