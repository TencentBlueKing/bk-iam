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

// SubjectActionAlterEvent ...
type SubjectActionAlterEvent struct {
	UUID string `db:"uuid"`

	Data       string `db:"data"` // [{subject_pk action_pk group_pks}]
	Status     int64  `db:"status"`
	CheckCount int64  `db:"check_count"`
}

// SubjectActionAlterEventManager ...
type SubjectActionAlterEventManager interface {
	Get(uuid string) (SubjectActionAlterEvent, error)
	BulkCreateWithTx(tx *sqlx.Tx, messages []SubjectActionAlterEvent) error
	BulkUpdateStatus(uuids []string, status int64) error
	Delete(uuid string) error
}

type subjectActionAlterEventManager struct {
	DB *sqlx.DB
}

// NewSubjectActionAlterEventManager ...
func NewSubjectActionAlterEventManager() SubjectActionAlterEventManager {
	return &subjectActionAlterEventManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *subjectActionAlterEventManager) Get(uuid string) (message SubjectActionAlterEvent, err error) {
	sql := `SELECT
		uuid,
		data,
		status,
		check_count
		FROM rbac_subject_action_alter_event
		WHERE uuid = ?
		LIMIT 1`
	err = database.SqlxGet(m.DB, &message, sql, uuid)
	return message, err
}

// BulkCreateWithTx ...
func (m *subjectActionAlterEventManager) BulkCreateWithTx(
	tx *sqlx.Tx,
	messages []SubjectActionAlterEvent,
) (err error) {
	if len(messages) == 0 {
		return nil
	}
	sql := `INSERT INTO rbac_subject_action_alter_event (
		uuid,
		data,
		status,
		check_count
	) VALUES (
		:uuid,
		:data,
		:status,
		:check_count
	)`
	return database.SqlxBulkInsertWithTx(tx, sql, messages)
}

// UpdateStatus ...
func (m *subjectActionAlterEventManager) BulkUpdateStatus(uuids []string, status int64) error {
	sql := `UPDATE rbac_subject_action_alter_event
		SET status = ?
		WHERE uuid IN (?)`
	return database.SqlxExec(m.DB, sql, status, uuids)
}

// Delete ...
func (m *subjectActionAlterEventManager) Delete(uuid string) error {
	sql := `DELETE FROM rbac_subject_action_alter_event
		WHERE uuid = ?`
	return database.SqlxExec(m.DB, sql, uuid)
}
