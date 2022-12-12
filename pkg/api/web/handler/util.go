/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"github.com/TencentBlueKing/gopkg/collection/set"

	"iam/pkg/util/json"
)

func validateFields(expected, actual string) bool {
	if actual == "" {
		return true
	}

	// 求差集
	s := set.SplitStringToSet(actual, ",").Diff(set.SplitStringToSet(expected, ","))
	return s.Size() == 0
}

// 过滤预期字段
func filterFields(set *set.StringSet, obj interface{}) (data map[string]interface{}, err error) {
	// 1. json.Marshal to json
	var m []byte
	m, err = json.Marshal(obj)
	if err != nil {
		return
	}

	// 2. json.UnMarshal from json to map[string]interface{}
	err = json.Unmarshal(m, &data)
	if err != nil {
		return
	}

	// 3. delete the key not in the set
	for key := range data {
		if !set.Has(key) {
			delete(data, key)
		}
	}
	return
}
