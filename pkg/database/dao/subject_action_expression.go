/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dao

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// SubjectActionExpression ...
type SubjectActionExpression struct {
	PK int64 `db:"pk"`

	SubjectPK  int64  `db:"subject_pk"` // 关联Subject表自增列
	ActionPK   int64  `db:"action_pk"`  // 关联Action表自增列
	Expression string `db:"expression"`

	// 策略有效期，unix time，单位秒(s)
	ExpiredAt int64 `db:"expired_at"`
}

// SubjectActionExpression ...
type SubjectActionExpressionManager interface {
	ListBySubjectAction(subjectPKs []int64, actionPK int64) ([]SubjectActionExpression, error)

	GetBySubjectAction(subjectPK, actionPK int64) (SubjectActionExpression, error)
	CreateWithTx(tx *sqlx.Tx, subjectActionExpression SubjectActionExpression) error
	UpdateExpressionExpiredAtWithTx(tx *sqlx.Tx, subjectActionExpression SubjectActionExpression) error
}

type subjectActionExpressionManager struct {
	DB *sqlx.DB
}

// NewSubjectActionExpressionManager ...
func NewSubjectActionExpressionManager() SubjectActionExpressionManager {
	return &subjectActionExpressionManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListBySubjectAction ...
func (m *subjectActionExpressionManager) ListBySubjectAction(
	subjectPKs []int64,
	actionPK int64,
) (subjectActionExpressions []SubjectActionExpression, err error) {
	if len(subjectPKs) == 0 {
		return
	}

	query := `SELECT 
		pk,
		subject_pk,
		action_pk,
		expression,
		expired_at
		FROM rbac_subject_action_expression
		WHERE subject_pk IN (?)
		AND action_pk = ?`
	err = database.SqlxSelect(m.DB, &subjectActionExpressions, query, subjectPKs, actionPK)
	if errors.Is(err, sql.ErrNoRows) {
		return subjectActionExpressions, nil
	}

	return
}

// GetBySubjectAction ...
func (m *subjectActionExpressionManager) GetBySubjectAction(
	subjectPK, actionPK int64,
) (subjectActionExpression SubjectActionExpression, err error) {
	query := `SELECT 
		pk,
		subject_pk,
		action_pk,
		expression,
		expired_at
		FROM rbac_subject_action_expression
		WHERE subject_pk = ?
		AND action_pk = ? LIMIT 1`
	err = database.SqlxGet(m.DB, &subjectActionExpression, query, subjectPK, actionPK)
	return
}

// CreateWithTx ...
func (m *subjectActionExpressionManager) CreateWithTx(
	tx *sqlx.Tx,
	subjectActionExpression SubjectActionExpression,
) error {
	sql := `INSERT INTO rbac_subject_action_expression (
		subject_pk,
		action_pk,
		expression,
		expired_at
	) VALUES (
		:subject_pk,
		:action_pk,
		:expression,
		:expired_at
	)`
	return database.SqlxInsertWithTx(tx, sql, subjectActionExpression)
}

// UpdateExpressionExpiredAtWithTx ...
func (m *subjectActionExpressionManager) UpdateExpressionExpiredAtWithTx(
	tx *sqlx.Tx,
	subjectActionExpression SubjectActionExpression,
) error {
	sql := `UPDATE rbac_subject_action_expression SET expression = :expression, expired_at = :expired_at WHERE pk = :pk`
	_, err := database.SqlxUpdateWithTx(tx, sql, subjectActionExpression)
	return err
}
