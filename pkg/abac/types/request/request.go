/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package request

import (
	"iam/pkg/abac/types"
	"iam/pkg/util"
)

/*
PDP模块鉴权上下文
*/

// Request 鉴权请求
type Request struct {
	System    string
	Subject   types.Subject
	Action    types.Action
	Resources []types.Resource
}

// NewRequest new request
func NewRequest() *Request {
	return &Request{
		Subject:   types.NewSubject(),
		Action:    types.NewAction(),
		Resources: []types.Resource{},
	}
}

func (r *Request) HasResources() bool {
	return len(r.Resources) > 0
}

// HasRemoteResources ...
func (r *Request) HasRemoteResources() bool {
	for i := range r.Resources {
		if r.System != r.Resources[i].System {
			return true
		}
	}
	return false
}

// GetRemoteResources 获取请求对远端依赖资源
func (r *Request) GetRemoteResources() []*types.Resource {
	resources := make([]*types.Resource, 0, len(r.Resources))

	for i := range r.Resources {
		if r.System != r.Resources[i].System {
			resources = append(resources, &r.Resources[i])
		}
	}
	return resources
}

// ValidateActionResource 检查鉴权传的资源与action关联的资源类型是否匹配
func (r *Request) ValidateActionResource() bool {
	typeSet := r.getActionResourceTypeIDSet()
	if typeSet.Size() != len(r.Resources) {
		return false
	}

	for _, resource := range r.Resources {
		key := r.genResourceTypeKey(resource.System, resource.Type)
		if !typeSet.Has(key) {
			return false
		}
	}
	return true
}

// ValidateActionRemoteResource 检查依赖资源是否与action关联资源类型匹配
func (r *Request) ValidateActionRemoteResource() bool {
	// 检查remote资源全覆盖, local资源部分覆盖
	resourceTypes, _ := r.Action.Attribute.GetResourceTypes()

	remoteTypeSet := util.NewStringSet()
	localTypeSet := util.NewStringSet()
	for _, rt := range resourceTypes {
		if rt.System == r.System {
			localTypeSet.Add(rt.Type)
		} else {
			remoteTypeSet.Add(rt.Type)
		}
	}

	var remoteCount int

	for _, resource := range r.Resources {
		// 不匹配的local资源
		if resource.System == r.System {
			if !localTypeSet.Has(resource.Type) {
				return false
			}
		} else {
			if !remoteTypeSet.Has(resource.Type) {
				return false
			}
			remoteCount++
		}
	}
	return remoteTypeSet.Size() == remoteCount
}

func (r *Request) getActionResourceTypeIDSet() *util.StringSet {
	resourceTypes, _ := r.Action.Attribute.GetResourceTypes()
	typeSet := util.NewStringSet()
	for _, rt := range resourceTypes {
		key := r.genResourceTypeKey(rt.System, rt.Type)
		typeSet.Add(key)
	}

	return typeSet
}

func (r *Request) genResourceTypeKey(system, _type string) string {
	return system + ":" + _type
}
