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
	"iam/pkg/database"

	"github.com/jmoiron/sqlx"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SubjectRole 角色
type SubjectRole struct {
	PK        int64  `db:"pk"`
	RoleType  string `db:"role_type"`
	System    string `db:"system_id"`
	SubjectPK int64  `db:"subject_pk"`
}

// SubjectRoleManager ...
type SubjectRoleManager interface {
	ListSubjectPKByRole(roleType, system string) ([]int64, error)
	ListSystemIDBySubjectPK(pk int64) ([]string, error)

	BulkCreate(roles []SubjectRole) error
	BulkDelete(roleType, system string, subjectPKs []int64) error
}

type subjectRoleManager struct {
	DB *sqlx.DB
}

// NewSubjectRoleManager New SubjectRoleManager
func NewSubjectRoleManager() SubjectRoleManager {
	return &subjectRoleManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListSubjectPKByRole ...
func (m *subjectRoleManager) ListSubjectPKByRole(roleType, system string) ([]int64, error) {
	var subjectPKs = []int64{}
	err := m.selectSubjectPKByRole(&subjectPKs, roleType, system)
	return subjectPKs, err
}

// ListSystemIDBySubjectPK ...
func (m *subjectRoleManager) ListSystemIDBySubjectPK(pk int64) ([]string, error) {
	var systemIDs = []string{}
	err := m.selectSubjectSystem(&systemIDs, pk)
	if errors.Is(err, sql.ErrNoRows) {
		return systemIDs, nil
	}
	return systemIDs, err
}

// BulkCreate ...
func (m *subjectRoleManager) BulkCreate(roles []SubjectRole) error {
	if len(roles) == 0 {
		return nil
	}
	return m.bulkInsert(roles)
}

// BulkDelete ...
func (m *subjectRoleManager) BulkDelete(roleType, system string, subjectPKs []int64) error {
	return m.bulkDelete(roleType, system, subjectPKs)
}

func (m *subjectRoleManager) selectSubjectPKByRole(subjectPKs *[]int64, roleType, system string) error {
	query := `SELECT
		subject_pk
		FROM subject_role
		WHERE role_type = ?
		AND system_id = ?`
	return database.SqlxSelect(m.DB, subjectPKs, query, roleType, system)
}

func (m *subjectRoleManager) bulkInsert(roles []SubjectRole) error {
	sql := `INSERT INTO subject_role (
		role_type,
		system_id,
		subject_pk
	) VALUES (:role_type,
		:system_id,
		:subject_pk)`
	return database.SqlxBulkInsert(m.DB, sql, roles)
}

func (m *subjectRoleManager) bulkDelete(roleType, system string, subjectPKs []int64) error {
	sql := `DELETE FROM subject_role WHERE role_type = ? AND system_id = ? AND subject_pk in (?)`
	_, err := database.SqlxDelete(m.DB, sql, roleType, system, subjectPKs)
	return err
}

func (m *subjectRoleManager) selectSubjectSystem(systemIDs *[]string, subjectPK int64) error {
	query := `SELECT
		system_id
		FROM subject_role
		WHERE (role_type = "system_manager" OR role_type = "super_manager")
		AND subject_pk = ?`
	return database.SqlxSelect(m.DB, systemIDs, query, subjectPK)
}
