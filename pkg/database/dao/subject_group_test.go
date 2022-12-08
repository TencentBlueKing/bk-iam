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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

func Test_subjectRelationManager_GetCount(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT COUNT(.*) FROM subject_relation WHERE parent_pk = `
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		cnt, err := manager.GetGroupMemberCount(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_List(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk, subject_pk, parent_pk, policy_expired_at, created_at FROM subject_relation
		 WHERE parent_pk = (.*) ORDER BY pk DESC LIMIT (.*) OFFSET (.*)`
		mockRows := sqlmock.NewRows(
			[]string{"pk", "subject_pk", "parent_pk", "policy_expired_at"},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), 0, 10).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		relations, err := manager.ListPagingGroupMember(int64(1), 0, 10)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_GetSubjectGroupCount(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT COUNT(.*) FROM subject_relation WHERE subject_pk = `
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		cnt, err := manager.GetSubjectGroupCount(int64(1))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_ListPagingRelation(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT pk, subject_pk, parent_pk, policy_expired_at, created_at FROM subject_relation
		 WHERE subject_pk = (.*) ORDER BY pk DESC LIMIT (.*) OFFSET (.*)`
		mockRows := sqlmock.NewRows(
			[]string{
				"pk", "subject_pk", "parent_pk", "policy_expired_at",
			},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(10), int64(0)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		relations, err := manager.ListPagingSubjectGroups(int64(1), 10, 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_BulkDeleteByMembersWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM subject_relation WHERE parent_pk=(.*) AND subject_pk in (.*)`).WithArgs(
			int64(1), int64(2),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectGroupManager{DB: db}
		cnt, err := manager.BulkDeleteByGroupMembersWithTx(tx, int64(1), []int64{2})

		tx.Commit()
		assert.NoError(t, err)
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_GetSubjectGroupCountBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT COUNT(.*) FROM subject_relation WHERE subject_pk = (.*) AND policy_expired_at < (.*)`
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(10)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		cnt, err := manager.GetSubjectGroupCountBeforeExpiredAt(int64(1), int64(10))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_ListPagingRelationBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		 pk,
		 subject_pk,
		 parent_pk,
		 policy_expired_at,
		 created_at
		 FROM subject_relation
		 WHERE subject_pk = (.*)
		 AND policy_expired_at < (.*)
		 ORDER BY policy_expired_at DESC, pk DESC
		 LIMIT (.*) OFFSET (.*)`
		mockRows := sqlmock.NewRows(
			[]string{
				"pk", "subject_pk", "parent_pk", "policy_expired_at",
			},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(1000), int64(10), int64(0)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		relations, err := manager.ListPagingSubjectGroupBeforeExpiredAt(int64(1), int64(1000), 10, 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_GetMemberCountBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT COUNT(.*) FROM subject_relation WHERE parent_pk = (.*) AND policy_expired_at < (.*)`
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(10)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		cnt, err := manager.GetGroupMemberCountBeforeExpiredAt(int64(1), int64(10))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_UpdateExpiredAtWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectPrepare(`^UPDATE subject_relation SET policy_expired_at = (.*) WHERE pk = (.*)`)
		mock.ExpectExec(`^UPDATE subject_relation SET policy_expired_at =`).WithArgs(
			int64(2), int64(1),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		subjects := []SubjectRelationForUpdateExpiredAt{{
			PK:        1,
			ExpiredAt: 2,
		}}

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectGroupManager{DB: db}
		err = manager.UpdateExpiredAtWithTx(tx, subjects)

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_subjectRelationManager_BulkCreateWithTx(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`^INSERT INTO subject_relation`).WithArgs(
			int64(2), int64(1), int64(3),
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		relations := []SubjectRelation{{
			SubjectPK: 2,
			GroupPK:   1,
			ExpiredAt: 3,
		}}

		tx, err := db.Beginx()
		assert.NoError(t, err)

		manager := &subjectGroupManager{DB: db}
		err = manager.BulkCreateWithTx(tx, relations)

		tx.Commit()

		assert.NoError(t, err)
	})
}

func Test_subjectRelationManager_ListThinRelationBySubjectPKGroupPKs(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT
		pk,
		subject_pk,
		parent_pk,
		policy_expired_at,
		created_at
		FROM subject_relation
		WHERE subject_pk = (.*)
		AND parent_pk in (.*)`
		mockRows := sqlmock.NewRows(
			[]string{
				"subject_pk",
				"parent_pk",
				"policy_expired_at",
			},
		).AddRow(int64(1), int64(2), int64(3))
		mock.ExpectQuery(mockQuery).WithArgs(int64(123), int64(1)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		relations, err := manager.ListRelationBySubjectPKGroupPKs(
			123,
			[]int64{1},
		)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_GetExpiredAtBySubjectGroup(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT policy_expired_at FROM subject_relation WHERE subject_pk = (.*) AND parent_pk = (.*)`
		mockRows := sqlmock.NewRows([]string{"policy_expired_at"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(10)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		expiredAt, err := manager.GetExpiredAtBySubjectGroup(int64(1), int64(10))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, expiredAt, int64(1))
	})
}

func Test_subjectRelationManager_GetSubjectSystemGroupCount(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT COUNT(.*) FROM subject_relation t LEFT JOIN group_system_auth_type s ON t.parent_pk = s.group_pk WHERE t.subject_pk =`
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), "demo").WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		cnt, err := manager.GetSubjectSystemGroupCount(int64(1), "demo")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_GetSubjectSystemGroupCountBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT COUNT(.*) FROM subject_relation t LEFT JOIN group_system_auth_type s ON t.parent_pk = s.group_pk WHERE t.subject_pk =`
		mockRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2), "demo").WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		cnt, err := manager.GetSubjectSystemGroupCountBeforeExpiredAt(int64(1), "demo", int64(2))

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, cnt, int64(1))
	})
}

