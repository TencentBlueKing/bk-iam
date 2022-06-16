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
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

const GroupResourcePolicySVC = "GroupResourcePolicySVC"

type GroupResourcePolicyService interface {
	Alter(
		tx *sqlx.Tx, groupPK, templateID int64, systemID string, resourceChangedContents []types.ResourceChangedContent,
	) error
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
				"jsoniter.UnmarshalFromString fail actionPKs=`%s`, err: %w", oldActionPKs, err,
			)
		}
	}

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
		return "", fmt.Errorf("jsoniter.MarshalToString fail actionPKs=`%v`, err: %w", actionPKs, err)
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
