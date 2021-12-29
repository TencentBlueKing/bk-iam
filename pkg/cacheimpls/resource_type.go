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
	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service"
	"iam/pkg/service/types"
)

// ResourceTypeCacheKey ...
type ResourceTypeCacheKey struct {
	SystemID       string
	ResourceTypeID string
}

// Key ...
func (k ResourceTypeCacheKey) Key() string {
	return k.SystemID + ":" + k.ResourceTypeID
}

func retrieveResourceType(key cache.Key) (resourceType interface{}, err error) {
	k := key.(ResourceTypeCacheKey)

	svc := service.NewResourceTypeService()
	return svc.Get(k.SystemID, k.ResourceTypeID)
}

// TODO: change to local?

// GetResourceType ...
func GetResourceType(systemID string, resourceTypeID string) (resourceType types.ResourceType, err error) {
	key := ResourceTypeCacheKey{
		SystemID:       systemID,
		ResourceTypeID: resourceTypeID,
	}

	err = ResourceTypeCache.GetInto(key, &resourceType, retrieveResourceType)
	err = errorx.Wrapf(err, CacheLayer, "GetResourceType",
		"ResourceTypeCache.Get key=`%s` fail", key.Key())
	return
}
