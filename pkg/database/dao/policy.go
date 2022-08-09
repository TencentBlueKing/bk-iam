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

// AuthPolicy ...
type AuthPolicy struct {
	PK           int64 `db:"pk"`
	SubjectPK    int64 `db:"subject_pk"`
	ExpressionPK int64 `db:"expression_pk"`
	ExpiredAt    int64 `db:"expired_at"`
}

// Policy ...
type Policy struct {
	PK int64 `db:"pk"`

	SubjectPK    int64 `db:"subject_pk"` // 关联Subject表自增列
	ActionPK     int64 `db:"action_pk"`  // 关联Action表自增列
	ExpressionPK int64 `db:"expression_pk"`

	// 策略有效期，unix time，单位秒(s)
	ExpiredAt  int64 `db:"expired_at"`
	TemplateID int64 `db:"template_id"`
}

// PolicyManager ...
type PolicyManager interface {
	// for auth

	ListAuthBySubjectAction(subjectPKs []int64, actionPK int64, expiredAt int64) ([]AuthPolicy, error)

	// for saas

	ListBySubjectPKAndPKs(subjectPK int64, pks []int64) ([]Policy, error)
	ListBySubjectActionTemplate(subjectPK int64, actionPKs []int64, templateID int64) ([]Policy, error)
	ListExpressionBySubjectsTemplate(subjectPKs []int64, templateID int64) ([]int64, error)
	ListBySubjectTemplateBeforeExpiredAt(subjectPK int64, templateID, expiredAt int64) ([]Policy, error)
	BulkCreateWithTx(tx *sqlx.Tx, policies []Policy) error
	BulkDeleteByTemplatePKsWithTx(tx *sqlx.Tx, subjectPK, templateID int64, pks []int64) (int64, error)
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error
	BulkUpdateExpressionPKWithTx(tx *sqlx.Tx, policies []Policy) error
	BulkUpdateExpiredAtWithTx(tx *sqlx.Tx, policies []Policy) error
	DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK, limit int64) (int64, error)
	// for model update

	HasAnyByActionPK(actionPK int64) (bool, error)

	// for query

	Get(pk int64) (Policy, error)
	GetCountByActionBeforeExpiredAt(actionPK int64, expiredAt int64) (int64, error)
	ListPagingByActionPKBeforeExpiredAt(actionPK int64, expiredAt int64, offset int64, limit int64) ([]Policy, error)
	ListByPKs(pks []int64) ([]Policy, error)
}

type policyManager struct {
	DB *sqlx.DB
}

