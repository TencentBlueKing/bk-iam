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

import (
	"database/sql"
	"errors"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/types"
	svcTypes "iam/pkg/service/types"
)

// ListSaaSBySubjectSystemTemplate 根据system和subject查询相关的policy的列表
func (m *policyManager) ListSaaSBySubjectSystemTemplate(
	system, subjectType, subjectID string,
	templateID int64,
) ([]types.SaaSPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "ListSaaSPolicyBySubjectSystemTemplate")

	// 查询subject pk
	pk, err := m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return nil, err
	}

	// 查询系统的所有action
	actions, err := m.actionService.ListThinActionBySystem(system)
	if err != nil {
		err = errorWrapf(err, "actionService.ListThinActionBySystem system=`%s` fail",
			system)
		return nil, err
	}

	if len(actions) == 0 {
		return []types.SaaSPolicy{}, nil
	}

	actionPKs := make([]int64, 0, len(actions))
	for _, ac := range actions {
		actionPKs = append(actionPKs, ac.PK)
	}

	// 查询subject相关的policies
	policies, err := m.policyService.ListThinBySubjectActionTemplate(pk, actionPKs, templateID)
	if (len(policies) == 0 && err == nil) || errors.Is(err, sql.ErrNoRows) {
		return []types.SaaSPolicy{}, nil
	}

	if err != nil {
		err = errorWrapf(
			err, "policyService.ListThinBySubjectActionTemplate pk=`%d`, actionPKs=`%+v`, templateID=`%d` fail",
			pk, actionPKs, templateID)
		return nil, err
	}

	// 转换数据结构
	return m.convertToSaaSPolicies(policies, actions), nil
}

// GetByActionTemplate ...
func (m *policyManager) GetByActionTemplate(
	system,
	subjectType,
	subjectID,
	actionID string,
	templateID int64,
) (policy types.AuthPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "GetCustomByAction")
	// 查询subject pk
	pk, err := m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return
	}

	actionPK, err := m.actionService.GetActionPK(system, actionID)
	if err != nil {
		err = errorWrapf(err, "actionService.Get system=`%s` actionID=`%s` fail", system, actionID)
		return
	}

	svcTypesPolicy, err := m.policyService.GetByActionTemplate(pk, actionPK, 0)
	if err != nil {
		err = errorWrapf(err, "policyService.GetByActionTemplate subjectPK=`%d`, actionPK=`%d` fail", pk, actionPK)
		return
	}
	policy = types.AuthPolicy{
		Version:    svcTypesPolicy.Version,
		ID:         svcTypesPolicy.ID,
		Expression: svcTypesPolicy.Expression,
		ExpiredAt:  svcTypesPolicy.ExpiredAt,
	}
	return policy, err
}

// ListSaaSBySubjectTemplateBeforeExpiredAt 根据system和subject查询相关的policy的列表
func (m *policyManager) ListSaaSBySubjectTemplateBeforeExpiredAt(
	subjectType, subjectID string,
	templateID, expiredAt int64,
) ([]types.SaaSPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "ListSaaSBySubjectTemplateBeforeExpiredAt")

	// 查询subject pk
	pk, err := m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return nil, err
	}

	// 查询subject相关的policies
	policies, err := m.policyService.ListThinBySubjectTemplateBeforeExpiredAt(pk, templateID, expiredAt)
	if (len(policies) == 0 && err == nil) || errors.Is(err, sql.ErrNoRows) {
		return []types.SaaSPolicy{}, nil
	}

	// 查询所有action PK的系统信息
	actionPKs := make([]int64, 0, len(policies))
	for _, p := range policies {
		actionPKs = append(actionPKs, p.ActionPK)
	}

	actions, err := m.actionService.ListThinActionByPKs(actionPKs)
	if err != nil {
		err = errorWrapf(err, "actionService.ListThinActionByPKs actionPKs=`%v` fail", actionPKs)
		return nil, err
	}

	// 转换数据结构
	return m.convertToSaaSPolicies(policies, actions), nil
}

func (m *policyManager) convertToSaaSPolicies(
	policies []svcTypes.ThinPolicy,
	actions []svcTypes.ThinAction,
) []types.SaaSPolicy {
	// 转换数据结构
	actionMap := make(map[int64]svcTypes.ThinAction, len(actions))
	for _, a := range actions {
		actionMap[a.PK] = a
	}

	saasPolicies := make([]types.SaaSPolicy, 0, len(policies))
	for _, p := range policies {
		saasPolicies = append(saasPolicies, types.SaaSPolicy{
			Version:   p.Version,
			ID:        p.ID,
			System:    actionMap[p.ActionPK].System,
			ActionID:  actionMap[p.ActionPK].ID,
			ExpiredAt: p.ExpiredAt,
		})
	}
	return saasPolicies
}
