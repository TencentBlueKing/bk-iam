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

// SaaSResourceType ...
type SaaSResourceType struct {
	database.AllowBlankFields

	PK             int64  `db:"pk"`
	System         string `db:"system_id"`
	ID             string `db:"id"`
	Name           string `db:"name"`
	NameEn         string `db:"name_en"`
	Description    string `db:"description"`
	DescriptionEn  string `db:"description_en"`
	Sensitivity    int64  `db:"sensitivity"`
	Parents        string `db:"parents"`         // JSON
	ProviderConfig string `db:"provider_config"` // JSON 'iam,saas_iam'
	Version        int64  `db:"version"`
}

// SaaSResourceTypeManager ...
type SaaSResourceTypeManager interface {
	ListBySystem(system string) ([]SaaSResourceType, error)

	BulkCreateWithTx(tx *sqlx.Tx, saasResourceTypes []SaaSResourceType) error
	Update(system, resourceTypeID string, sys SaaSResourceType) error
	BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error

	Get(system, resourceTypeID string) (SaaSResourceType, error)
}

type saasResourceTypeManager struct {
	DB *sqlx.DB
}

// NewSaaSResourceTypeManager ...
func NewSaaSResourceTypeManager() SaaSResourceTypeManager {
	return &saasResourceTypeManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListBySystem ...
func (m *saasResourceTypeManager) ListBySystem(system string) (saasResourceTypes []SaaSResourceType, err error) {
	err = m.selectBySystem(&saasResourceTypes, system)
	if errors.Is(err, sql.ErrNoRows) {
		return saasResourceTypes, nil
	}
	return
}

// Get ...
func (m *saasResourceTypeManager) Get(system string, resourceTypeID string) (SaaSResourceType, error) {
	srt := SaaSResourceType{}
	err := m.selectByID(&srt, system, resourceTypeID)
	return srt, err
}

// BulkCreateWithTx ...
func (m *saasResourceTypeManager) BulkCreateWithTx(tx *sqlx.Tx, saasResourceTypes []SaaSResourceType) error {
	if len(saasResourceTypes) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, saasResourceTypes)
}

// Update ...
func (m *saasResourceTypeManager) Update(system, resourceTypeID string, rt SaaSResourceType) error {
	// 1. parse the set sql string and update data
	expr, data, err := database.ParseUpdateStruct(rt, rt.AllowBlankFields)
	if err != nil {
		return fmt.Errorf("parse update struct fail. %w", err)
	}

	// 2. build sql
	sql := "UPDATE saas_resource_type SET " + expr + " WHERE system_id=:system_id AND id=:id"

	// 3. add the where data
	data["system_id"] = system
	data["id"] = resourceTypeID

	return m.update(sql, data)
}

// BulkDeleteWithTx ...
func (m *saasResourceTypeManager) BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, system, ids)
}

func (m *saasResourceTypeManager) bulkInsertWithTx(tx *sqlx.Tx, saasResourceTypes []SaaSResourceType) error {
	query := `INSERT INTO saas_resource_type (
		system_id,
		id,
		name,
		name_en,
		description,
		description_en,
		sensitivity,
		parents,
		provider_config,
		version
	) VALUES (
		:system_id,
		:id,
		:name,
		:name_en,
		:description,
		:description_en,
		:sensitivity,
		:parents,
		:provider_config,
		:version
	)`
	return database.SqlxBulkInsertWithTx(tx, query, saasResourceTypes)
}

func (m *saasResourceTypeManager) bulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	query := `DELETE FROM saas_resource_type WHERE system_id = ? AND id IN (?)`
	return database.SqlxDeleteWithTx(tx, query, system, ids)
}

func (m *saasResourceTypeManager) update(sql string, data map[string]interface{}) error {
	_, err := database.SqlxUpdate(m.DB, sql, data)
	if err != nil {
		return err
	}
	return nil
}

func (m *saasResourceTypeManager) selectBySystem(saasResourceTypes *[]SaaSResourceType, system string) error {
	query := `SELECT
		pk,
		system_id,
		id,
		name,
		name_en,
		description,
		description_en,
		sensitivity,
		parents,
		provider_config,
		version
		FROM saas_resource_type
		WHERE system_id = ?
		ORDER BY pk`
	return database.SqlxSelect(m.DB, saasResourceTypes, query, system)
}

func (m *saasResourceTypeManager) selectByID(srt *SaaSResourceType, system, id string) error {
	query := `SELECT
		pk,
		system_id,
		id,
		name,
		name_en,
		description,
		description_en,
		sensitivity,
		parents,
		provider_config,
		version
		FROM saas_resource_type
		WHERE system_id = ?
		AND id = ?`
	return database.SqlxGet(m.DB, srt, query, system, id)
}
