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
	"io"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

// 这里是通用的 FieldError 处理, 如果需要针对某些字段或struct做定制, 需要自行定义一个

// reference:
// 1. https://github.com/gin-gonic/gin/issues/430
// 2. https://medium.com/@seb.nyberg/better-validation-errors-in-go-gin-88f983564a3d

// ValidationFieldError ...
type ValidationFieldError struct {
	Err validator.FieldError
}

// String ...
func (v ValidationFieldError) String() string {
	e := v.Err

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s", e.Field(), e.Param())
	case "min":
		return fmt.Sprintf("%s must be longer than %s", e.Field(), e.Param())
	case "email":
		return "Invalid email format"
	case "len":
		return fmt.Sprintf("%s must be %s characters long", e.Field(), e.Param())
	case "gt":
		return fmt.Sprintf("%s must greater than %s", e.Field(), e.Param())
	case "gte":
		return fmt.Sprintf("%s must greater or equals to %s", e.Field(), e.Param())
	case "lt":
		return fmt.Sprintf("%s must less than %s", e.Field(), e.Param())
	case "lte":
		return fmt.Sprintf("%s must less or equals to %s", e.Field(), e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of '%s'", e.Field(), e.Param())
	}

	return fmt.Sprintf("%s is not valid, condition: %s", e.Field(), e.ActualTag())
}

// ValidationErrorMessage ...
func ValidationErrorMessage(err error) string {
	if err == io.EOF {
		return "EOF, json decode fail"
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		message := fmt.Sprintf("json decode or validate fail, err=%s", err)
		log.Info(message)
		return message
	}

	// currently, only return the first error
	for _, fieldErr := range validationErrs {
		return ValidationFieldError{fieldErr}.String()
	}

	return "validationErrs with no error message"
}
