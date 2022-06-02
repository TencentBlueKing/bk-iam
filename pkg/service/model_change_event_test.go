package service

import (
	"errors"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
	"iam/pkg/database/dao/mock"
)

var _ = Describe("ModelEventService", func() {
	var (
		ctl         *gomock.Controller
		mockManager *mock.MockModelChangeEventManager
		svc         ModelChangeEventService
	)
	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		mockManager = mock.NewMockModelChangeEventManager(ctl)
		svc = &modelChangeEventService{
			manager: mockManager,
		}

	})
	AfterEach(func() {
		ctl.Finish()
	})

	Context("DeleteByStatus", func() {
		var (
			now     int64
			patches *gomonkey.Patches
		)
		BeforeEach(func() {
			now = time.Now().Unix()
			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()
			patches = gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
		})
		AfterEach(func() {
			patches.Reset()
		})
		It("ok", func() {
			mockManager.EXPECT().
				DeleteByStatusWithTx(gomock.Any(), ModelChangeEventStatusFinished, now, int64(10)).
				Return(int64(10), nil)

			err := svc.DeleteByStatus(ModelChangeEventStatusFinished, now, int64(10))
			assert.NoError(GinkgoT(), err)
		})
		It("error", func() {
			mockManager.EXPECT().
				DeleteByStatusWithTx(gomock.Any(), ModelChangeEventStatusFinished, now, int64(10)).
				Return(int64(10), errors.New("error"))

			err := svc.DeleteByStatus(ModelChangeEventStatusFinished, now, int64(10))
			assert.Regexp(GinkgoT(), "manager.DeleteByStatusWithTx (.*) fail", err.Error())
		})
	})

})
