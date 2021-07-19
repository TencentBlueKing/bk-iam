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
)

// MapValueInterfaceToString ...
func MapValueInterfaceToString(input map[string]interface{}) (map[string]string, error) {
	data := make(map[string]string, len(input))
	for key, value := range input {
		valueStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("parse interface to string fail, the value of key=%s is not string", key)
		}

		data[key] = valueStr
	}
	return data, nil
}
