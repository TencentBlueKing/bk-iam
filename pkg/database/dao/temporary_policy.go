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

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// TemporaryPolicy ...
type TemporaryPolicy struct {
	PK int64 `db:"pk"`

	SubjectPK  int64  `db:"subject_pk"` // 关联Subject表自增列
	ActionPK   int64  `db:"action_pk"`  // 关联Action表自增列
	Expression string `db:"expression"`

	// 策略有效期，unix time，单位秒(s)
	ExpiredAt int64 `db:"expired_at"`
}

// ThinTemporaryPolicy ...
type ThinTemporaryPolicy struct {
	PK        int64 `db:"pk"`
	ExpiredAt int64 `db:"expired_at"`
}

// TemporaryPolicyManager ...
type TemporaryPolicyManager interface {
	// for auth
	ListThinBySubjectAction(subjectPK, actionPK, expiredAt int64) ([]ThinTemporaryPolicy, error)
	ListByPKs(pks []int64) ([]TemporaryPolicy, error)

	// for saas
	BulkCreateWithTx(tx *sqlx.Tx, policies []TemporaryPolicy) ([]int64, error)
	BulkDeleteByPKsWithTx(tx *sqlx.Tx, subjectPK int64, pks []int64) (int64, error)
	BulkDeleteBeforeExpiredAtWithTx(tx *sqlx.Tx, expiredAt, limit int64) (int64, error)
}

type temporaryPolicyManager struct {
	DB *sqlx.DB
}

// NewTemporaryPolicyManager ...
func NewTemporaryPolicyManager() TemporaryPolicyManager {
	return &temporaryPolicyManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListThinBySubjectAction ...
func (m *temporaryPolicyManager) ListThinBySubjectAction(
	subjectPK, actionPK, expiredAt int64) (policies []ThinTemporaryPolicy, err error) {
	err = m.selectThinBySubjectAction(&policies, subjectPK, actionPK, expiredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// ListByPKs ...
func (m *temporaryPolicyManager) ListByPKs(pks []int64) (policies []TemporaryPolicy, err error) {
	err = m.selectByPKs(&policies, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *temporaryPolicyManager) BulkCreateWithTx(tx *sqlx.Tx, policies []TemporaryPolicy) ([]int64, error) {
	if len(policies) == 0 {
		return []int64{}, nil
	}
	return m.bulkInsertWithTx(tx, policies)
}

// BulkDeleteByPKsWithTx ...
func (m *temporaryPolicyManager) BulkDeleteByPKsWithTx(
	tx *sqlx.Tx, subjectPK int64, pks []int64,
) (int64, error) {
	if len(pks) == 0 {
		return 0, nil
	}
	return m.bulkDeleteByPKsWithTx(tx, subjectPK, pks)
}

// BulkDeleteBeforeExpiredAtWithTx ...
func (m *temporaryPolicyManager) BulkDeleteBeforeExpiredAtWithTx(
	tx *sqlx.Tx, expiredAt, limit int64,
) (int64, error) {
	return m.bulkDeleteBeforeExpiredAtWithTx(tx, expiredAt, limit)
}

func (m *temporaryPolicyManager) selectByPKs(policies *[]TemporaryPolicy, pks []int64) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression,
		expired_at
		FROM temporary_policy
		WHERE pk IN (?)`
	return database.SqlxSelect(m.DB, policies, query, pks)
}

func (m *temporaryPolicyManager) selectThinBySubjectAction(
	policies *[]ThinTemporaryPolicy, subjectPK, actionPK, expiredAt int64) error {
	query := `SELECT
		pk,
		expired_at
		FROM temporary_policy
		WHERE subject_pk = ?
		AND action_pk = ?
		AND expired_at >= ?`
	return database.SqlxSelect(m.DB, policies, query, subjectPK, actionPK, expiredAt)
}

func (m *temporaryPolicyManager) bulkInsertWithTx(tx *sqlx.Tx, policies []TemporaryPolicy) ([]int64, error) {
	sql := `INSERT INTO temporary_policy (
		subject_pk,
		action_pk,
		expression,
		expired_at,
	) VALUES (
		:subject_pk,
		:action_pk,
		:expression,
		:expired_at)`
	return database.SqlxBulkInsertReturnIDWithTx(tx, sql, policies)
}

func (m *temporaryPolicyManager) bulkDeleteByPKsWithTx(
	tx *sqlx.Tx, subjectPK int64, pks []int64,
) (int64, error) {
	sql := `DELETE FROM temporary_policy WHERE subject_pk = ? AND pk IN (?)`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, subjectPK, pks)
}

func (m *temporaryPolicyManager) bulkDeleteBeforeExpiredAtWithTx(
	tx *sqlx.Tx, expiredAt, limit int64,
) (int64, error) {
	sql := `DELETE FROM temporary_policy WHERE expired_at < ? LIMIT ?`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, expiredAt, limit)
}
