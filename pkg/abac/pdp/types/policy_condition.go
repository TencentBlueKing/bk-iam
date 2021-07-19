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

// ResourceExpression keep the expression with fields:system/type
type ResourceExpression struct {
	System     string          `json:"system"`
	Type       string          `json:"type"`
	Expression PolicyCondition `json:"expression"`
}
