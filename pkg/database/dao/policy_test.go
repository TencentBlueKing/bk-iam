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

func Test_policyManager_ListAuthBySubjectAction(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			AuthPolicy{
				PK:           1,
				SubjectPK:    1,
				ExpressionPK: 1,
				ExpiredAt:    1,
			},
			AuthPolicy{
				PK:           2,
				SubjectPK:    2,
				ExpressionPK: 2,
				ExpiredAt:    2,
			},
		}
		mockQuery := `^SELECT pk, subject_pk, expression_pk, expired_at FROM policy WHERE subject_pk`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2), int64(1), 0).WillReturnRows(mockRows)

		manager := &policyManager{DB: db}
		policies, err := manager.ListAuthBySubjectAction([]int64{1, 2}, 1, 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(policies), 2)
		assert.Equal(t, policies[0], mockData[0].(AuthPolicy))
		assert.Equal(t, policies[1], mockData[1].(AuthPolicy))
	})
}

func Test_policyManager_BulkCreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO policy`).WithArgs(
			int64(1), int64(1), int64(1), int64(1), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		expr := Policy{
			SubjectPK:    1,
			ActionPK:     1,
			ExpressionPK: 1,
			ExpiredAt:    1,
			TemplateID:   1,
		}

		manager := &policyManager{DB: db}
		err = manager.BulkCreateWithTx(tx, []Policy{expr})

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_policyManager_BulkDeleteByTemplatePKsWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^DELETE FROM policy WHERE subject_pk = (.*) AND pk IN`).WithArgs(
			int64(1), int64(1), int64(2), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &policyManager{DB: db}
		rows, err := manager.BulkDeleteByTemplatePKsWithTx(tx, int64(1), int64(1), []int64{1, 2})

		tx.Commit()

		assert.NoError(t, err)
		assert.Equal(t, rows, int64(1))
	})
}

func Test_policyManager_ListByPKs(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			Policy{
				PK:           1,
				SubjectPK:    1,
				ExpressionPK: 1,
				ExpiredAt:    1,
			},
			Policy{
				PK:           2,
				SubjectPK:    2,
				ExpressionPK: 2,
				ExpiredAt:    2,
			},
		}
		mockQuery := `^SELECT pk, subject_pk, action_pk, expression_pk, expired_at, template_id FROM policy WHERE subject_pk`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(1), int64(2)).WillReturnRows(mockRows)

		manager := &policyManager{DB: db}
		policies, err := manager.ListBySubjectPKAndPKs(int64(1), []int64{1, 2})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(policies), 2)
		assert.Equal(t, policies[0], mockData[0].(Policy))
		assert.Equal(t, policies[1], mockData[1].(Policy))
	})
}

func Test_policyManager_BulkDeleteBySubjectPKsWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM policy WHERE subject_pk IN`).WithArgs(
			int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &policyManager{DB: db}
		err = manager.bulkDeleteBySubjectPKsWithTx(tx, []int64{1, 2})

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_policyManager_ListExpressionBySubjectsTemplate(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT expression_pk FROM policy WHERE subject_pk in (.*) AND template_id = ?`
		mockRows := sqlmock.NewRows([]string{"expression_pk"}).AddRow(
			int64(1)).AddRow(int64(2)).AddRow(int64(3)).AddRow(int64(4))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2), int64(0)).WillReturnRows(mockRows)

		manager := &policyManager{DB: db}
		expressionPKs, err := manager.ListExpressionBySubjectsTemplate([]int64{1, 2}, int64(0))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(expressionPKs), 4)
	})
}

func Test_policyManager_ListBySubjectTemplateBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			Policy{
				PK:           1,
				SubjectPK:    1,
				ExpressionPK: 1,
				ExpiredAt:    1,
			},
			Policy{
				PK:           2,
				SubjectPK:    2,
				ExpressionPK: 2,
				ExpiredAt:    2,
			},
		}
		mockQuery := `^SELECT pk, subject_pk, action_pk, expression_pk, expired_at, template_id FROM policy WHERE subject_pk`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(0), int64(1000)).WillReturnRows(mockRows)

		manager := &policyManager{DB: db}
		policies, err := manager.ListBySubjectTemplateBeforeExpiredAt(int64(1), int64(0), int64(1000))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(policies), 2)
		assert.Equal(t, policies[0], mockData[0].(Policy))
		assert.Equal(t, policies[1], mockData[1].(Policy))
	})
}

func Test_policyManager_UpdateExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE policy SET expired_at = (.*) WHERE pk = (.*)`)
		mock.ExpectExec(`UPDATE policy SET expired_at =`).WithArgs(
			int64(1), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		policies := []Policy{{
			PK:        1,
			ExpiredAt: 1,
		}}

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &policyManager{DB: db}
		err = manager.BulkUpdateExpiredAtWithTx(tx, policies)

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_policyManager_BulkUpdateExpressionPKWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE policy SET expression_pk=(.*) WHERE pk=(.*)`)
		mock.ExpectExec(`UPDATE policy SET expression_pk=`).WithArgs(
			int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		policy := Policy{
			PK:           2,
			ExpressionPK: 1,
		}

		manager := &policyManager{DB: db}
		err = manager.BulkUpdateExpressionPKWithTx(tx, []Policy{policy})

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_policyManager_DeleteByActionPKWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^DELETE FROM policy WHERE action_pk = (.*) LIMIT (.*)`).WithArgs(
			int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 2))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &policyManager{DB: db}
		count, err := manager.DeleteByActionPKWithTx(tx, 1, 2)

		tx.Commit()

		assert.NoError(t, err)
		assert.Equal(t, count, int64(2))
	})
}
