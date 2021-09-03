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

import "iam/pkg/abac/pdp/types"

// StringEqualsCondition 字符串相等
type StringEqualsCondition struct {
	baseCondition
}

//nolint:unparam
func newStringEqualsCondition(key string, values []interface{}) (Condition, error) {
	return &StringEqualsCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

// GetName 名称
func (c *StringEqualsCondition) GetName() string {
	return "StringEquals"
}

// Eval 求值
func (c *StringEqualsCondition) Eval(ctx types.AttributeGetter) bool {
	return c.forOr(ctx, func(a, b interface{}) bool {
		return a == b
	})
}
func (c *StringEqualsCondition) Translate() (map[string]interface{}, error) {
	exprCell := map[string]interface{}{
		"field": c.Key,
	}

	switch len(c.Value) {
	case 0:
		return nil, errMustNotEmpty
	case 1:
		exprCell["op"] = "eq"
		exprCell["value"] = c.Value[0]
	default:
		exprCell["op"] = "in"
		exprCell["value"] = c.Value
	}
	return exprCell, nil

}
