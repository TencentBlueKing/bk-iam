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
	"sort"
	"strings"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/TencentBlueKing/gopkg/stringx"
)

// RemoteResourceCacheKey ...
type RemoteResourceCacheKey struct {
	System string
	Type   string
	ID     string
	// a;b;c
	Fields string
}

// Key ...
func (k RemoteResourceCacheKey) Key() string {
	key := k.System + ":" + k.Type + ":" + k.ID + ":" + k.Fields
	return stringx.MD5Hash(key)
}

func retrieveRemoteResource(k cache.Key) (interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "retrieveRemoteResource")

	k1 := k.(RemoteResourceCacheKey)

	fields := strings.Split(k1.Fields, ";")

	resources, err := listRemoteResources(k1.System, k1.Type, []string{k1.ID}, fields)
	if err != nil {
		err = errorWrapf(
			err,
			"listRemoteResources systemID=`%s`, resourceTypeID=`%s`, resourceID=`%s`, fields=`%s` fail",
			k1.System,
			k1.Type,
			k1.ID,
			fields,
		)
		return nil, err
	}
	return resources[0], nil
}

// GetRemoteResource ...
func GetRemoteResource(system, _type, id string, fields []string) (remoteResource map[string]interface{}, err error) {
	// sort
	if len(fields) > 1 {
		sort.Strings(fields)
	}

	f := strings.Join(fields, ";")
	key := RemoteResourceCacheKey{
		System: system,
		Type:   _type,
		ID:     id,
		Fields: f,
	}

	err = RemoteResourceCache.GetInto(key, &remoteResource, retrieveRemoteResource)
	err = errorx.Wrapf(err, CacheLayer, "GetRemoteResource",
		"RemoteResourceCache.Get key=`%s` fail", key.Key())
	return
}
