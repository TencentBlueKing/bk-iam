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

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// GroupAlterEvent ...
type GroupAlterEvent struct {
	PK int64 `db:"pk"`

	GroupPK    int64  `db:"group_pk"`
	ActionPKs  string `db:"action_pks"`
	SubjectPKs string `db:"subject_pks"`
	Status     int64  `db:"status"`
}

// GroupAlterEventManager ...
type GroupAlterEventManager interface {
	Get(pk int64) (GroupAlterEvent, error)
	ListByGroupStatus(groupPK int64, status int64) ([]GroupAlterEvent, error)
	Create(groupAlterEvent GroupAlterEvent) error
	BulkCreateWithTx(tx *sqlx.Tx, groupAlterEvents []GroupAlterEvent) error
	UpdateStatus(pk int64, toStatus int64, fromStatus int64) (int64, error)
	DeleteWithTx(tx *sqlx.Tx, pk int64) error
}

type groupAlterEventManagerManager struct {
	DB *sqlx.DB
}

// GroupAlterEventManager ...
func NewGroupAlterEventManager() GroupAlterEventManager {
	return &groupAlterEventManagerManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

func (m *groupAlterEventManagerManager) Get(pk int64) (groupAlterEvent GroupAlterEvent, err error) {
	query := "SELECT pk, group_pk, action_pks, subject_pks, status FROM rbac_group_alter_event WHERE pk=?"
	err = database.SqlxGet(m.DB, &groupAlterEvent, query, pk)
	return
}

func (m *groupAlterEventManagerManager) ListByGroupStatus(
	groupPK int64,
	status int64,
) (groupAlterEvents []GroupAlterEvent, err error) {
	query := `SELECT
		pk,
		group_pk,
		action_pks,
		subject_pks,
		status 
		FROM rbac_group_alter_event
		WHERE group_pk=? AND status=?`
	err = database.SqlxSelect(m.DB, &groupAlterEvents, query, groupPK, status)
	if errors.Is(err, sql.ErrNoRows) {
		return groupAlterEvents, nil
	}
	return
}

func (m *groupAlterEventManagerManager) BulkCreateWithTx(tx *sqlx.Tx, groupAlterEvents []GroupAlterEvent) error {
	sql := `INSERT INTO rbac_group_alter_event (
		group_pk,
		action_pks,
		subject_pks,
		status
	) VALUES (
		:group_pk,
		:action_pks,
		:subject_pks,
		:status
	)`
	err := database.SqlxBulkInsertWithTx(tx, sql, groupAlterEvents)
	return err
}

func (m *groupAlterEventManagerManager) Create(groupAlterEvent GroupAlterEvent) error {
	sql := `INSERT INTO rbac_group_alter_event (
		group_pk,
		action_pks,
		subject_pks,
		status
	) VALUES (
		:group_pk,
		:action_pks,
		:subject_pks,
		:status
	)`
	err := database.SqlxBulkInsert(m.DB, sql, []GroupAlterEvent{groupAlterEvent})
	return err
}

func (m *groupAlterEventManagerManager) UpdateStatus(pk int64, toStatus int64, fromStatus int64) (int64, error) {
	sql := `UPDATE rbac_group_alter_event SET status=:to_status WHERE pk=:pk AND status=:from_status`
	return database.SqlxUpdate(m.DB, sql, map[string]interface{}{
		"pk":          pk,
		"to_status":   toStatus,
		"from_status": fromStatus,
	})
}

func (m *groupAlterEventManagerManager) DeleteWithTx(tx *sqlx.Tx, pk int64) error {
	sql := `DELETE FROM rbac_group_alter_event WHERE pk=?`
	err := database.SqlxDeleteWithTx(tx, sql, pk)
	return err
}
