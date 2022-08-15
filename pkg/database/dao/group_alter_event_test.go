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

func Test_groupAlterEventManagerManager_ListBeforeCreateAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT uuid, group_pk, action_pks, subject_pks FROM rbac_group_alter_event WHERE created_at<FROM_UNIXTIME(.*) LIMIT (.*)`
		mockRows := sqlmock.NewRows([]string{"uuid"}).AddRow("uuid")
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), sqlmock.AnyArg()).WillReturnRows(mockRows)

		manager := &groupAlterEventManagerManager{DB: db}
		events, err := manager.ListBeforeCreateAt(1, 2)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, events, 1)
	})
}

func Test_groupAlterEventManagerManager_Create(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`INSERT INTO rbac_group_alter_event`).WithArgs(
			"uuid", int64(1), "[1,2]", "[3,4]",
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &groupAlterEventManagerManager{DB: db}
		err := manager.Create(GroupAlterEvent{
			UUID: "uuid", GroupPK: int64(1), ActionPKs: "[1,2]", SubjectPKs: "[3,4]",
		})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_groupAlterEventManagerManager_BulkDeleteWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^DELETE FROM rbac_group_alter_event WHERE uuid IN`).WithArgs(
			"uuid1", "uuid2",
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)
		manager := &groupAlterEventManagerManager{DB: db}
		err = manager.BulkDeleteWithTx(tx, []string{"uuid1", "uuid2"})

		assert.NoError(t, err, "query from db fail.")
	})
}
