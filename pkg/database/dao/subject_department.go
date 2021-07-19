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

// SubjectDepartment subject的部门
type SubjectDepartment struct {
	PK            int64  `db:"pk"`
	SubjectPK     int64  `db:"subject_pk"`
	DepartmentPKs string `db:"department_pks"`
}

// SubjectDepartmentManager ...
type SubjectDepartmentManager interface {
	Get(subjectPK int64) (string, error)
	GetCount() (int64, error)
	ListPaging(limit, offset int64) ([]SubjectDepartment, error)

	BulkCreate(subjectDepartments []SubjectDepartment) error
	BulkUpdate(subjectDepartments []SubjectDepartment) error
	BulkDelete(subjectPKs []int64) error
	BulkDeleteWithTx(tx *sqlx.Tx, subjectPKs []int64) error
}

type subjectDepartmentManger struct {
	DB *sqlx.DB
}

// NewSubjectDepartmentManager New NewSubjectDepartmentManager
func NewSubjectDepartmentManager() SubjectDepartmentManager {
	return &subjectDepartmentManger{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *subjectDepartmentManger) Get(subjectPK int64) (departmentPKs string, err error) {
	err = m.getDepartmentPKs(&departmentPKs, subjectPK)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return
}

// GetCount ...
func (m *subjectDepartmentManger) GetCount() (count int64, err error) {
	err = m.getCount(&count)
	return
}

// BulkCreate ...
func (m *subjectDepartmentManger) BulkCreate(subjectDepartments []SubjectDepartment) error {
	if len(subjectDepartments) == 0 {
		return nil
	}
	return m.bulkInsert(subjectDepartments)
}

// BulkDelete ...
func (m *subjectDepartmentManger) BulkDelete(subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}
	return m.bulkDelete(subjectPKs)
}

// BulkDeleteWithTx ...
func (m *subjectDepartmentManger) BulkDeleteWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, subjectPKs)
}

// BulkUpdate ...
func (m *subjectDepartmentManger) BulkUpdate(subjectDepartments []SubjectDepartment) error {
	if len(subjectDepartments) == 0 {
		return nil
	}
	return m.bulkUpdate(subjectDepartments)
}

// ListPaging ...
func (m *subjectDepartmentManger) ListPaging(limit, offset int64) ([]SubjectDepartment, error) {
	subjectDepartments := []SubjectDepartment{}
	err := m.selectPaging(&subjectDepartments, limit, offset)
	if errors.Is(err, sql.ErrNoRows) {
		return subjectDepartments, nil
	}
	return subjectDepartments, err
}

func (m *subjectDepartmentManger) getDepartmentPKs(departmentPKs *string, subjectPK int64) error {
	query := `SELECT
		department_pks
		FROM subject_department
		WHERE subject_pk=?`
	return database.SqlxGet(m.DB, departmentPKs, query, subjectPK)
}

func (m *subjectDepartmentManger) getCount(count *int64) error {
	query := `SELECT
		COUNT(*)
		FROM subject_department`
	return database.SqlxGet(m.DB, count, query)
}

func (m *subjectDepartmentManger) bulkInsert(subjectDepartments []SubjectDepartment) error {
	sql := `INSERT INTO subject_department (
		subject_pk,
		department_pks
	) VALUES (
		:subject_pk,
		:department_pks)`
	return database.SqlxBulkInsert(m.DB, sql, subjectDepartments)
}

func (m *subjectDepartmentManger) bulkDelete(subjectPKs []int64) error {
	sql := `DELETE FROM subject_department WHERE subject_pk in (?)`
	_, err := database.SqlxDelete(m.DB, sql, subjectPKs)
	return err
}

func (m *subjectDepartmentManger) bulkDeleteWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	sql := `DELETE FROM subject_department WHERE subject_pk in (?)`
	return database.SqlxDeleteWithTx(tx, sql, subjectPKs)
}

func (m *subjectDepartmentManger) bulkUpdate(subjectDepartments []SubjectDepartment) error {
	sql := `UPDATE subject_department
		SET department_pks=:department_pks
		WHERE subject_pk=:subject_pk`
	return database.SqlxBulkUpdate(m.DB, sql, subjectDepartments)
}

func (m *subjectDepartmentManger) selectPaging(subjectDepartments *[]SubjectDepartment, limit, offset int64) error {
	query := `SELECT
		subject_pk,
		department_pks
		FROM subject_department
		ORDER BY subject_pk
		LIMIT ? OFFSET ?`
	return database.SqlxSelect(m.DB, subjectDepartments, query, limit, offset)
}
