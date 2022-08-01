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

import (
	"database/sql"
	"errors"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/prp/expression"
	"iam/pkg/abac/prp/policy"
	"iam/pkg/abac/prp/temporary"
	"iam/pkg/abac/types"
	svctypes "iam/pkg/service/types"
)

// NOTE: **important** / **重要**
//       curd中所有方法必须考虑删除policy缓存
//       curd中所有方法必须考虑删除policy缓存
//       curd中所有方法必须考虑删除policy缓存

var ErrActionNotExists = errors.New("action not exists")

func convertToServicePolicies(
	subjectPK int64, policies []types.Policy, actionMap map[string]int64,
) ([]svctypes.Policy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "convertServicePolicies")

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
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "convertToServiceTemporaryPolicies")

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

func (c *policyController) querySubjectActionForAlterPolicies(
	system, subjectType, subjectID string,
) (subjectPK int64, actionPKMap map[string]int64, actionPKWithResourceTypeSet *set.Int64Set, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "querySubjectActionForAlterPolicies")

	// 1. 查询 subject subjectPK
	subjectPK, err = c.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return
	}

	// 2. 查询操作列表
	actions, err := c.actionService.ListThinActionBySystem(system)
	if err != nil {
		err = errorWrapf(err, "actionService.ListThinActionBySystem system=`%s` fail", system)
		return
	}
	actionPKMap = make(map[string]int64, len(actions))
	for _, a := range actions {
		actionPKMap[a.ID] = a.PK
	}

	// 3. 查询关联了资源类型的操作pk set
	actionResourceTypes, err := c.actionService.ListActionResourceTypeIDByActionSystem(system)
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
func (c *policyController) DeleteByIDs(system string, subjectType, subjectID string, policyIDs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "DeletePoliciesByIDs")

	// 1. 查询 subject pk
	pk, err := c.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return err
	}
	// 判断policyIDs是否为空，避免执行无效SQL
	if len(policyIDs) > 0 {
		// NOTE: delete cache here => 可以查actionPK
		defer policy.DeleteSystemSubjectPKsFromCache(system, []int64{pk})

		err := c.policyService.DeleteByPKs(pk, policyIDs)
		if err != nil {
			err = errorWrapf(err, "policyService.DeleteByPKs pk=`%d`, policyIDs=`%+v` fail",
				pk, policyIDs)
			return err
		}
	}
	return nil
}

// AlterCustomPolicies alter subject custom policies
func (c *policyController) AlterCustomPolicies(
	system, subjectType, subjectID string,
	createPolicies, updatePolicies []types.Policy,
	deletePolicyIDs []int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "AlterPolicies")

	// 1. 查询subject action 相关的信息
	subjectPK, actionPKMap, actionPKWithResourceTypeSet, err := c.querySubjectActionForAlterPolicies(
		system, subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "c.querySubjectActionForAlterPolicies system=`%s` fail", system)
		return
	}

	// 2. 转换数据
	cps, err := convertToServicePolicies(subjectPK, createPolicies, actionPKMap)
	if err != nil {
		err = errorWrapf(
			err,
			"convertServicePolicies create policies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK,
			createPolicies,
			actionPKMap,
		)
		return
	}
	ups, err := convertToServicePolicies(subjectPK, updatePolicies, actionPKMap)
	if err != nil {
		err = errorWrapf(
			err,
			"convertServicePolicies update policies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK,
			updatePolicies,
			actionPKMap,
		)
		return
	}

	// NOTE: delete the policy cache before leave => 可以查actionPK
	defer policy.DeleteSystemSubjectPKsFromCache(system, []int64{subjectPK})

	// 3. service执行 create, update, delete
	updatedActionPKExpressionPKs, err := c.policyService.AlterCustomPolicies(
		subjectPK, cps, ups, deletePolicyIDs, actionPKWithResourceTypeSet)
	if err != nil {
		err = errorWrapf(err, "policyService.AlterPolicies system=`%s`, subjectPK=`%d` fail", system, subjectPK)
		return
	}

	defer expression.BatchDeleteExpressionsFromCache(updatedActionPKExpressionPKs)

	return nil
}

// DeleteByActionID 通过ActionID批量删除策略
func (c *policyController) DeleteByActionID(system, actionID string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "`DeleteByActionID`")

	// 1. 查询 action pk
	actionPK, err := c.actionService.GetActionPK(system, actionID)
	if err != nil {
		// if action already deleted, just return
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		err = errorWrapf(err, "actionService.GetActionPK system=`%s`, actionID=`%s` fail", system, actionID)
		return err
	}

	err = c.policyService.DeleteByActionPK(actionPK)
	if err != nil {
		err = errorWrapf(err, "policyService.DeleteByActionPK actionPk=`%d`` fail", actionPK)
		return err
	}

	// TODO: 删除Resource Group Policy

	return nil
}

// CreateTemporaryPolicies create subject temporary policies
func (c *policyController) CreateTemporaryPolicies(
	system, subjectType, subjectID string,
	policies []types.Policy,
) (pks []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "CreateTemporaryPolicies")

	// 1. 查询subject action 相关的信息
	subjectPK, actionPKMap, _, err := c.querySubjectActionForAlterPolicies(
		system, subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "c.querySubjectActionForAlterPolicies system=`%s` fail", system)
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
	defer temporary.DeleteTemporaryPolicyBySystemSubjectFromCache(system, subjectPK)

	// 3. 执行创建
	pks, err = c.temporaryPolicyService.Create(ps)
	if err != nil {
		err = errorWrapf(err, "temporaryPolicyService.Create system=`%s`, subjectPK=`%d` fail",
			system, subjectPK)
	}
	return pks, err
}

// DeleteTemporaryByIDs 通过IDs批量删除临时策略
func (c *policyController) DeleteTemporaryByIDs(system string, subjectType, subjectID string, policyIDs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "DeleteTemporaryByIDs")

	// 1. 查询 subject pk
	pk, err := c.subjectService.GetPK(subjectType, subjectID)
	if err != nil {
		err = errorWrapf(err, "subjectService.GetPK subjectType=`%s`, subjectID=`%s` fail",
			subjectType, subjectID)
		return err
	}
	// 判断policyIDs是否为空，避免执行无效SQL
	if len(policyIDs) > 0 {
		// NOTE: delete cache here
		defer temporary.DeleteTemporaryPolicyBySystemSubjectFromCache(system, pk)

		err := c.temporaryPolicyService.DeleteByPKs(pk, policyIDs)
		if err != nil {
			err = errorWrapf(err, "temporaryPolicyService.DeleteByPKs pk=`%d`, policyIDs=`%+v` fail",
				pk, policyIDs)
			return err
		}
	}
	return nil
}

// DeleteTemporaryBeforeExpiredAt 删除指定过期时间前的临时权限
func (c *policyController) DeleteTemporaryBeforeExpiredAt(expiredAt int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTL, "`DeleteTemporaryBeforeExpiredAt`")

	// NOTE: 这里删除一定要是已过期的策略, 不用清除缓存
	err := c.temporaryPolicyService.DeleteBeforeExpiredAt(expiredAt)
	if err != nil {
		err = errorWrapf(err, "temporaryPolicyService.DeleteBeforeExpiredAt expiredAt=`%d`` fail",
			expiredAt)
		return err
	}

	return nil
}
