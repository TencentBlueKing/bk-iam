/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	"fmt"
)

/*
policy 条件举例

{
	"StringEqual": {
		"id": ["1", "2"]
	}
}
*/

// PolicyCondition condition struct of policy single resource
type PolicyCondition map[string]map[string][]interface{}

func (p PolicyCondition) ToNewPolicyCondition(system, _type string) (PolicyCondition, error) {
	// case 1, binary operator
	//	{
	//		"StringEquals": {
	//			"system": ["linux"]
	//		}
	//	},
	// case 2, any
	// case 3, AND, OR
	//	{
	//		"AND": {
	//			"content": [
	//				{
	//					"StringEquals": {
	//						"system": ["linux"]
	//					}
	//				},
	//				{
	//					"StringPrefix": {
	//						"path": ["/biz,1/"]
	//					}
	//				}
	//			]
	//		}
	//	}

	// TODO: unittest
	// TODO: performance
	// NOTE: 这个对象被cache起来了, 所以理论上, 处理过一次之后, 不需要反复处理

	keyPrefix := system + "." + _type + "."
	pc := make(PolicyCondition, len(p))
	for op, c := range p {
		switch op {
		case "AND", "OR":
			content := c["content"]
			newContent := make([]interface{}, 0, len(c["content"]))
			for _, i := range content {
				pp, err := InterfaceToPolicyCondition(i)
				if err != nil {
					return nil, fmt.Errorf("convert fail %w", err)
				}

				npc, err2 := pp.ToNewPolicyCondition(system, _type)
				if err2 != nil {
					return nil, fmt.Errorf("convert fail2 %w", err)
				}

				newContent = append(newContent, npc)
			}
			pc[op] = map[string][]interface{}{
				"content": newContent,
			}
		// NOTE: any is the same as other operatos
		//case "Any":
		//	pc[op] = c
		default:
			pc[op] = make(map[string][]interface{}, len(c))
			for k, v := range c {
				if k != "" {
					key := keyPrefix + k
					pc[op][key] = v
				} else {
					// Any, field is empty
					pc[op][k] = v
				}

			}
		}
	}
	return pc, nil
}

// ResourceExpression keep the expression with fields:system/type
// will be removed later, DO NOT USE IT IN ANY NEW CODES
type ResourceExpression struct {
	System     string          `json:"system"`
	Type       string          `json:"type"`
	Expression PolicyCondition `json:"expression"`
}

func (r ResourceExpression) ToNewPolicyCondition() (PolicyCondition, error) {
	return r.Expression.ToNewPolicyCondition(r.System, r.Type)
}
