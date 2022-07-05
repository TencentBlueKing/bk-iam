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

// SubjectSystemGroup  用户-系统组关系
type SubjectSystemGroup struct {
	PK        int64  `db:"pk"`
	SystemID  string `db:"system_id"`
	SubjectPK int64  `db:"subject_pk"`
	Groups    string `db:"groups"`
	Reversion int64  `db:"reversion"` // 更新版本
}

// SubjectGroups with the minimum fields of the relationship: subject-group-expired_at
type SubjectGroups struct {
	SubjectPK int64  `db:"subject_pk"`
	Groups    string `db:"groups"`
}

// SubjectSystemGroupManager ...
type SubjectSystemGroupManager interface {
	ListSubjectGroups(systemID string, subjectPKs []int64) ([]SubjectGroups, error)

	GetBySystemSubject(systemID string, subjectPK int64) (SubjectSystemGroup, error)
	CreateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) error
	UpdateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) (int64, error)
	DeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error
}

type subjectSystemGroupManager struct {
	DB *sqlx.DB
}

// NewSubjectSystemGroupManager New NewSubjectSystemGroup
func NewSubjectSystemGroupManager() SubjectSystemGroupManager {
	return &subjectSystemGroupManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListSubjectGroups ...
func (m *subjectSystemGroupManager) ListSubjectGroups(
	systemID string,
	subjectPKs []int64,
) (groups []SubjectGroups, err error) {
	err = m.selectGroups(&groups, systemID, subjectPKs)
	// 不存在直接返回空
	if errors.Is(err, sql.ErrNoRows) {
		return groups, nil
	}
	return
}

// GetBySystemSubject ...
func (m *subjectSystemGroupManager) GetBySystemSubject(systemID string, subjectPK int64) (SubjectSystemGroup, error) {
	var subjectSystemGroup SubjectSystemGroup
	err := m.selectBySystemSubject(&subjectSystemGroup, systemID, subjectPK)
	return subjectSystemGroup, err
}

// CreateWithTx ...
func (m *subjectSystemGroupManager) CreateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) error {
	return m.insertWithTx(tx, &subjectSystemGroup)
}

// UpdateWithTx ...
func (m *subjectSystemGroupManager) UpdateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) (int64, error) {
	return m.updateWithTx(tx, &subjectSystemGroup)
}

// DeleteBySubjectPKsWithTx ...
func (m *subjectSystemGroupManager) DeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	return m.deleteBySubjectPKsWithTx(tx, subjectPKs)
}

func (m *subjectSystemGroupManager) selectGroups(
	groups *[]SubjectGroups,
	systemID string,
	subjectPKs []int64,
) error {
	query := `SELECT
		subject_pk,
		groups
		FROM subject_system_group
		WHERE system_id = ? AND subject_pk IN (?)`
	return database.SqlxSelect(m.DB, groups, query, systemID, subjectPKs)
}

func (m *subjectSystemGroupManager) selectBySystemSubject(
	subjectSystemGroup *SubjectSystemGroup,
	systemID string,
	subjectPK int64,
) error {
	query := `SELECT
		pk,
		system_id,
		subject_pk,
		groups,
		reversion
		FROM subject_system_group
		WHERE system_id = ? AND subject_pk = ?`
	return database.SqlxGet(m.DB, subjectSystemGroup, query, systemID, subjectPK)
}

func (m *subjectSystemGroupManager) insertWithTx(tx *sqlx.Tx, subjectSystemGroup *SubjectSystemGroup) error {
	sql := `INSERT INTO subject_system_group (
		system_id, 
		subject_pk, 
		groups
	) VALUES (
		:system_id,
		:subject_pk,
		:groups
	)`
	return database.SqlxInsertWithTx(tx, sql, subjectSystemGroup)
}

func (m *subjectSystemGroupManager) updateWithTx(tx *sqlx.Tx, subjectSystemGroup *SubjectSystemGroup) (int64, error) {
	sql := `UPDATE subject_system_group SET
		groups = :groups,
		reversion = reversion + 1 
		WHERE system_id = :system_id
		AND subject_pk = :subject_pk
		AND reversion = :reversion`
	return database.SqlxUpdateWithTx(tx, sql, subjectSystemGroup)
}

func (m *subjectSystemGroupManager) deleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	sql := `DELETE FROM subject_system_group WHERE subject_pk IN (?)`
	return database.SqlxDeleteWithTx(tx, sql, subjectPKs)
}
