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
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"iam/pkg/logging"
	"iam/pkg/util"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write ...
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// APILogger ...
func APILogger() gin.HandlerFunc {
	logger := logging.GetAPILogger()

	return func(c *gin.Context) {
		fields := logContextFields(c)
		logger.With(fields...).Info("-")
	}
}

// WebLogger ...
func WebLogger() gin.HandlerFunc {
	logger := logging.GetWebLogger()

	return func(c *gin.Context) {
		fields := logContextFields(c)
		logger.With(fields...).Info("-")
	}
}

func logContextFields(c *gin.Context) []zap.Field {
	start := time.Now()

	// request body
	var body string
	requestBody, err := util.ReadRequestBody(c.Request)
	if err != nil {
		body = ""
	} else {
		body = util.TruncateBytesToString(requestBody, 1024)
	}

	newWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = newWriter

	c.Next()

	duration := time.Since(start)
	// always add 1ms, in case the 0ms in log
	latency := float64(duration/time.Millisecond) + 1

	e, hasError := util.GetError(c)
	if !hasError {
		e = ""
	}

	params := util.TruncateString(c.Request.URL.RawQuery, 1024)
	fields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("params", params),
		zap.String("body", body),
		zap.Int("status", c.Writer.Status()),
		zap.Float64("latency", latency),
		zap.String("request_id", util.GetRequestID(c)),
		zap.String("client_id", util.GetClientID(c)),
		zap.String("client_ip", c.ClientIP()),
		zap.Any("error", e),
	}

	if hasError {
		fields = append(fields, zap.String("response_body", newWriter.body.String()))
	} else {
		fields = append(fields, zap.String("response_body", util.TruncateString(newWriter.body.String(), 1024)))
	}

	if hasError && e != nil {
		util.ReportToSentry(
			fmt.Sprintf("%s %s error", c.Request.Method, c.Request.URL.Path),
			map[string]interface{}{
				"fields": fields,
			},
		)
	}

	return fields
}
