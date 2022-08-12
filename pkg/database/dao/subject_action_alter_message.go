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
	"iam/pkg/database"

	"github.com/jmoiron/sqlx"
)

// SubjectActionAlterMessage ...
type SubjectActionAlterMessage struct {
	UUID string `db:"uuid"`

	Data       string `db:"data"`
	Status     int64  `db:"status"`
	CheckCount int64  `db:"check_count"`
}

// SubjectActionAlterMessageManager ...
type SubjectActionAlterMessageManager interface {
	Get(uuid string) (SubjectActionAlterMessage, error)
	BulkCreateWithTx(tx *sqlx.Tx, messages []SubjectActionAlterMessage) error
	BulkUpdateStatus(uuids []string, status int64) error
	Delete(uuid string) error
}

type subjectActionAlterMessageManager struct {
	DB *sqlx.DB
}

// NewSubjectActionAlterMessageManager ...
func NewSubjectActionAlterMessageManager() SubjectActionAlterMessageManager {
	return &subjectActionAlterMessageManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// Get ...
func (m *subjectActionAlterMessageManager) Get(uuid string) (message SubjectActionAlterMessage, err error) {
	sql := `SELECT
		uuid,
		data,
		status,
		check_count
		FROM rbac_subject_action_alter_message
		WHERE uuid = ?`
	err = database.SqlxGet(m.DB, &message, sql, uuid)
	return message, err
}

// BulkCreateWithTx ...
func (m *subjectActionAlterMessageManager) BulkCreateWithTx(
	tx *sqlx.Tx,
	messages []SubjectActionAlterMessage,
) (err error) {
	if len(messages) == 0 {
		return nil
	}
	sql := `INSERT INTO rbac_subject_action_alter_message (
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
func (m *subjectActionAlterMessageManager) BulkUpdateStatus(uuids []string, status int64) error {
	sql := `UPDATE rbac_subject_action_alter_message
		SET status = ?
		WHERE uuid IN (?)`
	return database.SqlxExec(m.DB, sql, status, uuids)
}

// Delete ...
func (m *subjectActionAlterMessageManager) Delete(uuid string) error {
	sql := `DELETE FROM rbac_subject_action_alter_message
		WHERE uuid = ?`
	return database.SqlxExec(m.DB, sql, uuid)
}
