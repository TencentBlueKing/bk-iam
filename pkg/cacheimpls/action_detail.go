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

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service"
	"iam/pkg/service/types"
)

func retrieveActionDetail(key cache.Key) (interface{}, error) {
	k := key.(ActionIDCacheKey)
	svc := service.NewActionService()

	pk, err := svc.GetActionPK(k.SystemID, k.ActionID)
	if err != nil {
		return nil, err
	}

	// 从saas_action拿 auth_type
	authType, err := svc.GetAuthType(k.SystemID, k.ActionID)
	if err != nil {
		return nil, err
	}

	resourceTypes, err := svc.ListThinActionResourceTypes(k.SystemID, k.ActionID)
	if err != nil {
		return nil, err
	}

	// NOTE: you should not add new field in ActionDetail, unless you know how to upgrade
	// 如果要加新成员, 必须变更cache名字, 防止从已有缓存数据拿不到对应的字段产生bug
	detail := types.ActionDetail{
		PK:            pk,
		AuthType:      authType,
		ResourceTypes: resourceTypes,
	}
	return detail, nil
}

func retrieveActionDetailForLocal(key cache.Key) (interface{}, error) {
	detail := types.ActionDetail{}

	// Note: 这里是三级缓存，LocalMemory失效后，从Redis Cache里获取，并不直接查询DB
	err := ActionDetailCache.GetInto(key, &detail, retrieveActionDetail)
	err = errorx.Wrapf(err, CacheLayer, "retrieveActionDetailForLocal",
		"ActionDetailCache.GetInto key=`%s` fail", key.Key())

	return detail, err
}

func GetActionDetail(systemID, actionID string) (detail types.ActionDetail, err error) {
	key := ActionIDCacheKey{
		SystemID: systemID,
		ActionID: actionID,
	}
	var value interface{}
	value, err = LocalActionDetailCache.Get(key)
	if err != nil {
		return detail, errorx.Wrapf(err, CacheLayer, "GetActionDetail",
			"LocalActionDetailCache.Get key=`%s` fail", key.Key())
	}

	var ok bool
	detail, ok = value.(types.ActionDetail)
	if !ok {
		err = errors.New("not types.ActionDetail in cache")
		return
	}
	return
}
