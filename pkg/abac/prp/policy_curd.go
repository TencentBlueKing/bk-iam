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

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/prp/expression"
	"iam/pkg/abac/prp/policy"
	"iam/pkg/abac/types"
	svctypes "iam/pkg/service/types"
)

// NOTE: **important** / **重要**
//       curd中所有方法必须考虑删除policy缓存
//       curd中所有方法必须考虑删除policy缓存
//       curd中所有方法必须考虑删除policy缓存

var (
	ErrActionNotExists = errors.New("action not exists")
)

func convertToServicePolicies(
	subjectPK int64, policies []types.Policy, actionMap map[string]int64,
) ([]svctypes.Policy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "convertServicePolicies")

	svcPolicies := make([]svctypes.Policy, 0, len(policies))
	for _, p := range policies {
		actionPK, ok := actionMap[p.Action.ID]
		if !ok {
			err := errorWrapf(ErrActionNotExists, "actionID=`%s` fail", p.Action.ID)
			return nil, err
		}
		svcPolicies = append(svcPolicies, svctypes.Policy{
			Version:    p.Version,
			ID:         p.ID,
			SubjectPK:  subjectPK,
			ActionPK:   actionPK,
			Expression: p.Expression,
			ExpiredAt:  p.ExpiredAt,
			TemplateID: p.TemplateID,
		})
	}
	return svcPolicies, nil
}

func convertToServiceTemporaryPolicies(
	subjectPK int64, policies []types.Policy, actionMap map[string]int64,
) ([]svctypes.TemporaryPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "convertToServiceTemporaryPolicies")

	svcTemporaryPolicies := make([]svctypes.TemporaryPolicy, 0, len(policies))
	for _, p := range policies {
		actionPK, ok := actionMap[p.Action.ID]
		if !ok {
			err := errorWrapf(ErrActionNotExists, "actionID=`%s` fail", p.Action.ID)
			return nil, err
		}
		svcTemporaryPolicies = append(svcTemporaryPolicies, svctypes.TemporaryPolicy{
			SubjectPK:  subjectPK,
			ActionPK:   actionPK,
			Expression: p.Expression,
			ExpiredAt:  p.ExpiredAt,
		})
	}
	return svcTemporaryPolicies, nil
}

func (m *policyManager) querySubjectActionForAlterPolicies(
	system, subjectType, subjectID string,
) (subjectPK int64, actionPKMap map[string]int64, actionPKWithResourceTypeSet *set.Int64Set, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "querySubjectActionForAlterPolicies")

	// 1. 查询 subject subjectPK
	subjectPK, err = m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return
	}

	// 2. 查询操作列表
	actions, err := m.actionService.ListThinActionBySystem(system)
	if err != nil {
		err = errorWrapf(err, "actionService.ListThinActionBySystem system=`%s` fail", system)
		return
	}
	actionPKMap = make(map[string]int64, len(actions))
	for _, a := range actions {
		actionPKMap[a.ID] = a.PK
	}

	// 3. 查询关联了资源类型的操作pk set
	actionResourceTypes, err := m.actionService.ListActionResourceTypeIDByActionSystem(system)
	if err != nil {
		err = errorWrapf(err, "actionService.ListActionResourceTypeIDByActionSystem system=`%s` fail", system)
		return
	}
	actionPKWithResourceTypeSet = set.NewInt64Set()
	for _, t := range actionResourceTypes {
		actionPKWithResourceTypeSet.Add(actionPKMap[t.ActionID])
	}

	return subjectPK, actionPKMap, actionPKWithResourceTypeSet, nil
}

