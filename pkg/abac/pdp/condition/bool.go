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

	"iam/pkg/abac/pdp/condition/operator"
	"iam/pkg/abac/pdp/types"
)

// BoolCondition bool计算
type BoolCondition struct {
	baseCondition
}

//nolint:unparam
func newBoolCondition(key string, values []interface{}) (Condition, error) {
	return &BoolCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

func NewBoolCondition(key string, value bool) Condition {
	return &BoolCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: []interface{}{value},
		},
	}
}

// GetName 名称
func (c *BoolCondition) GetName() string {
	return operator.Bool
}

// Eval 求值
func (c *BoolCondition) Eval(ctx types.EvalContextor) bool {
	attrValue, err := ctx.GetAttr(c.Key)
	if err != nil {
		return false
	}

	exprValues := c.GetValues()

	switch attrValue.(type) {
	case []interface{}:
		// bool 计算不支持多个值
		return false
	default:
		// bool 计算不支持多个值
		if len(exprValues) != 1 {
			return false
		}

		valueBool, ok := attrValue.(bool)
		if !ok {
			return false
		}

		exprBool, ok := exprValues[0].(bool)
		if !ok {
			return false
		}

		return valueBool == exprBool
	}
}

func (c *BoolCondition) Translate(withSystem bool) (map[string]interface{}, error) {
	key := c.Key
	if !withSystem {
		key = removeSystemFromKey(key)
	}

	if len(c.Value) != 1 {
		return nil, fmt.Errorf("bool not support multi value %+v", c.Value)
	}

	return map[string]interface{}{
		"op":    "eq",
		"field": key,
		"value": c.Value[0],
	}, nil
}
