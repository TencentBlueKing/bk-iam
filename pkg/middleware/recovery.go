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
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func panicLog(rval interface{}) {
	debug.PrintStack()
	rvalStr := fmt.Sprint(rval)
	err := errors.New(rvalStr)
	log.WithError(err).Error(fmt.Sprintf("system error %s", debug.Stack()))
}

func isBrokenPipeError(err interface{}) bool {
	if netErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := netErr.Err.(*os.SyscallError); ok {
			if strings.Contains(strings.ToLower(sysErr.Error()), "broken pipe") ||
				strings.Contains(strings.ToLower(sysErr.Error()), "connection reset by peer") {
				return true
			}
		}
	}
	return false
}

const sentryValuesKey = "sentry"

// Recovery ...
func Recovery(withSentry bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debug("Middleware: Recovery")

		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				brokenPipe := isBrokenPipeError(err)

				panicLog(err)

				if withSentry && !brokenPipe {
					hub := sentry.CurrentHub().Clone()
					hub.Scope().SetRequest(c.Request)
					c.Set(sentryValuesKey, hub)

					hub.RecoverWithContext(
						context.WithValue(c.Request.Context(), sentry.RequestContextKey, c.Request),
						err,
					)
				}

				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()

		c.Next()
	}
}