// NewPolicyManager ...
func NewPolicyManager() PolicyManager {
	return &policyManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListBySubjectPKAndPKs ...
func (m *policyManager) ListBySubjectPKAndPKs(subjectPK int64, pks []int64) (policies []Policy, err error) {
	if len(pks) == 0 {
		return
	}
	err = m.selectBySubjectPKAndPKs(&policies, subjectPK, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// ListAuthBySubjectAction ...
func (m *policyManager) ListAuthBySubjectAction(
	subjectPKs []int64, actionPK int64, expiredAt int64,
) (policies []AuthPolicy, err error) {
	if len(subjectPKs) == 0 {
		return
	}
	err = m.selectAuthBySubjectAction(&policies, subjectPKs, actionPK, expiredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// ListExpressionBySubjectsTemplate ...
func (m *policyManager) ListExpressionBySubjectsTemplate(subjectPKs []int64, templateID int64) (
	expressionPKs []int64, err error,
) {
	if len(subjectPKs) == 0 {
		return
	}
	err = m.selectExpressionPKBySubjectPKsTemplate(&expressionPKs, subjectPKs, templateID)
	if errors.Is(err, sql.ErrNoRows) {
		return expressionPKs, nil
	}
	return
}

// ListBySubjectTemplateBeforeExpiredAt ...
func (m *policyManager) ListBySubjectTemplateBeforeExpiredAt(
	subjectPK int64, templateID, expiredAt int64,
) (policies []Policy, err error) {
	err = m.selectBySubjectTemplateBeforeExpiredAt(&policies, subjectPK, templateID, expiredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// ListBySubjectActionTemplate ...
func (m *policyManager) ListBySubjectActionTemplate(
	subjectPK int64,
	actionPKs []int64,
	templateID int64,
) (policies []Policy, err error) {
	err = m.selectBySubjectActionTemplate(&policies, subjectPK, actionPKs, templateID)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *policyManager) BulkCreateWithTx(tx *sqlx.Tx, policies []Policy) error {
	if len(policies) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, policies)
}

// BulkDeleteByTemplatePKsWithTx ...
func (m *policyManager) BulkDeleteByTemplatePKsWithTx(
	tx *sqlx.Tx, subjectPK, templateID int64, pks []int64,
) (int64, error) {
	if len(pks) == 0 {
		return 0, nil
	}
	return m.bulkDeleteByTemplatePKsWithTx(tx, subjectPK, templateID, pks)
}

// BulkDeleteBySubjectPKsWithTx ...
func (m *policyManager) BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}
	return m.bulkDeleteBySubjectPKsWithTx(tx, subjectPKs)
}

// BulkUpdateExpressionPKWithTx ...
func (m *policyManager) BulkUpdateExpressionPKWithTx(tx *sqlx.Tx, policies []Policy) error {
	if len(policies) == 0 {
		return nil
	}
	return m.bulkUpdateExpressionPKWithTx(tx, policies)
}

// HasAnyByActionPK ...
func (m *policyManager) HasAnyByActionPK(actionPK int64) (exist bool, err error) {
	var pk int64
	err = m.selectExistenceByActionPK(&pk, actionPK)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Get ...
func (m *policyManager) Get(pk int64) (policy Policy, err error) {
	err = m.selectByPK(&policy, pk)
	return
}

// GetCountByActionBeforeExpiredAt ...
func (m *policyManager) GetCountByActionBeforeExpiredAt(actionPK int64, expiredAt int64) (count int64, err error) {
	err = m.selectCountByActionBeforeExpiredAt(&count, actionPK, expiredAt)
	return
}

// ListPagingByActionPKBeforeExpiredAt ...
func (m *policyManager) ListPagingByActionPKBeforeExpiredAt(
	actionPK int64,
	expiredAt int64,
	offset int64,
	limit int64,
) (policies []Policy, err error) {
	err = m.selectByActionPKOrderByPKAsc(&policies, actionPK, expiredAt, offset, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// ListByPKs ...
func (m *policyManager) ListByPKs(pks []int64) (policies []Policy, err error) {
	err = m.selectByPKs(&policies, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return policies, nil
	}
	return
}

// BulkUpdateExpiredAtWithTx ...
func (m *policyManager) BulkUpdateExpiredAtWithTx(tx *sqlx.Tx, policies []Policy) error {
	return m.updateExpiredAtWithTx(tx, policies)
}

// DeleteByActionPKWithTx ...
func (m *policyManager) DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK, limit int64) (int64, error) {
	return m.deleteByActionPKWithTx(tx, actionPK, limit)
}

func (m *policyManager) selectByPK(
	policy *Policy, pk int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
		FROM policy
		WHERE pk = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, policy, query, pk)
}

func (m *policyManager) selectByPKs(
	policy *[]Policy, pks []int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
		FROM policy
		WHERE pk in (?)`
	return database.SqlxSelect(m.DB, policy, query, pks)
}

func (m *policyManager) selectCountByActionBeforeExpiredAt(count *int64, actionPK int64, expiredAt int64) error {
	query := `SELECT
		count(*)
		FROM policy
		WHERE action_pk = ?
		AND expired_at > ?`
	return database.SqlxGet(m.DB, count, query, actionPK, expiredAt)
}

func (m *policyManager) selectByActionPKOrderByPKAsc(
	policies *[]Policy,
	actionPK int64,
	expiredAt int64,
	offset int64,
	limit int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
		FROM policy
		WHERE action_pk = ?
		AND expired_at > ?
		ORDER BY pk asc
		LIMIT ?, ?`

	// TODO: check when the performance is affect if the offset is greater than?
	if offset > 10000 {
		query = `SELECT
			t.pk,
			t.subject_pk,
			t.action_pk,
			t.expression_pk,
			t.expired_at,
			t.template_id
			FROM policy t
			INNER JOIN
			(
				SELECT pk
				FROM policy
				WHERE action_pk = ?
				AND expired_at > ?
				ORDER BY pk asc
				LIMIT ?, ?
			) p ON t.pk = p.pk`
	}

	return database.SqlxSelect(m.DB, policies, query, actionPK, expiredAt, offset, limit)
}

func (m *policyManager) selectBySubjectPKAndPKs(
	policies *[]Policy, subjectPK int64, pks []int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
		FROM policy
		WHERE subject_pk = ?
		AND pk IN (?)`
	return database.SqlxSelect(m.DB, policies, query, subjectPK, pks)
}

func (m *policyManager) selectAuthBySubjectAction(
	policies *[]AuthPolicy, subjectPKs []int64, actionPK int64, expiredAt int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		expression_pk,
		expired_at
		FROM policy
		WHERE subject_pk in (?)
		AND action_pk = ?
		AND expired_at >= ?`
	return database.SqlxSelect(m.DB, policies, query, subjectPKs, actionPK, expiredAt)
}

func (m *policyManager) selectExpressionPKBySubjectPKsTemplate(expressionPKs *[]int64,
	subjectPKs []int64, templateID int64,
) error {
	query := `SELECT
		expression_pk
		FROM policy
		WHERE subject_pk in (?)
		AND template_id = ?`
	return database.SqlxSelect(m.DB, expressionPKs, query, subjectPKs, templateID)
}

func (m *policyManager) selectBySubjectActionTemplate(
	policies *[]Policy, subjectPK int64, actionPKs []int64, templateID int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
		FROM policy
		WHERE subject_pk = ?
		AND action_pk in (?)
		AND template_id = ?`
	return database.SqlxSelect(m.DB, policies, query, subjectPK, actionPKs, templateID)
}

func (m *policyManager) selectBySubjectTemplateBeforeExpiredAt(
	policies *[]Policy, subjectPK int64, templateID int64, expiredAt int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
		FROM policy
		WHERE subject_pk = ?
		AND template_id = ?
		AND expired_at < ?
		ORDER BY expired_at DESC`
	return database.SqlxSelect(m.DB, policies, query, subjectPK, templateID, expiredAt)
}

func (m *policyManager) bulkInsertWithTx(tx *sqlx.Tx, policies []Policy) error {
	sql := `INSERT INTO policy (
		subject_pk,
		action_pk,
		expression_pk,
		expired_at,
		template_id
	) VALUES (
		:subject_pk,
		:action_pk,
		:expression_pk,
		:expired_at,
		:template_id)`
	return database.SqlxBulkInsertWithTx(tx, sql, policies)
}

func (m *policyManager) bulkDeleteByTemplatePKsWithTx(
	tx *sqlx.Tx, subjectPK, templateID int64, pks []int64,
) (int64, error) {
	sql := `DELETE FROM policy WHERE subject_pk = ? AND pk IN (?) AND template_id = ?`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, subjectPK, pks, templateID)
}

func (m *policyManager) bulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	sql := `DELETE FROM policy WHERE subject_pk IN (?)`
	return database.SqlxDeleteWithTx(tx, sql, subjectPKs)
}

func (m *policyManager) bulkUpdateExpressionPKWithTx(tx *sqlx.Tx, policies []Policy) error {
	sql := `UPDATE policy SET expression_pk=:expression_pk WHERE pk=:pk`
	return database.SqlxBulkUpdateWithTx(tx, sql, policies)
}

func (m *policyManager) selectExistenceByActionPK(pk *int64, actionPK int64) error {
	query := `SELECT
		pk
		FROM policy
		WHERE action_pk = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, pk, query, actionPK)
}

func (m *policyManager) updateExpiredAtWithTx(tx *sqlx.Tx, policies []Policy) error {
	sql := `UPDATE policy SET expired_at = :expired_at WHERE pk = :pk`

	return database.SqlxBulkUpdateWithTx(tx, sql, policies)
}

func (m *policyManager) deleteByActionPKWithTx(tx *sqlx.Tx, actionPK, limit int64) (int64, error) {
	sql := `DELETE FROM policy WHERE action_pk = ? LIMIT ?`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, actionPK, limit)
}
