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
		mockQuery := `^SELECT pk, uuid, group_pk, action_pks, subject_pks, check_times FROM rbac_group_alter_event WHERE pk=(.*)`
		mockRows := sqlmock.NewRows([]string{"pk", "uuid", "group_pk", "action_pks", "subject_pks", "check_times"}).
			AddRow(
				int64(1), "", int64(1), "[1,2]", "[3,4]", int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &groupAlterEventManagerManager{DB: db}
		evnet, err := manager.Get(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, GroupAlterEvent{
			PK: int64(1), GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]", CheckTimes: int64(0),
		}, evnet)
	})
}

func Test_groupAlterEventManagerManager_ListByGroupCheckTimes(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk, uuid, group_pk, action_pks, subject_pks, check_times FROM rbac_group_alter_event WHERE group_pk=(.*) AND check_times<(.*)`
		mockRows := sqlmock.NewRows([]string{"pk", "uuid", "group_pk", "action_pks", "subject_pks", "check_times"}).
			AddRow(
				int64(1), "", int64(1), "[1,2]", "[3,4]", int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(0), sqlmock.AnyArg()).WillReturnRows(mockRows)

		manager := &groupAlterEventManagerManager{DB: db}
		evnets, err := manager.ListByGroupCheckTimes(int64(1), 0, 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, []GroupAlterEvent{
			{PK: int64(1), GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]", CheckTimes: int64(0)},
		}, evnets)
	})
}

func Test_groupAlterEventManagerManager_BulkCreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(`^INSERT INTO rbac_group_alter_event`).ExpectExec().WithArgs(
			"", int64(1), "[1,2]", "[3,4]", int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO rbac_group_alter_event`).WithArgs(
			"", int64(1), "[1,2]", "[3,4]", int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &groupAlterEventManagerManager{DB: db}
		ids, err := manager.BulkCreateWithTx(tx, []GroupAlterEvent{
			{GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]", CheckTimes: int64(0)},
		})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, []int64{1}, ids)
	})
}

func Test_groupAlterEventManagerManager_Delete(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`^DELETE FROM rbac_group_alter_event WHERE pk=`).WithArgs(
			int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &groupAlterEventManagerManager{DB: db}
		err := manager.Delete(1)

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_groupAlterEventManagerManager_IncrCheckTimes(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`^UPDATE rbac_group_alter_event SET check_times=`).WithArgs(
			int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &groupAlterEventManagerManager{DB: db}
		err := manager.IncrCheckTimes(1)

		assert.NoError(t, err, "query from db fail.")
	})
}
