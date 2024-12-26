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

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// SubjectActionGroupResource ...
type SubjectActionGroupResource struct {
	PK int64 `db:"pk"`

	SubjectPK     int64  `db:"subject_pk"`
	ActionPK      int64  `db:"action_pk"`
	GroupResource string `db:"group_resource"`
}

// SubjectActionGroupResourceManager ...
type SubjectActionGroupResourceManager interface {
	GetBySubjectAction(subjectPK, actionPK int64) (SubjectActionGroupResource, error)
	CreateWithTx(tx *sqlx.Tx, subjectActionResourceGroup SubjectActionGroupResource) error
	UpdateGroupResourceWithTx(tx *sqlx.Tx, pk int64, groupResource string) error
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error
	DeleteBySubjectActionWithTx(tx *sqlx.Tx, subjectPK, actionPK int64) error

	HasAnyByActionPK(actionPK int64) (exist bool, err error)
	DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK, limit int64) (int64, error)
}

type subjectActionGroupResourceManager struct {
	DB *sqlx.DB
}

// NewSubjectActionGroupResourceManager ...
func NewSubjectActionGroupResourceManager() SubjectActionGroupResourceManager {
	return &subjectActionGroupResourceManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// GetBySubjectAction ...
func (m *subjectActionGroupResourceManager) GetBySubjectAction(
	subjectPK, actionPK int64,
) (subjectActionResourceGroup SubjectActionGroupResource, err error) {
	query := `SELECT 
		pk,
		subject_pk,
		action_pk,
		group_resource
		FROM rbac_subject_action_group_resource
		WHERE subject_pk = ?
		AND action_pk = ? LIMIT 1`
	err = database.SqlxGet(m.DB, &subjectActionResourceGroup, query, subjectPK, actionPK)
	return
}

// CreateWithTx ...
func (m *subjectActionGroupResourceManager) CreateWithTx(
	tx *sqlx.Tx,
	subjectActionResourceGroup SubjectActionGroupResource,
) error {
	sql := `INSERT INTO rbac_subject_action_group_resource (
		subject_pk,
		action_pk,
		group_resource
	) VALUES (
		:subject_pk,
		:action_pk,
		:group_resource
	)`
	return database.SqlxInsertWithTx(tx, sql, subjectActionResourceGroup)
}

// UpdateGroupResourceWithTx ...
func (m *subjectActionGroupResourceManager) UpdateGroupResourceWithTx(
	tx *sqlx.Tx,
	pk int64,
	groupResource string,
) error {
	sql := `UPDATE rbac_subject_action_group_resource SET group_resource = :group_resource WHERE pk = :pk`
	_, err := database.SqlxUpdateWithTx(tx, sql, map[string]interface{}{
		"pk":             pk,
		"group_resource": groupResource,
	})
	return err
}

// BulkDeleteBySubjectPKsWithTx ...
func (m *subjectActionGroupResourceManager) BulkDeleteBySubjectPKsWithTx(
	tx *sqlx.Tx,
	subjectPKs []int64,
) error {
	if len(subjectPKs) == 0 {
		return nil
	}
	sql := `DELETE FROM rbac_subject_action_group_resource WHERE subject_pk IN (?)`
	return database.SqlxDeleteWithTx(tx, sql, subjectPKs)
}

// HasAnyByActionPK ...
func (m *subjectActionGroupResourceManager) HasAnyByActionPK(actionPK int64) (exist bool, err error) {
	var pk int64
	query := `SELECT
		pk
		FROM rbac_subject_action_group_resource
		WHERE action_pk = ?
		LIMIT 1`
	err = database.SqlxGet(m.DB, &pk, query, actionPK)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteByActionPKWithTx ...
func (m *subjectActionGroupResourceManager) DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK, limit int64) (int64, error) {
	sql := `DELETE FROM rbac_subject_action_group_resource WHERE action_pk = ? LIMIT ?`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, actionPK, limit)
}

// DeleteBySubjectActionWithTx ...
func (m *subjectActionGroupResourceManager) DeleteBySubjectActionWithTx(tx *sqlx.Tx, subjectPK, actionPK int64) error {
	sql := `DELETE FROM rbac_subject_action_group_resource WHERE subject_pk = ? AND action_pk = ?`
	return database.SqlxDeleteWithTx(tx, sql, subjectPK, actionPK)
}
