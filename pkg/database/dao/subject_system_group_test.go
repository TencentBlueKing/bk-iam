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

func Test_subjectSystemGroupManager_ListGroups(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		subject_pk,
		groups
		FROM subject_system_group
		WHERE system_id = (.*) AND subject_pk IN (.*)`
		mockRows := sqlmock.NewRows([]string{"subject_pk", "groups"}).AddRow(int64(1), "test")
		mock.ExpectQuery(mockQuery).WithArgs("system", int64(1)).WillReturnRows(mockRows)

		manager := &subjectSystemGroupManager{DB: db}
		groups, err := manager.ListSubjectGroups("system", []int64{1})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, groups, []SubjectGroups{{Groups: "test", SubjectPK: int64(1)}})
	})
}

func Test_subjectSystemGroupManager_DeleteBySystemSubjectWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^DELETE FROM subject_system_group WHERE subject_pk IN (.*)`).WithArgs(
			int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectSystemGroupManager{DB: db}
		err = manager.DeleteBySubjectPKsWithTx(tx, []int64{1})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectSystemGroupManager_GetBySystemSubject(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		pk,
		system_id,
		subject_pk,
		groups,
		reversion
		FROM subject_system_group
		WHERE system_id = (.*) AND subject_pk = (.*)`
		mockRows := sqlmock.NewRows([]string{"system_id", "subject_pk", "groups", "reversion"}).
			AddRow("test", int64(1), "[]", int64(2))
		mock.ExpectQuery(mockQuery).WithArgs("system", int64(1)).WillReturnRows(mockRows)

		manager := &subjectSystemGroupManager{DB: db}
		groups, err := manager.GetBySystemSubject("system", int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, SubjectSystemGroup{
			SystemID:  "test",
			SubjectPK: int64(1),
			Groups:    "[]",
			Reversion: int64(2),
		}, groups)
	})
}

func Test_subjectSystemGroupManager_CreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^INSERT INTO subject_system_group`).WithArgs(
			"system", int64(1), "[]",
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectSystemGroupManager{DB: db}
		err = manager.CreateWithTx(tx, SubjectSystemGroup{
			SystemID:  "system",
			SubjectPK: int64(1),
			Groups:    "[]",
		})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectSystemGroupManager_UpdateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^UPDATE subject_system_group SET groups = (.*)`).WithArgs(
			"[]", "system", int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectSystemGroupManager{DB: db}
		rows, err := manager.UpdateWithTx(tx, SubjectSystemGroup{
			SystemID:  "system",
			SubjectPK: int64(1),
			Groups:    "[]",
			Reversion: int64(2),
		})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, int64(1), rows)
	})
}
