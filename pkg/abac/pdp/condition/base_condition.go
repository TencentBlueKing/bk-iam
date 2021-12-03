/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package condition

import (
	"strings"

	"iam/pkg/abac/pdp/types"
	abacTypes "iam/pkg/abac/types"
)

type baseCondition struct {
	Key   string
	Value []interface{}
}

// GetKeys 返回条件中属性key值
func (c *baseCondition) GetKeys() []string {
	return []string{c.Key}
}

func (c *baseCondition) HasEnv() bool {
	return strings.Contains(c.Key, abacTypes.IamEnvSuffix)
}

func (c *baseCondition) GetEnvTz() (tz string, ok bool) {
	if strings.HasSuffix(c.Key, abacTypes.IamEnvTzSuffix) {
		if len(c.Value) != 1 {
			return
		}

		tz, ok = c.Value[0].(string)
		return
	}
	return
}

// GetValues 如果Value中有参数, 获取参数的值
func (c *baseCondition) GetValues() []interface{} {
	return c.Value
}

// forOr value之间or关系遍历
// ? 需要注意 对slice的操作都是OR的关系, 如果需要其它的关系, 使用forOR
func (c *baseCondition) forOr(ctx types.EvalContextor, fn func(interface{}, interface{}) bool) bool {
	attrValue, err := ctx.GetAttr(c.Key)
	if err != nil {
		return false
	}

	exprValues := c.GetValues()

	switch vs := attrValue.(type) {
	case []interface{}: // 处理属性为array的情况
		for _, av := range vs {
			for _, v := range exprValues {
				if fn(av, v) {
					return true
				}
			}
		}
	default:
		for _, v := range exprValues {
			if fn(attrValue, v) {
				return true
			}
		}
	}
	return false
}
