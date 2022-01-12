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

func Test_expressionManager_ListAuthByPKs(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			AuthExpression{
				PK:         1,
				Expression: "test",
				Signature:  "test",
			},
			AuthExpression{
				PK:         2,
				Expression: "test2",
				Signature:  "test2",
			},
		}
		mockQuery := `^SELECT pk, expression, signature FROM expression WHERE pk IN`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2)).WillReturnRows(mockRows)

		manager := &expressionManager{DB: db}
		expressions, err := manager.ListAuthByPKs([]int64{1, 2})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(expressions), 2)
		assert.Equal(t, mockData[0].(AuthExpression), expressions[0])
		assert.Equal(t, mockData[1].(AuthExpression), expressions[1])
	})
}

func Test_expressionManager_BulkCreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(`INSERT INTO expression`)
		mock.ExpectExec(`INSERT INTO expression`).WithArgs(
			int64(1), "expression", "test",
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		expr := Expression{
			Type:       1,
			Expression: "expression",
			Signature:  "test",
		}

		manager := &expressionManager{DB: db}
		ids, err := manager.BulkCreateWithTx(tx, []Expression{expr})

		tx.Commit()

		assert.NoError(t, err)
		assert.Equal(t, ids, []int64{1})
	})
}

func Test_expressionManager_BulkUpdateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE expression SET expression=`)
		mock.ExpectExec(`UPDATE expression SET expression=`).WithArgs(
			"expression", "test", int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		expr := Expression{
			PK:         1,
			Type:       1,
			Expression: "expression",
			Signature:  "test",
		}

		manager := &expressionManager{DB: db}
		err = manager.BulkUpdateWithTx(tx, []Expression{expr})

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_expressionManager_BulkDeleteByPKsWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM expression WHERE pk IN`).WithArgs(
			int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &expressionManager{DB: db}
		rows, err := manager.BulkDeleteByPKsWithTx(tx, []int64{1, 2})

		tx.Commit()

		assert.NoError(t, err)
		assert.Equal(t, rows, int64(1))
	})
}

func Test_expressionManager_ListDistinctBySignaturesType(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			Expression{
				PK:         1,
				Type:       1,
				Expression: "test",
				Signature:  "test",
			},
			Expression{
				PK:         2,
				Type:       1,
				Expression: "test2",
				Signature:  "test2",
			},
		}
		mockQuery := `^SELECT pk, type, expression, signature FROM expression WHERE pk IN`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs("a", "b", int64(1)).WillReturnRows(mockRows)

		manager := &expressionManager{DB: db}
		expressions, err := manager.ListDistinctBySignaturesType([]string{"a", "b"}, 1)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(expressions), 2)
		assert.Equal(t, mockData[0].(Expression), expressions[0])
		assert.Equal(t, mockData[1].(Expression), expressions[1])
	})
}

func Test_expressionManager_ChangeUnreferencedExpressionType(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE expression SET type=`).WithArgs(
			int64(-1), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &expressionManager{DB: db}
		err = manager.ChangeUnreferencedExpressionType(1, -1)

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_expressionManager_ChangeReferencedExpressionTypeBeforeUpdateAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE expression SET type=`).WithArgs(
			int64(1), int64(-1), int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &expressionManager{DB: db}
		err = manager.ChangeReferencedExpressionTypeBeforeUpdateAt(-1, 1, 0)

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_expressionManager_DeleteByTypeBeforeUpdateAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM expression WHERE type=`).WithArgs(
			int64(-1), int64(0),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &expressionManager{DB: db}
		err = manager.DeleteByTypeBeforeUpdateAt(-1, 0)

		tx.Commit()

		assert.NoError(t, err)
	})
}
