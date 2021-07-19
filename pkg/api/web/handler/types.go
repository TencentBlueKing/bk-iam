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

import "fmt"

type queryViaFields struct {
	Fields string `form:"fields" json:"fields" binding:"omitempty"`
}

func (q *queryViaFields) validateFields(supportFields, defaultFields string) (bool, string) {
	// 如果为空，则设置默认值
	if q.Fields == "" {
		q.Fields = defaultFields
	}
	// 校验是否存在不支持的查询字段
	if !validateFields(supportFields, q.Fields) {
		return false, fmt.Sprintf("fields only support `%s`", supportFields)
	}
	return true, "valid"
}
