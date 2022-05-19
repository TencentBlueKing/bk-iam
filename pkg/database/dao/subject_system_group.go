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
	"time"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SubjectSystemGroup  用户-系统组关系
type SubjectSystemGroup struct {
	PK        int64     `db:"pk"`
	SystemID  string    `db:"system_id"`
	SubjectPK int64     `db:"subject_pk"`
	Groups    string    `db:"groups"`
	Updates   int64     `db:"updates"` // 更新版本
	CreateAt  time.Time `db:"created_at"`
}

// SubjectSystemGroup ...
type SubjectSystemGroupManager interface {
	GetGroups(systemID string, subjectPK int64) (string, error)

	GetBySystemSubject(systemID string, subjectPK int64) (SubjectSystemGroup, error)
	CreateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) error
	UpdateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) (int64, error)
	DeleteBySystemSubjectWithTx(tx *sqlx.Tx, systemID string, subjectPK int64) error
}

type subjectSystemGroupManager struct {
	DB *sqlx.DB
}

// NewSubjectSystemGroup New NewSubjectSystemGroup
func NewSubjectSystemGroup() SubjectSystemGroupManager {
	return &subjectSystemGroupManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// GetGroups ...
func (m *subjectSystemGroupManager) GetGroups(systemID string, subjectPK int64) (groups string, err error) {
	err = m.selectGroups(&groups, systemID, subjectPK)
	return
}

func (m *subjectSystemGroupManager) selectGroups(groups *string, systemID string, subjectPK int64) error {
	query := `SELECT
		groups
		FROM subject_system_group
		WHERE system_id = ? AND subject_pk = ?`
	return database.SqlxGet(m.DB, groups, query, systemID, subjectPK)
}

// GetBySystemSubject ...
func (m *subjectSystemGroupManager) GetBySystemSubject(systemID string, subjectPK int64) (SubjectSystemGroup, error) {
	var subjectSystemGroup SubjectSystemGroup
	err := m.selectBySystemSubject(&subjectSystemGroup, systemID, subjectPK)
	return subjectSystemGroup, err
}

// Create ...
func (m *subjectSystemGroupManager) CreateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) error {
	return m.insertWithTx(tx, &subjectSystemGroup)
}

// Update ...
func (m *subjectSystemGroupManager) UpdateWithTx(tx *sqlx.Tx, subjectSystemGroup SubjectSystemGroup) (int64, error) {
	return m.updateWithTx(tx, &subjectSystemGroup)
}

// DeleteBySystemSubject ...
func (m *subjectSystemGroupManager) DeleteBySystemSubjectWithTx(tx *sqlx.Tx, systemID string, subjectPK int64) error {
	return m.deleteBySystemSubjectWithTx(tx, systemID, subjectPK)
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
		updates,
		created_at
		FROM subject_system_group
		WHERE system_id = ? AND subject_pk = ?`
	return database.SqlxGet(m.DB, subjectSystemGroup, query, systemID, subjectPK)
}

func (m *subjectSystemGroupManager) insertWithTx(tx *sqlx.Tx, subjectSystemGroup *SubjectSystemGroup) error {
	sql := `INSERT INTO subject_system_group (
		system_id, 
		subject_pk, 
		groups, 
		created_at
	) VALUES (
		:system_id,
		:subject_pk,
		:groups,
		:created_at
	)`
	return database.SqlxInsertWithTx(tx, sql, subjectSystemGroup)
}

func (m *subjectSystemGroupManager) updateWithTx(tx *sqlx.Tx, subjectSystemGroup *SubjectSystemGroup) (int64, error) {
	sql := `UPDATE subject_system_group SET
		groups = :groups,
		updates = updates + 1 
		WHERE system_id = :system_id
		AND subject_pk = :subject_pk
		AND updates = :updates`
	return database.SqlxUpdateWithTx(tx, sql, subjectSystemGroup)
}

func (m *subjectSystemGroupManager) deleteBySystemSubjectWithTx(tx *sqlx.Tx, systemID string, subjectPK int64) error {
	sql := `DELETE FROM subject_system_group WHERE system_id = ? AND subject_pk = ?`
	return database.SqlxDeleteWithTx(tx, sql, systemID, subjectPK)
}
