package dao

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

var _ = Describe("GroupResourcePolicyManager", func() {
	var (
		mock    sqlmock.Sqlmock
		db      *sqlx.DB
		manager GroupResourcePolicyManager
	)
	BeforeEach(func() {
		db, mock = database.NewMockSqlxDB()
		manager = &groupResourcePolicyManager{DB: db}
	})

	It("GetPKAndActionPKs", func() {
		mockData := []interface{}{GroupResourcePolicyPKActionPKs{PK: int64(1), ActionPKs: "[1,2,3]"}}
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(
			"^SELECT pk, action_pks FROM group_resource_policy WHERE group_pk = (.*) AND template_id = (.*) AND system_id = (.*) AND action_related_resource_type_pk = (.*) AND resource_type_pk = (.*) AND resource_id = (.*) LIMIT 1$",
		).WithArgs(
			int64(1), int64(2), "test", int64(3), int64(4), "resource_id",
		).WillReturnRows(mockRows)

		pkActionPKs, err := manager.GetPKAndActionPKs(int64(1), int64(2), "test", int64(3), int64(4), "resource_id")

		assert.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), int64(1), pkActionPKs.PK)
		assert.Equal(GinkgoT(), "[1,2,3]", pkActionPKs.ActionPKs)
	})

	It("BulkInsertWithTx", func() {
		mock.ExpectBegin()
		mock.ExpectExec(
			`INSERT INTO group_resource_policy`,
		).WithArgs(
			int64(1), int64(2), "test", "[1,2,3]", int64(3), int64(4), "resource_id",
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(GinkgoT(), err)

		err = manager.BulkInsertWithTx(tx, []GroupResourcePolicy{
			{
				GroupPK:                     int64(1),
				TemplateID:                  int64(2),
				SystemID:                    "test",
				ActionPKs:                   "[1,2,3]",
				ActionRelatedResourceTypePK: int64(3),
				ResourceTypePK:              int64(4),
				ResourceID:                  "resource_id",
			},
		})
		tx.Commit()

		assert.NoError(GinkgoT(), err)
	})

	It("BulkUpdateActionPKsWithTx", func() {
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE group_resource_policy SET action_pks = (.*) WHERE pk = (.*)`)
		mock.ExpectExec(`UPDATE group_resource_policy SET action_pks = (.*) WHERE pk = (.*)`).
			WithArgs("[1,2,3]", int64(1)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(GinkgoT(), err)

		err = manager.BulkUpdateActionPKsWithTx(tx, []GroupResourcePolicyPKActionPKs{
			{
				PK:        int64(1),
				ActionPKs: "[1,2,3]",
			},
		})
		tx.Commit()

		assert.NoError(GinkgoT(), err)
	})

	It("BulkDeleteByPKsWithTx", func() {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM group_resource_policy WHERE pk IN (.*)`).
			WithArgs(int64(1), int64(2)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(GinkgoT(), err)

		err = manager.BulkDeleteByPKsWithTx(tx, []int64{int64(1), int64(2)})
		tx.Commit()

		assert.NoError(GinkgoT(), err)
	})
})