// DeleteByIDs 通过IDs批量删除策略
func (m *policyManager) DeleteByIDs(system string, subjectType, subjectID string, policyIDs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "DeletePoliciesByIDs")

	// 1. 查询 subject pk
	pk, err := m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return err
	}
	// 判断policyIDs是否为空，避免执行无效SQL
	if len(policyIDs) > 0 {
		// NOTE: delete cache here => 可以查actionPK
		defer policy.DeleteSystemSubjectPKsFromCache(system, []int64{pk})

		err := m.policyService.DeleteByPKs(pk, policyIDs)
		if err != nil {
			err = errorWrapf(err, "policyService.DeleteByPKs pk=`%d`, policyIDs=`%+v` fail",
				pk, policyIDs)
			return err
		}
	}
	return nil
}

// AlterCustomPolicies alter subject custom policies
func (m *policyManager) AlterCustomPolicies(
	system, subjectType, subjectID string,
	createPolicies, updatePolicies []types.Policy,
	deletePolicyIDs []int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "AlterPolicies")

	// 1. 查询subject action 相关的信息
	subjectPK, actionPKMap, actionPKWithResourceTypeSet, err := m.querySubjectActionForAlterPolicies(
		system, subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "m.querySubjectActionForAlterPolicies system=`%s` fail", system)
		return
	}

	// 2. 转换数据
	cps, err := convertToServicePolicies(subjectPK, createPolicies, actionPKMap)
	if err != nil {
		err = errorWrapf(err, "convertServicePolicies create policies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK, createPolicies, actionPKMap)
		return
	}
	ups, err := convertToServicePolicies(subjectPK, updatePolicies, actionPKMap)
	if err != nil {
		err = errorWrapf(err, "convertServicePolicies update policies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK, updatePolicies, actionPKMap)
		return
	}

	// NOTE: delete the policy cache before leave => 可以查actionPK
	defer policy.DeleteSystemSubjectPKsFromCache(system, []int64{subjectPK})

	// 3. service执行 create, update, delete
	updatedActionPKExpressionPKs, err := m.policyService.AlterCustomPolicies(
		subjectPK, cps, ups, deletePolicyIDs, actionPKWithResourceTypeSet)
	if err != nil {
		err = errorWrapf(err, "policyService.AlterPolicies system=`%s`, subjectPK=`%d` fail", system, subjectPK)
		return
	}

	defer expression.BatchDeleteExpressionsFromCache(updatedActionPKExpressionPKs)

	return nil
}

// CreateAndDeleteTemplatePolicies create and delete subject template policies
func (m *policyManager) CreateAndDeleteTemplatePolicies(
	system, subjectType, subjectID string, templateID int64,
	createPolicies []types.Policy, deletePolicyIDs []int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "CreateAndDeleteTemplatePolicies")

	// 1. 查询subject action 相关的信息
	subjectPK, actionPKMap, actionPKWithResourceTypeSet, err := m.querySubjectActionForAlterPolicies(
		system, subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "m.querySubjectActionForAlterPolicies system=`%s` fail", system)
		return
	}

	// 2. 转换数据
	cps, err := convertToServicePolicies(subjectPK, createPolicies, actionPKMap)
	if err != nil {
		err = errorWrapf(err, "convertServicePolicies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK, createPolicies, actionPKMap)
		return
	}

	// NOTE: delete the policy cache before leave
	defer policy.DeleteSystemSubjectPKsFromCache(system, []int64{subjectPK})

	// 3. service执行 create, delete
	err = m.policyService.CreateAndDeleteTemplatePolicies(
		subjectPK, templateID, cps, deletePolicyIDs, actionPKWithResourceTypeSet)
	if err != nil {
		err = errorWrapf(err, "policyService.CreateAndDeleteTemplatePolicies system=`%s`, subjectPK=`%d` fail",
			system, subjectPK)
		return
	}

	return nil
}

