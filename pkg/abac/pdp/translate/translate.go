/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package translate

import (
	"errors"
	"fmt"

	"iam/pkg/abac/pdp/types"
	"iam/pkg/abac/pdp/util"
)

/*
表达式转换
*/

// 1. 解析policy表达式到对外的条件表达式
// 2. 对操作符需要有一个映射关系
// 3. 遍历实现

var errMustNotEmpty = errors.New("value must not be empty")

// ExprCell 表达式基本单元
type ExprCell map[string]interface{}

// Op return the operator of ExprCell
func (c ExprCell) Op() string {
	return c["op"].(string)
}

type translateFunc func(field string, value []interface{}) (ExprCell, error)

var translateFactories map[string]translateFunc

func init() {
	translateFactories = map[string]translateFunc{
		"AND":           andTranslate,
		"OR":            orTranslate,
		"Any":           anyTranslate,
		"StringEquals":  stringEqualsTranslate,
		"StringPrefix":  stringPrefixTranslate,
		"NumericEquals": numericEqualsTranslate,
		"Bool":          boolTranslate,
	}
}

func singleTranslate(expression types.PolicyCondition, _type string) (ExprCell, error) {
	for operator, option := range expression {
		tf, ok := translateFactories[operator]
		if !ok {
			return nil, fmt.Errorf("can not support operator %s", operator)
		}

		for field, value := range option {
			switch operator {
			case "OR", "AND":
				return tf(_type, value)
			default:
				//typeField := fmt.Sprintf("%s.%s", _type, field)
				typeField := _type + "." + field
				return tf(typeField, value)
			}
		}
	}
	return nil, errMustNotEmpty
}

func andTranslate(_type string, value []interface{}) (ExprCell, error) {
	content := make([]interface{}, 0, len(value))

	for _, v := range value {
		m, err := util.InterfaceToPolicyCondition(v)
		if err != nil {
			return nil, err
		}
		condition, err2 := singleTranslate(m, _type)
		if err2 != nil {
			return nil, err2
		}

		content = append(content, condition)
	}

	return map[string]interface{}{
		"op":      "AND",
		"content": content,
	}, nil
}

func orTranslate(_type string, value []interface{}) (ExprCell, error) {
	content := make([]interface{}, 0, len(value))

	for _, v := range value {
		m, err := util.InterfaceToPolicyCondition(v)
		if err != nil {
			return nil, err
		}
		condition, err2 := singleTranslate(m, _type)
		if err2 != nil {
			return nil, err2
		}

		content = append(content, condition)
	}

	return map[string]interface{}{
		"op":      "OR",
		"content": content,
	}, nil
}

//nolint:unparam
func anyTranslate(field string, value []interface{}) (ExprCell, error) {
	return map[string]interface{}{
		"op":    "any",
		"field": field,
		"value": value,
	}, nil
}

func stringEqualsTranslate(field string, value []interface{}) (ExprCell, error) {
	exprCell := map[string]interface{}{
		"field": field,
	}

	switch len(value) {
	case 0:
		return nil, errMustNotEmpty
	case 1:
		exprCell["op"] = "eq"
		exprCell["value"] = value[0]
	default:
		exprCell["op"] = "in"
		exprCell["value"] = value
	}
	return exprCell, nil
}

func stringPrefixTranslate(field string, value []interface{}) (ExprCell, error) {
	content := make([]map[string]interface{}, 0, len(value))
	for _, v := range value {
		content = append(content, map[string]interface{}{
			"op":    "starts_with",
			"field": field,
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

func numericEqualsTranslate(field string, value []interface{}) (ExprCell, error) {
	exprCell := map[string]interface{}{
		"field": field,
	}

	switch len(value) {
	case 0:
		return nil, errMustNotEmpty
	case 1:
		exprCell["op"] = "eq"
		exprCell["value"] = value[0]
	default:
		exprCell["op"] = "in"
		exprCell["value"] = value
	}
	return exprCell, nil
}

func boolTranslate(field string, value []interface{}) (ExprCell, error) {
	if len(value) != 1 {
		return nil, fmt.Errorf("bool not support multi value %+v", value)
	}

	return map[string]interface{}{
		"op":    "eq",
		"field": field,
		"value": value[0],
	}, nil
}
