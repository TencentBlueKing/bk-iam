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

	jsoniter "github.com/json-iterator/go"

	"iam/pkg/abac/pdp/condition"
	pdptypes "iam/pkg/abac/pdp/types"
	"iam/pkg/errorx"
)

// Translate ...
const Translate = "Translate"

var errMustNotEmpty = errors.New("value must not be empty")

// ExprCell 表达式基本单元
type ExprCell map[string]interface{}

// Op return the operator of ExprCell
func (c ExprCell) Op() string {
	return c["op"].(string)
}

// PoliciesTranslate 策略列表转换为QL表达式
func PoliciesTranslate(
	policies []condition.Condition,
) (map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(Translate, "PoliciesTranslate")

	content := make([]ExprCell, 0, len(policies))
	for _, policy := range policies {
		// TODO: 可以优化的点, expression + resourceTypeSet => Condition的local cache
		condition, err := policyTranslate(policy)
		if err != nil {
			err = errorWrapf(err, "policyTranslate condition=`%+v` fail", policy)
			return nil, err
		}

		// NOTE: if got an `any`, return `any`!
		if condition.Op() == "any" {
			return ExprCell{
				"op":    "any",
				"field": "",
				"value": []string{},
			}, nil
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

func policyTranslate(
	cond condition.Condition,
) (ExprCell, error) {
	return cond.Translate()
}

func PolicyStringTranslate(resourceExpression string) (ExprCell, error) {

	// TODO: 代码重复
	//pkg/cache/impls/local_unmarshaled_expression.go

	expressions := []pdptypes.ResourceExpression{}
	err := jsoniter.UnmarshalFromString(resourceExpression, &expressions)
	// 无效的policy条件表达式, 容错
	if err != nil {
		err = fmt.Errorf("cache UnmarshalExpression unmarshal %s error: %w",
			resourceExpression, err)
		return nil, err
	}

	content := make([]condition.Condition, 0, len(expressions))
	for _, expression := range expressions {
		// NOTE: change the expression
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

	if len(content) == 1 {
		return content[0].Translate()
	} else {
		return condition.NewAndCondition(content).Translate()
	}
}

// PolicyTranslate ...
//func PolicyTranslate(
//	cond condition.Condition,
//) (ExprCell, error) {
//	return cond.Translate()
//errorWrapf := errorx.NewLayerFunctionErrorWrapf(Translate, "PolicyTranslate")

//expressions := []pdptypes.ResourceExpression{}

// TODO: newExpression, do translate here
//       需要支持, 将新版本表达式搞过来, 支持translate
// NOTE: if expression == "" or expression == "[]", all return any
// if action without resource_types, the expression is ""
//if resourceExpression != "" {
//	err := jsoniter.UnmarshalFromString(resourceExpression, &expressions)
//	if err != nil {
//		err = errorWrapf(err, "unmarshal resourceExpression=`%s` fail", resourceExpression)
//		return nil, err
//	}
//}

// 注意, 如果resourceType不匹配, 那么最终会返回any => 这里有没有问题? 两阶段计算?
//content := make([]ExprCell, 0, len(expressions))
//for _, expression := range expressions {
//	key := expression.System + ":" + expression.Type
//	if resourceTypeSet.Has(key) {
//		expr, err := singleTranslate(expression.Expression, expression.Type)
//		if err != nil {
//			err = errorWrapf(err, "pdp PolicyTranslate expression: %s", expression.Expression)
//			return nil, err
//		}
//		content = append(content, expr)
//	}
//}

//switch len(content) {
//// content为空, 说明policy的操作不关联资源, 返回any
//case 0:
//	return ExprCell{
//		"op":    "any",
//		"field": "",
//		"value": []string{},
//	}, nil
//case 1:
//	return content[0], nil
//default:
//	// NOTE: 这里是满足 一个操作依赖两个资源的场景, 所以是 AND => 两阶段计算
//	return ExprCell{
//		"op":      "AND",
//		"content": content,
//	}, nil
//}
//}

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
