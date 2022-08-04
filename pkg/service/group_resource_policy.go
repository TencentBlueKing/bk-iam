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
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

const GroupResourcePolicySVC = "GroupResourcePolicySVC"

type GroupResourcePolicyService interface {
	Alter(
		tx *sqlx.Tx, groupPK, templateID int64, systemID string, resourceChangedContents []types.ResourceChangedContent,
	) error

	GetAuthorizedActionGroupMap(
		systemID string,
		actionResourceTypePK, resourceTypePK int64,
		resourceTypeID string,
	) (map[int64][]int64, error)
	ListResourceByGroupAction(
		groupPK int64,
		systemID string,
		actionPK, actionResourceTypePK int64,
	) ([]types.Resource, error)
	BulkDeleteByGroupPKsWithTx(tx *sqlx.Tx, groupPKs []int64) error

	DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK int64) error
}

type groupResourcePolicyService struct {
	manager dao.GroupResourcePolicyManager
}

func NewGroupResourcePolicyService() GroupResourcePolicyService {
	return &groupResourcePolicyService{
		manager: dao.NewGroupResourcePolicyManager(),
	}
}

// calculateSignature : 用于计算出group resource policy的唯一索引Signature
func (s *groupResourcePolicyService) calculateSignature(
	groupPK, templateID int64,
	systemID string,
	rcc types.ResourceChangedContent,
) string {
	// Note: 由于字符串类型的system_id并不会出现冒号分隔符，所以
	//  前缀groupPK, templateID, systemID, rcc.ActionRelatedResourceTypePK, rcc.ResourceTypePK可保证唯一性
	//  即使resource_id包含了冒号分隔符，由于前缀已保证了唯一性，所以最终整个字符串也是唯一的
	signature := fmt.Sprintf(
		"%d:%d:%s:%d:%d:%s",
		groupPK, templateID, systemID, rcc.ActionRelatedResourceTypePK, rcc.ResourceTypePK, rcc.ResourceID,
	)
	return stringx.MD5Hash(signature)
}

// calculateChangedActionPKs : 使用旧的ActionPKs和要变更的内容，计算出最终变更的ActionPKs
func (s *groupResourcePolicyService) calculateChangedActionPKs(
	oldActionPKs string, rcc types.ResourceChangedContent,
) (string, error) {
	// 将ActionPKs从Json字符串转为列表格式
	var oldActionPKList []int64
	if len(oldActionPKs) > 0 {
		err := jsoniter.UnmarshalFromString(oldActionPKs, &oldActionPKList)
		if err != nil {
			return "", fmt.Errorf(
				"jsoniter.UnmarshalFromString actionPKs=`%s` fail, err: %w", oldActionPKs, err,
			)
		}
	}

	// TODO: 判断现有的oldActionPKList是否存在, 剔除不存在的action_pk

	// 使用set对新增和删除的Action进行变更
	actionPKSet := set.NewInt64SetWithValues(oldActionPKList)
	// 添加需要新增的操作
	actionPKSet.Append(rcc.CreatedActionPKs...)
	// 移除将被删除的操作
	for _, actionPK := range rcc.DeletedActionPKs {
		if actionPKSet.Has(actionPK) {
			// TODO: 由于TencentBlueKing gopkg的int64Set不支持remove操作，待提PR发版后调整
			delete(actionPKSet.Data, actionPK)
		}
	}

	// 如果被删空了，则直接返回，不需要再进行json dumps
	if actionPKSet.Size() == 0 {
		return "", nil
	}

	actionPKs, err := jsoniter.MarshalToString(actionPKSet.ToSlice())
	if err != nil {
		return "", fmt.Errorf("jsoniter.MarshalToString actionPKs=`%v` fail, err: %w", actionPKs, err)
	}

	return actionPKs, nil
}

