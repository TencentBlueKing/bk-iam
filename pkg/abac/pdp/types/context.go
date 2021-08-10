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
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
)

/*
PDP模块表达式求值
*/

// think about the Context? it's request + index of the resource

// ExprContext 表达式求值上下文
// 只有一个Resource的信息
type ExprContext struct {
	*request.Request
	Resource *types.Resource
}

// NewExprContext new context
func NewExprContext(ctx *request.Request, resource *types.Resource) *ExprContext {
	return &ExprContext{
		Request:  ctx,
		Resource: resource,
	}
}

// GetAttr 获取资源的属性值
func (c *ExprContext) GetAttr(name string) (interface{}, error) {
	return c.getResourceAttr(name)
}

func (c *ExprContext) getResourceAttr(name string) (interface{}, error) {
	switch name {
	case "id":
		return c.Resource.ID, nil
	default:
		value, _ := c.Resource.Attribute.Get(name)
		return value, nil
	}
}

// GetFullNameAttr 获取带前缀的属性值
//func (c *ExprContext) GetFullNameAttr(name string) (interface{}, error) {
//	// 属性name格式 resource.id 使用 . 分割
//	parts := strings.Split(name, ".")
//	if len(parts) != 2 {
//		return nil, fmt.Errorf("name format error %s", name)
//	}
//
//	switch parts[0] {
//	case "resource":
//		return c.getResourceAttr(parts[1])
//	case "action":
//		return c.getActionAttr(parts[1])
//	case "subject":
//		return c.getSubjectAttr(parts[1])
//	default:
//		return nil, fmt.Errorf("name not support %s", name)
//	}
//}

//func (c *ExprContext) getActionAttr(name string) (interface{}, error) {
//	switch name {
//	case "id":
//		return c.Action.ID, nil
//	default: // action 的公用属性暂时只有group相关信息, 存储为interface{}, 暂不处理
//		return nil, nil
//	}
//}
//
//func (c *ExprContext) getSubjectAttr(name string) (interface{}, error) {
//	switch name {
//	case "type":
//		return c.Subject.Type, nil
//	case "id":
//		return c.Subject.ID, nil
//	default: // subject 的公用属性暂时只有group相关信息, 存储为interface{}, 暂不处理
//		return nil, nil
//	}
//}
