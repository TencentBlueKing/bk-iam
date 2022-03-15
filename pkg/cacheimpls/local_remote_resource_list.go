/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"errors"
	"sort"
	"strings"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/TencentBlueKing/gopkg/stringx"

	"iam/pkg/component"
)

// RemoteResourceListCacheKey ...
type RemoteResourceListCacheKey struct {
	System string
	Type   string
	// 1;2;3
	IDs string
	// a;b;c
	Fields string
}

// Key ...
func (k RemoteResourceListCacheKey) Key() string {
	key := k.System + ":" + k.Type + ":" + k.IDs + ":" + k.Fields
	return stringx.MD5Hash(key)
}

func retrieveRemoteResourceList(k cache.Key) (interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "retrieveRemoteResourceList")

	k1 := k.(RemoteResourceListCacheKey)

	fields := strings.Split(k1.Fields, ";")
	ids := strings.Split(k1.IDs, ";")
	systemID := k1.System
	_type := k1.Type

	resources, err := listRemoteResources(systemID, _type, ids, fields)
	if err != nil {
		err = errorWrapf(err,
			"pip.ListRemoteResources systemID=`%s`, resourceTypeID=`%s`, resourceIDs=`%+v`, fields=`%s` fail",
			k1.System, k1.Type, ids, fields)
		return nil, err
	}
	return resources, nil
}

// listRemoteResources 批量获取资源的属性信息, without cache
func listRemoteResources(systemID, _type string, ids []string, fields []string) ([]map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "listRemoteResources")

	// 1. get system and resourceType
	system, err := GetSystem(systemID)
	if err != nil {
		err = errorWrapf(err, "cacheimpls.GetSystem systemID=`%s` fail", systemID)
		return nil, err
	}
	resourceType, err := GetResourceType(systemID, _type)
	if err != nil {
		err = errorWrapf(err, "cacheimpls.GetResourceType systemID=`%s`, resourceTypeID=`%s` fail", systemID, _type)
		return nil, err
	}

	// 2. make request
	req, err := component.PrepareRequest(system, resourceType)
	if err != nil {
		err = errorWrapf(err, "component.PrepareRequest systemID=`%s`, resourceTypeID=`%s` fail", systemID, _type)
		return nil, err
	}

	resources, err := component.BkRemoteResource.GetResources(req, systemID, _type, ids, fields)
	if err != nil {
		err = errorWrapf(
			err, "BkRemoteResource.GetResource systemID=`%s`, resourceTypeID=`%s`, ids length=`%d`, fields=`%s` fail",
			systemID, _type, len(ids), fields)
		return nil, err
	}

	return resources, nil
}

// ListRemoteResources ...
func ListRemoteResources(
	system string,
	_type string,
	ids []string,
	fields []string,
) (remoteResourceList []map[string]interface{}, err error) {
	// sort
	if len(ids) > 1 {
		sort.Strings(ids)
	}
	i := strings.Join(ids, ";")

	if len(fields) > 1 {
		sort.Strings(fields)
	}
	f := strings.Join(fields, ";")

	key := RemoteResourceListCacheKey{
		System: system,
		Type:   _type,
		IDs:    i,
		Fields: f,
	}

	var value interface{}
	value, err = LocalRemoteResourceListCache.Get(key)
	if err != nil {
		return
	}

	var ok bool
	remoteResourceList, ok = value.([]map[string]interface{})
	if !ok {
		err = errors.New("not []map[string]interface{} in cache")
		return
	}
	return remoteResourceList, nil
}
