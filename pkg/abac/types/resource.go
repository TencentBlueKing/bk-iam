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

// Resource 内部模块的资源, 包括资源的各种属性
type Resource struct {
	System    string
	Type      string
	ID        string
	Attribute Attribute // 路径, 业务集, 业务都直接存在map中
}

// ExtResource 附加属性查询的资源
type ExtResource struct {
	System string
	Type   string
	IDs    []string
}

// Instance ...
type Instance struct {
	ID        string    `json:"id"`
	Attribute Attribute `json:"attribute"`
}

// ExtResourceWithAttribute 填充的属性附加查询资源
type ExtResourceWithAttribute struct {
	System    string     `json:"system"`
	Type      string     `json:"type"`
	Instances []Instance `json:"instances"`
}

// ResourceNode 用于RBAC鉴权
type ResourceNode struct {
	System string
	Type   string
	ID     string
	TypePK int64
}

func (r *ResourceNode) String() string {
	return r.System + ":" + r.Type + ":" + r.ID
}
