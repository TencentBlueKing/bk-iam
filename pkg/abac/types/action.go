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

// Action 操作
type Action struct {
	ID        string
	Attribute *ActionAttribute
}

// NewAction ...
func NewAction() Action {
	// 问题: 如果忘了 FillAttributes -> SetResourceTypes, 直接调用WithoutResourceType这个函数会True
	// 所以, 应该放个空的进去?
	// 核心问题: => Fill初始化必须强制!!!!!!!!!!!!
	return Action{
		Attribute: NewActionAttribute(),
	}
}

// FillAttributes 填充action属性
func (a *Action) FillAttributes(pk int64, actionResourceTypes []ActionResourceType) {
	a.Attribute.SetPK(pk)
	a.Attribute.SetResourceTypes(actionResourceTypes)
}

// WithoutResourceType 判断操作关联的Resource Type是否为空
func (a *Action) WithoutResourceType() bool {
	actionResourceTypes, err := a.Attribute.GetResourceTypes()

	if err != nil || len(actionResourceTypes) == 0 {
		return true
	}

	return false
}

// ActionResourceType 操作关联的资源类型
type ActionResourceType struct {
	System string
	Type   string
}
