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

	"iam/pkg/database"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func Test_subjectRoleManager_ListSubjectPKByRole(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT subject_pk FROM subject_role WHERE role_type`
		mockRows := sqlmock.NewRows([]string{"subject_pk"}).AddRow(int64(1)).AddRow(int64(2))
		mock.ExpectQuery(mockQuery).WithArgs("super_manger", "").WillReturnRows(mockRows)

		manager := &subjectRoleManager{DB: db}
		subjectPKs, err := manager.ListSubjectPKByRole("super_manger", "")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, subjectPKs, []int64{1, 2})
	})
}

func Test_subjectRoleManager_BulkCreate(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^INSERT INTO subject_role`
		mock.ExpectExec(mockQuery).WithArgs("super_manager", "", 1).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &subjectRoleManager{DB: db}
		err := manager.BulkCreate([]SubjectRole{{
			RoleType:  "super_manager",
			System:    "",
			SubjectPK: 1,
		}})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectRoleManager_BulkDelete(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^DELETE FROM subject_role WHERE role_type`
		mock.ExpectExec(mockQuery).WithArgs("super_manger", "", int64(1), int64(2)).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &subjectRoleManager{DB: db}
		err := manager.BulkDelete("super_manger", "", []int64{1, 2})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectRoleManager_ListSystemIDBySubjectPK(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_role`
		mockRows := sqlmock.NewRows([]string{"system_id"}).AddRow("bk_cmdb").AddRow("bk_job")
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &subjectRoleManager{DB: db}
		systems, err := manager.ListSystemIDBySubjectPK(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, systems, []string{"bk_cmdb", "bk_job"})
	})
}
