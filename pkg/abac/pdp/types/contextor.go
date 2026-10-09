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

// EvalContextor 属性获取接口
type EvalContextor interface {
	// GetAttr get the attr like id / type / name, currently only support resource
	GetAttr(name string) (interface{}, error)

	HasResource(_type string) bool
}

// AttributeExistenceContextor is implemented by contexts that can distinguish
// an absent attribute from an attribute with a nil value.
type AttributeExistenceContextor interface {
	HasAttr(name string) bool
}

// CanPartialEvalAttribute reports whether a condition can be evaluated with
// the current context.
//
// The legacy last-dot resource lookup is attempted first and keeps its old
// behavior. For dotted attribute IDs, the canonical resource type is used as a
// fallback. A canonical condition is evaluated only when the attribute itself
// is present; otherwise it must remain in the partial-evaluation result.
func CanPartialEvalAttribute(ctx EvalContextor, key string) bool {
	parts := splitAttributeKey(key)
	for index, part := range parts {
		if !ctx.HasResource(part.objectType) {
			continue
		}

		// Preserve legacy behavior for the original split result.
		if index == 0 {
			return true
		}

		if attrCtx, ok := ctx.(AttributeExistenceContextor); ok {
			return attrCtx.HasAttr(key)
		}
		// Custom/test contexts predating HasAttr retain their old behavior.
		return true
	}
	return false
}
