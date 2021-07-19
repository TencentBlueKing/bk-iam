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
	"iam/pkg/errorx"
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

// HasSingleLocalResource 是否只有一个本地依赖资源
func (r *Request) HasSingleLocalResource() bool {
	resourceTypes, _ := r.Action.Attribute.GetResourceTypes()

	return len(resourceTypes) == 1 && resourceTypes[0].System == r.System
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

// GetSortedResources ...
func (r *Request) GetSortedResources() []*types.Resource {
	resources := make([]*types.Resource, 0, len(r.Resources))

	remoteResources := make([]*types.Resource, 0, len(r.Resources))

	for i := range r.Resources {
		if r.System == r.Resources[i].System {
			resources = append(resources, &r.Resources[i])
		} else {
			remoteResources = append(remoteResources, &r.Resources[i])
		}
	}

	resources = append(resources, remoteResources...)
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

// GetQueryResourceTypes ...
func (r *Request) GetQueryResourceTypes() (queryResourceTypes []types.ActionResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Request", "GetActionResourceTypes")

	resourceTypes, err := r.Action.Attribute.GetResourceTypes()
	if err != nil {
		err = errorWrapf(err, "action GetResourceTypes fail")
		return
	}

	existingResourceSet := util.NewStringSet()
	for _, resource := range r.Resources {
		key := r.genResourceTypeKey(resource.System, resource.Type)
		existingResourceSet.Add(key)
	}

	// 只查找参数中不存在的资源类型
	for _, rt := range resourceTypes {
		key := r.genResourceTypeKey(rt.System, rt.Type)
		if !existingResourceSet.Has(key) {
			queryResourceTypes = append(queryResourceTypes, rt)
		}
	}

	return
}

func (r *Request) genResourceTypeKey(system, _type string) string {
	return system + ":" + _type
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
