/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"iam/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSystemExists(t *testing.T) {
	t.Parallel()

	r := gin.Default()
	r.Use(SystemExists())
	util.NewTestRouter(r)

	// 1. no system_id in url
	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	body, err := ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(body), "1901400"))

	// 2. TODO: mock systemSvc.Exists
}

func TestSystemExistsAndClientValid(t *testing.T) {
	t.Parallel()

	r := gin.Default()
	r.Use(SystemExistsAndClientValid())
	util.NewTestRouter(r)

	// 1. no system_id in url
	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	body, err := ioutil.ReadAll(w.Body)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(body), "1901400"))

	// 2. TODO: mock systemSvc.Exists
}
