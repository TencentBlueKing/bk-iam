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
	"fmt"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SaaSAction ...
type SaaSAction struct {
	database.AllowBlankFields

	PK             int64  `db:"pk"`
	System         string `db:"system_id"`
	ID             string `db:"id"`
	Name           string `db:"name"`
	NameEn         string `db:"name_en"`
	Description    string `db:"description"`
	DescriptionEn  string `db:"description_en"`
	RelatedActions string `db:"related_actions"`
	Type           string `db:"type"`
	Version        int64  `db:"version"`
}

// SaaSActionManager ...
type SaaSActionManager interface {
	Get(system, actionID string) (SaaSAction, error)
	ListBySystem(system string) ([]SaaSAction, error)

	BulkCreateWithTx(tx *sqlx.Tx, saasActions []SaaSAction) error
	Update(tx *sqlx.Tx, system, actionID string, saasAction SaaSAction) error
	BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error
}

type saasActionManager struct {
	DB *sqlx.DB
}

// NewSaaSActionManager ...
func NewSaaSActionManager() SaaSActionManager {
	return &saasActionManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *saasActionManager) Get(system, actionID string) (action SaaSAction, err error) {
	err = m.getByActionID(&action, system, actionID)
	return
}

// ListBySystem ...
func (m *saasActionManager) ListBySystem(system string) (saasAction []SaaSAction, err error) {
	err = m.selectBySystem(&saasAction, system)
	if errors.Is(err, sql.ErrNoRows) {
		return saasAction, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *saasActionManager) BulkCreateWithTx(tx *sqlx.Tx, saasActions []SaaSAction) error {
	if len(saasActions) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, saasActions)
}

// Update ...
func (m *saasActionManager) Update(tx *sqlx.Tx, system, actionID string,
	saasAction SaaSAction) error {
	// 1. parse the set sql string and update data
	expr, data, err := database.ParseUpdateStruct(saasAction, saasAction.AllowBlankFields)
	if err != nil {
		return fmt.Errorf("parse update struct fail. %w", err)
	}

	// 2. build sql
	sql := "UPDATE saas_action SET " + expr + " WHERE system_id=:system_id AND id=:id"

	// 3. add the where data
	data["system_id"] = system
	data["id"] = actionID

	return m.updateWithTx(tx, sql, data)
}

// BulkDeleteWithTx ...
func (m *saasActionManager) BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, system, ids)
}

func (m *saasActionManager) bulkInsertWithTx(tx *sqlx.Tx, saasActions []SaaSAction) error {
	query := `INSERT INTO saas_action (
		system_id,
		id,
		name,
		name_en,
		description,
		description_en,
		related_actions,
		type,
		version
	) VALUES (:system_id, :id, :name, :name_en, :description, :description_en, :related_actions, :type, :version)`
	return database.SqlxBulkInsertWithTx(tx, query, saasActions)
}

func (m *saasActionManager) bulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	query := `DELETE FROM saas_action WHERE system_id = ? AND id IN (?)`
	return database.SqlxDeleteWithTx(tx, query, system, ids)
}

func (m *saasActionManager) updateWithTx(tx *sqlx.Tx, sql string, data map[string]interface{}) error {
	_, err := database.SqlxUpdateWithTx(tx, sql, data)
	if err != nil {
		return err
	}
	return nil
}

func (m *saasActionManager) selectBySystem(saasAction *[]SaaSAction, system string) error {
	query := `SELECT
		pk,
		system_id,
		id,
		name,
		name_en,
		description,
		description_en,
		related_actions,
		type,
		version
		FROM saas_action
		WHERE system_id = ?
		ORDER BY pk`
	return database.SqlxSelect(m.DB, saasAction, query, system)
}

func (m *saasActionManager) getByActionID(saasAction *SaaSAction, system, actionID string) error {
	query := `SELECT
		pk,
		system_id,
		id,
		name,
		name_en,
		type,
		version
		FROM saas_action
		WHERE system_id = ?
		AND id = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, saasAction, query, system, actionID)
}
