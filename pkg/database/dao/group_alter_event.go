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
	PK   int64  `db:"pk"`
	UUID string `db:"uuid"`

	GroupPK    int64  `db:"group_pk"`
	ActionPKs  string `db:"action_pks"`
	SubjectPKs string `db:"subject_pks"`
	CheckTimes int64  `db:"check_times"`
}

// GroupAlterEventManager ...
type GroupAlterEventManager interface {
	Get(pk int64) (GroupAlterEvent, error)
	ListPKByCheckTimesBeforeCreateAt(checkTimes int64, createdAt int64) ([]int64, error)
	BulkCreateWithTx(tx *sqlx.Tx, groupAlterEvents []GroupAlterEvent) ([]int64, error)
	Delete(pk int64) error
	IncrCheckTimes(pk int64) error
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

// Get ...
func (m *groupAlterEventManagerManager) Get(pk int64) (groupAlterEvent GroupAlterEvent, err error) {
	query := "SELECT pk, uuid, group_pk, action_pks, subject_pks, check_times FROM rbac_group_alter_event WHERE pk=?"
	err = database.SqlxGet(m.DB, &groupAlterEvent, query, pk)
	return
}

// ListByGroupCheckTimes ...
func (m *groupAlterEventManagerManager) ListPKByCheckTimesBeforeCreateAt(
	checkTimes int64,
	createdAt int64,
) (pks []int64, err error) {
	query := `SELECT
		pk
		FROM rbac_group_alter_event
		WHERE check_times<?
		AND created_at<FROM_UNIXTIME(?)`
	err = database.SqlxSelect(m.DB, &pks, query, checkTimes, createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return pks, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *groupAlterEventManagerManager) BulkCreateWithTx(
	tx *sqlx.Tx,
	groupAlterEvents []GroupAlterEvent,
) ([]int64, error) {
	sql := `INSERT INTO rbac_group_alter_event (
		uuid,
		group_pk,
		action_pks,
		subject_pks,
		check_times
	) VALUES (
		:uuid,
		:group_pk,
		:action_pks,
		:subject_pks,
		:check_times
	)`
	ids, err := database.SqlxBulkInsertReturnIDWithTx(tx, sql, groupAlterEvents)
	return ids, err
}

// Delete ...
func (m *groupAlterEventManagerManager) Delete(pk int64) error {
	sql := `DELETE FROM rbac_group_alter_event WHERE pk=?`
	err := database.SqlxExec(m.DB, sql, pk)
	return err
}

// IncrCheckTimes ...
func (m *groupAlterEventManagerManager) IncrCheckTimes(pk int64) error {
	sql := `UPDATE rbac_group_alter_event SET check_times=check_times+1 WHERE pk=?`
	err := database.SqlxExec(m.DB, sql, pk)
	return err
}
