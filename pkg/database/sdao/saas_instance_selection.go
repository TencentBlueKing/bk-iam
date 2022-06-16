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

// SaaSInstanceSelection ...
type SaaSInstanceSelection struct {
	database.AllowBlankFields

	PK                int64  `db:"pk"`
	System            string `db:"system_id"`
	ID                string `db:"id"`
	Name              string `db:"name"`
	NameEn            string `db:"name_en"`
	IsDynamic         bool   `db:"is_dynamic"`
	ResourceTypeChain string `db:"resource_type_chain"` // JSON
}

// SaaSInstanceSelectionManager ...
type SaaSInstanceSelectionManager interface {
	Get(system, id string) (SaaSInstanceSelection, error)
	ListBySystem(system string) ([]SaaSInstanceSelection, error)

	BulkCreateWithTx(tx *sqlx.Tx, saasInstanceSelections []SaaSInstanceSelection) error
	Update(system, instanceSelectionID string, sis SaaSInstanceSelection) error
	BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error
}

type saasInstanceSelectionManager struct {
	DB *sqlx.DB
}

// NewSaaSInstanceSelectionManager ...
func NewSaaSInstanceSelectionManager() SaaSInstanceSelectionManager {
	return &saasInstanceSelectionManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *saasInstanceSelectionManager) Get(system, id string) (instanceSelection SaaSInstanceSelection, err error) {
	err = m.getByID(&instanceSelection, system, id)
	return
}

// ListBySystem ...
func (m *saasInstanceSelectionManager) ListBySystem(system string) (
	saasInstanceSelections []SaaSInstanceSelection, err error,
) {
	err = m.selectBySystem(&saasInstanceSelections, system)
	if errors.Is(err, sql.ErrNoRows) {
		return saasInstanceSelections, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *saasInstanceSelectionManager) BulkCreateWithTx(
	tx *sqlx.Tx,
	saasInstanceSelections []SaaSInstanceSelection,
) error {
	if len(saasInstanceSelections) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, saasInstanceSelections)
}

// Update ...
func (m *saasInstanceSelectionManager) Update(system, instanceSelectionID string, sis SaaSInstanceSelection) error {
	// 1. parse the set sql string and update data
	expr, data, err := database.ParseUpdateStruct(sis, sis.AllowBlankFields)
	if err != nil {
		return fmt.Errorf("parse update struct fail. %w", err)
	}
	// if all fields are blank, the parsed expr will be empty string, return, otherwise will SQL syntax error
	if expr == "" {
		return nil
	}

	// 2. build sql
	sql := "UPDATE saas_instance_selection SET " + expr + " WHERE system_id=:system_id AND id=:id"

	// 3. add the where data
	data["system_id"] = system
	data["id"] = instanceSelectionID

	return m.update(sql, data)
}

// BulkDeleteWithTx ...
func (m *saasInstanceSelectionManager) BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, system, ids)
}

func (m *saasInstanceSelectionManager) getByID(instanceSelection *SaaSInstanceSelection, system, id string) error {
	query := `SELECT
		pk,
		system_id,
		id,
		name,
		name_en,
		is_dynamic,
		resource_type_chain
		FROM saas_instance_selection
		WHERE system_id = ?
		AND id = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, instanceSelection, query, system, id)
}

func (m *saasInstanceSelectionManager) selectBySystem(
	saasInstanceSelections *[]SaaSInstanceSelection,
	system string,
) error {
	query := `SELECT
		pk,
		system_id,
		id,
		name,
		name_en,
		is_dynamic,
		resource_type_chain
		FROM saas_instance_selection
		WHERE system_id = ?`
	return database.SqlxSelect(m.DB, saasInstanceSelections, query, system)
}

func (m *saasInstanceSelectionManager) bulkInsertWithTx(
	tx *sqlx.Tx,
	saasInstanceSelections []SaaSInstanceSelection,
) error {
	query := `INSERT INTO saas_instance_selection (
		system_id,
		id,
		name,
		name_en,
		is_dynamic,
		resource_type_chain
	) VALUES (:system_id, :id, :name, :name_en, :is_dynamic, :resource_type_chain)`
	return database.SqlxBulkInsertWithTx(tx, query, saasInstanceSelections)
}

func (m *saasInstanceSelectionManager) update(sql string, data map[string]interface{}) error {
	_, err := database.SqlxUpdate(m.DB, sql, data)
	if err != nil {
		return err
	}
	return nil
}

func (m *saasInstanceSelectionManager) bulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	query := `DELETE FROM saas_instance_selection WHERE system_id = ? AND id IN (?)`
	return database.SqlxDeleteWithTx(tx, query, system, ids)
}
