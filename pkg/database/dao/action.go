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

// Action 操作
type Action struct {
	PK     int64  `db:"pk"`
	System string `db:"system_id"`
	ID     string `db:"id"`
}

// ActionManager ...
type ActionManager interface {
	GetPK(system string, id string) (int64, error)
	Get(pk int64) (Action, error)
	ListByPKs(pks []int64) ([]Action, error)
	ListBySystem(system string) ([]Action, error)
	ListPKBySystem(system string) (actionPKs []int64, err error)

	BulkCreateWithTx(tx *sqlx.Tx, actions []Action) error
	BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error
}

type actionManager struct {
	DB *sqlx.DB
}

// NewActionManager ...
func NewActionManager() ActionManager {
	return &actionManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// GetPK ...
func (m *actionManager) GetPK(system, id string) (int64, error) {
	var pk int64
	err := m.selectPK(&pk, system, id)
	return pk, err
}

// Get ...
func (m *actionManager) Get(pk int64) (action Action, err error) {
	err = m.selectByPK(&action, pk)
	return
}

// ListByPKs ...
func (m *actionManager) ListByPKs(pks []int64) (actions []Action, err error) {
	if len(pks) == 0 {
		return
	}
	err = m.selectByPKs(&actions, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return actions, nil
	}
	return
}

// ListBySystem ...
func (m *actionManager) ListBySystem(system string) (actions []Action, err error) {
	err = m.selectBySystem(&actions, system)
	if errors.Is(err, sql.ErrNoRows) {
		return actions, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *actionManager) BulkCreateWithTx(tx *sqlx.Tx, actions []Action) error {
	if len(actions) == 0 {
		return nil
	}
	return m.bulkInsertWithTx(tx, actions)
}

// BulkDeleteWithTx ...
func (m *actionManager) BulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return m.bulkDeleteWithTx(tx, system, ids)
}

// ListPKBySystem ...
func (m *actionManager) ListPKBySystem(system string) (actionPKs []int64, err error) {
	query := `SELECT
		pk
		FROM action
		WHERE system_id=?`
	err = database.SqlxSelect(m.DB, &actionPKs, query, system)
	return
}

func (m *actionManager) selectPK(pk *int64, system, id string) error {
	query := `SELECT
		pk
		FROM action
		WHERE system_id = ?
		AND id = ?
		LIMIT 1`
	return database.SqlxGet(m.DB, pk, query, system, id)
}

func (m *actionManager) selectByPK(action *Action, pk int64) error {
	query := `SELECT
		pk,
		system_id,
		id
		FROM action
		WHERE pk = ?`
	return database.SqlxGet(m.DB, action, query, pk)
}

func (m *actionManager) selectByPKs(actions *[]Action, pks []int64) error {
	query := `SELECT
		pk,
		system_id,
		id
		FROM action
		WHERE pk in (?)`
	return database.SqlxSelect(m.DB, actions, query, pks)
}

func (m *actionManager) selectBySystem(actions *[]Action, system string) error {
	query := `SELECT
		pk,
		system_id,
		id
		FROM action
		WHERE system_id=?`
	return database.SqlxSelect(m.DB, actions, query, system)
}

func (m *actionManager) bulkInsertWithTx(tx *sqlx.Tx, actions []Action) error {
	query := `INSERT INTO action (system_id, id) VALUES (:system_id, :id)`
	return database.SqlxBulkInsertWithTx(tx, query, actions)
}

func (m *actionManager) bulkDeleteWithTx(tx *sqlx.Tx, system string, ids []string) error {
	query := `DELETE FROM action WHERE system_id = ? AND id IN (?)`
	return database.SqlxDeleteWithTx(tx, query, system, ids)
}
