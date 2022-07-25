/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package sdao

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SaaSActionResourceType ...
type SaaSActionResourceType struct {
	PK                        int64  `db:"pk"`
	ActionSystem              string `db:"action_system_id"`
	ActionID                  string `db:"action_id"`
	ResourceTypeSystem        string `db:"resource_type_system_id"`
	ResourceTypeID            string `db:"resource_type_id"`
	NameAlias                 string `db:"name_alias"`
	NameAliasEn               string `db:"name_alias_en"`
	SelectionMode             string `db:"selection_mode"`
	RelatedInstanceSelections string `db:"related_instance_selections"` // JSON
}

// SaaSActionResourceTypeManager ...
type SaaSActionResourceTypeManager interface {
	ListByActionSystem(actionSystem string) ([]SaaSActionResourceType, error)
	ListByActionID(system, actionID string) ([]SaaSActionResourceType, error)

	BulkCreateWithTx(tx *sqlx.Tx, saasActionResourceTypes []SaaSActionResourceType) error
	BulkDeleteWithTx(tx *sqlx.Tx, actionSystem string, actionIDs []string) error
}

type saasActionResourceTypeManager struct {
	DB *sqlx.DB
}

// NewSaaSActionResourceTypeManager ...
func NewSaaSActionResourceTypeManager() SaaSActionResourceTypeManager {
	return &saasActionResourceTypeManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListByActionSystem ...
func (m *saasActionResourceTypeManager) ListByActionSystem(actionSystem string) (
	saaSActionResourceTypes []SaaSActionResourceType, err error,
) {
	err = m.selectByActionSystem(&saaSActionResourceTypes, actionSystem)
	if errors.Is(err, sql.ErrNoRows) {
		return saaSActionResourceTypes, nil
	}
	return
}

// ListByActionID ...
func (m *saasActionResourceTypeManager) ListByActionID(system, actionID string) (
	saaSActionResourceTypes []SaaSActionResourceType, err error,
) {
	err = m.selectByActionID(&saaSActionResourceTypes, system, actionID)
	if errors.Is(err, sql.ErrNoRows) {
		return saaSActionResourceTypes, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *saasActionResourceTypeManager) BulkCreateWithTx(
	tx *sqlx.Tx, saasActionResourceTypes []SaaSActionResourceType,
) error {
	if len(saasActionResourceTypes) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, saasActionResourceTypes)
}

// BulkDeleteWithTx ...
func (m *saasActionResourceTypeManager) BulkDeleteWithTx(tx *sqlx.Tx, actionSystem string, actionIDs []string) error {
	if len(actionIDs) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, actionSystem, actionIDs)
}

func (m *saasActionResourceTypeManager) bulkInsertWithTx(
	tx *sqlx.Tx, saasActionResourceTypes []SaaSActionResourceType,
) error {
	query := `INSERT INTO saas_action_resource_type (
		action_system_id,
		action_id,
		resource_type_system_id,
		resource_type_id,
		name_alias,
		name_alias_en,
		selection_mode,
		related_instance_selections
	) VALUES (
		:action_system_id,
		:action_id,
		:resource_type_system_id,
		:resource_type_id,
		:name_alias,
		:name_alias_en,
		:selection_mode,
		:related_instance_selections)`
	return database.SqlxBulkInsertWithTx(tx, query, saasActionResourceTypes)
}

func (m *saasActionResourceTypeManager) bulkDeleteWithTx(tx *sqlx.Tx, actionSystem string, actionIDs []string) error {
	query := `DELETE FROM saas_action_resource_type WHERE action_system_id = ? AND action_id IN (?)`
	return database.SqlxDeleteWithTx(tx, query, actionSystem, actionIDs)
}

func (m *saasActionResourceTypeManager) selectByActionSystem(
	saaSActionResourceTypes *[]SaaSActionResourceType, actionSystem string,
) error {
	query := `SELECT
		pk,
		action_system_id,
		action_id,
		resource_type_system_id,
		resource_type_id,
		name_alias,
		name_alias_en,
		selection_mode,
		related_instance_selections
		FROM saas_action_resource_type
		WHERE action_system_id = ?`
	return database.SqlxSelect(m.DB, saaSActionResourceTypes, query, actionSystem)
}

func (m *saasActionResourceTypeManager) selectByActionID(
	saaSActionResourceTypes *[]SaaSActionResourceType, system, actionID string,
) error {
	query := `SELECT
		pk,
		action_system_id,
		action_id,
		resource_type_system_id,
		resource_type_id,
		name_alias,
		name_alias_en,
		selection_mode,
		related_instance_selections
		FROM saas_action_resource_type
		WHERE action_system_id = ?
		AND action_id = ?`
	return database.SqlxSelect(m.DB, saaSActionResourceTypes, query, system, actionID)
}
