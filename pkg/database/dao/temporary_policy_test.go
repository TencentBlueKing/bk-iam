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

func Test_temporaryPolicyManager_ListThinBySubjectAction(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk, expired_at FROM temporary_policy WHERE subject_pk`
		mockRows := sqlmock.NewRows([]string{"pk", "expired_at"}).AddRow(1, 2)
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2), int64(3)).WillReturnRows(mockRows)

		manager := &temporaryPolicyManager{DB: db}
		ps, err := manager.ListThinBySubjectAction(1, 2, 3)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, ps, []ThinTemporaryPolicy{{
			PK:        1,
			ExpiredAt: 2,
		}})
	})
}

func Test_temporaryPolicyManager_ListByPKs(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk, subject_pk, action_pk, expression, expired_at FROM temporary_policy WHERE pk IN`
		mockRows := sqlmock.NewRows([]string{"pk", "subject_pk", "action_pk", "expression", "expired_at"}).AddRow(
			1, 2, 3, "", 4)
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2)).WillReturnRows(mockRows)

		manager := &temporaryPolicyManager{DB: db}
		ps, err := manager.ListByPKs([]int64{1, 2})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, ps, []TemporaryPolicy{{
			PK:         1,
			SubjectPK:  2,
			ActionPK:   3,
			Expression: "",
			ExpiredAt:  4,
		}})
	})
}

func Test_temporaryPolicyManager_BulkCreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO temporary_policy`)
		mock.ExpectExec(`INSERT INTO temporary_policy`).WithArgs(
			int64(1), int64(1), "", int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		expr := TemporaryPolicy{
			SubjectPK:  1,
			ActionPK:   1,
			Expression: "",
			ExpiredAt:  1,
		}

		manager := &temporaryPolicyManager{DB: db}
		pks, err := manager.BulkCreateWithTx(tx, []TemporaryPolicy{expr})

		tx.Commit()

		assert.NoError(t, err)
		assert.Equal(t, pks, []int64{1})
	})
}

func Test_temporaryPolicyManager_BulkDeleteByPKsWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectExec(`DELETE FROM temporary_policy WHERE subject_pk`).WithArgs(
			int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &temporaryPolicyManager{DB: db}
		l, err := manager.BulkDeleteByPKs(1, []int64{2})

		assert.NoError(t, err)
		assert.Equal(t, l, int64(1))
	})
}

func Test_temporaryPolicyManager_BulkDeleteBeforeExpiredAtWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM temporary_policy WHERE expired_at`).WithArgs(
			int64(1), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &temporaryPolicyManager{DB: db}
		l, err := manager.BulkDeleteBeforeExpiredAtWithTx(tx, 1, 1)

		tx.Commit()

		assert.NoError(t, err)
		assert.Equal(t, l, int64(1))
	})
}
