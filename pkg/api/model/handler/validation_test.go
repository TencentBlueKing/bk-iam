/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Validation", func() {
	Describe("validateDeleteViaID", func() {
		var c *gin.Context
		BeforeEach(func() {
			c, _ = gin.CreateTestContext(httptest.NewRecorder())
		})

		It("not json body", func() {
			c.Request = &http.Request{
				Body: ioutil.NopCloser(bytes.NewBuffer([]byte("hello"))),
			}

			_, err := validateDeleteViaID(c)
			assert.Error(GinkgoT(), err)
		})

		It("json body but not array", func() {
			c.Request = &http.Request{
				Body: ioutil.NopCloser(bytes.NewBuffer([]byte("{}"))),
			}
			_, err := validateDeleteViaID(c)
			assert.Error(GinkgoT(), err)
		})

		It("array but empty", func() {
			c.Request = &http.Request{
				Body: ioutil.NopCloser(bytes.NewBuffer([]byte("[]"))),
			}
			_, err := validateDeleteViaID(c)
			assert.Error(GinkgoT(), err)
		})

		It("array not empty", func() {
			c.Request = &http.Request{
				Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`[{"id": "123"}]`))),
			}
			_, err := validateDeleteViaID(c)
			assert.NoError(GinkgoT(), err)
		})
	})
})
