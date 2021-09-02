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

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/types"
	"iam/pkg/abac/pdp/util"
)

/*
条件的反序列化

1. 条件之间没有隐含的关系 每个条件都是 {"operator": {"filed": values}}
*/

const iamPath = "_bk_iam_path_"

// conditionFunc define the func which keyword match to func be called
type conditionFunc func(key string, values []interface{}) (Condition, error)

var conditionFactories map[string]conditionFunc

func init() {
	conditionFactories = map[string]conditionFunc{
		new(AndCondition).GetName():           newAndCondition,
		new(OrCondition).GetName():            newOrCondition,
		new(AnyCondition).GetName():           newAnyCondition,
		new(StringEqualsCondition).GetName():  newStringEqualsCondition,
		new(StringPrefixCondition).GetName():  newStringPrefixCondition,
		new(NumericEqualsCondition).GetName(): newNumericEqualsCondition,
		new(BoolCondition).GetName():          newBoolCondition,
	}
}

// Condition 条件接口
type Condition interface {
	GetName() string

	Eval(ctx types.AttributeGetter) bool

	GetKeys() []string // 返回条件中包含的所有属性key
}

// NewConditionByJSON 解析条件
func NewConditionByJSON(data []byte) (Condition, error) {
	condition := types.PolicyCondition{}
	err := jsoniter.Unmarshal(data, &condition)
	if err != nil {
		return nil, fmt.Errorf("json data %s unmarshal error: %w", data, err)
	}

	return NewConditionFromPolicyCondition(condition)
}

func newConditionFromInterface(value interface{}) (Condition, error) {
	var err error
	var pd types.PolicyCondition
	var ok bool

	pd, ok = value.(types.PolicyCondition)
	if !ok {
		pd, err = util.InterfaceToPolicyCondition(value)
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
	return nil, fmt.Errorf("can not support data %v", data)
}

// ================== conditions ==================

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
	var (
		condition Condition
		err       error
	)
	for _, v := range values {
		condition, err = newConditionFromInterface(v)
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

// Eval 求值
func (c *AndCondition) Eval(ctx types.AttributeGetter) bool {
	for _, condition := range c.content {
		if !condition.Eval(ctx) {
			return false
		}
	}

	return true
}

// GetKeys 返回嵌套条件中所有包含的属性key
func (c *AndCondition) GetKeys() []string {
	keys := make([]string, 0, len(c.content))
	for _, condition := range c.content {
		keys = append(keys, condition.GetKeys()...)
	}
	return keys
}

// OrCondition 逻辑OR
type OrCondition struct {
	content []Condition
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

// Eval 求值
func (c *OrCondition) Eval(ctx types.AttributeGetter) bool {
	for _, condition := range c.content {
		if condition.Eval(ctx) {
			return true
		}
	}
	return false
}

// GetKeys 返回嵌套条件中所有包含的属性key
func (c *OrCondition) GetKeys() []string {
	keys := make([]string, 0, len(c.content))
	for _, condition := range c.content {
		keys = append(keys, condition.GetKeys()...)
	}
	return keys
}

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

// GetName 名称
func (c *AnyCondition) GetName() string {
	return "Any"
}

// Eval 求值
func (c *AnyCondition) Eval(ctx types.AttributeGetter) bool {
	return true
}

// GetKeys 属性key
func (c *AnyCondition) GetKeys() []string {
	return []string{}
}

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
	return "StringPrefix"
}

// Eval 求值
func (c *StringPrefixCondition) Eval(ctx types.AttributeGetter) bool {
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
		if c.Key == iamPath && strings.HasSuffix(bStr, ",*/") {
			bStr = bStr[0 : len(bStr)-2]
		}

		return strings.HasPrefix(aStr, bStr)
	})
}

// NumericEqualsCondition Number相等
type NumericEqualsCondition struct {
	baseCondition
}

//nolint:unparam
func newNumericEqualsCondition(key string, values []interface{}) (Condition, error) {
	return &NumericEqualsCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

// GetName 名称
func (c *NumericEqualsCondition) GetName() string {
	return "NumericEquals"
}

// Eval 求值
func (c *NumericEqualsCondition) Eval(ctx types.AttributeGetter) bool {
	return c.forOr(ctx, func(a, b interface{}) bool {
		return a == b
	})
}

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

// GetName 名称
func (c *BoolCondition) GetName() string {
	return "Bool"
}

// Eval 求值
func (c *BoolCondition) Eval(ctx types.AttributeGetter) bool {
	attrValue, err := ctx.GetAttr(c.Key)
	if err != nil {
		log.Debugf("get attr %s from ctx %v error %v", c.Key, ctx, err)
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
