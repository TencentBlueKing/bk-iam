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

import "errors"

/*
带有逻辑条件的操作符会有嵌套的内容, 比如:
{
	"AND": {
		"content": [
			{
				"StringEqual": {
					"id": ["1", "2"]
				}
			},
			{
				"Bool": {
					"online": [true]
				}
			}
		]
	}
}

目标:

把逻辑表达式中content嵌套的单个表达式从interface{}类型转换成实际的map[string]map[string][]interface{}

输入:

类型 interface{}

内容 逻辑条件中嵌套的单个条件
{
	"StringEqual": {
		"id": ["1", "2"]
	}
}

输出:

map[string]map[string][]interface{}{
	"StringEqual": map[string][]interface{}{
		"id": []interface{}{"1", "2"}
	}
}

*/

var ErrTypeAssertFail = errors.New("type assert fail")

// InterfaceToPolicyCondition 嵌套的条件interface换行为可解析的类型
func InterfaceToPolicyCondition(value interface{}) (PolicyCondition, error) {
	// 从interface{}转换为操作符key的map
	operatorMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, ErrTypeAssertFail
	}

	// 函数返回的解析好的条件map
	// map[string]map[string][]interface{}
	conditionMap := make(PolicyCondition, len(operatorMap))

	// 解析第一层map, key为操作符
	for operator, options := range operatorMap {
		// 操作附加的属性选项
		options, ok := options.(map[string]interface{})
		if !ok {
			return nil, ErrTypeAssertFail
		}

		// condition中的属性map
		attributeMap := make(map[string][]interface{}, len(options))
		// 解析第二层map, key为属性名称
		for k, v := range options {
			attributeMap[k], ok = v.([]interface{}) // 属性的值转换为数组
			if !ok {
				return nil, ErrTypeAssertFail
			}
		}

		conditionMap[operator] = attributeMap
	}
	return conditionMap, nil
}
