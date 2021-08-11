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
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Response ...
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// DebugResponse ...
type DebugResponse struct {
	Response
	Debug interface{} `json:"debug"`
}

// BaseJSONResponse make the response more Explicit
func BaseJSONResponse(c *gin.Context, status int, code int, message string, data interface{}) {
	// 通过 code = 0 或 非0, 确认是否成功, 不增加result字段
	body := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.JSON(status, body)
}

// BaseErrorJSONResponse ...
func BaseErrorJSONResponse(c *gin.Context, code int, message string) {
	BaseJSONResponse(c, http.StatusOK, code, message, gin.H{})
}

// SuccessJSONResponse ...
func SuccessJSONResponse(c *gin.Context, message string, data interface{}) {
	BaseJSONResponse(c, http.StatusOK, NoError, message, data)
}

// SuccessJSONResponseWithDebug ...
func SuccessJSONResponseWithDebug(c *gin.Context, message string, data interface{}, debug interface{}) {
	if debug == nil || reflect.ValueOf(debug).IsNil() {
		SuccessJSONResponse(c, message, data)
		return
	}

	body := DebugResponse{
		Response: Response{
			Code:    NoError,
			Message: message,
			Data:    data,
		},
		Debug: debug,
	}
	c.JSON(http.StatusOK, body)
}

// =============== impls of some common error response ===============

// NewErrorJSONResponse ...
func NewErrorJSONResponse(errorCode int, defaultMessage string) func(c *gin.Context, message string) {
	return func(c *gin.Context, message string) {
		msg := defaultMessage
		if message != "" {
			msg = fmt.Sprintf("%s:%s", msg, message)
		}
		BaseErrorJSONResponse(c, errorCode, msg)
	}
}

// BadRequestErrorJSONResponse ...
var (
	BadRequestErrorJSONResponse = NewErrorJSONResponse(BadRequestError, "bad request")
	ParamErrorJSONResponse      = NewErrorJSONResponse(ParamError, "param error")
	ForbiddenJSONResponse       = NewErrorJSONResponse(ForbiddenError, "no permission")
	UnauthorizedJSONResponse    = NewErrorJSONResponse(UnauthorizedError, "unauthorized")
	NotFoundJSONResponse        = NewErrorJSONResponse(NotFoundError, "not found")
	ConflictJSONResponse        = NewErrorJSONResponse(ConflictError, "conflict")
	TooManyRequestsJSONResponse = NewErrorJSONResponse(TooManyRequests, "too many requests")
)

// SystemErrorJSONResponse ...
func SystemErrorJSONResponse(c *gin.Context, err error) {
	message := fmt.Sprintf("system error[request_id=%s]: %s", GetRequestID(c), err.Error())
	SetError(c, err)
	BaseErrorJSONResponse(c, SystemError, message)
}

// SystemErrorJSONResponseWithDebug ...
func SystemErrorJSONResponseWithDebug(c *gin.Context, err error, debug interface{}) {
	if debug == nil || reflect.ValueOf(debug).IsNil() {
		SystemErrorJSONResponse(c, err)
		return
	}

	message := fmt.Sprintf("system error[request_id=%s]: %s", GetRequestID(c), err.Error())
	SetError(c, err)

	body := DebugResponse{
		Response: Response{
			Code:    SystemError,
			Message: message,
			Data:    gin.H{},
		},
		Debug: debug,
	}
	c.JSON(http.StatusOK, body)
}
