/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-权限中心Go SDK(iam-go-sdk) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

/* MIT License
 * Copyright (c) 2012-2020 Mat Ryer, Tyler Bunnell and contributors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/* NOTE: copied from https://github.com/stretchr/testify/assert/assertions.go and modified
 *  The original versions of the files are MIT licensed
 */

package eval

import (
	"bytes"
	"reflect"
)

// Equal return true if expected equals to actual
func Equal(expected, actual interface{}) bool {
	if err := validateEqualArgs(expected, actual); err != nil {
		return false
	}

	if !ObjectsAreEqual(expected, actual) {
		return false
	}

	return true
}

// NotEqual return true if expected not equals to actual
func NotEqual(expected, actual interface{}) bool {
	if err := validateEqualArgs(expected, actual); err != nil {
		return false
	}

	if ObjectsAreEqual(expected, actual) {
		return false
	}

	return true
}

// validateEqualArgs checks whether provided arguments can be safely used in the
// Equal/NotEqual functions.
func validateEqualArgs(expected, actual interface{}) error {
	if expected == nil && actual == nil {
		return nil
	}

	// NOTE: in iam policy expression, we guarantee the value will not be Function!
	// if isFunction(expected) || isFunction(actual) {
	// 	return errors.New("cannot take func type as argument")
	// }
	return nil
}

func isFunction(arg interface{}) bool {
	if arg == nil {
		return false
	}

	return reflect.TypeOf(arg).Kind() == reflect.Func
}

// ObjectsAreEqual determines if two objects are considered equal.
// This function does no assertion of any kind.
func ObjectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}

	if exp == nil || act == nil {
		return exp == nil && act == nil
	}

	return bytes.Equal(exp, act)
}