func (s *groupResourcePolicyService) Alter(
	tx *sqlx.Tx, groupPK, templateID int64, systemID string, resourceChangedContents []types.ResourceChangedContent,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupResourcePolicySVC, "Alter")

	// Note: 底层这里没有保证并发同时修改的问题，而是顶层调用者SaaS通过分布式锁保证的
	// 1. 计算每条策略的signature=md5(group_pk:template_id:system_id:action_related_resource_type_pk:resource_type_pk:resource_id)
	signatures := make([]string, 0, len(resourceChangedContents))
	for _, rcc := range resourceChangedContents {
		signature := s.calculateSignature(groupPK, templateID, systemID, rcc)
		signatures = append(signatures, signature)
	}

	// 2. 查询每条策略
	policies, err := s.manager.ListBySignatures(signatures)
	if err != nil {
		return errorWrapf(
			err,
			"manager.ListBySignatures fail, groupPK=`%d`, templateID=`%d`, systemID=`%s`, resourceChangedContents=`%v`",
			groupPK, templateID, systemID, resourceChangedContents,
		)
	}
	signatureToPolicyMap := make(map[string]dao.GroupResourcePolicy, len(policies))
	for _, p := range policies {
		signatureToPolicyMap[p.Signature] = p
	}

	// 3. 遍历策略，根据要变更的内容，分析计算出要创建、更新、删除的策略
	createdPolicies := make([]dao.GroupResourcePolicy, 0, len(resourceChangedContents))
	updatedPolicies := make([]dao.GroupResourcePolicy, 0, len(resourceChangedContents))
	deletedPolicyPKs := make([]int64, 0, len(resourceChangedContents))
	for _, rcc := range resourceChangedContents {
		signature := s.calculateSignature(groupPK, templateID, systemID, rcc)
		policy, found := signatureToPolicyMap[signature]

		// 根据变更内容，计算出变更后的ActionPKs Json字符串
		actionPKs, err := s.calculateChangedActionPKs(policy.ActionPKs, rcc)
		if err != nil {
			return errorWrapf(
				err,
				"calculateChangedActionPKs fail, signature=`%s`, groupPK=`%d`, templateID=`%d`,"+
					" systemID=`%s`, resourceChangedContent=`%v`",
				signature,
				groupPK,
				templateID,
				systemID,
				rcc,
			)
		}

		// 3.1 找不到，则需要新增记录
		if !found {
			createdPolicies = append(createdPolicies, dao.GroupResourcePolicy{
				Signature:                   signature,
				GroupPK:                     groupPK,
				TemplateID:                  templateID,
				SystemID:                    systemID,
				ActionPKs:                   actionPKs,
				ActionRelatedResourceTypePK: rcc.ActionRelatedResourceTypePK,
				ResourceTypePK:              rcc.ResourceTypePK,
				ResourceID:                  rcc.ResourceID,
			})
			continue
		}

		// 3.2 若变更后的ActionPKs为空，则删除整条策略
		if actionPKs == "" {
			deletedPolicyPKs = append(deletedPolicyPKs, policy.PK)
			continue
		}

		// 3.3 只更新ActionPKs
		policy.ActionPKs = actionPKs
		updatedPolicies = append(updatedPolicies, policy)
	}

	// 4. 变更
	// 增
	if len(createdPolicies) > 0 {
		err := s.manager.BulkCreateWithTx(tx, createdPolicies)
		if err != nil {
			return errorWrapf(err, "manager.BulkCreateWithTx fail policies=`%v`", createdPolicies)
		}
	}
	// 删
	if len(deletedPolicyPKs) > 0 {
		err := s.manager.BulkDeleteByPKsWithTx(tx, deletedPolicyPKs)
		if err != nil {
			return errorWrapf(err, "manager.BulkDeleteByPKsWithTx fail pks=`%v`", deletedPolicyPKs)
		}
	}
	// 改
	if len(updatedPolicies) > 0 {
		err := s.manager.BulkUpdateActionPKsWithTx(tx, updatedPolicies)
		if err != nil {
			return errorWrapf(err, "manager.BulkUpdateActionPKsWithTx fail policies=`%v`", updatedPolicies)
		}
	}

	return nil
}

