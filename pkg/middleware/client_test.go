/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cacheimpls"
	"iam/pkg/config"
	"iam/pkg/util"
)

func TestClientAuthMiddleware(t *testing.T) {
	t.Parallel()

	cacheimpls.InitVerifyAppCodeAppSecret(false)

	// 1. without appCode appSecret
	r := gin.Default()
	r.Use(ClientAuthMiddleware([]byte("")))
	util.NewTestRouter(r)

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	body, err := ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(body), "1901401"))

	// 2. appCode and appSecret not empty, but not in cache
	// TODO

	// 3. valid
	// TODO

	// 4. apigateway
}

func TestSuperClientMiddleware(t *testing.T) {
	t.Parallel()

	// 1. right
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	util.SetClientID(c, "bk_iam_app")
	config.InitSuperAppCode("bk_iam_app,bk_iam")

	SuperClientMiddleware()(c)

	assert.Equal(t, 200, w.Code)

	body, err := ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(body), "1901401"))

	// 2. wrong
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	util.SetClientID(c, "abc")
	config.InitSuperAppCode("bk_iam_app,bk_iam")

	SuperClientMiddleware()(c)

	assert.Equal(t, 200, w.Code)

	body, err = ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(body), "1901401"))
}
