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

func Test_subjectActionExpressionManager_ListBySubjectAction(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT 
		pk,
		subject_pk,
		action_pk,
		expression,
		expired_at
		FROM rbac_subject_action_expression
		WHERE subject_pk IN (.*)
		AND action_pk = (.*)
		AND expired_at >= (.*)`
		mockRows := sqlmock.NewRows([]string{"pk", "subject_pk", "action_pk", "expression", "expired_at"}).AddRow(
			int64(1), int64(1), int64(3), "{}", int64(10)).AddRow(
			int64(2), int64(2), int64(3), "{}", int64(10))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2), int64(3), int64(0)).WillReturnRows(mockRows)

		manager := &subjectActionExpressionManager{DB: db}
		subjectActionExpressions, err := manager.ListBySubjectAction([]int64{1, 2}, int64(3), int64(0))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, []SubjectActionExpression{
			{PK: int64(1), SubjectPK: int64(1), ActionPK: int64(3), Expression: "{}", ExpiredAt: int64(10)},
			{PK: int64(2), SubjectPK: int64(2), ActionPK: int64(3), Expression: "{}", ExpiredAt: int64(10)},
		}, subjectActionExpressions)
	})
}

func Test_subjectActionExpressionManager_GetBySubjectAction(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT 
		pk,
		subject_pk,
		action_pk,
		expression,
		expired_at
		FROM rbac_subject_action_expression
		WHERE subject_pk = (.*)
		AND action_pk = (.*)`
		mockRows := sqlmock.NewRows([]string{"pk", "subject_pk", "action_pk", "expression", "expired_at"}).AddRow(
			int64(1), int64(1), int64(3), "{}", int64(10))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(3)).WillReturnRows(mockRows)

		manager := &subjectActionExpressionManager{DB: db}
		subjectActionExpression, err := manager.GetBySubjectAction(1, 3)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(
			t,
			SubjectActionExpression{
				PK:         int64(1),
				SubjectPK:  int64(1),
				ActionPK:   int64(3),
				Expression: "{}",
				ExpiredAt:  int64(10),
			},
			subjectActionExpression,
		)
	})
}

func Test_subjectActionExpressionManager_CreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO rbac_subject_action_expression`).WithArgs(
			int64(1), int64(2), "{}", int64(10),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectActionExpressionManager{DB: db}
		err = manager.CreateWithTx(tx, SubjectActionExpression{
			SubjectPK:  int64(1),
			ActionPK:   int64(2),
			Expression: "{}",
			ExpiredAt:  int64(10),
		})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectActionExpressionManager_UpdateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE rbac_subject_action_expression SET expression = `).WithArgs(
			"{}", int64(10), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectActionExpressionManager{DB: db}
		err = manager.UpdateWithTx(tx, SubjectActionExpression{
			PK:         int64(1),
			SubjectPK:  int64(1),
			ActionPK:   int64(2),
			Expression: "{}",
			ExpiredAt:  int64(10),
		})

		assert.NoError(t, err, "query from db fail.")
	})
}
