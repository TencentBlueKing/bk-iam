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
	"github.com/TencentBlueKing/iam-go-sdk/expression/eval"

	"iam/pkg/abac/pdp/condition/operator"
	"iam/pkg/abac/pdp/types"
)

type numericCompareFunc func(interface{}, interface{}) bool

// NumericCompareCondition Number相等
type NumericCompareCondition struct {
	baseCondition
	name              string
	translateOperator string
	compareFunc       numericCompareFunc
}

//nolint:unparam
func newNumericCompareCondition(key string, values []interface{},
	name string, translateOperator string, compareFunc numericCompareFunc) (Condition, error) {
	return &NumericCompareCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},

		name:              name,
		translateOperator: translateOperator,
		compareFunc:       compareFunc,
	}, nil
}

func newNumericGreaterThanCondition(key string, values []interface{}) (Condition, error) {
	return newNumericCompareCondition(key, values, operator.NumericGreaterThan, "gt", eval.Greater)
}

func newNumericGreaterThanEqualsCondition(key string, values []interface{}) (Condition, error) {
	return newNumericCompareCondition(key, values, operator.NumericGreaterThanEquals, "gte", eval.GreaterOrEqual)
}

func newNumericLessThanCondition(key string, values []interface{}) (Condition, error) {
	return newNumericCompareCondition(key, values, operator.NumericLessThan, "lt", eval.Less)
}

func newNumericLessThanEqualsCondition(key string, values []interface{}) (Condition, error) {
	return newNumericCompareCondition(key, values, operator.NumericLessThanEquals, "lte", eval.LessOrEqual)
}

// GetName 名称
func (c *NumericCompareCondition) GetName() string {
	return c.name
}

// Eval 求值
func (c *NumericCompareCondition) Eval(ctx types.EvalContextor) bool {
	return c.forOr(ctx, func(a, b interface{}) bool {
		return c.compareFunc(a, b)
	})
}

func (c *NumericCompareCondition) Translate(withSystem bool) (map[string]interface{}, error) {
	key := c.Key
	if !withSystem {
		key = removeSystemFromKey(key)
	}

	exprCell := map[string]interface{}{
		"field": key,
	}

	switch len(c.Value) {
	case 0:
		return nil, errMustNotEmpty
	case 1:
		exprCell["op"] = c.translateOperator
		exprCell["value"] = c.Value[0]
		return exprCell, nil
	default:
		exprCell["op"] = c.translateOperator
		exprCell["value"] = c.Value
		return exprCell, nil
	}
}
