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

// AndCondition 逻辑AND
type AndCondition struct {
	content []Condition
}

func NewAndCondition(content []Condition) Condition {
	return &AndCondition{content: content}
}

func newAndCondition(key string, values []interface{}) (Condition, error) {
	if key != "content" {
		return nil, fmt.Errorf("and condition not support key %s", key)
	}

	conditions := make([]Condition, 0, len(values))

	for _, v := range values {
		condition, err := newConditionFromInterface(v)
		if err != nil {
			return nil, fmt.Errorf("and condition parser error: %w", err)
		}

		conditions = append(conditions, condition)
	}

	return &AndCondition{content: conditions}, nil
}

// GetName 名称
func (c *AndCondition) GetName() string {
	return "AND"
}

// GetKeys 返回嵌套条件中所有包含的属性key
func (c *AndCondition) GetKeys() []string {
	keys := make([]string, 0, len(c.content))
	for _, condition := range c.content {
		keys = append(keys, condition.GetKeys()...)
	}
	return keys
}

// Eval 求值
func (c *AndCondition) Eval(ctx types.AttributeGetter) bool {
	for _, condition := range c.content {
		if !condition.Eval(ctx) {
			return false
		}
	}

	return true
}

func (c *AndCondition) Translate() (map[string]interface{}, error) {
	content := make([]map[string]interface{}, 0, len(c.content))
	for _, ci := range c.content {
		ct, err := ci.Translate()
		if err != nil {
			return nil, err
		}
		content = append(content, ct)
	}

	return map[string]interface{}{
		"op":      "AND",
		"content": content,
	}, nil

}

func (c *AndCondition) PartialEval(ctx types.AttributeGetter) (bool, Condition) {
	// NOTE: If allowed=False, condition should be nil
	// once got False=> return

	// TODO: get from sync.Pool
	remainContent := make([]Condition, 0, len(c.content))
	for _, condition := range c.content {
		if condition.GetName() == "AND" || condition.GetName() == "OR" {
			ok, ci := condition.(LogicalCondition).PartialEval(ctx)
			if !ok {
				return false, nil
			}

			// 如果残留单独一个any, any always=True, 则没有必要append
			if ci.GetName() != "Any" {
				remainContent = append(remainContent, ci)
			}
		} else {
			key := condition.GetKeys()[0]
			dotIdx := strings.LastIndexByte(key, '.')
			if dotIdx == -1 {
				//panic("should contain dot in key")
				return false, nil
			}
			_type := key[:dotIdx]

			if ctx.HasResource(_type) {
				// resource exists and eval fail, no remain content
				if !condition.Eval(ctx) {
					return false, nil
				}
			} else {
				remainContent = append(remainContent, condition)
			}
		}
	}

	// all eval success, 全部执行成功, 理论上只剩any
	switch len(remainContent) {
	case 0:
		return true, NewAnyCondition()
	case 1:
		return true, remainContent[0]
	default:
		return true, NewAndCondition(remainContent)
	}
}
