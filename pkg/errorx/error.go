/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package errorx

import (
	"errors"
	"fmt"
)

// IAMError is a wrapped struct for err
type IAMError struct {
	message string
	err     error
}

// Error show the error message
func (e IAMError) Error() string {
	return e.message
}

// Is check if the error is target
func (e IAMError) Is(target error) bool {
	if target == nil || e.err == nil {
		return e.err == target
	}

	return errors.Is(e.err, target)
}

// Unwrap will unwrap the wrapped error
func (e *IAMError) Unwrap() error {
	u, ok := e.err.(interface {
		Unwrap() error
	})
	if !ok {
		return e.err
	}

	return u.Unwrap()
}

func makeMessage(err error, layer, function, msg string) string {
	var message string
	var e IAMError
	if errors.As(err, &e) {
		message = fmt.Sprintf("[%s:%s] %s => %s", layer, function, msg, err.Error())
	} else {
		// jsoniter.marshal/unmarshal error not print, others?
		message = fmt.Sprintf("[%s:%s] %s => [Raw:Error] %v", layer, function, msg, err.Error())
	}

	return message
}

// Wrap will wrap the error with layer, function and message
func Wrap(err error, layer string, function string, message string) error {
	if err == nil {
		return nil
	}

	return IAMError{
		message: makeMessage(err, layer, function, message),
		err:     err,
	}
}

// Wrapf will wrap the error with layer, function, and format the message with args
func Wrapf(err error, layer string, function string, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)

	return IAMError{
		message: makeMessage(err, layer, function, msg),
		err:     err,
	}
}

// WrapFuncWithLayerFunction is a type alias for Wrap func
type WrapFuncWithLayerFunction func(err error, message string) error

// WrapfFuncWithLayerFunction is a type alias for Wrapf func
type WrapfFuncWithLayerFunction func(err error, format string, args ...interface{}) error

// NewLayerFunctionErrorWrap will create a Wrap func with specific layer and function
func NewLayerFunctionErrorWrap(layer string, function string) WrapFuncWithLayerFunction {
	return func(err error, message string) error {
		return Wrap(err, layer, function, message)
	}
}

// NewLayerFunctionErrorWrapf will create a Wrapf func with specific layer and function
func NewLayerFunctionErrorWrapf(layer string, function string) WrapfFuncWithLayerFunction {
	return func(err error, format string, args ...interface{}) error {
		return Wrapf(err, layer, function, format, args...)
	}
}
