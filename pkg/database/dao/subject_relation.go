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
	PK int64 `db:"pk"`
	// 冗余存储，便于鉴权查询
	SubjectPK   int64  `db:"subject_pk"`
	SubjectType string `db:"subject_type"`
	SubjectID   string `db:"subject_id"`
	// 冗余存储，便于鉴权查询
	ParentPK   int64  `db:"parent_pk"`
	ParentType string `db:"parent_type"`
	ParentID   string `db:"parent_id"`
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
	ListRelation(_type, id string) ([]SubjectRelation, error)
	ListRelationBySubjectPK(subjectPK int64) ([]SubjectRelation, error)
	ListThinRelationBySubjectPK(subjectPK int64) ([]ThinSubjectRelation, error)
	ListEffectRelationBySubjectPKs(subjectPKs []int64) ([]EffectSubjectRelation, error)
	ListRelationBeforeExpiredAt(_type, id string, expiredAt int64) ([]SubjectRelation, error)

	ListPagingMember(_type, id string, limit, offset int64) ([]SubjectRelation, error)
	ListPagingMemberBeforeExpiredAt(
		_type string, id string, expiredAt int64, limit, offset int64,
	) (members []SubjectRelation, err error)
	ListMember(_type, id string) ([]SubjectRelation, error)
	GetMemberCount(_type, id string) (int64, error)
	GetMemberCountBeforeExpiredAt(_type string, id string, expiredAt int64) (int64, error)
	ListParentIDsBeforeExpiredAt(_type string, ids []string, expiredAt int64) ([]string, error)

	UpdateExpiredAt(relations []SubjectRelationPKPolicyExpiredAt) error

	BulkDeleteByMembersWithTx(tx *sqlx.Tx, _type, id, subjectType string, subjectIDs []string) (int64, error)
	BulkCreate(relations []SubjectRelation) error
	BulkDeleteBySubjectPKs(tx *sqlx.Tx, subjectPKs []int64) error
	BulkDeleteByParentPKs(tx *sqlx.Tx, parentPKs []int64) error
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
func (m *subjectRelationManager) ListRelation(_type, id string) (relations []SubjectRelation, err error) {
	err = m.selectRelation(&relations, _type, id)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListRelationBeforeExpiredAt ...
func (m *subjectRelationManager) ListRelationBeforeExpiredAt(
	_type, id string, expiredAt int64,
) (relations []SubjectRelation, err error) {
	err = m.selectRelationBeforeExpiredAt(&relations, _type, id, expiredAt)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListRelationBySubjectPK ...
func (m *subjectRelationManager) ListRelationBySubjectPK(subjectPK int64) (relations []SubjectRelation, err error) {
	err = m.selectRelationBySubjectPK(&relations, subjectPK)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListThinRelationBySubjectPK ...
func (m *subjectRelationManager) ListThinRelationBySubjectPK(
	subjectPK int64,
) (relations []ThinSubjectRelation, err error) {
	err = m.selectThinRelationBySubjectPK(&relations, subjectPK)
	// 吞掉记录不存在的错误, subject本身是可以不加入任何用户组和组织的
	if errors.Is(err, sql.ErrNoRows) {
		return relations, nil
	}
	return
}

// ListEffectRelationBySubjectPKs ...
func (m *subjectRelationManager) ListEffectRelationBySubjectPKs(subjectPKs []int64) (
	relations []EffectSubjectRelation, err error) {
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
func (m *subjectRelationManager) ListPagingMember(_type, id string, limit, offset int64) (
	members []SubjectRelation, err error) {
	err = m.selectPagingMembers(&members, _type, id, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// ListMember ...
func (m *subjectRelationManager) ListMember(_type, id string) (members []SubjectRelation, err error) {
	err = m.selectMembers(&members, _type, id)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// GetMemberCount ...
func (m *subjectRelationManager) GetMemberCount(_type, id string) (int64, error) {
	var cnt int64
	err := m.getMemberCount(&cnt, _type, id)
	return cnt, err
}

// BulkDeleteByMembersWithTx ...
func (m *subjectRelationManager) BulkDeleteByMembersWithTx(
	tx *sqlx.Tx, _type, id, subjectType string, subjectIDs []string) (int64, error) {
	if len(subjectIDs) == 0 {
		return 0, nil
	}
	return m.bulkDeleteByMembersWithTx(tx, _type, id, subjectType, subjectIDs)
}

// BulkCreate ...
func (m *subjectRelationManager) BulkCreate(relations []SubjectRelation) error {
	if len(relations) == 0 {
		return nil
	}
	return m.bulkInsert(relations)
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
func (m *subjectRelationManager) UpdateExpiredAt(relations []SubjectRelationPKPolicyExpiredAt) error {
	return m.updateExpiredAt(relations)
}

// GetMemberCountBeforeExpiredAt ...
func (m *subjectRelationManager) GetMemberCountBeforeExpiredAt(
	_type string, id string, expiredAt int64,
) (int64, error) {
	var cnt int64
	err := m.getMemberCountBeforeExpiredAt(&cnt, _type, id, expiredAt)
	return cnt, err
}

// ListPagingMemberBeforeExpiredAt ...
func (m *subjectRelationManager) ListPagingMemberBeforeExpiredAt(
	_type string, id string, expiredAt int64, limit, offset int64,
) (members []SubjectRelation, err error) {
	err = m.selectPagingMembersBeforeExpiredAt(&members, _type, id, expiredAt, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return members, nil
	}
	return
}

// ListParentIDsBeforeExpiredAt get the group ids before timestamp(expiredAt)
func (m *subjectRelationManager) ListParentIDsBeforeExpiredAt(
	_type string, ids []string, expiredAt int64,
) ([]string, error) {
	expiredParentIDs := []string{}
	err := m.listParentIDsBeforeExpiredAt(&expiredParentIDs, _type, ids, expiredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return expiredParentIDs, nil
	}
	return expiredParentIDs, err
}

func (m *subjectRelationManager) selectRelation(relations *[]SubjectRelation, _type, id string) error {
	query := `SELECT
		pk,
		subject_pk,
		subject_type,
		subject_id,
		parent_pk,
		parent_type,
		parent_id,
		policy_expired_at,
		created_at
		FROM subject_relation
		WHERE subject_type = ?
		AND subject_id = ?`
	return database.SqlxSelect(m.DB, relations, query, _type, id)
}

func (m *subjectRelationManager) selectRelationBeforeExpiredAt(
	relations *[]SubjectRelation, _type, id string, expiredAt int64,
) error {
	query := `SELECT
		pk,
		subject_pk,
		subject_type,
		subject_id,
		parent_pk,
		parent_type,
		parent_id,
		policy_expired_at,
		created_at
		FROM subject_relation
		WHERE subject_type = ?
		AND subject_id = ?
		AND policy_expired_at < ?
		ORDER BY policy_expired_at DESC`
	return database.SqlxSelect(m.DB, relations, query, _type, id, expiredAt)
}

func (m *subjectRelationManager) selectRelationBySubjectPK(relations *[]SubjectRelation, pk int64) error {
	query := `SELECT
		pk,
		subject_pk,
		subject_type,
		subject_id,
		parent_pk,
		parent_type,
		parent_id,
		policy_expired_at,
		created_at
		FROM subject_relation
		WHERE subject_pk = ?`
	return database.SqlxSelect(m.DB, relations, query, pk)
}

func (m *subjectRelationManager) selectThinRelationBySubjectPK(relations *[]ThinSubjectRelation, pk int64) error {
	query := `SELECT
		parent_pk,
		policy_expired_at
		FROM subject_relation
		WHERE subject_pk = ?`
	return database.SqlxSelect(m.DB, relations, query, pk)
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
	members *[]SubjectRelation, _type, id string, limit, offset int64) error {
	query := `SELECT
		pk,
		subject_pk,
		subject_type,
		subject_id,
		parent_pk,
		parent_type,
		parent_id,
		policy_expired_at,
		created_at
		FROM subject_relation
		WHERE parent_type = ?
		AND parent_id = ?
		ORDER BY pk DESC
		LIMIT ? OFFSET ?`
	return database.SqlxSelect(m.DB, members, query, _type, id, limit, offset)
}

func (m *subjectRelationManager) selectPagingMembersBeforeExpiredAt(
	members *[]SubjectRelation, _type string, id string, expiredAt int64, limit, offset int64) error {
	query := `SELECT
		pk,
		subject_pk,
		subject_type,
		subject_id,
		parent_pk,
		parent_type,
		parent_id,
		policy_expired_at,
		created_at
		FROM subject_relation
		WHERE parent_type = ?
		AND parent_id = ?
		AND policy_expired_at < ?
		ORDER BY policy_expired_at DESC, pk DESC
		LIMIT ? OFFSET ?`
	return database.SqlxSelect(m.DB, members, query, _type, id, expiredAt, limit, offset)
}

func (m *subjectRelationManager) selectMembers(
	members *[]SubjectRelation, _type, id string) error {
	query := `SELECT
		pk,
		subject_pk,
		subject_type,
		subject_id,
		parent_pk,
		parent_type,
		parent_id,
		policy_expired_at,
		created_at
		FROM subject_relation
		WHERE parent_type = ?
		AND parent_id = ?`
	return database.SqlxSelect(m.DB, members, query, _type, id)
}

func (m *subjectRelationManager) getMemberCount(cnt *int64, _type, id string) error {
	query := `SELECT
		COUNT(*)
		FROM subject_relation
		WHERE parent_type = ?
		AND parent_id = ?`
	return database.SqlxGet(m.DB, cnt, query, _type, id)
}

func (m *subjectRelationManager) getMemberCountBeforeExpiredAt(
	cnt *int64, _type string, id string, expiredAt int64,
) error {
	query := `SELECT
		COUNT(*)
		FROM subject_relation
		WHERE parent_type = ?
		AND parent_id = ?
		AND policy_expired_at < ?`
	return database.SqlxGet(m.DB, cnt, query, _type, id, expiredAt)
}

func (m *subjectRelationManager) bulkDeleteByMembersWithTx(
	tx *sqlx.Tx, _type, id, subjectType string, subjectIDs []string) (int64, error) {
	sql := `DELETE FROM subject_relation WHERE parent_type=? AND parent_id=? AND subject_type=? AND subject_id in (?)`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, _type, id, subjectType, subjectIDs)
}

func (m *subjectRelationManager) bulkInsert(relations []SubjectRelation) error {
	sql := `INSERT INTO subject_relation (
		subject_pk,
		subject_type,
		subject_id,
		parent_pk,
		parent_type,
		parent_id,
		policy_expired_at,
		created_at
	) VALUES (:subject_pk,
		:subject_type,
		:subject_id,
		:parent_pk,
		:parent_type,
		:parent_id,
		:policy_expired_at,
		:created_at)`
	return database.SqlxBulkInsert(m.DB, sql, relations)
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

func (m *subjectRelationManager) updateExpiredAt(relations []SubjectRelationPKPolicyExpiredAt) error {
	sql := `UPDATE subject_relation SET policy_expired_at = :policy_expired_at WHERE pk = :pk`

	return database.SqlxBulkUpdate(m.DB, sql, relations)
}

func (m *subjectRelationManager) listParentIDsBeforeExpiredAt(
	parentIDs *[]string, _type string, ids []string, expiredAt int64,
) error {
	query := `SELECT
		DISTINCT parent_id
		FROM subject_relation
		WHERE parent_type = ?
		AND parent_id IN (?)
		AND policy_expired_at < ?`
	return database.SqlxSelect(m.DB, parentIDs, query, _type, ids, expiredAt)
}
