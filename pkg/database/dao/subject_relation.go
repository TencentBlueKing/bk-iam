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
	"time"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// SubjectRelation  用户-组/部门-组关系表
type SubjectRelation struct {
	PK        int64 `db:"pk"`
	SubjectPK int64 `db:"subject_pk"`
	ParentPK  int64 `db:"parent_pk"`
	// 策略有效期，unix time，单位秒(s)
	PolicyExpiredAt int64     `db:"policy_expired_at"`
	CreateAt        time.Time `db:"created_at"`
}

// SubjectRelationPKPolicyExpiredAt keep the PrimaryKey and expired_at
type SubjectRelationPKPolicyExpiredAt struct {
	PK              int64 `db:"pk"`
	PolicyExpiredAt int64 `db:"policy_expired_at"`
}

// ThinSubjectRelation keep the groupPK with a expired_at
type ThinSubjectRelation struct {
	ParentPK        int64 `db:"parent_pk"`
	PolicyExpiredAt int64 `db:"policy_expired_at"`
}

// EffectSubjectRelation with the minimum fields of the relationship: subject-group-expired_at
type EffectSubjectRelation struct {
	SubjectPK       int64 `db:"subject_pk"`
	ParentPK        int64 `db:"parent_pk"`
	PolicyExpiredAt int64 `db:"policy_expired_at"`
}

// SubjectRelationManager ...
type SubjectRelationManager interface {
	ListEffectThinRelationBySubjectPK(subjectPK int64) ([]ThinSubjectRelation, error)
	ListEffectRelationBySubjectPKs(subjectPKs []int64) ([]EffectSubjectRelation, error)

	ListRelation(subjectPK int64) ([]SubjectRelation, error)
	ListRelationBeforeExpiredAt(subjectPK int64, expiredAt int64) ([]SubjectRelation, error)
	ListParentPKsBeforeExpiredAt(parentPKs []int64, expiredAt int64) ([]int64, error)

	UpdateExpiredAtWithTx(tx *sqlx.Tx, relations []SubjectRelationPKPolicyExpiredAt) error
	BulkCreateWithTx(tx *sqlx.Tx, relations []SubjectRelation) error
	BulkDeleteBySubjectPKs(tx *sqlx.Tx, subjectPKs []int64) error
	BulkDeleteByParentPKs(tx *sqlx.Tx, parentPKs []int64) error

	ListPagingMember(parentPK int64, limit, offset int64) ([]SubjectRelation, error)
	ListPagingMemberBeforeExpiredAt(
		parentPK int64, expiredAt int64, limit, offset int64,
	) (members []SubjectRelation, err error)
	ListMember(parentPK int64) ([]SubjectRelation, error)
	GetMemberCount(parentPK int64) (int64, error)
	GetMemberCountBeforeExpiredAt(parentPK int64, expiredAt int64) (int64, error)

	BulkDeleteByMembersWithTx(tx *sqlx.Tx, parentPK int64, subjectPKs []int64) (int64, error)
}

type subjectRelationManager struct {
	DB *sqlx.DB
}

