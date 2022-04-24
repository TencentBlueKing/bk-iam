package cacheimpls

import (
	"errors"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/redis"
	"iam/pkg/service"
	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("ActionListCache", func() {
	var ctl *gomock.Controller
	var patches *gomonkey.Patches

	BeforeEach(func() {
		var (
			expiration = 30 * time.Minute
			mockCache  = redis.NewMockCache("mockCache", expiration)
		)
		ActionListCache = mockCache
		ctl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctl.Finish()
		patches.Reset()
	})
	Context("ListActionBySystem", func() {
		It("ActionListCache GetInto ok", func() {
			mockService := mock.NewMockActionService(ctl)
			mockService.EXPECT().ListBySystem("test").Return([]types.Action{{ID: "test"}}, nil).AnyTimes()

			patches = gomonkey.ApplyFunc(service.NewActionService,
				func() service.ActionService {
					return mockService
				})

			actions, err := ListActionBySystem("test")
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), actions, 1)
			assert.Equal(GinkgoT(), actions[0].ID, "test")

		})
		It("ActionListCache GetInto fail", func() {
			mockService := mock.NewMockActionService(ctl)
			mockService.EXPECT().ListBySystem("test").Return(nil, errors.New("error")).AnyTimes()

			patches = gomonkey.ApplyFunc(service.NewActionService,
				func() service.ActionService {
					return mockService
				})

			actions, err := ListActionBySystem("test")
			assert.Error(GinkgoT(), err)
			assert.Len(GinkgoT(), actions, 0)
		})
	})
})
