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
	"database/sql"
	"errors"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// SubjectActionExpressionSVC ...
const SubjectActionExpressionSVC = "SubjectActionExpressionSVC"

// SubjectActionExpressionService ...
type SubjectActionExpressionService interface {
	ListBySubjectAction(subjectPKs []int64, actionPK int64) ([]types.SubjectActionExpression, error)
	CreateOrUpdateWithTx(tx *sqlx.Tx, expression types.SubjectActionExpression) error
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error
}

type subjectActionExpressionService struct {
	manager dao.SubjectActionExpressionManager
}

// NewSubjectActionExpressionService ...
func NewSubjectActionExpressionService() SubjectActionExpressionService {
	return &subjectActionExpressionService{
		manager: dao.NewSubjectActionExpressionManager(),
	}
}

/*
	NOTE: SubjectActionExpression不会被主动删除, 存在expiredAt==0的数据, 实际没有作用

	原因: 保持subjectActionGroupResource与subjectActionExpression数据的一致性
	前提: 查询时使用 subject_pk 与 department_pks 的总数量不会太多, 并且每个action_pk只有一条数据
*/

// CreateOrUpdateWithTx ...
func (s *subjectActionExpressionService) CreateOrUpdateWithTx(
	tx *sqlx.Tx,
	expression types.SubjectActionExpression,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionExpressionSVC, "CreateOrUpdateWithTx")

	daoExpression, err := s.manager.GetBySubjectAction(expression.SubjectPK, expression.ActionPK)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		err = errorWrapf(err,
			"manager.GetBySubjectAction fail, subjectPK=`%d`, actionPK=`%d`",
			expression.SubjectPK, expression.ActionPK,
		)

		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		// create
		daoExpression = dao.SubjectActionExpression{
			SubjectPK:  expression.SubjectPK,
			ActionPK:   expression.ActionPK,
			Expression: expression.Expression,
			Signature:  expression.Signature,
			ExpiredAt:  expression.ExpiredAt,
		}

		err = s.manager.CreateWithTx(tx, daoExpression)
		if err != nil {
			err = errorWrapf(err, "manager.CreateWithTx fail, daoExpression=`%+v`", daoExpression)
			return err
		}
	} else {
		// update
		err = s.manager.UpdateExpressionExpiredAtWithTx(
			tx,
			daoExpression.PK,
			expression.Expression,
			expression.Signature,
			expression.ExpiredAt,
		)
		if err != nil {
			err = errorWrapf(err, "manager.UpdateWithTx fail, daoExpression=`%+v`", daoExpression)
			return err
		}
	}

	return nil
}

// ListBySubjectAction ...
func (s *subjectActionExpressionService) ListBySubjectAction(
	subjectPKs []int64,
	actionPK int64,
) ([]types.SubjectActionExpression, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionExpressionSVC, "ListAuthBySubjectAction")
	daoExpressions, err := s.manager.ListBySubjectAction(subjectPKs, actionPK)
	if err != nil {
		return nil, errorWrapf(
			err, "manager.ListBySubjectAction subjectPKs=`%+v`, actionPK=`%d`",
			subjectPKs, actionPK,
		)
	}

	expressions := make([]types.SubjectActionExpression, 0, len(daoExpressions))
	for _, e := range daoExpressions {
		// NOTE: 这里可能有已经过期的数据, 由上层处理更新事件
		expressions = append(expressions, types.SubjectActionExpression{
			PK:         e.PK,
			SubjectPK:  e.SubjectPK,
			ActionPK:   e.ActionPK,
			Expression: e.Expression,
			Signature:  e.Signature,
			ExpiredAt:  e.ExpiredAt,
		})
	}
	return expressions, nil
}

// BulkDeleteBySubjectPKsWithTx ...
func (s *subjectActionExpressionService) BulkDeleteBySubjectPKsWithTx(
	tx *sqlx.Tx,
	subjectPKs []int64,
) error {
	return s.manager.BulkDeleteBySubjectPKsWithTx(tx, subjectPKs)
}