// UpdateTemplatePolicies update subject template policies
func (m *policyManager) UpdateTemplatePolicies(
	system, subjectType, subjectID string, policies []types.Policy,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "UpdateTemplatePolicies")

	// 1. 查询subject action 相关的信息
	subjectPK, actionMap, actionPKWithResourceTypeSet, err := m.querySubjectActionForAlterPolicies(
		system, subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "m.querySubjectActionForAlterPolicies system=`%s` fail", system)
		return
	}

	// 2. 类型转换
	ups, err := convertToServicePolicies(subjectPK, policies, actionMap)
	if err != nil {
		err = errorWrapf(err, "convertServicePolicies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK, policies, actionMap)
		return
	}

	// NOTE: delete the policy cache before leave => 可以查actionPK
	defer policy.DeleteSystemSubjectPKsFromCache(system, []int64{subjectPK})

	// 3. service执行 update
	err = m.policyService.UpdateTemplatePolicies(subjectPK, ups, actionPKWithResourceTypeSet)
	if err != nil {
		err = errorWrapf(err, "policyService.UpdateTemplatePolicies system=`%s`, subjectPK=`%d` fail", system, subjectPK)
		return
	}

	return nil
}

// DeleteTemplatePolicies delete subject template policies
func (m *policyManager) DeleteTemplatePolicies(system, subjectType, subjectID string, templateID int64) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "DeleteTemplatePolicies")

	// 1. 查询 subject subjectPK
	subjectPK, err := m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return
	}

	// NOTE: delete the policy cache before leave
	defer policy.DeleteSystemSubjectPKsFromCache(system, []int64{subjectPK})

	// 2. service执行 delete
	err = m.policyService.DeleteTemplatePolicies(subjectPK, templateID)
	if err != nil {
		err = errorWrapf(err, "policyService.DeleteTemplatePolicies subjectPK=`%d`, templateID=`%s` fail",
			subjectPK, templateID)
		return
	}

	return nil
}

// UpdateSubjectPoliciesExpiredAt 更新过期时间
func (m *policyManager) UpdateSubjectPoliciesExpiredAt(
	subjectType, subjectID string, policies []types.PolicyPKExpiredAt,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "RenewExpiredAtByIDs")

	// 1. 查询 subject pk
	subjectPK, err := m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return err
	}

	if len(policies) == 0 {
		return nil
	}

	pks := make([]int64, 0, len(policies))
	idExpiredAtMap := make(map[int64]int64, len(policies))
	for _, p := range policies {
		pks = append(pks, p.PK)
		idExpiredAtMap[p.PK] = p.ExpiredAt
	}

	ps, err := m.policyService.ListQueryByPKs(pks)
	if err != nil {
		err = errorWrapf(err, "policyService.ListQueryByPKs pks=`%+v` fail", pks)
		return err
	}

	updatePolicies := make([]svctypes.QueryPolicy, 0, len(ps))

	for _, p := range ps {
		if p.SubjectPK == subjectPK && (p.ExpiredAt < idExpiredAtMap[p.PK]) {
			p.ExpiredAt = idExpiredAtMap[p.PK]
			updatePolicies = append(updatePolicies, p)
		}
	}

	if len(updatePolicies) == 0 {
		return nil
	}

	// 按系统删除缓存
	systemSet, err := m.queryPoliciesSystemSet(updatePolicies)
	if err != nil {
		err = errorWrapf(err, "getSystemSetFromPolicyIDs pks=`%v` fail", pks)
		return err
	}

	// 清理缓存 => NOTE: 这里是可以知道actionPK的!!!!1
	defer policy.BatchDeleteSystemSubjectPKsFromCache(systemSet.ToSlice(), []int64{subjectPK})

	err = m.policyService.UpdateExpiredAt(updatePolicies)
	if err != nil {
		err = errorWrapf(err, "policyService.UpdateExpiredAt policies=`%+v` fail", ps)
		return err
	}

	return nil
}

