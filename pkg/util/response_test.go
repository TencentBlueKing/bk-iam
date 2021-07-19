/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"testing"

	"iam/pkg/logging/debug"

	"iam/pkg/util"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

func readResponse(w *httptest.ResponseRecorder) util.Response {
	var got util.Response
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(GinkgoT(), err)
	return got
}

var _ = Describe("Response", func() {

	var c *gin.Context
	//var r *gin.Engine
	var w *httptest.ResponseRecorder
	BeforeEach(func() {
		w = httptest.NewRecorder()
		gin.SetMode(gin.ReleaseMode)
		//gin.DefaultWriter = ioutil.Discard
		c, _ = gin.CreateTestContext(w)
		//c, r = gin.CreateTestContext(w)
		//r.Use(gin.Recovery())
	})

	It("BaseJSONResponse", func() {
		util.BaseJSONResponse(c, 200, 10000, "ok", nil)

		assert.Equal(GinkgoT(), 200, w.Code)

		got := readResponse(w)
		assert.Equal(GinkgoT(), 10000, got.Code)
		assert.Equal(GinkgoT(), "ok", got.Message)
	})

	It("BaseErrorJSONResponse", func() {
		util.BaseErrorJSONResponse(c, 1901000, "error")
		assert.Equal(GinkgoT(), 200, c.Writer.Status())

		got := readResponse(w)
		assert.Equal(GinkgoT(), 1901000, got.Code)
		assert.Equal(GinkgoT(), "error", got.Message)
	})

	It("SuccessJSONResponse", func() {
		util.SuccessJSONResponse(c, "ok", nil)
		assert.Equal(GinkgoT(), 200, c.Writer.Status())

		got := readResponse(w)
		assert.Equal(GinkgoT(), 0, got.Code)
		assert.Equal(GinkgoT(), "ok", got.Message)
	})

	Context("SuccessJSONResponseWithDebug", func() {

		It("debug is nil", func() {
			util.SuccessJSONResponseWithDebug(c, "ok", nil, nil)
			assert.Equal(GinkgoT(), 200, c.Writer.Status())

			got := readResponse(w)
			assert.Equal(GinkgoT(), util.NoError, got.Code)
		})

		It("debug is not nil", func() {
			util.SuccessJSONResponseWithDebug(c, "ok", nil, map[string]interface{}{"hello": "world"})
			assert.Equal(GinkgoT(), 200, c.Writer.Status())

			got := readResponse(w)
			assert.Equal(GinkgoT(), util.NoError, got.Code)
		})
	})

	It("BadRequestErrorJSONResponse", func() {
		util.BadRequestErrorJSONResponse(c, "error")
		assert.Equal(GinkgoT(), 200, c.Writer.Status())

		got := readResponse(w)
		assert.Equal(GinkgoT(), util.BadRequestError, got.Code)
		assert.Equal(GinkgoT(), "bad request:error", got.Message)
	})

	It("SystemErrorJSONResponse", func() {
		util.SystemErrorJSONResponse(c, errors.New("anError"))
		assert.Equal(GinkgoT(), 200, c.Writer.Status())

		got := readResponse(w)
		assert.Equal(GinkgoT(), util.SystemError, got.Code)
		assert.Contains(GinkgoT(), got.Message, "system error")
	})

	Context("SystemErrorJSONResponseWithDebug", func() {

		It("debug is nil", func() {
			util.SystemErrorJSONResponseWithDebug(c, errors.New("anError"), nil)
			assert.Equal(GinkgoT(), 200, c.Writer.Status())

			got := readResponse(w)
			assert.Equal(GinkgoT(), util.SystemError, got.Code)
		})

		It("debug is not nil", func() {
			util.SystemErrorJSONResponseWithDebug(c, errors.New("anError"), map[string]interface{}{"hello": "world"})
			assert.Equal(GinkgoT(), 200, c.Writer.Status())

			got := readResponse(w)
			assert.Equal(GinkgoT(), util.SystemError, got.Code)
		})
	})

})

func BenchmarkSuccessJSONResponseWithDebugNilInterface(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	var entry *debug.Entry

	for i := 0; i < b.N; i++ {
		reflect.ValueOf(entry).IsNil()
	}
}
