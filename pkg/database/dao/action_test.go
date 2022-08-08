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

func Test_actionManager_GetPK(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk FROM action WHERE system_id = (.*) AND id = (.*)$`
		mockRows := sqlmock.NewRows([]string{"pk"}).
			AddRow(1)
		mock.ExpectQuery(mockQuery).WithArgs("iam", "edit").WillReturnRows(mockRows)

		manager := &actionManager{DB: db}
		pk, err := manager.GetPK("iam", "edit")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, pk, int64(1))
	})
}

func Test_actionManager_ListBySystem(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT (.*) FROM action WHERE system_id=(.*)`
		mockRows := sqlmock.NewRows([]string{"pk", "system_id", "id"}).
			AddRow(1, "test", "action")
		mock.ExpectQuery(mockQuery).WithArgs("test").WillReturnRows(mockRows)

		manager := &actionManager{DB: db}
		actions, err := manager.ListBySystem("test")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, actions, []Action{{
			PK:     int64(1),
			System: "test",
			ID:     "action",
		}})
	})
}

func Test_actionManager_ListPKBySystem(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk FROM action WHERE system_id=(.*)`
		mockRows := sqlmock.NewRows([]string{"pk"}).
			AddRow(int64(1)).AddRow(int64(2))
		mock.ExpectQuery(mockQuery).WithArgs("test").WillReturnRows(mockRows)

		manager := &actionManager{DB: db}
		pks, err := manager.ListPKBySystem("test")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, []int64{1, 2}, pks)
	})
}
