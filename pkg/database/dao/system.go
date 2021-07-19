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
	"iam/pkg/database"

	"github.com/jmoiron/sqlx"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// System ...
type System struct {
	ID string `db:"id"`
}

// SystemManager ...
type SystemManager interface {
	Get(id string) (System, error)

	CreateWithTx(tx *sqlx.Tx, system System) error
}

type systemManager struct {
	DB *sqlx.DB
}

// NewSystemManager ...
func NewSystemManager() SystemManager {
	return &systemManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *systemManager) Get(id string) (system System, err error) {
	err = m.selectOne(&system, id)
	return system, err
}

// CreateWithTx ...
func (m *systemManager) CreateWithTx(tx *sqlx.Tx, system System) error {
	return m.insertWithTx(tx, system)
}

func (m *systemManager) insertWithTx(tx *sqlx.Tx, system System) error {
	query := `INSERT INTO system_info (id) VALUES (:id)`
	return database.SqlxInsertWithTx(tx, query, system)
}

func (m *systemManager) selectOne(system *System, id string) error {
	query := `SELECT id FROM system_info where id = ? LIMIT 1`
	return database.SqlxGet(m.DB, system, query, id)
}
