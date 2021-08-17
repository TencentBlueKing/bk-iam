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
	jsoniter "github.com/json-iterator/go"

	pdptypes "iam/pkg/abac/pdp/types"
	"iam/pkg/abac/types"
	"iam/pkg/errorx"
	"iam/pkg/util"
)

// Translate ...
const Translate = "Translate"

// PoliciesTranslate 策略列表转换为QL表达式
func PoliciesTranslate(
	policies []types.AuthPolicy,
	resourceTypes []types.ActionResourceType,
) (map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(Translate, "PoliciesTranslate")

	// make a resourceTypeSet
	resourceTypeSet := util.NewFixedLengthStringSet(len(resourceTypes))
	for _, rt := range resourceTypes {
		key := rt.System + ":" + rt.Type
		resourceTypeSet.Add(key)
	}

	// 对每一条policy转换成一个条件表达式, 再组合成一个 OR 关系表达式
	content := make([]ExprCell, 0, len(policies))
	for _, policy := range policies {
		// TODO: 可以优化的点, expression + resourceTypeSet => Condition的local cache
		condition, err := PolicyTranslate(policy.Expression, resourceTypeSet)
		if err != nil {
			err = errorWrapf(err, "PolicyTranslate policyID=`%d` expression=`%s` resourceType=`%+v`",
				policy.ID, policy.Expression, resourceTypeSet)
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

// PolicyTranslate ...
func PolicyTranslate(
	resourceExpression string,
	resourceTypeSet *util.StringSet,
) (ExprCell, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(Translate, "PolicyTranslate")

	// TODO: newExpression, do translate here
	expressions := []pdptypes.ResourceExpression{}

	// NOTE: if expression == "" or expression == "[]", all return any
	// if action without resource_types, the expression is ""
	if resourceExpression != "" {
		err := jsoniter.UnmarshalFromString(resourceExpression, &expressions)
		if err != nil {
			err = errorWrapf(err, "unmarshal resourceExpression=`%s` fail", resourceExpression)
			return nil, err
		}
	}

	// 注意, 如果resourceType不匹配, 那么最终会返回any => 这里有没有问题? 两阶段计算?
	content := make([]ExprCell, 0, len(expressions))
	for _, expression := range expressions {
		key := expression.System + ":" + expression.Type
		if resourceTypeSet.Has(key) {
			expr, err := singleTranslate(expression.Expression, expression.Type)
			if err != nil {
				err = errorWrapf(err, "pdp PolicyTranslate expression: %s", expression.Expression)
				return nil, err
			}
			content = append(content, expr)
		}
	}

	switch len(content) {
	// content为空, 说明policy的操作不关联资源, 返回any
	case 0:
		return ExprCell{
			"op":    "any",
			"field": "",
			"value": []string{},
		}, nil
	case 1:
		return content[0], nil
	default:
		// NOTE: 这里是满足 一个操作依赖两个资源的场景, 所以是 AND => 两阶段计算
		return ExprCell{
			"op":      "AND",
			"content": content,
		}, nil
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
