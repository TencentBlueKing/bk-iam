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

func Test_subjectManager_GetPK(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk FROM subject WHERE type=.* AND id=.* LIMIT 1`
		mockRows := sqlmock.NewRows([]string{"pk"}).AddRow(1)
		mock.ExpectQuery(mockQuery).WithArgs("user", "admin").WillReturnRows(mockRows)

		manager := &subjectManager{DB: db}
		pk, err := manager.GetPK("user", "admin")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, pk, int64(1))
	})
}

func Test_subjectManager_ListPaging(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			Subject{
				PK:   1,
				Type: "group",
				ID:   "1",
				Name: "group1",
			},
			Subject{
				PK:   3,
				Type: "group",
				ID:   "3",
				Name: "group3",
			},
		}

		mockQuery := `^SELECT pk, type, id, name FROM subject WHERE type=.* LIMIT .* OFFSET .*`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs("group", 0, 10).WillReturnRows(mockRows)

		manager := &subjectManager{DB: db}
		subject, err := manager.ListPaging("group", 0, 10)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(subject), 2)
	})
}

//func Test_subjectManager_Delete(t *testing.T) {
//	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
//		mockQuery := `^DELETE FROM subject`
//		mock.ExpectExec(mockQuery).WithArgs("type", "id").WillReturnResult(sqlmock.NewResult(1, 1))
//
//		manager := &subjectManager{DB: db}
//		err := manager.Delete(Subject{Type: "type", ID: "id", Name: "name"})
//
//		assert.NoError(t, err, "query from db fail.")
//	})
//}

func Test_subjectManager_BulkCreate(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^INSERT INTO subject`
		mock.ExpectExec(mockQuery).WithArgs("type", "id", "name").WillReturnResult(sqlmock.NewResult(1, 1))

		manager := &subjectManager{DB: db}
		err := manager.BulkCreate([]Subject{{Type: "type", ID: "id", Name: "name"}})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectManager_ListByPKs(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			Subject{
				PK:   1,
				Type: "group",
				ID:   "1",
				Name: "group1",
			},
			Subject{
				PK:   3,
				Type: "group",
				ID:   "3",
				Name: "group3",
			},
		}

		mockQuery := `^SELECT pk, type, id, name FROM subject WHERE pk IN`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2)).WillReturnRows(mockRows)

		manager := &subjectManager{DB: db}
		subject, err := manager.ListByPKs([]int64{1, 2})

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(subject), 2)
	})
}

func Test_subjectManager_BulkUpdate(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^UPDATE subject`
		mock.ExpectBegin()
		mock.ExpectPrepare(mockQuery)
		mock.ExpectExec(mockQuery).WithArgs("name", "type", "id").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		manager := &subjectManager{DB: db}
		err := manager.BulkUpdate([]Subject{{Type: "type", ID: "id", Name: "name"}})

		assert.NoError(t, err, "query from db fail.")
	})
}
