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

	"iam/pkg/database"

	"github.com/jmoiron/sqlx"
)

// Subject 被授权人
type Subject struct {
	PK   int64  `db:"pk" json:"pk"`
	Type string `db:"type" json:"type"`
	ID   string `db:"id" json:"id"`
	// 仅用于”查询有某个资源的某个权限的用户列表“，
	Name string `db:"name" json:"_"`
}

// SubjectManager 获取subject属性的相关方法
type SubjectManager interface {
	Get(pk int64) (Subject, error)
	GetPK(_type, id string) (int64, error)
	ListByPKs(pks []int64) (subjects []Subject, err error)
	ListByIDs(_type string, ids []string) ([]Subject, error)
	ListPaging(_type string, limit, offset int64) ([]Subject, error)
	GetCount(_type string) (int64, error)

	BulkCreate(subjects []Subject) error
	BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error
	BulkUpdate(subjects []Subject) error
}

type subjectManager struct {
	DB *sqlx.DB
}

// NewSubjectManager New NewSubjectManager
func NewSubjectManager() SubjectManager {
	return &subjectManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *subjectManager) Get(pk int64) (subject Subject, err error) {
	err = m.selectOne(&subject, pk)
	return
}

// ListByPKs ...
func (m *subjectManager) ListByPKs(pks []int64) (subjects []Subject, err error) {
	if len(pks) == 0 {
		return
	}
	err = m.selectByPKs(&subjects, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return subjects, nil
	}
	return
}

// GetPK ...
func (m *subjectManager) GetPK(_type, id string) (int64, error) {
	var pk int64
	err := m.selectPK(&pk, _type, id)
	return pk, err
}

// ListByIDs ...
func (m *subjectManager) ListByIDs(_type string, ids []string) (subjects []Subject, err error) {
	if len(ids) == 0 {
		return
	}
	err = m.selectSubjectsByIDs(&subjects, _type, ids)
	if errors.Is(err, sql.ErrNoRows) {
		return subjects, nil
	}
	return
}

// ListPaging ...
func (m *subjectManager) ListPaging(_type string, limit, offset int64) (subjects []Subject, err error) {
	err = m.selectPagingSubjects(&subjects, _type, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return subjects, nil
	}
	return
}

// GetCount ...
func (m *subjectManager) GetCount(_type string) (int64, error) {
	var cnt int64
	err := m.getCount(&cnt, _type)
	return cnt, err
}

// BulkCreate ...
func (m *subjectManager) BulkCreate(subjects []Subject) error {
	if len(subjects) == 0 {
		return nil
	}
	return m.bulkInsert(subjects)
}

// BulkDeleteByPKsWithTx ...
func (m *subjectManager) BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	if len(pks) == 0 {
		return nil
	}
	return m.bulkDeleteByPKsWithTx(tx, pks)
}

// BulkUpdate ...
func (m *subjectManager) BulkUpdate(subjects []Subject) error {
	if len(subjects) == 0 {
		return nil
	}
	return m.bulkUpdate(subjects)
}

func (m *subjectManager) selectOne(subject *Subject, pk int64) error {
	query := `SELECT
		pk,
		type,
		id,
		name
		FROM subject
		WHERE pk = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, subject, query, pk)
}

func (m *subjectManager) selectPK(pk *int64, _type string, id string) error {
	query := `SELECT
		pk
		FROM subject
		WHERE type=?
		AND id=?
		LIMIT 1`
	return database.SqlxGet(m.DB, pk, query, _type, id)
}

func (m *subjectManager) selectByPKs(subjects *[]Subject, pks []int64) error {
	query := `SELECT
		pk,
		type,
		id,
		name
		FROM subject
		WHERE pk IN (?)`
	return database.SqlxSelect(m.DB, subjects, query, pks)
}

func (m *subjectManager) selectSubjectsByIDs(subjects *[]Subject, _type string, ids []string) error {
	query := `SELECT
		pk,
		type,
		id,
		name
		FROM subject
		WHERE type=?
		AND id IN (?)`
	return database.SqlxSelect(m.DB, subjects, query, _type, ids)
}

func (m *subjectManager) selectPagingSubjects(subjects *[]Subject, _type string, limit, offset int64) error {
	query := `SELECT
		pk,
		type,
		id,
		name
		FROM subject
		WHERE type=?
		ORDER BY pk asc
		LIMIT ? OFFSET ?`

	if offset > 10000 {
		query = `SELECT
			t.pk,
			t.type,
			t.id,
			t.name
			FROM subject t
			INNER JOIN
			(
				SELECT pk
				FROM subject
				WHERE type=?
				ORDER BY pk asc
				LIMIT ? OFFSET ?
			) s ON t.pk = s.pk`
	}
	return database.SqlxSelect(m.DB, subjects, query, _type, limit, offset)
}

func (m *subjectManager) getCount(cnt *int64, _type string) error {
	query := `SELECT
		COUNT(*)
		FROM subject
		WHERE type = ?`
	return database.SqlxGet(m.DB, cnt, query, _type)
}

func (m *subjectManager) bulkInsert(subjects []Subject) error {
	sql := "INSERT INTO subject (type, id, name) VALUES (:type, :id, :name)"
	return database.SqlxBulkInsert(m.DB, sql, subjects)
}

func (m *subjectManager) bulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	sql := `DELETE FROM subject WHERE pk in (?)`
	return database.SqlxDeleteWithTx(tx, sql, pks)
}

func (m *subjectManager) bulkUpdate(subjects []Subject) error {
	sql := "UPDATE subject SET name=:name WHERE type=:type AND id=:id"
	return database.SqlxBulkUpdate(m.DB, sql, subjects)
}
