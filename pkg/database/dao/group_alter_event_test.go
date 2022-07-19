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
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

func Test_groupAlterEventManagerManager_Get(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk, group_pk, action_pks, subject_pks, status FROM rbac_group_alter_event WHERE pk=(.*)`
		mockRows := sqlmock.NewRows([]string{"pk", "group_pk", "action_pks", "subject_pks", "status"}).AddRow(
			int64(1), int64(1), "[1,2]", "[3,4]", int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &groupAlterEventManagerManager{DB: db}
		evnet, err := manager.Get(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, GroupAlterEvent{
			PK: int64(1), GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]", Status: int64(0),
		}, evnet)
	})
}

func Test_groupAlterEventManagerManager_ListByGroupStatus(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk, group_pk, action_pks, subject_pks, status FROM rbac_group_alter_event WHERE group_pk=(.*) AND status=(.*)`
		mockRows := sqlmock.NewRows([]string{"pk", "group_pk", "action_pks", "subject_pks", "status"}).AddRow(
			int64(1), int64(1), "[1,2]", "[3,4]", int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(0)).WillReturnRows(mockRows)

		manager := &groupAlterEventManagerManager{DB: db}
		evnets, err := manager.ListByGroupStatus(int64(1), 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, []GroupAlterEvent{
			{PK: int64(1), GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]", Status: int64(0)},
		}, evnets)
	})
}

func Test_groupAlterEventManagerManager_BulkCreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO rbac_group_alter_event`).WithArgs(
			int64(1), "[1,2]", "[3,4]", int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &groupAlterEventManagerManager{DB: db}
		err = manager.BulkCreateWithTx(tx, []GroupAlterEvent{
			{GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]", Status: int64(0)},
		})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_groupAlterEventManagerManager_DeleteWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^DELETE FROM rbac_group_alter_event WHERE pk=`).WithArgs(
			int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &groupAlterEventManagerManager{DB: db}
		err = manager.DeleteWithTx(tx, 1)

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_groupAlterEventManagerManager_UpdateStatus(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`^UPDATE rbac_group_alter_event SET status=(.*) WHERE pk=(.*) AND status=(.*)`).WithArgs(
			int64(1), int64(1), int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &groupAlterEventManagerManager{DB: db}
		count, err := manager.UpdateStatus(1, 1, 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, int64(1), count)
	})
}

func Test_groupAlterEventManagerManager_Create(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`INSERT INTO rbac_group_alter_event`).WithArgs(
			int64(1), "[1,2]", "[3,4]", int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &groupAlterEventManagerManager{DB: db}
		err := manager.Create(GroupAlterEvent{
			GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]", Status: int64(0),
		})

		assert.NoError(t, err, "query from db fail.")
	})
}