func (m *policyManager) queryPoliciesSystemSet(policies []svctypes.QueryPolicy) (*set.StringSet, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "RenewExpiredAtByIDs")

	actionPKs := make([]int64, 0, len(policies))
	for _, p := range policies {
		actionPKs = append(actionPKs, p.ActionPK)
	}

	// 2. 查询所有的action
	actions, err := m.actionService.ListThinActionByPKs(actionPKs)
	if err != nil {
		err = errorWrapf(err, "actionService.ListThinActionByPKs actionPKs=`%+v` fail", actionPKs)
		return nil, err
	}

	// 3. 得到涉及的系统id
	systemSet := set.NewStringSet()
	for _, ac := range actions {
		systemSet.Add(ac.System)
	}

	return systemSet, nil
}

// DeleteByActionID 通过ActionID批量删除策略
func (m *policyManager) DeleteByActionID(system, actionID string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "`DeleteByActionID`")

	// 1. 查询 action pk
	actionPK, err := m.actionService.GetActionPK(system, actionID)
	if err != nil {
		// if action already deleted, just return
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		err = errorWrapf(err, "actionService.GetActionPK system=`%s`, actionID=`%s` fail", system, actionID)
		return err
	}

	err = m.policyService.DeleteByActionPK(actionPK)
	if err != nil {
		err = errorWrapf(err, "policyService.DeleteByActionPK actionPk=`%d`` fail", actionPK)
		return err
	}

	return nil
}

// CreateTemporaryPolicies create subject temporary policies
func (m *policyManager) CreateTemporaryPolicies(
	system, subjectType, subjectID string,
	policies []types.Policy,
) (pks []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "CreateTemporaryPolicies")

	// 1. 查询subject action 相关的信息
	subjectPK, actionPKMap, _, err := m.querySubjectActionForAlterPolicies(
		system, subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "m.querySubjectActionForAlterPolicies system=`%s` fail", system)
		return
	}

	// 2. 转换数据
	ps, err := convertToServiceTemporaryPolicies(subjectPK, policies, actionPKMap)
	if err != nil {
		err = errorWrapf(err, "convertToServiceTemporaryPolicies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK, policies, actionPKMap)
		return
	}

	// NOTE: delete the temporary policy cache before leave
	defer newTemporaryPolicyRedisCache(
		system, m.temporaryPolicyService,
	).DeleteBySubject(subjectPK)

	// 3. 执行创建
	pks, err = m.temporaryPolicyService.Create(ps)
	if err != nil {
		err = errorWrapf(err, "temporaryPolicyService.Create system=`%s`, subjectPK=`%d` fail",
			system, subjectPK)
	}
	return pks, err
}

// DeleteTemporaryByIDs 通过IDs批量删除临时策略
func (m *policyManager) DeleteTemporaryByIDs(system string, subjectType, subjectID string, policyIDs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "DeleteTemporaryByIDs")

	// 1. 查询 subject pk
	pk, err := m.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return err
	}
	// 判断policyIDs是否为空，避免执行无效SQL
	if len(policyIDs) > 0 {
		// NOTE: delete cache here
		defer newTemporaryPolicyRedisCache(
			system, m.temporaryPolicyService,
		).DeleteBySubject(pk)

		err := m.temporaryPolicyService.DeleteByPKs(pk, policyIDs)
		if err != nil {
			err = errorWrapf(err, "temporaryPolicyService.DeleteByPKs pk=`%d`, policyIDs=`%+v` fail",
				pk, policyIDs)
			return err
		}
	}
	return nil
}

// DeleteTemporaryBeforeExpiredAt 删除指定过期时间前的临时权限
func (m *policyManager) DeleteTemporaryBeforeExpiredAt(expiredAt int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "`DeleteTemporaryBeforeExpiredAt`")

	// NOTE: 这里删除一定要是已过期的策略, 不用清除缓存
	err := m.temporaryPolicyService.DeleteBeforeExpireAt(expiredAt)
	if err != nil {
		err = errorWrapf(err, "temporaryPolicyService.DeleteBeforeExpireAt expiredAt=`%d`` fail",
			expiredAt)
		return err
	}

	return nil
}