// NewSubjectRelationManager New SubjectRelationManager
func NewSubjectRelationManager() SubjectRelationManager {
	return &subjectRelationManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListRelation ...
func (m *subjectRelationManager) ListRelation(subjectPK int64) (relations []SubjectRelation, err error) {
	err = m.selectRelation(&relations, subjectPK)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListRelationBeforeExpiredAt ...
func (m *subjectRelationManager) ListRelationBeforeExpiredAt(
	subjectPK int64, expiredAt int64,
) (relations []SubjectRelation, err error) {
	err = m.selectRelationBeforeExpiredAt(&relations, subjectPK, expiredAt)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListEffectThinRelationBySubjectPK ...
func (m *subjectRelationManager) ListEffectThinRelationBySubjectPK(subjectPK int64) (
	relations []ThinSubjectRelation, err error,
) {
	// 过期时间必须大于当前时间
	now := time.Now().Unix()

	err = m.selectEffectThinRelationBySubjectPK(&relations, subjectPK, now)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListEffectRelationBySubjectPKs ...
func (m *subjectRelationManager) ListEffectRelationBySubjectPKs(subjectPKs []int64) (
	relations []EffectSubjectRelation, err error,
) {
	if len(subjectPKs) == 0 {
		return
	}

	// 过期时间必须大于当前时间
	now := time.Now().Unix()

	err = m.selectEffectRelationBySubjectPKs(&relations, subjectPKs, now)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListPagingMember ...
func (m *subjectRelationManager) ListPagingMember(parentPK int64, limit, offset int64) (
	members []SubjectRelation, err error,
) {
	err = m.selectPagingMembers(&members, parentPK, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// ListMember ...
func (m *subjectRelationManager) ListMember(parentPK int64) (members []SubjectRelation, err error) {
	err = m.selectMembers(&members, parentPK)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// GetMemberCount ...
func (m *subjectRelationManager) GetMemberCount(parentPK int64) (int64, error) {
	var count int64
	err := m.getMemberCount(&count, parentPK)
	return count, err
}

// BulkDeleteByMembersWithTx ...
func (m *subjectRelationManager) BulkDeleteByMembersWithTx(
	tx *sqlx.Tx, parentPK int64, subjectPKs []int64,
) (int64, error) {
	if len(subjectPKs) == 0 {
		return 0, nil
	}
	return m.bulkDeleteByMembersWithTx(tx, parentPK, subjectPKs)
}

// BulkCreateWithTx ...
func (m *subjectRelationManager) BulkCreateWithTx(tx *sqlx.Tx, relations []SubjectRelation) error {
	if len(relations) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, relations)
}

// BulkDeleteBySubjectPKs ...
func (m *subjectRelationManager) BulkDeleteBySubjectPKs(tx *sqlx.Tx, subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}
	return m.bulkDeleteBySubjectPKs(tx, subjectPKs)
}

// BulkDeleteByParentPKs ...
func (m *subjectRelationManager) BulkDeleteByParentPKs(tx *sqlx.Tx, parentPKs []int64) error {
	if len(parentPKs) == 0 {
		return nil
	}
	return m.bulkDeleteByParentPKs(tx, parentPKs)
}

// UpdateExpiredAt ...
func (m *subjectRelationManager) UpdateExpiredAtWithTx(
	tx *sqlx.Tx,
	relations []SubjectRelationPKPolicyExpiredAt,
) error {
	return m.updateExpiredAtWithTx(tx, relations)
}

// GetMemberCountBeforeExpiredAt ...
func (m *subjectRelationManager) GetMemberCountBeforeExpiredAt(
	parentPK int64, expiredAt int64,
) (int64, error) {
	var count int64
	err := m.getMemberCountBeforeExpiredAt(&count, parentPK, expiredAt)
	return count, err
}

// ListPagingMemberBeforeExpiredAt ...
func (m *subjectRelationManager) ListPagingMemberBeforeExpiredAt(
	parentPK int64, expiredAt int64, limit, offset int64,
) (members []SubjectRelation, err error) {
	err = m.selectPagingMembersBeforeExpiredAt(&members, parentPK, expiredAt, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// ListParentPKsBeforeExpiredAt get the group pks before timestamp(expiredAt)
func (m *subjectRelationManager) ListParentPKsBeforeExpiredAt(parentPKs []int64, expiredAt int64) ([]int64, error) {
	expiredParentPKs := []int64{}
	err := m.listParentPKsBeforeExpiredAt(&expiredParentPKs, parentPKs, expiredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return expiredParentPKs, nil
	}
	return expiredParentPKs, err
}

func (m *subjectRelationManager) selectRelation(relations *[]SubjectRelation, subjectPK int64) error {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE subject_pk = ?`
	return database.SqlxSelect(m.DB, relations, query, subjectPK)
}

func (m *subjectRelationManager) selectRelationBeforeExpiredAt(
	relations *[]SubjectRelation, subjectPK int64, expiredAt int64,
) error {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE subject_pk = ?
		 AND policy_expired_at < ?
		 ORDER BY policy_expired_at DESC`
	return database.SqlxSelect(m.DB, relations, query, subjectPK, expiredAt)
}

func (m *subjectRelationManager) selectEffectThinRelationBySubjectPK(
	relations *[]ThinSubjectRelation,
	pk int64,
	now int64,
) error {
	query := `SELECT
		 parent_pk,
		 policy_expired_at
		 FROM subject_relation
		 WHERE subject_pk = ?
		 AND policy_expired_at > ?`
	return database.SqlxSelect(m.DB, relations, query, pk, now)
}

func (m *subjectRelationManager) selectEffectRelationBySubjectPKs(
	relations *[]EffectSubjectRelation,
	pks []int64,
	now int64,
) error {
	query := `SELECT
		 subject_pk,
		 parent_pk,
		 policy_expired_at
		 FROM subject_relation
		 WHERE subject_pk in (?)
		 AND policy_expired_at > ?`
	return database.SqlxSelect(m.DB, relations, query, pks, now)
}

func (m *subjectRelationManager) selectPagingMembers(
	members *[]SubjectRelation, parentPK int64, limit, offset int64,
) error {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE parent_pk = ?
		 ORDER BY pk DESC
		 LIMIT ? OFFSET ?`
	return database.SqlxSelect(m.DB, members, query, parentPK, limit, offset)
}

func (m *subjectRelationManager) selectPagingMembersBeforeExpiredAt(
	members *[]SubjectRelation, parentPK int64, expiredAt int64, limit, offset int64,
) error {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE parent_pk = ?
		 AND policy_expired_at < ?
		 ORDER BY policy_expired_at DESC, pk DESC
		 LIMIT ? OFFSET ?`
	return database.SqlxSelect(m.DB, members, query, parentPK, expiredAt, limit, offset)
}

func (m *subjectRelationManager) selectMembers(
	members *[]SubjectRelation, parentPK int64,
) error {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE parent_pk = ?`
	return database.SqlxSelect(m.DB, members, query, parentPK)
}

func (m *subjectRelationManager) getMemberCount(count *int64, parentPK int64) error {
	query := `SELECT
		 COUNT(*)
		 FROM subject_relation
		 WHERE parent_pk = ?`
	return database.SqlxGet(m.DB, count, query, parentPK)
}

func (m *subjectRelationManager) getMemberCountBeforeExpiredAt(
	count *int64, parentPK int64, expiredAt int64,
) error {
	query := `SELECT
		 COUNT(*)
		 FROM subject_relation
		 WHERE parent_pk = ?
		 AND policy_expired_at < ?`
	return database.SqlxGet(m.DB, count, query, parentPK, expiredAt)
}

func (m *subjectRelationManager) bulkDeleteByMembersWithTx(
	tx *sqlx.Tx, parentPK int64, subjectPKs []int64,
) (int64, error) {
	sql := `DELETE FROM subject_relation WHERE parent_pk=? AND subject_pk in (?)`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, parentPK, subjectPKs)
}

func (m *subjectRelationManager) bulkInsertWithTx(tx *sqlx.Tx, relations []SubjectRelation) error {
	sql := `INSERT INTO subject_relation (
		subject_pk,
		parent_pk,
		policy_expired_at
	) VALUES (:subject_pk,
		:parent_pk,
		:policy_expired_at)`
	return database.SqlxBulkInsertWithTx(tx, sql, relations)
}

func (m *subjectRelationManager) bulkDeleteBySubjectPKs(tx *sqlx.Tx, subjectPKs []int64) error {
	sql := `DELETE FROM subject_relation WHERE subject_pk in (?)`
	return database.SqlxDeleteWithTx(tx, sql, subjectPKs)
}

func (m *subjectRelationManager) bulkDeleteByParentPKs(tx *sqlx.Tx, parentPKs []int64) error {
	// TODO: 可能的全表扫描
	sql := `DELETE FROM subject_relation WHERE parent_pk in (?)`
	return database.SqlxDeleteWithTx(tx, sql, parentPKs)
}

func (m *subjectRelationManager) updateExpiredAtWithTx(
	tx *sqlx.Tx,
	relations []SubjectRelationPKPolicyExpiredAt,
) error {
	sql := `UPDATE subject_relation SET policy_expired_at = :policy_expired_at WHERE pk = :pk`
	return database.SqlxBulkUpdateWithTx(tx, sql, relations)
}

func (m *subjectRelationManager) listParentPKsBeforeExpiredAt(
	expiredParentPKs *[]int64, parentPKs []int64, expiredAt int64,
) error {
	// TODO: DISTINCT 大表很慢
	query := `SELECT
		 DISTINCT parent_pk
		 FROM subject_relation
		 WHERE parent_pk IN (?)
		 AND policy_expired_at < ?`
	return database.SqlxSelect(m.DB, expiredParentPKs, query, parentPKs, expiredAt)
}
