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

// SaaSSystem ...
type SaaSSystem struct {
	database.AllowBlankFields

	ID             string `db:"id"`
	Name           string `db:"name"`
	NameEn         string `db:"name_en"`
	Description    string `db:"description"`
	DescriptionEn  string `db:"description_en"`
	Clients        string `db:"clients"`         // 逗号分隔
	ProviderConfig string `db:"provider_config"` // JSON 'iam,saas_iam'
}

// SaaSSystemManager ...
type SaaSSystemManager interface {
	Get(id string) (SaaSSystem, error)
	ListAll() ([]SaaSSystem, error)

	CreateWithTx(tx *sqlx.Tx, system SaaSSystem) error
	Update(id string, system SaaSSystem) error
}

type saasSystemManager struct {
	DB *sqlx.DB
}

// NewSaaSSystemManager ...
func NewSaaSSystemManager() SaaSSystemManager {
	return &saasSystemManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *saasSystemManager) Get(id string) (saasSystem SaaSSystem, err error) {
	err = m.selectOne(&saasSystem, id)
	return
}

// ListAll ...
func (m *saasSystemManager) ListAll() (saasSystems []SaaSSystem, err error) {
	err = m.selectAll(&saasSystems)
	if errors.Is(err, sql.ErrNoRows) {
		return saasSystems, nil
	}
	return
}

// CreateWithTx ...
func (m *saasSystemManager) CreateWithTx(tx *sqlx.Tx, system SaaSSystem) error {
	return m.insertWithTx(tx, system)
}

// Update ...
func (m *saasSystemManager) Update(id string, system SaaSSystem) error {
	// 1. parse the set sql string and update data
	expr, data, err := database.ParseUpdateStruct(system, system.AllowBlankFields)
	if err != nil {
		return fmt.Errorf("parse update struct fail. %w", err)
	}
	// if all fields are blank, the parsed expr will be empty string, return, otherwise will SQL syntax error
	if expr == "" {
		return nil
	}

	// 2. build sql
	sql := "UPDATE saas_system_info SET " + expr + " WHERE id=:id"

	// 3. add the where data
	data["id"] = id

	return m.update(sql, data)
}

func (m *saasSystemManager) insertWithTx(tx *sqlx.Tx, system SaaSSystem) error {
	query := `INSERT INTO saas_system_info (
		id,
		name,
		name_en,
        description,
        description_en,
		clients,
		provider_config
	) VALUES (:id, :name, :name_en, :description, :description_en, :clients, :provider_config)`
	return database.SqlxInsertWithTx(tx, query, system)
}

func (m *saasSystemManager) update(sql string, data map[string]interface{}) error {
	_, err := database.SqlxUpdate(m.DB, sql, data)
	if err != nil {
		return err
	}
	return nil
}

func (m *saasSystemManager) selectOne(saasSystem *SaaSSystem, id string) error {
	query := `SELECT
		id,
		name,
		name_en,
		description,
		description_en,
		clients,
		provider_config
		FROM saas_system_info
		WHERE id = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, saasSystem, query, id)
}

func (m *saasSystemManager) selectAll(saasSystems *[]SaaSSystem) error {
	query := `SELECT
		id,
		name,
		name_en,
		description,
		description_en,
		clients,
		provider_config
		FROM saas_system_info ORDER BY created_at`
	return database.SqlxSelect(m.DB, saasSystems, query)
}
