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
	"strconv"
)

// ToInt64 casts a interface to an int64
func ToInt64(i interface{}) (int64, error) {
	switch s := i.(type) {
	case int:
		return int64(s), nil
	case int64:
		return s, nil
	case string:
		v, err := strconv.ParseInt(s, 0, 64)
		if err == nil {
			return v, nil
		}
		return 0, fmt.Errorf("unable to cast %#v to int64, %w", i, err)
	case float64:
		return int64(s), nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v to int64, unsupported type", i)
	}
}
