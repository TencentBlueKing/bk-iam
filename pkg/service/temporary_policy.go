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
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// TemporaryPolicySVC ...
const TemporaryPolicySVC = "TemporaryPolicySVC"

// TemporaryPolicyService ...
type TemporaryPolicyService interface {
	// for auth
	ListThinBySubjectAction(subjectPK, actionPK int64) ([]types.ThinTemporaryPolicy, error)
	ListByPKs(pks []int64) ([]types.TemporaryPolicy, error)

	// for saas
	Create(policies []types.TemporaryPolicy) (pks []int64, err error)
	DeleteByPKs(subjectPK int64, pks []int64) error
	DeleteBeforeExpireAt(expiredAt int64) error
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

// Create subject temporary policies
func (s *temporaryPolicyService) Create(
	policies []types.TemporaryPolicy,
) (pks []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "Create")

	daoPolicies := make([]dao.TemporaryPolicy, 0, len(policies))
	for _, p := range policies {
		daoPolicies = append(daoPolicies, dao.TemporaryPolicy{
			SubjectPK:  p.SubjectPK,
			ActionPK:   p.ActionPK,
			Expression: p.Expression,
			ExpiredAt:  p.ExpiredAt,
		})
	}

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		err = errorWrapf(err, "define tx fail")
		return
	}

	pks, err = s.manager.BulkCreateWithTx(tx, daoPolicies)
	if err != nil {
		err = errorWrapf(err, "manager.BulkCreateWithTx policies=`%+v`", daoPolicies)
		return
	}

	err = tx.Commit()
	return pks, err
}

// DeleteByPKs ...
func (s *temporaryPolicyService) DeleteByPKs(subjectPK int64, pks []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "DeleteByPKs")

	tx, err := database.GenerateDefaultDBTx()
	if err != nil {
		return errorWrapf(err, "define tx fail")
	}
	defer database.RollBackWithLog(tx)

	_, err = s.manager.BulkDeleteByPKsWithTx(tx, subjectPK, pks)
	if err != nil {
		return errorWrapf(err, "manager.BulkDeleteByPKsWithTx subjectPK=`%d`, pks=`%+v`",
			subjectPK, pks)
	}

	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx.Commit fail")
	}
	return err
}

// DeleteBeforeExpireAt ...
func (s *temporaryPolicyService) DeleteBeforeExpireAt(expiredAt int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "DeleteBeforeExpireAt")
	tx, err := database.GenerateDefaultDBTx()
	if err != nil {
		return errorWrapf(err, "define tx fail")
	}
	defer database.RollBackWithLog(tx)

	// NOTE: 由于删除时可能数量较大，耗时长，锁行数据较多，影响鉴权，所以需要循环删除，限制每次删除的记录数，以及最多执行删除多少次
	// NOTE: 即使一次没删除完也没关系, 下一个周期的定时任务还是会继续删除
	rowLimit := int64(10000)
	maxAttempts := 10 // 相当于最多删除10万数据

	for i := 0; i < maxAttempts; i++ {
		rowsAffected, err1 := s.manager.BulkDeleteBeforeExpiredAtWithTx(tx, expiredAt, rowLimit)
		if err1 != nil {
			return errorWrapf(err1, "manager.BulkDeleteBeforeExpiredAtWithTx expiredAt=`%d`", expiredAt)
		}
		// 如果已经没有需要删除的了，就停止
		if rowsAffected < rowLimit {
			break
		}
	}

	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx.Commit fail")
	}
	return err
}
