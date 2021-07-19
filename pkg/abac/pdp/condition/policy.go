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

	"iam/pkg/cache/impls"

	"iam/pkg/abac/types"
	"iam/pkg/util"
)

// GetPoliciesAttrKeys 条件中的属性key
func GetPoliciesAttrKeys(
	resource *types.Resource,
	policies []types.AuthPolicy,
) ([]string, error) {
	// 查询policies相关的属性key
	conditions, err := parseResourceConditionFromPolicies(resource, policies)
	if err != nil {
		return nil, fmt.Errorf("parseResourceConditionFromPolicies error: %w", err)
	}

	keySet := util.NewFixedLengthStringSet(len(conditions))
	for _, condition := range conditions {
		for _, key := range condition.GetKeys() {
			keySet.Add(key)
		}
	}

	return keySet.ToSlice(), nil
}

// parseResourceConditionFromPolicies 从policies中解析出resource相关的conditions数组
func parseResourceConditionFromPolicies(
	resource *types.Resource,
	policies []types.AuthPolicy,
) ([]Condition, error) {
	conditions := make([]Condition, 0, len(policies))

	// 查询policies的key
	for _, policy := range policies {
		condition, err := ParseResourceConditionFromExpression(resource, policy.Expression, policy.ExpressionSignature)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, condition)
	}

	return conditions, nil
}

// ParseResourceConditionFromExpression ...
func ParseResourceConditionFromExpression(
	resource *types.Resource,
	policyExpression string,
	policyExpressionSignature string,
) (Condition, error) {
	expressions, err := impls.GetUnmarshalledResourceExpression(policyExpression, policyExpressionSignature)
	if err != nil {
		err = fmt.Errorf("pdp impls.GetUnmarshalledResourceExpression expression=`%s`,signature=`%s` fail %w",
			policyExpression, policyExpressionSignature, err)
		return nil, err
	}

	// NOTE: 这里只会返回第一个condition
	for _, expression := range expressions {
		if resource.System == expression.System && resource.Type == expression.Type {
			condition, err := NewConditionFromPolicyCondition(expression.Expression)
			// 表达式解析出错, 容错
			if err != nil {
				return nil, fmt.Errorf("expression parser error: %w", err)
			}
			return condition, err
		}
	}
	return nil, fmt.Errorf("resource not match expression")
}
