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

// EvalContext 表达式求值上下文
// 只有一个Resource的信息
type EvalContext struct {
	*request.Request
	objSet ObjectSetInterface
}

// NewEvalContext new context
func NewEvalContext(req *request.Request) *EvalContext {
	objSet := NewObjectSet()

	for _, r := range req.Resources {
		// maybe nil here
		if r.Attribute == nil {
			r.Attribute = types.Attribute{}
		}
		// set id into attributes
		r.Attribute.Set("id", r.ID)

		// bk_job.script => attributes
		_type := r.System + "." + r.Type
		objSet.Set(_type, r.Attribute)

	}
	// TODO: 需要限制接入系统资源id字段不能配置为attribute; 因为会被覆盖
	return &EvalContext{
		Request: req,
		objSet:  objSet,
	}
}

// GetAttr 获取资源的属性值
func (c *EvalContext) GetAttr(name string) (interface{}, error) {
	// name should be {system}.{resource_type}.{attr_key}
	return c.objSet.GetAttribute(name), nil
}

func (c *EvalContext) HasResource(_type string) bool {
	// has {system}.{resource_type}
	return c.objSet.Has(_type)
}
