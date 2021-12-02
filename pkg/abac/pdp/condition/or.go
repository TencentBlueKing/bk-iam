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

	"iam/pkg/abac/pdp/condition/operator"
	"iam/pkg/abac/pdp/types"
)

// OrCondition 逻辑OR
type OrCondition struct {
	content []Condition
}

func NewOrCondition(content []Condition) Condition {
	return &OrCondition{content: content}
}

func newOrCondition(field string, values []interface{}) (Condition, error) {
	if field != "content" {
		return nil, fmt.Errorf("or condition not support field %s", field)
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
	return operator.OR
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
func (c *OrCondition) Eval(ctx types.EvalContextor) bool {
	for _, condition := range c.content {
		if condition.Eval(ctx) {
			return true
		}
	}
	return false
}

func (c *OrCondition) Translate(withSystem bool) (map[string]interface{}, error) {
	content := make([]map[string]interface{}, 0, len(c.content))
	for _, ci := range c.content {
		ct, err := ci.Translate(withSystem)
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

// PartialEval 使用传递的部分资源执行表达式, 并返回剩余的部分
func (c *OrCondition) PartialEval(ctx types.EvalContextor) (bool, Condition) {
	// NOTE: If allowed=False, condition should be nil
	// once got True => return
	remainedContent := make([]Condition, 0, len(c.content))
	for _, condition := range c.content {
		switch condition.GetName() {
		case operator.AND, operator.OR:
			// NOTE: true的时候, 可能还有剩余的表达式
			ok, ci := condition.(LogicalCondition).PartialEval(ctx)
			if ok {
				// if true and remain condition is ANY, return True
				if ci.GetName() == operator.ANY {
					return true, NewAnyCondition()
				} else {
					remainedContent = append(remainedContent, ci)
				}
			}

			// a OR b, if a false, do nothing!
		case operator.ANY:
			// if any, it's always true, return true
			return true, NewAnyCondition()
		default:
			key := condition.GetKeys()[0]
			dotIdx := strings.LastIndexByte(key, '.')
			if dotIdx == -1 {
				//panic("should contain dot in key")
				return false, nil
			}
			_type := key[:dotIdx]

			if ctx.HasResource(_type) {
				// a OR b, if a true, return True
				if condition.Eval(ctx) {
					return true, NewAnyCondition()
				}
				// if hasKey = true and eval fail, do nothing!
			} else {
				// request has no resource, so append to remain
				remainedContent = append(remainedContent, condition)
			}
		}
	}

	switch len(remainedContent) {
	case 0:
		// all false, return False
		return false, nil
	case 1:
		return true, remainedContent[0]
	default:
		return true, NewOrCondition(remainedContent)
	}
}
