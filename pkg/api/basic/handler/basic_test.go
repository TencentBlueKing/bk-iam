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
	"net/http"
	"testing"

	"iam/pkg/util"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

// GinkgoT() not support testing.TB yet https://github.com/onsi/ginkgo/issues/582
func TestPing(t *testing.T) {
	t.Parallel()

	r := util.SetupRouter()
	r.GET("/ping", Ping)

	apitest.New().
		Handler(r).
		Get("/ping").
		Expect(t).
		Body(`{"message":"pong"}`).
		Status(http.StatusOK).
		End()
}

func TestVersion(t *testing.T) {
	t.Parallel()

	r := util.SetupRouter()
	r.GET("/version", Version)

	apitest.New().
		Handler(r).
		Get("/version").
		Expect(t).
		Assert(util.NewJSONAssertFunc(t, func(m map[string]interface{}) error {
			assert.Contains(t, m, "version")
			assert.Contains(t, m, "commit")
			assert.Contains(t, m, "buildTime")
			assert.Contains(t, m, "goVersion")
			assert.Contains(t, m, "env")
			return nil
		})).
		Status(http.StatusOK).
		End()
}
