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
	"iam/pkg/abac/pdp/condition/operator"
	"iam/pkg/abac/pdp/types"
)

// AnyCondition 任意条件
type AnyCondition struct {
	baseCondition
}

//nolint:unparam
func newAnyCondition(key string, values []interface{}) (Condition, error) {
	return &AnyCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

func NewAnyCondition() Condition {
	return &AnyCondition{
		baseCondition: baseCondition{
			Key:   "",
			Value: []interface{}{},
		},
	}
}

// GetName 名称
func (c *AnyCondition) GetName() string {
	return operator.ANY
}

// GetKeys 属性key
func (c *AnyCondition) GetKeys() []string {
	return []string{}
}

// Eval 求值
func (c *AnyCondition) Eval(ctx types.EvalContextor) bool {
	return true
}

func (c *AnyCondition) Translate(withSystem bool) (map[string]interface{}, error) {
	key := c.Key
	if !withSystem {
		key = removeSystemFromKey(key)
	}

	return map[string]interface{}{
		"op":    "any",
		"field": key,
		"value": c.Value,
	}, nil
}
