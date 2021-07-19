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
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"iam/pkg/logging"
	"iam/pkg/util"
)

// only record `change` method audit log
var auditRequestMethod = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
	"PATCH":  true,
}

// Audit ...
func Audit() gin.HandlerFunc {
	logger := logging.GetAuditLogger()

	return func(c *gin.Context) {
		log.Debug("Middleware: Audit")

		_, willAudit := auditRequestMethod[c.Request.Method]

		start := time.Now()

		// request body
		var body string
		requestBody, err := util.ReadRequestBody(c.Request)
		if err != nil {
			body = ""
		} else {
			// NOTE: no truncation
			body = string(requestBody)
		}

		newWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = newWriter

		c.Next()

		// check error
		e, hasError := util.GetError(c)
		if !hasError {
			e = ""
		}

		if willAudit || hasError {
			duration := time.Since(start)
			// always add 1ms, in case the 0ms in log
			latency := float64(duration/time.Millisecond) + 1

			params := util.TruncateString(c.Request.URL.RawQuery, 1024)
			fields := log.Fields{
				"method":        c.Request.Method,
				"path":          c.Request.URL.Path,
				"params":        params,
				"body":          body,
				"response_body": util.TruncateString(newWriter.body.String(), 1024),
				"status":        c.Writer.Status(),
				"latency":       latency,

				"request_id": util.GetRequestID(c),
				"client_id":  util.GetClientID(c),
				"client_ip":  c.ClientIP(),

				"error": e,
			}

			// 审计日志只记录增删改 => 是否成功
			if willAudit {
				logger.WithFields(fields).Info("-")
			}
			// 增删改查日志是gbd => log to system default logger
			if hasError {
				logging.GetSystemLogger().WithFields(fields).Error("-")
			}
		}
	}
}
