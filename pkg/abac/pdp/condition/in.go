/*
 * TencentBlueKing is pleased to support the open source community by making BlueKing-IAM available.
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

// InCondition checks whether a scalar or any element of an array is in the policy values.
type InCondition struct {
	baseCondition
}

func newInCondition(key string, values []interface{}) (Condition, error) {
	return &InCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

func (c *InCondition) GetName() string {
	return operator.IN
}

func (c *InCondition) Eval(ctx types.EvalContextor) bool {
	return c.forOr(ctx, eval.ValueEqual)
}

func (c *InCondition) Translate(withSystem bool) (map[string]interface{}, error) {
	if len(c.Value) == 0 {
		return nil, errMustNotEmpty
	}

	key := c.Key
	if !withSystem {
		key = removeSystemFromKey(key)
	}

	return map[string]interface{}{
		"op":    "in",
		"field": key,
		"value": c.Value,
	}, nil
}
