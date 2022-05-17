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
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

func Test_subjectRelationManager_GetCount(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_relation`
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &subjectRelationManager{DB: db}
		cnt, err := manager.GetMemberCount(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_List(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_relation`
		mockRows := sqlmock.NewRows(
			[]string{"pk", "subject_pk", "parent_pk", "policy_expired_at"},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), 0, 10).WillReturnRows(mockRows)

		manager := &subjectRelationManager{DB: db}
		relations, err := manager.ListPagingMember(int64(1), 0, 10)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_ListRelation(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_relation`
		mockRows := sqlmock.NewRows(
			[]string{
				"pk", "subject_pk", "parent_pk", "policy_expired_at"},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &subjectRelationManager{DB: db}
		relations, err := manager.ListRelation(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_ListEffectRelationBySubjectPKs(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_relation`
		mockRows := sqlmock.NewRows(
			[]string{"subject_pk", "parent_pk", "policy_expired_at"},
		).AddRow(int64(1), int64(1), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), time.Now().Unix()).WillReturnRows(mockRows)

		manager := &subjectRelationManager{DB: db}
		relations, err := manager.ListEffectRelationBySubjectPKs([]int64{1})

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_BulkDeleteByMembersWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM subject_relation WHERE parent_pk=`).WithArgs(
			int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectRelationManager{DB: db}
		cnt, err := manager.BulkDeleteByMembersWithTx(tx, int64(1), []int64{2})

		tx.Commit()
		assert.NoError(t, err)
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_ListRelationBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_relation`
		mockRows := sqlmock.NewRows(
			[]string{
				"pk", "subject_pk", "parent_pk", "policy_expired_at"},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(1000)).WillReturnRows(mockRows)

		manager := &subjectRelationManager{DB: db}
		relations, err := manager.ListRelationBeforeExpiredAt(int64(1), int64(1000))

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_GetMemberCountBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM subject_relation`
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(10)).WillReturnRows(mockRows)

		manager := &subjectRelationManager{DB: db}
		cnt, err := manager.GetMemberCountBeforeExpiredAt(int64(1), int64(10))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}
