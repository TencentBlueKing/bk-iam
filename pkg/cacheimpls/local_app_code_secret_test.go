package cacheimpls_test

import (
	"errors"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	gocache "github.com/wklken/go-cache"

	"iam/pkg/cacheimpls"
	"iam/pkg/component"
	mock2 "iam/pkg/component/mock"
	"iam/pkg/database/edao"
	"iam/pkg/database/edao/mock"
)

var _ = Describe("LocalAppCodeSecret", func() {
	Describe("VerifyAppCodeAppSecret", func() {
		var ctl *gomock.Controller
		var mockManager *mock.MockAppSecretManager
		var patches *gomonkey.Patches
		BeforeEach(func() {
			cacheimpls.LocalAppCodeAppSecretCache = gocache.New(12*time.Hour, 5*time.Minute)

			ctl = gomock.NewController(GinkgoT())
			mockManager = mock.NewMockAppSecretManager(ctl)
		})

		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("hit", func() {
			cacheimpls.LocalAppCodeAppSecretCache.Set("app:123", true, 12*time.Hour)
			ok := cacheimpls.VerifyAppCodeAppSecretFromDB("app", "123")
			assert.True(GinkgoT(), ok)
		})

		It("miss, get from database error", func() {
			mockManager.EXPECT().
				Exists(gomock.Any(), gomock.Any()).
				Return(false, errors.New("errror happend")).
				AnyTimes()
			patches = gomonkey.ApplyFunc(edao.NewAppSecretManager,
				func() edao.AppSecretManager {
					return mockManager
				})

			ok := cacheimpls.VerifyAppCodeAppSecretFromDB("app", "123")
			assert.False(GinkgoT(), ok)
		})

		It("miss, get from database valid", func() {
			mockManager.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
			patches = gomonkey.ApplyFunc(edao.NewAppSecretManager,
				func() edao.AppSecretManager {
					return mockManager
				})

			ok := cacheimpls.VerifyAppCodeAppSecretFromDB("app", "123")
			assert.False(GinkgoT(), ok)
		})

		It("miss, get from database invalid", func() {
			mockManager.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
			patches = gomonkey.ApplyFunc(edao.NewAppSecretManager,
				func() edao.AppSecretManager {
					return mockManager
				})

			ok := cacheimpls.VerifyAppCodeAppSecretFromDB("app", "123")
			assert.True(GinkgoT(), ok)
		})
	})

	Describe("VerifyAppCodeAppSecretFromAuth", func() {
		var ctl *gomock.Controller
		var mockCli *mock2.MockAuthClient
		BeforeEach(func() {
			cacheimpls.LocalAuthAppAccessKeyCache = gocache.New(12*time.Hour, 5*time.Minute)

			ctl = gomock.NewController(GinkgoT())
			mockCli = mock2.NewMockAuthClient(ctl)
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("hit", func() {
			cacheimpls.LocalAuthAppAccessKeyCache.Set("app:123", true, 12*time.Hour)
			ok := cacheimpls.VerifyAppCodeAppSecretFromAuth("app", "123")
			assert.True(GinkgoT(), ok)
		})

		It("miss, get from bkauth error", func() {
			mockCli.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, errors.New("errror happend")).AnyTimes()
			component.BkAuth = mockCli

			ok := cacheimpls.VerifyAppCodeAppSecretFromAuth("app", "123")
			assert.False(GinkgoT(), ok)
		})

		It("miss, get from bkauth valid", func() {
			mockCli.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
			component.BkAuth = mockCli

			ok := cacheimpls.VerifyAppCodeAppSecretFromAuth("app", "123")
			assert.False(GinkgoT(), ok)
		})

		It("miss, get from bkauth invalid", func() {
			mockCli.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
			component.BkAuth = mockCli

			ok := cacheimpls.VerifyAppCodeAppSecretFromAuth("app", "123")
			assert.True(GinkgoT(), ok)
		})
	})
})
