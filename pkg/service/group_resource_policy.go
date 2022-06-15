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
	"database/sql"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
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

// preprocessChangedContent : 预处理分析出对于单个资源权限的变更，是创建、更新还是删除 DB记录
func (s *groupResourcePolicyService) preprocessChangedContent(
	groupPK, templateID int64,
	systemID string,
	rcc types.ResourceChangedContent,
) (
	createdPolicy *dao.GroupResourcePolicy,
	updatedActionPKs *dao.GroupResourcePolicyPKActionPKs,
	deletedPK int64,
	err error,
) {
	// 查询是否有记录
	pkActionPKs, err := s.manager.GetPKAndActionPKs(
		groupPK, templateID, systemID, rcc.ActionRelatedResourceTypePK, rcc.ResourceTypePK, rcc.ResourceID,
	)

	// 无则创建
	if errors.Is(err, sql.ErrNoRows) {
		// 不存在记录，其deleted action必然为空
		actionPKs, err := jsoniter.MarshalToString(rcc.CreatedActionPKs)
		if err != nil {
			return nil, nil, 0, fmt.Errorf(
				"jsoniter.MarshalToString for createdRecord fail actionPKs=`%v`, err: %w", rcc.CreatedActionPKs, err,
			)
		}
		createdPolicy = &dao.GroupResourcePolicy{
			GroupPK:                     groupPK,
			TemplateID:                  templateID,
			SystemID:                    systemID,
			ActionPKs:                   actionPKs,
			ActionRelatedResourceTypePK: rcc.ActionRelatedResourceTypePK,
			ResourceTypePK:              rcc.ResourceTypePK,
			ResourceID:                  rcc.ResourceID,
		}
		return createdPolicy, nil, 0, nil
	}

	// 对ActionPKs进行增删Action的变更
	var oldActionPKs []int64
	err = jsoniter.UnmarshalFromString(pkActionPKs.ActionPKs, &oldActionPKs)
	if err != nil {
		return nil, nil, 0, fmt.Errorf(
			"jsoniter.UnmarshalFromString for updateRecord fail actionPKs=`%s`, err: %w", pkActionPKs.ActionPKs, err,
		)
	}
	actionPKSet := set.NewInt64SetWithValues(oldActionPKs)
	// 添加需要新增的操作
	actionPKSet.Append(rcc.CreatedActionPKs...)
	// 移除将被删除的操作
	for _, actionPK := range rcc.DeletedActionPKs {
		if actionPKSet.Has(actionPK) {
			// TODO: 由于TencentBlueKing gopkg的int64Set不支持remove操作，待提PR发版后调整
			delete(actionPKSet.Data, actionPK)
		}
	}

	// 如果被删空了，则需要删除记录
	if actionPKSet.Size() == 0 {
		return nil, nil, pkActionPKs.PK, nil
	}

	actionPKs, err := jsoniter.MarshalToString(actionPKSet.ToSlice())
	if err != nil {
		return nil, nil, 0, fmt.Errorf(
			"jsoniter.MarshalToString for updatedRecord fail actionPKs=`%v`, err: %w",
			actionPKs,
			err,
		)
	}

	updatedActionPKs = &dao.GroupResourcePolicyPKActionPKs{
		PK:        pkActionPKs.PK,
		ActionPKs: actionPKs,
	}
	return nil, updatedActionPKs, 0, nil
}

func (s *groupResourcePolicyService) Alter(
	tx *sqlx.Tx, groupPK, templateID int64, systemID string, resourceChangedContents []types.ResourceChangedContent,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupResourcePolicySVC, "Alter")

	// Note: 底层这里没有保证并发同时修改的问题，而是顶层调用者SaaS通过分布式锁保证的
	// 1. 遍历每个资源其修改的Action，查询记录并变更ActionPKs
	// TODO: 看看是否有批量查询的优化空间
	createdGroupResourcePolicies := make([]dao.GroupResourcePolicy, 0, len(resourceChangedContents))
	updatedActionPKss := make([]dao.GroupResourcePolicyPKActionPKs, 0, len(resourceChangedContents))
	deletedPKs := make([]int64, 0, len(resourceChangedContents))
	for _, rcc := range resourceChangedContents {
		// 分析出该用户组的资源权限需要如何变更
		createdPolicy, updatedActionPKs, deletedPK, err := s.preprocessChangedContent(
			groupPK,
			templateID,
			systemID,
			rcc,
		)
		if err != nil {
			return errorWrapf(
				err,
				"preprocessChangedContent fail groupPK=`%d`, templateID=`%d`, systemID=`%s`, resourceChangedContent=`%v`",
				groupPK,
				templateID,
				systemID,
				rcc,
			)
		}
		// 不进行单独变更，集中批量变更
		if createdPolicy != nil {
			createdGroupResourcePolicies = append(createdGroupResourcePolicies, *createdPolicy)
		}
		if updatedActionPKs != nil {
			updatedActionPKss = append(updatedActionPKss, *updatedActionPKs)
		}
		if deletedPK != 0 {
			deletedPKs = append(deletedPKs, deletedPK)
		}
	}

	// 增
	if len(createdGroupResourcePolicies) > 0 {
		err := s.manager.BulkCreateWithTx(tx, createdGroupResourcePolicies)
		if err != nil {
			return errorWrapf(err, "manager.BulkCreateWithTx fail policies=`%v`", createdGroupResourcePolicies)
		}
	}
	// 删
	if len(deletedPKs) > 0 {
		err := s.manager.BulkDeleteByPKsWithTx(tx, deletedPKs)
		if err != nil {
			return errorWrapf(err, "manager.BulkDeleteByPKsWithTx fail pks=`%v`", deletedPKs)
		}
	}
	// 改
	if len(updatedActionPKss) > 0 {
		err := s.manager.BulkUpdateActionPKsWithTx(tx, updatedActionPKss)
		if err != nil {
			return errorWrapf(err, "manager.BulkUpdateActionPKsWithTx fail pkActionPKss=`%v`", updatedActionPKss)
		}
	}

	return nil
}
