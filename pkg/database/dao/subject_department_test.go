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
	"iam/pkg/database"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func Test_subjectDepartmentManger_Get(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_department`
		mockRows := sqlmock.NewRows([]string{"department_pks"}).AddRow("1")
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &subjectDepartmentManger{DB: db}
		pks, err := manager.Get(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, pks, "1")
	})
}

func Test_subjectDepartmentManger_GetCount(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_department`
		mockRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WillReturnRows(mockRows)

		manager := &subjectDepartmentManger{DB: db}
		cnt, err := manager.GetCount()

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectDepartmentManger_BulkCreate(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^INSERT INTO subject_department`
		mock.ExpectExec(mockQuery).WithArgs(int64(1), "1").WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &subjectDepartmentManger{DB: db}
		err := manager.BulkCreate([]SubjectDepartment{{
			SubjectPK:     int64(1),
			DepartmentPKs: "1",
		}})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectDepartmentManger_BulkDelete(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^DELETE FROM subject_department`
		mock.ExpectExec(mockQuery).WithArgs(int64(1), int64(2)).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &subjectDepartmentManger{DB: db}
		err := manager.BulkDelete([]int64{1, 2})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectDepartmentManger_BulkUpdate(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^UPDATE subject_department`
		mock.ExpectBegin()
		mock.ExpectPrepare(mockQuery)
		mock.ExpectExec(mockQuery).WithArgs("1", int64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		manager := &subjectDepartmentManger{DB: db}
		err := manager.BulkUpdate([]SubjectDepartment{{
			SubjectPK:     int64(1),
			DepartmentPKs: "1",
		}})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectDepartmentManger_ListPaging(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_department`
		mockRows := sqlmock.NewRows([]string{"subject_pk", "department_pks"}).AddRow(
			int64(1), "1").AddRow(int64(2), "2")
		mock.ExpectQuery(mockQuery).WithArgs(int64(2), 0).WillReturnRows(mockRows)

		manager := &subjectDepartmentManger{DB: db}
		subjectDepartments, err := manager.ListPaging(int64(2), 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, subjectDepartments, 2)
	})
}
