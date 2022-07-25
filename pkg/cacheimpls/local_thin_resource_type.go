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
	"strconv"

	"github.com/TencentBlueKing/gopkg/cache"

	"iam/pkg/service"
	"iam/pkg/service/types"
)

// ResourceTypePKCacheKey ...
type ResourceTypePKCacheKey struct {
	PK int64
}

// Key ...
func (k ResourceTypePKCacheKey) Key() string {
	return strconv.FormatInt(k.PK, 10)
}

func retrieveThinResourceType(key cache.Key) (interface{}, error) {
	k := key.(ResourceTypePKCacheKey)
	svc := service.NewResourceTypeService()
	return svc.GetByPK(k.PK)
}

// GetThinResourceType ...
func GetThinResourceType(pk int64) (resourceType types.ThinResourceType, err error) {
	key := ResourceTypePKCacheKey{
		PK: pk,
	}
	var value interface{}
	value, err = LocalThinResourceTypeCache.Get(key)
	if err != nil {
		return
	}

	var ok bool
	resourceType, ok = value.(types.ThinResourceType)
	if !ok {
		err = errors.New("not svctypes.ThinResourceType in cache")
		return
	}
	return
}
