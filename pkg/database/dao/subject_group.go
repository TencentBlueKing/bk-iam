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
	// NOTE: map parent_pk to GroupPK in dao
	GroupPK int64 `db:"parent_pk"`
	// 策略有效期，unix time，单位秒(s)
	// NOTE: map policy_expired_at to ExpiredAt in dao
	ExpiredAt int64     `db:"policy_expired_at"`
	CreatedAt time.Time `db:"created_at"`
}

// SubjectRelationForUpdateExpiredAt keep the PrimaryKey and expired_at
type SubjectRelationForUpdateExpiredAt struct {
	PK int64 `db:"pk"`
	// NOTE: map policy_expired_at to ExpiredAt in dao
	ExpiredAt int64 `db:"policy_expired_at"`
}

// ThinSubjectRelation with the minimum fields of the relationship: subject-group-expired_at
type ThinSubjectRelation struct {
	SubjectPK int64 `db:"subject_pk"`
	// NOTE: map parent_pk to GroupPK in dao
	GroupPK int64 `db:"parent_pk"`
	// NOTE: map policy_expired_at to ExpiredAt in dao
	ExpiredAt int64 `db:"policy_expired_at"`
}

// SubjectGroupManager ...
type SubjectGroupManager interface {
	ListThinRelationAfterExpiredAtBySubjectPKs(subjectPKs []int64, expiredAt int64) ([]ThinSubjectRelation, error)

	GetSubjectGroupCount(subjectPK int64) (int64, error)
	GetSubjectGroupCountBeforeExpiredAt(subjectPK int64, expiredAt int64) (int64, error)
	ListPagingSubjectGroups(subjectPK, limit, offset int64) ([]SubjectRelation, error)
	ListPagingSubjectGroupBeforeExpiredAt(subjectPK, expiredAt, limit, offset int64) ([]SubjectRelation, error)

	GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (int64, error)
	ListPagingGroupSubjectBeforeExpiredAt(expiredAt int64, limit, offset int64) (members []SubjectRelation, err error)

	FilterGroupPKsHasMemberBeforeExpiredAt(groupPKs []int64, expiredAt int64) ([]int64, error)
	FilterSubjectPKsExistGroupPKsAfterExpiredAt(subjectPKs []int64, groupPKs []int64, expiredAt int64) ([]int64, error)

	UpdateExpiredAtWithTx(tx *sqlx.Tx, relations []SubjectRelationForUpdateExpiredAt) error
	BulkCreateWithTx(tx *sqlx.Tx, relations []SubjectRelation) error
	BulkDeleteBySubjectPKs(tx *sqlx.Tx, subjectPKs []int64) error
	BulkDeleteByGroupPKs(tx *sqlx.Tx, groupPKs []int64) error

	ListGroupMember(groupPK int64) ([]SubjectRelation, error)
	ListPagingGroupMember(groupPK int64, limit, offset int64) ([]SubjectRelation, error)
	ListPagingGroupMemberBeforeExpiredAt(
		groupPK int64, expiredAt int64, limit, offset int64,
	) (members []SubjectRelation, err error)
	GetGroupMemberCount(groupPK int64) (int64, error)
	GetGroupMemberCountBeforeExpiredAt(groupPK int64, expiredAt int64) (int64, error)
	GetExpiredAtBySubjectGroup(subjectPK, groupPK int64) (int64, error)

	BulkDeleteByGroupMembersWithTx(tx *sqlx.Tx, groupPK int64, subjectPKs []int64) (int64, error)
}

type subjectGroupManager struct {
	DB *sqlx.DB
}

