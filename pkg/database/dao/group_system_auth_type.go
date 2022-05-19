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

// GroupSystemAuthType 用户组-系统权限类型关系
type GroupSystemAuthType struct {
	PK       int64     `db:"pk"`
	SystemID string    `db:"system_id"`
	GroupPK  int64     `db:"group_pk"`
	AuthType int64     `db:"auth_type"`
	CreateAt time.Time `db:"created_at"`
}

// GroupAuthType 用于鉴权查询
type GroupAuthType struct {
	GroupPK  int64 `db:"group_pk"`
	AuthType int64 `db:"auth_type"`
}

// GroupSystemAuthTypeManager ...
type GroupSystemAuthTypeManager interface {
	ListAuthTypeBySystemGroups(systemID string, groupPKs []int64) ([]GroupAuthType, error)

	CreateWithTx(tx *sqlx.Tx, groupSystemAuthType GroupSystemAuthType) error
	UpdateWithTx(tx *sqlx.Tx, groupSystemAuthType GroupSystemAuthType) (int64, error)
	DeleteBySystemGroupWithTx(tx *sqlx.Tx, systemID string, groupPK int64) (int64, error)
}

type groupSystemAuthTypeManager struct {
	DB *sqlx.DB
}

// NewGroupSystemAuthTypeManager New NewGroupSystemAuthTypeManager
func NewGroupSystemAuthTypeManager() GroupSystemAuthTypeManager {
	return &groupSystemAuthTypeManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListAuthTypeBySystemGroups ...
func (m *groupSystemAuthTypeManager) ListAuthTypeBySystemGroups(
	systemID string,
	groupPKs []int64,
) ([]GroupAuthType, error) {
	if len(groupPKs) == 0 {
		return nil, nil
	}

	var authTypes []GroupAuthType
	err := m.selectAuthTypeBySystemGroups(&authTypes, systemID, groupPKs)
	return authTypes, err
}

// CreateWithTx ...
func (m *groupSystemAuthTypeManager) CreateWithTx(tx *sqlx.Tx, groupSystemAuthType GroupSystemAuthType) error {
	return m.insertWithTx(tx, &groupSystemAuthType)
}

// UpdateWithTx ...
func (m *groupSystemAuthTypeManager) UpdateWithTx(tx *sqlx.Tx, groupSystemAuthType GroupSystemAuthType) (int64, error) {
	return m.updateWithTx(tx, &groupSystemAuthType)
}

// DeleteByGroupSystem ..
func (m *groupSystemAuthTypeManager) DeleteBySystemGroupWithTx(tx *sqlx.Tx, systemID string, groupPK int64) (int64, error) {
	return m.deleteBySystemGroupWithTx(tx, systemID, groupPK)
}

func (m *groupSystemAuthTypeManager) selectAuthTypeBySystemGroups(
	authTypes *[]GroupAuthType,
	systemID string,
	groupPKs []int64,
) error {
	query := `SELECT
		group_pk,
		auth_type
		FROM group_system_auth_type
		WHERE system_id = ? AND group_pk in (?)`
	return database.SqlxSelect(m.DB, authTypes, query, systemID, groupPKs)
}

func (m *groupSystemAuthTypeManager) insertWithTx(tx *sqlx.Tx, groupSystemAuthType *GroupSystemAuthType) error {
	sql := `INSERT INTO group_system_auth_type (
		system_id,
		group_pk,
		auth_type,
		created_at
	) VALUES (
		:system_id,
		:group_pk,
		:auth_type,
		:created_at)`
	return database.SqlxInsertWithTx(tx, sql, groupSystemAuthType)
}

func (m *groupSystemAuthTypeManager) updateWithTx(tx *sqlx.Tx, groupSystemAuthType *GroupSystemAuthType) (int64, error) {
	sql := `UPDATE group_system_auth_type SET
		auth_type = :auth_type 
		WHERE system_id = :system_id 
		AND group_pk = :group_pk 
		AND auth_type <> :auth_type`
	return database.SqlxUpdateWithTx(tx, sql, groupSystemAuthType)
}

func (m *groupSystemAuthTypeManager) deleteBySystemGroupWithTx(tx *sqlx.Tx, systemID string, groupPK int64) (int64, error) {
	query := `DELETE FROM group_system_auth_type WHERE system_id = ? AND group_pk = ?`
	rows, err := database.SqlxDeleteReturnRowsWithTx(tx, query, systemID, groupPK)
	return rows, err
}
