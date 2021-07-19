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
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPanicLog(t *testing.T) {
	t.Parallel()

	err1 := errors.New("test")
	panicLog(err1)
}

func TestIsBrokenPipeError(t *testing.T) {
	t.Parallel()

	err1 := errors.New("test")
	assert.False(t, isBrokenPipeError(err1))

	err2 := &net.OpError{
		Op:     "",
		Net:    "",
		Source: nil,
		Addr:   nil,
		Err: &os.SyscallError{
			Syscall: "",
			Err:     errors.New("broken pipe"),
		},
	}
	assert.True(t, isBrokenPipeError(err2))
}

func TestRecovery(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	r := gin.Default()
	r.Use(Recovery(true))

	r.GET("/ping", func(c *gin.Context) {
		b := 0
		a := 1 / b
		fmt.Println(a, b)
		c.String(200, "pong")
	})

	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}
