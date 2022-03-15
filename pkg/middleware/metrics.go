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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"iam/pkg/metric"
	"iam/pkg/util"
)

// Metrics ...
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Debug("Middleware: Metrics")

		start := time.Now()

		c.Next()

		duration := time.Since(start)

		clientID := util.GetClientID(c)
		status := strconv.Itoa(c.Writer.Status())

		e := "0"
		if _, hasError := util.GetError(c); hasError {
			e = "1"
		}

		// request count
		metric.RequestCount.With(prometheus.Labels{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"status":    status,
			"error":     e,
			"client_id": clientID,
		}).Inc()

		// request duration, in ms
		metric.RequestDuration.With(prometheus.Labels{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"status":    status,
			"client_id": clientID,
		}).Observe(float64(duration / time.Millisecond))
	}
}
