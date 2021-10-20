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
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"iam/pkg/abac/pdp/condition"
	pdptypes "iam/pkg/abac/pdp/types"
	"iam/pkg/errorx"
)

/*
request 中:
```
ObjectSet:
    {system}.{type}   => {attr: abc}

```

表达式中:

```
{system}.{type}.{attr} eq abc
```

转换给到接入系统的表达式(translate)

1. 当前policy版本v1:  `{type}.{attr} eq abc`
2. 未来policy版本v2:  `{system}.{type}.{attr} eq abc`
*/

// NOTE: currently, policy v1 should remove system from field, withSystem=False
// TODO: when we upgrade policy to v2, we should support withSystem=True

// Translate ...
const Translate = "Translate"

const defaultWithSystem = false

// ExprCell 表达式基本单元
type ExprCell map[string]interface{}

// Op return the operator of ExprCell
func (c ExprCell) Op() string {
	return c["op"].(string)
}

// ConditionsTranslate 策略列表转换为QL表达式
func ConditionsTranslate(
	conditions []condition.Condition,
) (map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(Translate, "ConditionsTranslate")

	content := make([]ExprCell, 0, len(conditions))
	for _, c := range conditions {
		condition, err := c.Translate(defaultWithSystem)
		if err != nil {
			err = errorWrapf(err, "conditionTranslate condition=`%+v` fail", c)
			return nil, err
		}

		// NOTE: if got an `any`, return `any`!
		if condition["op"].(string) == "any" {
			return condition, nil
			//return ExprCell{
			//	"op":    "any",
			//	"field": "",
			//	"value": []interface{}{},
			//}, nil
		}

		content = append(content, condition)
	}

	// merge same field `eq` and `in`; to `in`
	if len(content) > 1 {
		// 合并条件中field相同, op为eq, in的条件
		content = mergeContentField(content)
	}

	switch len(content) {
	case 1:
		return content[0], nil
	default:
		return ExprCell{
			"op":      "OR",
			"content": content,
		}, nil
	}
}

func expressionToConditions(expr string) ([]condition.Condition, error) {
	expressions := []pdptypes.ResourceExpression{}

	// NOTE: if expression == "" or expression == "[]", all return any
	// if action without resource_types, the expression is ""
	if expr != "" {
		err := jsoniter.UnmarshalFromString(expr, &expressions)
		// 无效的policy条件表达式, 容错
		if err != nil {
			err = fmt.Errorf("unmarshalFromString fail expr: %s error: %w",
				expr, err)
			return nil, err
		}
	}

	content := make([]condition.Condition, 0, len(expressions))
	for _, expression := range expressions {
		// NOTE: change the expr
		pc, err1 := expression.ToNewPolicyCondition()
		if err1 != nil {
			return nil, fmt.Errorf("toNewPolicyCondition error: %w", err1)
		}

		c, err2 := condition.NewConditionFromPolicyCondition(pc)
		// 表达式解析出错, 容错
		if err2 != nil {
			return nil, fmt.Errorf("newConditionFromPolicyCondition error: %w", err2)
		}
		content = append(content, c)
	}

	if len(content) == 0 {
		content = append(content, condition.NewAnyCondition())
	}

	return content, nil

}

func PolicyExpressionTranslate(expr string) (ExprCell, error) {
	content, err := expressionToConditions(expr)
	if err != nil {
		return nil, err
	}

	if len(content) == 1 {
		return content[0].Translate(defaultWithSystem)
	} else {
		return condition.NewAndCondition(content).Translate(defaultWithSystem)
	}
}

func PolicyExpressionToCondition(expr string) (condition.Condition, error) {
	content, err := expressionToConditions(expr)
	if err != nil {
		return nil, err
	}

	if len(content) == 1 {
		return content[0], nil
	} else {
		return condition.NewAndCondition(content), nil
	}
}

func mergeContentField(content []ExprCell) []ExprCell {
	mergeableExprs := map[string][]ExprCell{}
	newContent := make([]ExprCell, 0, len(content))

	for _, expr := range content {
		switch expr.Op() {
		case "eq", "in":
			field := expr["field"].(string)
			exprs, ok := mergeableExprs[field]
			if ok {
				exprs = append(exprs, expr)
				mergeableExprs[field] = exprs
			} else {
				mergeableExprs[field] = []ExprCell{expr}
			}
		default:
			newContent = append(newContent, expr)
		}
	}

	for field, exprs := range mergeableExprs {
		if len(exprs) == 1 {
			newContent = append(newContent, exprs[0])
		} else {
			values := make([]interface{}, 0, len(exprs))

			// 合并
			for _, expr := range exprs {
				switch expr.Op() {
				case "eq":
					values = append(values, expr["value"])
				case "in":
					values = append(values, expr["value"].([]interface{})...)
				}
			}

			newContent = append(newContent, ExprCell{
				"op":    "in",
				"field": field,
				"value": values,
			})
		}
	}

	return newContent
}
