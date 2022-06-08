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

// ActionResourceType 操作与类型的关联
type ActionResourceType struct {
	PK                 int64  `db:"pk"`
	ActionSystem       string `db:"action_system_id"`
	ActionID           string `db:"action_id"`
	ResourceTypeSystem string `db:"resource_type_system_id"`
	ResourceTypeID     string `db:"resource_type_id"`
	// TODO: should remove ScopeExpression from this table
}

// ActionResourceTypeManager ...
type ActionResourceTypeManager interface {
	ListByActionSystem(actionSystem string) ([]ActionResourceType, error)
	ListResourceTypeByAction(actionSystem string, actionID string) ([]ActionResourceType, error)
	ListByResourceTypeSystem(resourceTypeSystem string) ([]ActionResourceType, error)

	BulkCreateWithTx(tx *sqlx.Tx, actionResourceTypes []ActionResourceType) error
	BulkDeleteWithTx(tx *sqlx.Tx, actionSystem string, actionIDs []string) error
}

type actionResourceTypeManager struct {
	DB *sqlx.DB
}

// NewActionResourceTypeManager ...
func NewActionResourceTypeManager() ActionResourceTypeManager {
	return &actionResourceTypeManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListByActionSystem ...
func (m *actionResourceTypeManager) ListByActionSystem(actionSystem string) (
	actionResourceTypes []ActionResourceType, err error,
) {
	err = m.selectByActionSystem(&actionResourceTypes, actionSystem)
	// 吞掉记录不存在的错误, action本身是可以不关联任何resource type
	if errors.Is(err, sql.ErrNoRows) {
		return actionResourceTypes, nil
	}
	return
}

// ListResourceTypeByAction ...
func (m *actionResourceTypeManager) ListResourceTypeByAction(
	actionSystem, actionID string,
) (actionResourceTypes []ActionResourceType, err error) {
	err = m.selectResourceTypeByAction(&actionResourceTypes, actionSystem, actionID)
	// 吞掉记录不存在的错误, action本身是可以不关联任何resource type
	if errors.Is(err, sql.ErrNoRows) {
		return actionResourceTypes, nil
	}
	return
}

// ListByResourceTypeSystem ...
func (m *actionResourceTypeManager) ListByResourceTypeSystem(resourceTypeSystem string) (
	actionResourceTypes []ActionResourceType, err error,
) {
	err = m.selectByResourceTypeSystem(&actionResourceTypes, resourceTypeSystem)
	// 吞掉记录不存在的错误, action本身是可以不关联任何resource type
	if errors.Is(err, sql.ErrNoRows) {
		return actionResourceTypes, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *actionResourceTypeManager) BulkCreateWithTx(tx *sqlx.Tx, actionResourceTypes []ActionResourceType) error {
	if len(actionResourceTypes) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, actionResourceTypes)
}

// BulkDeleteWithTx ...
func (m *actionResourceTypeManager) BulkDeleteWithTx(tx *sqlx.Tx, actionSystem string, actionIDs []string) error {
	if len(actionIDs) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, actionSystem, actionIDs)
}

func (m *actionResourceTypeManager) selectResourceTypeByAction(
	actionResourceTypes *[]ActionResourceType, actionSystem, actionID string,
) error {
	query := `SELECT
		resource_type_system_id,
		resource_type_id
		FROM action_resource_type
		WHERE action_system_id = ?
		AND action_id = ?`
	return database.SqlxSelect(m.DB, actionResourceTypes, query, actionSystem, actionID)
}

func (m *actionResourceTypeManager) bulkInsertWithTx(tx *sqlx.Tx, actionResourceTypes []ActionResourceType) error {
	query := `INSERT INTO action_resource_type (
		action_system_id,
		action_id,
		resource_type_system_id,
		resource_type_id
	) VALUES (:action_system_id, :action_id, :resource_type_system_id, :resource_type_id)`
	return database.SqlxBulkInsertWithTx(tx, query, actionResourceTypes)
}

func (m *actionResourceTypeManager) bulkDeleteWithTx(tx *sqlx.Tx, actionSystem string, actionIDs []string) error {
	query := `DELETE FROM action_resource_type WHERE action_system_id = ? AND action_id IN (?)`
	return database.SqlxDeleteWithTx(tx, query, actionSystem, actionIDs)
}

func (m *actionResourceTypeManager) selectByActionSystem(
	actionResourceTypes *[]ActionResourceType, actionSystem string,
) error {
	query := `SELECT
		pk,
		action_system_id,
		action_id,
		resource_type_system_id,
		resource_type_id
		FROM action_resource_type
		WHERE action_system_id = ?
		ORDER BY pk`
	return database.SqlxSelect(m.DB, actionResourceTypes, query, actionSystem)
}

func (m *actionResourceTypeManager) selectByResourceTypeSystem(
	actionResourceTypes *[]ActionResourceType, resourceTypeSystem string,
) error {
	query := `SELECT
		pk,
		action_system_id,
		action_id,
		resource_type_system_id,
		resource_type_id
		FROM action_resource_type
		WHERE resource_type_system_id = ?`
	return database.SqlxSelect(m.DB, actionResourceTypes, query, resourceTypeSystem)
}