// NewSubjectGroupManager New SubjectGroupManager
func NewSubjectGroupManager() SubjectGroupManager {
	return &subjectGroupManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// GetSubjectGroupCount ...
func (m *subjectGroupManager) GetSubjectGroupCount(subjectPK int64) (int64, error) {
	var count int64
	query := `SELECT
		 COUNT(*)
		 FROM subject_relation
		 WHERE subject_pk = ?`

	err := database.SqlxGet(m.DB, &count, query, subjectPK)
	return count, err
}

// ListPagingSubjectGroups ...
func (m *subjectGroupManager) ListPagingSubjectGroups(
	subjectPK, limit, offset int64,
) (relations []SubjectRelation, err error) {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE subject_pk = ?
		 ORDER BY pk DESC
		 LIMIT ? OFFSET ?`
	err = database.SqlxSelect(m.DB, &relations, query, subjectPK, limit, offset)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// GetSubjectGroupCountBeforeExpiredAt ...
func (m *subjectGroupManager) GetSubjectGroupCountBeforeExpiredAt(subjectPK int64, expiredAt int64) (int64, error) {
	var count int64
	query := `SELECT
		 COUNT(*)
		 FROM subject_relation
		 WHERE subject_pk = ?
		 AND policy_expired_at < ?`
	err := database.SqlxGet(m.DB, &count, query, subjectPK, expiredAt)

	return count, err
}

// GetGroupSubjectCountBeforeExpiredAt ...
func (m *subjectGroupManager) GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (int64, error) {
	var count int64
	query := `SELECT
		 COUNT(*)
		 FROM subject_relation
		 WHERE policy_expired_at < ?`
	err := database.SqlxGet(m.DB, &count, query, expiredAt)

	return count, err
}

// ListPagingSubjectGroupBeforeExpiredAt ...
func (m *subjectGroupManager) ListPagingSubjectGroupBeforeExpiredAt(
	subjectPK, expiredAt, limit, offset int64,
) (relations []SubjectRelation, err error) {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE subject_pk = ?
		 AND policy_expired_at < ?
		 ORDER BY policy_expired_at DESC, pk DESC
		 LIMIT ? OFFSET ?`
	err = database.SqlxSelect(m.DB, &relations, query, subjectPK, expiredAt, limit, offset)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListThinRelationAfterExpiredAtBySubjectPKs ...
func (m *subjectGroupManager) ListThinRelationAfterExpiredAtBySubjectPKs(subjectPKs []int64, expiredAt int64) (
	relations []ThinSubjectRelation, err error,
) {
	if len(subjectPKs) == 0 {
		return
	}

	query := `SELECT
		 subject_pk,
		 parent_pk,
		 policy_expired_at
		 FROM subject_relation
		 WHERE subject_pk in (?)
		 AND policy_expired_at > ?`
	err = database.SqlxSelect(m.DB, &relations, query, subjectPKs, expiredAt)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListPagingGroupMember ...
func (m *subjectGroupManager) ListPagingGroupMember(groupPK int64, limit, offset int64) (
	members []SubjectRelation, err error,
) {
	err = m.selectPagingMembers(&members, groupPK, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// ListGroupMember ...
func (m *subjectGroupManager) ListGroupMember(groupPK int64) (members []SubjectRelation, err error) {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE parent_pk = ?`
	err = database.SqlxSelect(m.DB, &members, query, groupPK)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// GetExpiredAtBySubjectGroup ...
func (m *subjectGroupManager) GetExpiredAtBySubjectGroup(subjectPK, groupPK int64) (int64, error) {
	var expiredAt int64
	query := `SELECT
		 policy_expired_at
		 FROM subject_relation
		 WHERE subject_pk = ?
		 AND parent_pk = ?`
	err := database.SqlxGet(m.DB, &expiredAt, query, subjectPK, groupPK)

	return expiredAt, err
}

// GetGroupMemberCount ...
func (m *subjectGroupManager) GetGroupMemberCount(groupPK int64) (int64, error) {
	var count int64
	err := m.getMemberCount(&count, groupPK)
	return count, err
}

// BulkDeleteByGroupMembersWithTx ...
func (m *subjectGroupManager) BulkDeleteByGroupMembersWithTx(
	tx *sqlx.Tx, groupPK int64, subjectPKs []int64,
) (int64, error) {
	if len(subjectPKs) == 0 {
		return 0, nil
	}
	return m.bulkDeleteByMembersWithTx(tx, groupPK, subjectPKs)
}

// BulkCreateWithTx ...
func (m *subjectGroupManager) BulkCreateWithTx(tx *sqlx.Tx, relations []SubjectRelation) error {
	if len(relations) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, relations)
}

// BulkDeleteBySubjectPKs ...
func (m *subjectGroupManager) BulkDeleteBySubjectPKs(tx *sqlx.Tx, subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}
	return m.bulkDeleteBySubjectPKs(tx, subjectPKs)
}

// BulkDeleteByGroupPKs ...
func (m *subjectGroupManager) BulkDeleteByGroupPKs(tx *sqlx.Tx, groupPKs []int64) error {
	if len(groupPKs) == 0 {
		return nil
	}
	return m.bulkDeleteByGroupPKs(tx, groupPKs)
}

// UpdateExpiredAtWithTx ...
func (m *subjectGroupManager) UpdateExpiredAtWithTx(
	tx *sqlx.Tx,
	relations []SubjectRelationForUpdateExpiredAt,
) error {
	sql := `UPDATE subject_relation SET policy_expired_at = :policy_expired_at WHERE pk = :pk`
	return database.SqlxBulkUpdateWithTx(tx, sql, relations)
}

// GetGroupMemberCountBeforeExpiredAt ...
func (m *subjectGroupManager) GetGroupMemberCountBeforeExpiredAt(
	groupPK int64, expiredAt int64,
) (int64, error) {
	var count int64
	err := m.getMemberCountBeforeExpiredAt(&count, groupPK, expiredAt)
	return count, err
}

// ListPagingGroupMemberBeforeExpiredAt ...
func (m *subjectGroupManager) ListPagingGroupMemberBeforeExpiredAt(
	groupPK int64, expiredAt int64, limit, offset int64,
) (members []SubjectRelation, err error) {
	err = m.selectPagingMembersBeforeExpiredAt(&members, groupPK, expiredAt, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// ListPagingGroupSubjectBeforeExpiredAt ...
func (m *subjectGroupManager) ListPagingGroupSubjectBeforeExpiredAt(
	expiredAt int64, limit, offset int64,
) (members []SubjectRelation, err error) {
	err = m.selectPagingGroupSubjectBeforeExpiredAt(&members, expiredAt, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// FilterGroupPKsHasMemberBeforeExpiredAt get the group pks before timestamp(expiredAt)
func (m *subjectGroupManager) FilterGroupPKsHasMemberBeforeExpiredAt(
	groupPKs []int64,
	expiredAt int64,
) ([]int64, error) {
	expiredGroupPKs := []int64{}
	// TODO: DISTINCT 大表很慢
	query := `SELECT
		 DISTINCT parent_pk
		 FROM subject_relation
		 WHERE parent_pk IN (?)
		 AND policy_expired_at < ?`
	err := database.SqlxSelect(m.DB, &expiredGroupPKs, query, groupPKs, expiredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return expiredGroupPKs, nil
	}
	return expiredGroupPKs, err
}

func (m *subjectGroupManager) FilterSubjectPKsExistGroupPKsAfterExpiredAt(
	subjectPKs []int64,
	groupPKs []int64,
	expiredAt int64,
) ([]int64, error) {
	existGroupPKs := []int64{}

	query := `SELECT
		 parent_pk
		 FROM subject_relation
		 WHERE subject_pk in (?)
		 AND parent_pk in (?)
		 AND policy_expired_at > ?`

	err := database.SqlxSelect(m.DB, &existGroupPKs, query, subjectPKs, groupPKs, expiredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return groupPKs, nil
	}

	return groupPKs, err
}

func (m *subjectGroupManager) selectPagingMembers(
	members *[]SubjectRelation, groupPK int64, limit, offset int64,
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
	return database.SqlxSelect(m.DB, members, query, groupPK, limit, offset)
}

func (m *subjectGroupManager) selectPagingMembersBeforeExpiredAt(
	members *[]SubjectRelation, groupPK int64, expiredAt int64, limit, offset int64,
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
	return database.SqlxSelect(m.DB, members, query, groupPK, expiredAt, limit, offset)
}

func (m *subjectGroupManager) selectPagingGroupSubjectBeforeExpiredAt(
	members *[]SubjectRelation, expiredAt int64, limit, offset int64,
) error {
	query := `SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE policy_expired_at < ?
		 ORDER BY policy_expired_at DESC, pk DESC
		 LIMIT ? OFFSET ?`
	return database.SqlxSelect(m.DB, members, query, expiredAt, limit, offset)
}

func (m *subjectGroupManager) getMemberCount(count *int64, groupPK int64) error {
	query := `SELECT
		 COUNT(*)
		 FROM subject_relation
		 WHERE parent_pk = ?`
	return database.SqlxGet(m.DB, count, query, groupPK)
}

func (m *subjectGroupManager) getMemberCountBeforeExpiredAt(
	count *int64, groupPK int64, expiredAt int64,
) error {
	query := `SELECT
		 COUNT(*)
		 FROM subject_relation
		 WHERE parent_pk = ?
		 AND policy_expired_at < ?`
	return database.SqlxGet(m.DB, count, query, groupPK, expiredAt)
}

func (m *subjectGroupManager) bulkDeleteByMembersWithTx(
	tx *sqlx.Tx, groupPK int64, subjectPKs []int64,
) (int64, error) {
	sql := `DELETE FROM subject_relation WHERE parent_pk=? AND subject_pk in (?)`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, groupPK, subjectPKs)
}

func (m *subjectGroupManager) bulkInsertWithTx(tx *sqlx.Tx, relations []SubjectRelation) error {
	sql := `INSERT INTO subject_relation (
		subject_pk,
		parent_pk,
		policy_expired_at
	) VALUES (:subject_pk,
		:parent_pk,
		:policy_expired_at)`
	return database.SqlxBulkInsertWithTx(tx, sql, relations)
}

func (m *subjectGroupManager) bulkDeleteBySubjectPKs(tx *sqlx.Tx, subjectPKs []int64) error {
	sql := `DELETE FROM subject_relation WHERE subject_pk in (?)`
	return database.SqlxDeleteWithTx(tx, sql, subjectPKs)
}

func (m *subjectGroupManager) bulkDeleteByGroupPKs(tx *sqlx.Tx, groupPKs []int64) error {
	// TODO: 可能的全表扫描
	sql := `DELETE FROM subject_relation WHERE parent_pk in (?)`
	return database.SqlxDeleteWithTx(tx, sql, groupPKs)
}
