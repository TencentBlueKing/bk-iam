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

func Test_groupSystemAuthTypeManager_ListAuthTypeBySystemGroups(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		group_pk,
		auth_type
		FROM group_system_auth_type
		WHERE system_id = (.*) AND group_pk in (.*)`
		mockRows := sqlmock.NewRows([]string{"group_pk", "auth_type"}).AddRow(int64(1), int(2)).AddRow(int64(2), int(3))
		mock.ExpectQuery(mockQuery).WithArgs("system", int64(1), int64(2)).WillReturnRows(mockRows)

		manager := &groupSystemAuthTypeManager{DB: db}
		authTypes, err := manager.ListAuthTypeBySystemGroups("system", []int64{1, 2})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, []GroupAuthType{
			{GroupPK: int64(1), AuthType: int64(2)},
			{GroupPK: int64(2), AuthType: int64(3)},
		}, authTypes)
	})
}

func Test_groupSystemAuthTypeManager_ListByGroup(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		pk,
		system_id,
		group_pk,
		auth_type,
		reversion
		FROM group_system_auth_type
		WHERE group_pk = (.*)`
		mockRows := sqlmock.NewRows([]string{"system_id", "group_pk", "auth_type"}).AddRow("1", int64(1), int(2)).AddRow("2", int64(1), int(3))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &groupSystemAuthTypeManager{DB: db}
		authTypes, err := manager.ListByGroup(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, []GroupSystemAuthType{
			{SystemID: "1", GroupPK: int64(1), AuthType: int64(2)},
			{SystemID: "2", GroupPK: int64(1), AuthType: int64(3)},
		}, authTypes)
	})
}

func Test_groupSystemAuthTypeManager_GetBySystemGroup(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		pk,
		system_id,
		group_pk,
		auth_type,
		reversion
		FROM group_system_auth_type
		WHERE system_id = (.*) AND group_pk = (.*)`
		mockRows := sqlmock.NewRows([]string{"system_id", "group_pk", "auth_type"}).AddRow("1", int64(1), int(2))
		mock.ExpectQuery(mockQuery).WithArgs("system", int64(1)).WillReturnRows(mockRows)

		manager := &groupSystemAuthTypeManager{DB: db}
		authType, err := manager.GetBySystemGroup("system", int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, GroupSystemAuthType{SystemID: "1", GroupPK: int64(1), AuthType: int64(2)}, authType)
	})
}

func Test_groupSystemAuthTypeManager_DeleteByGroupSystemWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^DELETE FROM group_system_auth_type WHERE system_id = (.*) AND group_pk = (.*)`).WithArgs(
			"system", int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &groupSystemAuthTypeManager{DB: db}
		rows, err := manager.deleteBySystemGroupWithTx(tx, "system", int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, int64(1), rows)
	})
}

func Test_groupSystemAuthTypeManager_CreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^INSERT INTO group_system_auth_type`).WithArgs(
			"system", int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &groupSystemAuthTypeManager{DB: db}
		err = manager.CreateWithTx(tx, GroupSystemAuthType{GroupPK: int64(1), AuthType: int64(2), SystemID: "system"})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_groupSystemAuthTypeManager_UpdateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^UPDATE group_system_auth_type SET auth_type =`).WithArgs(
			int64(2), "system", int64(1), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &groupSystemAuthTypeManager{DB: db}
		rows, err := manager.UpdateWithTx(tx, GroupSystemAuthType{GroupPK: int64(1), AuthType: int64(2), SystemID: "system", Reversion: int64(1)})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, int64(1), rows)
	})
}
