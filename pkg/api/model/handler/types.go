/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

type referenceResourceType struct {
	SystemID string `json:"system_id" structs:"system_id" binding:"required" example:"bk_cmdb"`
	ID       string `json:"id"        structs:"id"        binding:"required" example:"host"`
}

type referenceInstanceSelection struct {
	SystemID      string `json:"system_id"       structs:"system_id"       binding:"required"  example:"bk_cmdb"`
	ID            string `json:"id"              structs:"id"              binding:"required"  example:"host_view"`
	IgnoreIAMPath bool   `json:"ignore_iam_path" structs:"ignore_iam_path" binding:"omitempty" example:"false"`
}

type deleteViaID struct {
	ID string `json:"id" binding:"required" example:"192.168.1.1"`
}

// AllBaseInfo ...
type AllBaseInfo struct {
	IDSet     map[string]string
	NameSet   map[string]string
	NameEnSet map[string]string
}

// ContainsID ...
func (a *AllBaseInfo) ContainsID(id string) bool {
	_, ok := a.IDSet[id]
	return ok
}

// ContainsName ...
func (a *AllBaseInfo) ContainsName(name string) bool {
	_, ok := a.NameSet[name]
	return ok
}

// ContainsNameExcludeSelf ...
func (a *AllBaseInfo) ContainsNameExcludeSelf(name, baseID string) bool {
	id, ok := a.NameSet[name]
	// 不存在则直接返回
	if !ok {
		return false
	}
	// 存在则需要对比是否自身名称，若是则排除掉，若不是则表示是其他的名称
	return id != baseID
}

// ContainsNameEn ...
func (a *AllBaseInfo) ContainsNameEn(nameEn string) bool {
	_, ok := a.NameEnSet[nameEn]
	return ok
}

// ContainsNameEnExcludeSelf ...
func (a *AllBaseInfo) ContainsNameEnExcludeSelf(nameEn, baseID string) bool {
	id, ok := a.NameEnSet[nameEn]
	// 不存在则直接返回
	if !ok {
		return false
	}
	// 存在则需要对比是否自身名称，若是则排除掉，若不是则表示是其他的名称
	return id != baseID
}
