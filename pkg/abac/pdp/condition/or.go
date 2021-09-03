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
	"fmt"
	"strings"

	"iam/pkg/abac/pdp/types"
)

// OrCondition 逻辑OR
type OrCondition struct {
	content []Condition
}

func NewOrCondition(content []Condition) Condition {
	return &OrCondition{content: content}
}

func newOrCondition(key string, values []interface{}) (Condition, error) {
	if key != "content" {
		return nil, fmt.Errorf("or condition not support key %s", key)
	}

	conditions := make([]Condition, 0, len(values))
	var (
		condition Condition
		err       error
	)

	for _, v := range values {
		condition, err = newConditionFromInterface(v)
		if err != nil {
			return nil, fmt.Errorf("or condition parser error: %w", err)
		}

		conditions = append(conditions, condition)
	}

	return &OrCondition{content: conditions}, nil
}

// GetName 名称
func (c *OrCondition) GetName() string {
	return "OR"
}

// GetKeys 返回嵌套条件中所有包含的属性key
func (c *OrCondition) GetKeys() []string {
	keys := make([]string, 0, len(c.content))
	for _, condition := range c.content {
		keys = append(keys, condition.GetKeys()...)
	}
	return keys
}

// Eval 求值
func (c *OrCondition) Eval(ctx types.AttributeGetter) bool {
	for _, condition := range c.content {
		if condition.Eval(ctx) {
			return true
		}
	}
	return false
}

func (c *OrCondition) Translate() (map[string]interface{}, error) {
	content := make([]interface{}, 0, len(c.content))
	for _, c := range c.content {
		ct, err := c.Translate()
		if err != nil {
			return nil, err
		}
		content = append(content, ct)
	}

	return map[string]interface{}{
		"op":      "OR",
		"content": content,
	}, nil

}

func (c *OrCondition) PartialEval(ctx types.AttributeGetter) (bool, Condition) {
	remainContent := make([]Condition, 0, len(c.content))
	for _, condition := range c.content {
		if condition.GetName() == "AND" || condition.GetName() == "OR" {
			ok, ci := condition.(LogicalCondition).PartialEval(ctx)
			if ok {
				// NOTE: here, return any condition?
				return true, NewAnyCondition()
			} else {
				remainContent = append(remainContent, ci)
			}
		} else {
			key := condition.GetKeys()[0]
			dotIdx := strings.LastIndexByte(key, '.')
			if dotIdx == -1 {
				panic("should contains dot in key")
			}
			_type := key[:dotIdx]

			if !ctx.HasKey(_type) {
				remainContent = append(remainContent, condition)
			}
		}
	}

	return true, NewOrCondition(remainContent)
}
