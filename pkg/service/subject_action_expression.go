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
	CreateOrUpdateWithTx(tx *sqlx.Tx, expression types.SubjectActionExpression) error
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
			ExpiredAt:  expression.ExpiredAt,
		}

		err = s.manager.CreateWithTx(tx, daoExpression)
		if err != nil {
			err = errorWrapf(err, "manager.CreateWithTx fail, daoExpression=`%+v`", daoExpression)
			return err
		}
	} else {
		// update
		daoExpression.Expression = expression.Expression
		daoExpression.ExpiredAt = expression.ExpiredAt
		err = s.manager.UpdateWithTx(tx, daoExpression)
		if err != nil {
			err = errorWrapf(err, "manager.UpdateWithTx fail, daoExpression=`%+v`", daoExpression)
			return err
		}
	}

	return nil
}