func Test_subjectRelationManager_ListPagingSubjectSystemGroups(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT t.pk, t.subject_pk, t.parent_pk, t.policy_expired_at, t.created_at
		 FROM subject_relation t LEFT JOIN group_system_auth_type s ON t.parent_pk = s.group_pk
		 WHERE t.subject_pk = (.*) AND s.system_id = (.*) ORDER BY t.pk DESC LIMIT (.*) OFFSET (.*)`
		mockRows := sqlmock.NewRows(
			[]string{
				"pk", "subject_pk", "parent_pk", "policy_expired_at",
			},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), "demo", int64(10), int64(0)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		relations, err := manager.ListPagingSubjectSystemGroups(int64(1), "demo", 10, 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}

func Test_subjectRelationManager_ListPagingSubjectSystemGroupBeforeExpiredAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT t.pk, t.subject_pk, t.parent_pk, t.policy_expired_at, t.created_at
		 FROM subject_relation t LEFT JOIN group_system_auth_type s ON t.parent_pk = s.group_pk
		 WHERE t.subject_pk = (.*) AND t.policy_expired_at < (.*) AND s.system_id = (.*) ORDER BY t.pk DESC LIMIT (.*) OFFSET (.*)`
		mockRows := sqlmock.NewRows(
			[]string{
				"pk", "subject_pk", "parent_pk", "policy_expired_at",
			},
		).AddRow(int64(1), int64(2), int64(3), int64(0))
		mock.ExpectQuery(mockQuery).WithArgs(int64(1), int64(2), "demo", int64(10), int64(0)).WillReturnRows(mockRows)

		manager := &subjectGroupManager{DB: db}
		relations, err := manager.ListPagingSubjectSystemGroupBeforeExpiredAt(int64(1), "demo", int64(2), 10, 0)

		assert.NoError(t, err, "query from db fail.")
		assert.Len(t, relations, 1)
	})
}
