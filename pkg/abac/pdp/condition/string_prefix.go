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

	"iam/pkg/abac/pdp/condition/operator"
	"iam/pkg/abac/pdp/types"
	abacTypes "iam/pkg/abac/types"
)

// StringPrefixCondition 字符串前缀匹配
type StringPrefixCondition struct {
	baseCondition
}

func newStringPrefixCondition(key string, values []interface{}) (Condition, error) {
	return &StringPrefixCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

// GetName 名称
func (c *StringPrefixCondition) GetName() string {
	return operator.StringPrefix
}

// Eval 求值
func (c *StringPrefixCondition) Eval(ctx types.EvalContextor) bool {
	return c.forOr(ctx, func(a, b interface{}) bool {
		aStr, ok := a.(string)
		if !ok {
			return false
		}

		bStr, ok := b.(string)
		if !ok {
			return false
		}

		// 支持表达式中最后一个节点为任意
		// /biz,1/set,*/ -> /biz,1/set,
		if strings.HasSuffix(c.Key, abacTypes.IamPathSuffix) && strings.HasSuffix(bStr, ",*/") {
			bStr = bStr[0 : len(bStr)-2]
		}

		return strings.HasPrefix(aStr, bStr)
	})
}

func (c *StringPrefixCondition) Translate(withSystem bool) (map[string]interface{}, error) {
	key := c.Key
	if !withSystem {
		key = removeSystemFromKey(key)
	}

	// NOTE: starts_with/ends_with/not_starts_with/not_ends_with should be
	// 1. single value like: a starts_with x
	// 2. multiple value like: a starts_with x OR a starts_with y
	// NEVER BE `a starts_with [x, y]`
	content := make([]map[string]interface{}, 0, len(c.Value))
	for _, v := range c.Value {
		content = append(content, map[string]interface{}{
			"op":    "starts_with",
			"field": key,
			"value": v,
		})
	}

	switch len(content) {
	case 0:
		return nil, errMustNotEmpty
	case 1:
		return content[0], nil
	default:
		return map[string]interface{}{
			"op":      "OR",
			"content": content,
		}, nil
	}
}
