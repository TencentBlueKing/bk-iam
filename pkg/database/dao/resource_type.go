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
	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// ResourceTypeManager ...
type ResourceTypeManager interface {
	GetPK(system string, id string) (pk int64, err error)
	GetByPK(pk int64) (ResourceType, error)
	BulkCreateWithTx(tx *sqlx.Tx, resourceTypes []ResourceType) error
	BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error
}

type resourceTypeManager struct {
	DB *sqlx.DB
}

// ResourceType 资源类型
type ResourceType struct {
	PK     int64  `db:"pk"`
	System string `db:"system_id"`
	ID     string `db:"id"`
}

// NewResourceTypeManager ...
func NewResourceTypeManager() ResourceTypeManager {
	return &resourceTypeManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// GetPK ...
func (m *resourceTypeManager) GetPK(system string, id string) (pk int64, err error) {
	query := "SELECT pk FROM resource_type WHERE system_id = ? AND id = ?"
	err = database.SqlxGet(m.DB, &pk, query, system, id)
	return
}

// GetByPK ...
func (m *resourceTypeManager) GetByPK(pk int64) (resourceType ResourceType, err error) {
	query := "SELECT pk, system_id, id FROM resource_type WHERE pk=?"
	err = database.SqlxGet(m.DB, &resourceType, query, pk)
	return
}

// BulkCreateWithTx ...
func (m *resourceTypeManager) BulkCreateWithTx(tx *sqlx.Tx, resourceTypes []ResourceType) error {
	if len(resourceTypes) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, resourceTypes)
}

// BulkDeleteWithTx ...
func (m *resourceTypeManager) BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, system, ids)
}

func (m *resourceTypeManager) bulkInsertWithTx(tx *sqlx.Tx, resourceTypes []ResourceType) error {
	query := `INSERT INTO resource_type (system_id, id) VALUES (:system_id, :id)`
	return database.SqlxBulkInsertWithTx(tx, query, resourceTypes)
}

func (m *resourceTypeManager) bulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	query := `DELETE FROM resource_type WHERE system_id = ? AND id IN (?)`
	return database.SqlxDeleteWithTx(tx, query, system, ids)
}
