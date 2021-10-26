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
	"errors"
	"fmt"
	"strings"

	"iam/pkg/abac/pdp/condition/operator"
	"iam/pkg/abac/pdp/types"
)

/*
条件的反序列化

1. 条件之间没有隐含的关系 每个条件都是 {"operator": {"filed": values}}
*/

const iamPath = "_bk_iam_path_"

var errMustNotEmpty = errors.New("value must not be empty")

// conditionFunc define the func which keyword match to func be called
type conditionFunc func(key string, values []interface{}) (Condition, error)

var conditionFactories map[string]conditionFunc

func init() {
	conditionFactories = map[string]conditionFunc{
		operator.AND:           newAndCondition,
		operator.OR:            newOrCondition,
		operator.ANY:           newAnyCondition,
		operator.StringEquals:  newStringEqualsCondition,
		operator.StringPrefix:  newStringPrefixCondition,
		operator.Bool:          newBoolCondition,
		operator.NumericEquals: newNumericEqualsCondition,
		operator.NumericGt:     newNumericGreaterThanCondition,
		operator.NumericGte:    newNumericGreaterThanEqualsCondition,
		operator.NumericLt:     newNumericLessThanCondition,
		operator.NumericLte:    newNumericLessThanEqualsCondition,
	}
}

// Condition 条件接口
type Condition interface {
	GetName() string
	GetKeys() []string // 返回条件中包含的所有属性key

	Eval(ctx types.EvalContextor) bool
	Translate(withSystem bool) (map[string]interface{}, error)
}

type LogicalCondition interface {
	Condition
	PartialEval(ctx types.EvalContextor) (bool, Condition)
}

func newConditionFromInterface(value interface{}) (Condition, error) {
	var err error
	var pd types.PolicyCondition
	var ok bool

	pd, ok = value.(types.PolicyCondition)
	if !ok {
		pd, err = types.InterfaceToPolicyCondition(value)
		if err != nil {
			return nil, err
		}
	}

	return NewConditionFromPolicyCondition(pd)
}

// NewConditionFromPolicyCondition will create condition from types.PolicyCondition
func NewConditionFromPolicyCondition(data types.PolicyCondition) (Condition, error) {
	for operator, options := range data {
		newConditionFunc, ok := conditionFactories[operator]
		if !ok {
			return nil, fmt.Errorf("can not support operator %s", operator)
		}

		for k, v := range options {
			return newConditionFunc(k, v)
		}
	}
	return nil, fmt.Errorf("not support data %v", data)
}

func removeSystemFromKey(key string) string {
	idx := strings.IndexByte(key, '.')
	if idx == -1 {
		return key
	}

	lidx := strings.LastIndexByte(key, '.')
	if idx == lidx {
		return key
	}

	return key[idx+1:]
}
