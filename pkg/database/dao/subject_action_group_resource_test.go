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

func Test_subjectActionResourceGroupManager_GetBySubjectAction(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT 
		 pk,
		 subject_pk,
		 action_pk,
		 group_resource
		 FROM rbac_subject_action_group_resource
		 WHERE subject_pk = (.*)
		 AND action_pk = (.*) LIMIT 1`
		mockRows := sqlmock.NewRows([]string{"pk", "subject_pk", "action_pk", "group_resource"}).AddRow(
			int64(1), int64(1), int64(3), "{}")
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(3)).WillReturnRows(mockRows)

		manager := &subjectActionGroupResourceManager{DB: db}
		subjectActionGroupResource, err := manager.GetBySubjectAction(1, 3)

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(
			t,
			SubjectActionGroupResource{
				PK:            int64(1),
				SubjectPK:     int64(1),
				ActionPK:      int64(3),
				GroupResource: "{}",
			},
			subjectActionGroupResource,
		)
	})
}

func Test_subjectActionResourceGroupManager_CreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO rbac_subject_action_group_resource`).WithArgs(
			int64(1), int64(2), "{}",
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectActionGroupResourceManager{DB: db}
		err = manager.CreateWithTx(tx, SubjectActionGroupResource{
			SubjectPK:     int64(1),
			ActionPK:      int64(2),
			GroupResource: "{}",
		})

		assert.NoError(t, err, "query from db fail.")
	})
}

func Test_subjectActionResourceGroupManager_UpdateGroupResourceWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE rbac_subject_action_group_resource SET group_resource = `).WithArgs(
			"{}", int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectActionGroupResourceManager{DB: db}
		err = manager.UpdateGroupResourceWithTx(tx, int64(1), "{}")

		assert.NoError(t, err, "query from db fail.")
	})
}
