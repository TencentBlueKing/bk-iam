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
	UUID string `db:"uuid"`

	GroupPK    int64  `db:"group_pk"`
	ActionPKs  string `db:"action_pks"`
	SubjectPKs string `db:"subject_pks"`
}

// GroupAlterEventManager ...
type GroupAlterEventManager interface {
	ListBeforeCreateAt(createdAt int64, limit int64) ([]GroupAlterEvent, error)
	Create(event GroupAlterEvent) error
	BulkDeleteWithTx(tx *sqlx.Tx, uuids []string) error
}

type groupAlterEventManagerManager struct {
	DB *sqlx.DB
}

// NewGroupAlterEventManager ...
func NewGroupAlterEventManager() GroupAlterEventManager {
	return &groupAlterEventManagerManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListBeforeCreateAt ...
func (m *groupAlterEventManagerManager) ListBeforeCreateAt(
	createdAt int64,
	limit int64,
) (events []GroupAlterEvent, err error) {
	query := `SELECT
		uuid,
		group_pk,
		action_pks,
		subject_pks
		FROM rbac_group_alter_event
		WHERE created_at<FROM_UNIXTIME(?)
		LIMIT ?`
	err = database.SqlxSelect(m.DB, &events, query, createdAt, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return events, nil
	}
	return
}

// Create ...
func (m *groupAlterEventManagerManager) Create(
	event GroupAlterEvent,
) error {
	sql := `INSERT INTO rbac_group_alter_event (
		uuid,
		group_pk,
		action_pks,
		subject_pks
	) VALUES (
		:uuid,
		:group_pk,
		:action_pks,
		:subject_pks
	)`
	return database.SqlxBulkInsert(m.DB, sql, []GroupAlterEvent{event})
}

// BulkDelete ...
func (m *groupAlterEventManagerManager) BulkDeleteWithTx(tx *sqlx.Tx, uuids []string) error {
	sql := `DELETE FROM rbac_group_alter_event WHERE uuid IN (?)`
	return database.SqlxDeleteWithTx(tx, sql, uuids)
}
