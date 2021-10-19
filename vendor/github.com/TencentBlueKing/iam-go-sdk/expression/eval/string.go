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

package eval

import "strings"

func convertArgsToString(v1, v2 interface{}) (s1, s2 string, ok bool) {
	s1, ok = v1.(string)
	if !ok {
		return "", "", false
	}

	s2, ok = v2.(string)
	if !ok {
		return "", "", false
	}

	return s1, s2, true
}

// StartsWith return true if v1 startswith v2
func StartsWith(v1 interface{}, v2 interface{}) bool {
	s1, s2, ok := convertArgsToString(v1, v2)
	if !ok {
		return false
	}

	return strings.HasPrefix(s1, s2)
}

// NotStartsWith return true if v1 not startswith v2
func NotStartsWith(v1 interface{}, v2 interface{}) bool {
	s1, s2, ok := convertArgsToString(v1, v2)
	if !ok {
		return false
	}

	return !strings.HasPrefix(s1, s2)
}

// EndsWith return true if v1 endswith v2
func EndsWith(v1 interface{}, v2 interface{}) bool {
	s1, s2, ok := convertArgsToString(v1, v2)
	if !ok {
		return false
	}

	return strings.HasSuffix(s1, s2)
}

// NotEndsWith return true if v1 not endswith v2
func NotEndsWith(v1 interface{}, v2 interface{}) bool {
	s1, s2, ok := convertArgsToString(v1, v2)
	if !ok {
		return false
	}

	return !strings.HasSuffix(s1, s2)
}
