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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

var _ = Describe("Request", func() {

	Describe("ReadRequestBody", func() {

		It("nil body", func() {

			r := &http.Request{Body: nil}
			body, err := util.ReadRequestBody(r)
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), body)
		})

		It("Empty response", func() {
			// read empty body
			r, _ := util.NewRequestEmptyResponse()
			body, err := util.ReadRequestBody(r)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []byte(""), body)
		})

		It("Error response", func() {
			// read error body, will error
			r, _ := util.NewRequestErrorResponse()
			_, err := util.ReadRequestBody(r)
			assert.Error(GinkgoT(), err)

		})

		It("read test content from body", func() {
			// read test content from body
			r, _ := util.NewRequestResponse()
			body, err := util.ReadRequestBody(r)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), util.TestingContent, body)
		})

		It("read twice", func() {
			// test read twice
			r, _ := util.NewRequestResponse()
			body, err := util.ReadRequestBody(r)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), util.TestingContent, body)

			body, err = util.ReadRequestBody(r)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), util.TestingContent, body)
		})

	})

	Describe("RequestID", func() {
		var c *gin.Context
		BeforeEach(func() {
			c = &gin.Context{}
		})

		It("GetRequestID", func() {
			id := util.GetRequestID(c)
			assert.Equal(GinkgoT(), "", id)
		})

		It("SetRequestID", func() {
			util.SetRequestID(c, "123")

			id := util.GetRequestID(c)
			assert.Equal(GinkgoT(), "123", id)
		})

	})

	Describe("ClientID", func() {
		var c *gin.Context
		BeforeEach(func() {
			c = &gin.Context{}
		})

		It("GetClientID", func() {
			id := util.GetClientID(c)
			assert.Equal(GinkgoT(), "", id)
		})

		It("SetClientID", func() {
			util.SetClientID(c, "test")

			id := util.GetClientID(c)
			assert.Equal(GinkgoT(), "test", id)
		})
	})

	Describe("Error", func() {
		var c *gin.Context
		BeforeEach(func() {
			c = &gin.Context{}
		})

		It("GetError", func() {
			_, ok := util.GetError(c)
			assert.False(GinkgoT(), ok)
		})

		It("SetError", func() {
			expected := errors.New("test")
			util.SetError(c, expected)
			err, ok := util.GetError(c)
			assert.True(GinkgoT(), ok)
			assert.Equal(GinkgoT(), expected, err)
		})

	})

	Describe("BasicAuthAuthorizationHeader", func() {
		It("valid basic auth", func() {
			assert.Equal(GinkgoT(), "Basic YWRtaW46cGFzc3dvcmQ=", util.BasicAuthAuthorizationHeader("admin", "password"))
		})

	})
})
