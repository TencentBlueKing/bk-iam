/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-gopkg available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package conv

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) (b []byte) {
	// nolint:govet
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	// nolint:govet
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ToString casts a interface to a string.
func ToString(i interface{}) string {
	switch s := i.(type) {
	case string:
		return s
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(i.(float32)), 'f', -1, 64)
	case int:
		return strconv.FormatInt(int64(i.(int)), 10)
	case int8:
		return strconv.FormatInt(int64(i.(int8)), 10)
	case int16:
		return strconv.FormatInt(int64(i.(int16)), 10)
	case int32:
		return strconv.FormatInt(int64(i.(int32)), 10)
	case int64:
		return strconv.FormatInt(i.(int64), 10)
	case uint:
		return strconv.FormatUint(uint64(i.(uint)), 10)
	case uint8:
		return strconv.FormatUint(uint64(i.(uint8)), 10)
	case uint16:
		return strconv.FormatUint(uint64(i.(uint16)), 10)
	case uint32:
		return strconv.FormatUint(uint64(i.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(i.(uint64), 10)
	case []byte:
		return string(s)
	case nil:
		return ""
	case error:
		return s.Error()
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprint(i)
	}
}
