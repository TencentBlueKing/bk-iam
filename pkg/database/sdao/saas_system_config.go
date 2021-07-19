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
	"fmt"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// SaaSSystemConfig ...
type SaaSSystemConfig struct {
	database.AllowBlankFields

	PK     int64  `db:"pk"`
	System string `db:"system_id"`
	Name   string `db:"name"`
	Type   string `db:"type"`
	Value  string `db:"value"`
}

// SaaSSystemConfigManager ...
type SaaSSystemConfigManager interface {
	Get(system string, name string) (SaaSSystemConfig, error)

	Create(systemConfig SaaSSystemConfig) error
	Update(systemConfig SaaSSystemConfig) error
}

type saasSystemConfigManager struct {
	DB *sqlx.DB
}

// NewSaaSSystemConfigManager ...
func NewSaaSSystemConfigManager() SaaSSystemConfigManager {
	return &saasSystemConfigManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *saasSystemConfigManager) Get(system string, name string) (systemConfig SaaSSystemConfig, err error) {
	err = m.selectOne(&systemConfig, system, name)
	return
}

// Create ...
func (m *saasSystemConfigManager) Create(systemConfig SaaSSystemConfig) error {
	return m.insert(systemConfig)
}

func (m *saasSystemConfigManager) selectOne(systemConfig *SaaSSystemConfig, system string, name string) error {
	query := `SELECT
		pk,
		system_id,
		name,
		type,
		value
		FROM saas_system_config
		WHERE system_id = ?
		AND name = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, systemConfig, query, system, name)
}

// Update ...
func (m *saasSystemConfigManager) Update(systemConfig SaaSSystemConfig) error {
	// 1. parse the set sql string and update data
	expr, data, err := database.ParseUpdateStruct(systemConfig, systemConfig.AllowBlankFields)
	if err != nil {
		return fmt.Errorf("parse update struct fail. %w", err)
	}

	// 2. build sql
	sql := "UPDATE saas_system_config SET " + expr + " WHERE system_id=:system_id AND name=:name"

	// 3. add the where data
	//data["system_id"] = system
	//data["name"] = name

	return m.update(sql, data)
}

func (m *saasSystemConfigManager) insert(systemConfig SaaSSystemConfig) error {
	query := `INSERT INTO saas_system_config (
		system_id,
		name,
		type,
		value
	) VALUES (:system_id, :name, :type, :value)`
	return database.SqlxBulkInsert(m.DB, query, []SaaSSystemConfig{systemConfig})
}

func (m *saasSystemConfigManager) update(sql string, data map[string]interface{}) error {
	_, err := database.SqlxUpdate(m.DB, sql, data)
	if err != nil {
		return err
	}
	return nil
}