// GetAuthorizedActionGroupMap 查询有权限的用户组的操作
func (s *groupResourcePolicyService) GetAuthorizedActionGroupMap(
	systemID string,
	actionResourceTypePK, resourceTypePK int64,
	resourceID string,
) (map[int64][]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupResourcePolicySVC, "GetAuthorizedActionGroupMap")

	daoGroupResourcePolicies, err := s.manager.ListThinByResource(
		systemID,
		actionResourceTypePK,
		resourceTypePK,
		resourceID,
	)
	if err != nil {
		return nil, errorWrapf(
			err,
			"manager.ListThinByResource fail, systemID=`%s`, actionResourceTypePK=`%d`, resourceTypePK=`%d`, resourceID=`%s`",
			systemID,
			actionResourceTypePK,
			resourceTypePK,
			resourceID,
		)
	}

	actionGroupPKsSets := make(map[int64]*set.Int64Set, 5)
	for _, daoGroupResourcePolicy := range daoGroupResourcePolicies {
		var actionPKs []int64
		if err := jsoniter.UnmarshalFromString(daoGroupResourcePolicy.ActionPKs, &actionPKs); err != nil {
			return nil, errorWrapf(
				err,
				"jsoniter.UnmarshalFromString fail, actionPKs=`%s`",
				daoGroupResourcePolicy.ActionPKs,
			)
		}

		for _, actionPK := range actionPKs {
			if _, found := actionGroupPKsSets[actionPK]; !found {
				actionGroupPKsSets[actionPK] = set.NewInt64Set()
			}

			actionGroupPKsSets[actionPK].Add(daoGroupResourcePolicy.GroupPK)
		}
	}

	actionGroupPKs := make(map[int64][]int64, len(actionGroupPKsSets))
	for actionPK, actionGroupPKsSet := range actionGroupPKsSets {
		actionGroupPKs[actionPK] = actionGroupPKsSet.ToSlice()
	}

	return actionGroupPKs, nil
}

func (s *groupResourcePolicyService) ListResourceByGroupAction(
	groupPK int64,
	systemID string,
	actionPK, actionResourceTypePK int64,
) ([]types.Resource, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupResourcePolicySVC, "ListResourceByGroupAction")

	policies, err := s.manager.ListByGroupSystemActionRelatedResourceType(
		groupPK, systemID, actionResourceTypePK,
	)
	if err != nil {
		return nil, errorWrapf(
			err,
			"manager.ListByGroupSystemActionRelatedResourceType fail,"+
				" groupPK=`%d`, systemID=`%s`, actionRelatedResourceTypePK=`%d`",
			groupPK,
			systemID,
			actionResourceTypePK,
		)
	}

	resources := make([]types.Resource, 0, len(policies)/2)
	for _, policy := range policies {
		var actionPKs []int64
		if err := jsoniter.UnmarshalFromString(policy.ActionPKs, &actionPKs); err != nil {
			return nil, errorWrapf(
				err,
				"jsoniter.UnmarshalFromString fail, pk=`%d` actionPKs=`%s`",
				policy.PK, policy.ActionPKs,
			)
		}

		actionSet := set.NewInt64SetWithValues(actionPKs)
		if !actionSet.Has(actionPK) {
			continue
		}

		resources = append(resources, types.Resource{
			ResourceTypePK: policy.ResourceTypePK,
			ResourceID:     policy.ResourceID,
		})
	}

	return resources, nil
}

// BulkDeleteByGroupPKsWithTx ...
func (s *groupResourcePolicyService) BulkDeleteByGroupPKsWithTx(
	tx *sqlx.Tx,
	groupPKs []int64,
) error {
	return s.manager.BulkDeleteByGroupPKsWithTx(tx, groupPKs)
}

// DeleteByActionPK ...
// NOTE: 这里只删除action_pks中只有一个action_pk的记录, 变更的时候再检查对应的记录是否需要删除存在的action_pk
func (s *groupResourcePolicyService) DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "DeleteByActionPK")

	actionPKs, err := jsoniter.MarshalToString([]int64{actionPK})
	if err != nil {
		return errorWrapf(err, "jsoniter.MarshalToString fail, actionPK=`%d`", actionPK)
	}

	// 由于删除时可能数量较大，耗时长，锁行数据较多，影响鉴权，所以需要循环删除，限制每次删除的记录数，以及最多执行删除多少次
	rowLimit := int64(10000)
	maxAttempts := 100 // 相当于最多删除100万数据

	for i := 0; i < maxAttempts; i++ {
		rowsAffected, err := s.manager.DeleteByActionPKsWithTx(tx, actionPKs, rowLimit)
		if err != nil {
			return errorWrapf(err, "manager.DeleteByActionPKWithTx actionPK=`%d`", actionPK)
		}
		// 如果已经没有需要删除的了，就停止
		if rowsAffected == 0 {
			break
		}
	}

	return nil
}
