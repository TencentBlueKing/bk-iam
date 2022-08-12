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

func Test_subjectActionAlterMessageManager_Get(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		uuid,
		data,
		status,
		check_count
		FROM rbac_subject_action_alter_message
		WHERE uuid = (.*)`
		mockRows := sqlmock.NewRows([]string{"uuid"}).AddRow("uuid")
		mock.ExpectQuery(mockQuery).WithArgs("uuid").WillReturnRows(mockRows)

		manager := &subjectActionAlterMessageManager{DB: db}
		message, err := manager.Get("uuid")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, "uuid", message.UUID)
	})
}

func Test_subjectActionAlterMessageManager_BulkCreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^INSERT INTO rbac_subject_action_alter_message`).WithArgs(
			"uuid", "test", int64(0), int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)
		manager := &subjectActionAlterMessageManager{DB: db}
		err = manager.BulkCreateWithTx(tx, []SubjectActionAlterMessage{
			{
				UUID:       "uuid",
				Data:       "test",
				Status:     0,
				CheckCount: 0,
			},
		})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectActionAlterMessageManager_BulkUpdateStatus(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`UPDATE rbac_subject_action_alter_message SET status =`).WithArgs(
			int64(1), "uuid1", "uuid2",
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &subjectActionAlterMessageManager{DB: db}
		err := manager.BulkUpdateStatus([]string{"uuid1", "uuid2"}, 1)

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectActionAlterMessageManager_Delete(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`DELETE FROM rbac_subject_action_alter_message WHERE uuid =`).WithArgs(
			"uuid",
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &subjectActionAlterMessageManager{DB: db}
		err := manager.Delete("uuid")

		assert.NoError(t, err, "query from db fail.")
	})
}
