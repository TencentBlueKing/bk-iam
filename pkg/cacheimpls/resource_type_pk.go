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
)

func retrieveResourceTypePK(key cache.Key) (pk interface{}, err error) {
	k := key.(ResourceTypeCacheKey)

	svc := service.NewResourceTypeService()
	return svc.GetPK(k.SystemID, k.ResourceTypeID)
}

// GetResourceTypePK ...
func GetResourceTypePK(systemID string, resourceTypeID string) (resourceTypePK int64, err error) {
	key := ResourceTypeCacheKey{
		SystemID:       systemID,
		ResourceTypeID: resourceTypeID,
	}

	err = ResourceTypePKCache.GetInto(key, &resourceTypePK, retrieveResourceTypePK)
	err = errorx.Wrapf(err, CacheLayer, "GetResourceTypePK",
		"ResourceTypePKCache.Get key=`%s` fail", key.Key())
	return
}
