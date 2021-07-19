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

import "fmt"

// Attribute 属性map
type Attribute map[string]interface{}

// Set ...
func (a Attribute) Set(key string, values interface{}) {
	a[key] = values
}

// Delete ...
func (a Attribute) Delete(key string) {
	delete(a, key)
}

// Has 是否有name的属性
func (a Attribute) Has(key string) (ok bool) {
	_, ok = a[key]
	return
}

// Keys ...
func (a Attribute) Keys() []string {
	if len(a) == 0 {
		return []string{}
	}

	keys := make([]string, 0, len(a))
	for key := range a {
		keys = append(keys, key)
	}
	return keys
}

// Get 获取name的值
func (a Attribute) Get(key string) (values interface{}, ok bool) {
	values, ok = a[key]
	return
}

// GetInt64 ...
func (a Attribute) GetInt64(key string) (int64, error) {
	v, ok := a[key]
	if !ok {
		return 0, fmt.Errorf("key %s not exists", key)
	}
	vInt64, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("value %+v of key %s can not convert to int64", v, key)
	}
	return vInt64, nil
}

// GetInt64Slice ...
func (a Attribute) GetInt64Slice(key string) ([]int64, error) {
	v, ok := a[key]
	if !ok {
		return nil, fmt.Errorf("key %s not exists", key)
	}
	vInt64Slice, ok := v.([]int64)
	if !ok {
		return nil, fmt.Errorf("value %+v of key %s can not convert to []int64", v, key)
	}
	return vInt64Slice, nil
}

// GetString ...
func (a Attribute) GetString(key string) (string, error) {
	v, ok := a[key]
	if !ok {
		return "", fmt.Errorf("key %s not exists", key)
	}
	vStr, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value %+v of key %s can not convert to string", v, key)
	}
	return vStr, nil
}

// ActionAttribute 操作的属性
type ActionAttribute struct {
	Attribute
}

// NewActionAttribute new ActionAttribute
func NewActionAttribute() *ActionAttribute {
	return &ActionAttribute{
		Attribute: make(map[string]interface{}),
	}
}

// GetPK 获取pk
func (a *ActionAttribute) GetPK() (int64, error) {
	return a.GetInt64(PKAttrName)
}

// SetPK 设置pk
func (a *ActionAttribute) SetPK(pk int64) {
	a.Set(PKAttrName, pk)
}

// GetResourceTypes 获取资源类型
func (a *ActionAttribute) GetResourceTypes() ([]ActionResourceType, error) {
	key := ResourceTypeAttrName

	actionResourceTypes, ok := a.Get(key)
	if !ok {
		return nil, fmt.Errorf("key %s not exists", key)
	}

	resourceTypes, ok := actionResourceTypes.([]ActionResourceType)
	if !ok {
		return nil, fmt.Errorf("value %+v of key %s can not convert to []ActionResourceType", actionResourceTypes, key)
	}
	return resourceTypes, nil
}

// SetResourceTypes 设置操作的资源类型
func (a *ActionAttribute) SetResourceTypes(resourceTypes []ActionResourceType) {
	a.Set(ResourceTypeAttrName, resourceTypes)
}

// SubjectAttribute subject 的属性
type SubjectAttribute struct {
	Attribute
}

// NewSubjectAttribute new SubjectAttribute
func NewSubjectAttribute() *SubjectAttribute {
	return &SubjectAttribute{
		Attribute: make(map[string]interface{}),
	}
}

// GetPK 获取pk
func (a *SubjectAttribute) GetPK() (int64, error) {
	return a.GetInt64(PKAttrName)
}

// SetPK 设置pk
func (a *SubjectAttribute) SetPK(pk int64) {
	a.Set(PKAttrName, pk)
}

// GetGroups 获取subject属于的组
func (a *SubjectAttribute) GetGroups() ([]SubjectGroup, error) {
	groups, ok := a.Get(GroupAttrName)
	if !ok {
		return nil, fmt.Errorf("key %s not exists", GroupAttrName)
	}
	subjectGroups, ok := groups.([]SubjectGroup)
	if !ok {
		return nil, fmt.Errorf("value %+v of key %s can not convert to []SubjectGroup", groups, GroupAttrName)
	}
	return subjectGroups, nil
}

// SetGroups 设置用户组
func (a *SubjectAttribute) SetGroups(groups []SubjectGroup) {
	a.Set(GroupAttrName, groups)
}

// GetDepartments 获取subject属于的部门
func (a *SubjectAttribute) GetDepartments() ([]int64, error) {
	return a.GetInt64Slice(DeptAttrName)
}

// SetDepartments 设置部门
func (a *SubjectAttribute) SetDepartments(department []int64) {
	a.Set(DeptAttrName, department)
}
