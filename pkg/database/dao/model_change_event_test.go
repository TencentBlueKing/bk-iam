package dao

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

var _ = Describe("ModelEventManager", func() {
	var (
		mock    sqlmock.Sqlmock
		db      *sqlx.DB
		manager ModelChangeEventManager
	)
	BeforeEach(func() {
		db, mock = database.NewMockSqlxDB()
		manager = &modelChangeEventManager{DB: db}
	})

	It("DeleteByStatusWithTx", func() {
		mock.ExpectBegin()
		mock.ExpectExec(
			`^DELETE FORM model_change_event WHERE status = (.*) AND updated_at <= FROM_UNIXTIME((.*)) LIMIT (.*)$`,
		).WithArgs(
			"finished", int64(4102444800), int64(10),
		).WillReturnResult(sqlmock.NewResult(0, 10))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(GinkgoT(), err)

		rowsAffected, err := manager.DeleteByStatusWithTx(tx, "finished", int64(4102444800), int64(10))
		tx.Commit()

		assert.NoError(GinkgoT(), err)
		assert.Equal(GinkgoT(), rowsAffected, int64(10))
	})
})
