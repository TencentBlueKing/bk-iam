/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package component

import (
	"net/http"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"moul.io/http2curl"

	"iam/pkg/logging"
	"iam/pkg/metric"
	"iam/pkg/util"
)

const (
	maxResponseBodyLength = 10240
)

var (
	BkRemoteResource RemoteResourceClient
	BkAuth           AuthClient
)

// InitBkRemoteResourceClient ...
func InitBkRemoteResourceClient() {
	BkRemoteResource = NewRemoteResourceClient()
}

func InitBkAuthClient(bkAuthHost, bkAppCode, bkAppSecret string) {
	BkAuth = NewAuthClient(bkAuthHost, bkAppCode, bkAppSecret)
}

// CallbackFunc ...
type CallbackFunc func(response gorequest.Response, v interface{}, body []byte, errs []error)

// NewMetricCallback ...
func NewMetricCallback(system string, start time.Time) CallbackFunc {
	return func(response gorequest.Response, v interface{}, body []byte, errs []error) {
		duration := time.Since(start)
		metric.ComponentRequestDuration.With(prometheus.Labels{
			"method":    response.Request.Method,
			"path":      response.Request.URL.Path,
			"status":    strconv.Itoa(response.StatusCode),
			"component": system,
		}).Observe(float64(duration / time.Millisecond))
	}
}

// AsCurlCommand returns a string representing the runnable `curl' command
// version of the request.
func AsCurlCommand(request *gorequest.SuperAgent) (string, error) {
	req, err := request.MakeRequest()
	if err != nil {
		return "", err
	}

	// 脱敏, 去掉-H 中 Authorization
	req.Header.Del("Authorization")

	cmd, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return "", err
	}
	return cmd.String(), nil
}

func logFailHTTPRequest(
	start time.Time,
	request *gorequest.SuperAgent,
	response gorequest.Response,
	respBody []byte,
	errs []error,
	data responseStruct,
) {
	logger := logging.GetComponentLogger()

	responseBodyErr := data.Error()
	// check will log or not?
	willLog := logger.GetLevel() == log.DebugLevel ||
		len(errs) != 0 || response.StatusCode != http.StatusOK || responseBodyErr != nil

	if !willLog {
		return
	}

	// duration
	duration := time.Since(start)

	// curl dump
	dump, err := AsCurlCommand(request)
	if err != nil {
		logger.Error("component request AsCurlCommand fail")
	}

	// status
	status := -1
	if response != nil {
		status = response.StatusCode
	}

	// message
	message := ""
	if responseBodyErr != nil {
		message = responseBodyErr.Error()
	}
	// response body
	respBodyStr := ""
	if respBody != nil {
		respBodyStr = util.TruncateBytesToString(respBody, maxResponseBodyLength)
	}

	fields := log.Fields{
		"status":        status,
		"errs":          errs,
		"request":       dump,
		"response_body": respBodyStr,
		"latency":       float64(duration / time.Millisecond),
	}

	entry := logger.WithFields(fields)
	if logger.GetLevel() == log.DebugLevel {
		entry.Debug(message)
	} else {
		entry.Error(message)

		// report to sentry
		util.ReportToSentry("component error", fields)
	}
}
