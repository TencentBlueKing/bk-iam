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
	"reflect"
	"strings"
)

// containsElement try loop over the list check if the list includes the element.
// return (false, false) if impossible.
// return (true, false) if element was not found.
// return (true, true) if element was found.
func includeElement(list interface{}, element interface{}) (ok, found bool) {

	listValue := reflect.ValueOf(list)
	listKind := reflect.TypeOf(list).Kind()
	defer func() {
		if e := recover(); e != nil {
			ok = false
			found = false
		}
	}()

	if listKind == reflect.String {
		elementValue := reflect.ValueOf(element)
		return true, strings.Contains(listValue.String(), elementValue.String())
	}

	if listKind == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if ObjectsAreEqual(mapKeys[i].Interface(), element) {
				return true, true
			}
		}
		return true, false
	}

	for i := 0; i < listValue.Len(); i++ {
		if ObjectsAreEqual(listValue.Index(i).Interface(), element) {
			return true, true
		}
	}
	return true, false

}

// Contains asserts that the specified string, list(array, slice...) or map contains the
// specified substring or element.
//
//    assert.Contains(t, "Hello World", "World")
//    assert.Contains(t, ["Hello", "World"], "World")
//    assert.Contains(t, {"Hello": "World"}, "Hello")
func Contains(list, element interface{}) bool {

	ok, found := includeElement(list, element)
	if !ok {
		return false
		//return Fail(t, fmt.Sprintf("%#v could not be applied builtin len()", list), msgAndArgs...)
	}
	if !found {
		return false
		//return Fail(t, fmt.Sprintf("%#v does not contain %#v", list, element), msgAndArgs...)
	}
	return true
}

// NotContains asserts that the specified string, list(array, slice...) or map does NOT contain the
// specified substring or element.
//
//    assert.NotContains(t, "Hello World", "Earth")
//    assert.NotContains(t, ["Hello", "World"], "Earth")
//    assert.NotContains(t, {"Hello": "World"}, "Earth")
func NotContains(list, element interface{}, msgAndArgs ...interface{}) bool {
	ok, found := includeElement(list, element)
	if !ok {
		return false
		//return Fail(t, fmt.Sprintf("\"%list\" could not be applied builtin len()", list), msgAndArgs...)
	}
	if found {
		return false
		//return Fail(t, fmt.Sprintf("\"%list\" should not contain \"%list\"", list, element), msgAndArgs...)
	}
	return true
}

// In return true if element in list
func In(element, list interface{}) bool {
	listValue := reflect.ValueOf(list)
	listKind := reflect.TypeOf(list).Kind()
	if listKind != reflect.String && listKind != reflect.Map {

		// try to speed up the `string in []string`
		if reflect.ValueOf(element).Kind() == reflect.String &&
			listValue.Len() > 0 &&
			listValue.Index(0).Kind() == reflect.String {
			elementStrValue := reflect.ValueOf(element).String()

			for i := 0; i < listValue.Len(); i++ {
				if listValue.Index(i).String() == elementStrValue {
					return true
				}
			}
			return false
		}
	}

	return Contains(list, element)
}

// NotIn return true if element not in list
func NotIn(element, list interface{}) bool {
	listValue := reflect.ValueOf(list)
	listKind := reflect.TypeOf(list).Kind()
	if listKind != reflect.String && listKind != reflect.Map {

		// try to speed up the `string in []string`
		if reflect.ValueOf(element).Kind() == reflect.String &&
			listValue.Len() > 0 &&
			listValue.Index(0).Kind() == reflect.String {
			elementStrValue := reflect.ValueOf(element).String()

			for i := 0; i < listValue.Len(); i++ {
				if listValue.Index(i).String() == elementStrValue {
					return false
				}
			}
			return true
		}
	}

	return NotContains(list, element)
}
