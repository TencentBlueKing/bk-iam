/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util

import (
	"time"

	"github.com/getsentry/sentry-go"

	"iam/pkg/errorx"
)

// Error Codes
const (
	NoError           = 0
	ParamError        = 1901002
	BadRequestError   = 1901400
	UnauthorizedError = 1901401
	ForbiddenError    = 1901403
	NotFoundError     = 1901404
	ConflictError     = 1901409
	SystemError       = 1901500
	TooManyRequests   = 1901429
)

// ReportToSentry is a shortcut to build and send an event to sentry
func ReportToSentry(message string, extra map[string]interface{}) {
	// report to sentry
	ev := sentry.NewEvent()
	ev.Message = message
	ev.Level = "error"
	ev.Timestamp = time.Now()
	ev.Extra = extra
	errorx.ReportEvent(ev)
}
