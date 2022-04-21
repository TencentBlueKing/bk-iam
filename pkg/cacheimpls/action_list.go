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

func retrieveActionList(key cache.Key) (interface{}, error) {
	k := key.(SystemIDCacheKey)
	svc := service.NewActionService()
	return svc.ListBySystem(k.SystemID)
}

func ListActionBySystem(systemID string) (actions []types.Action, err error) {
	key := SystemIDCacheKey{
		SystemID: systemID,
	}

	err = ActionListCache.GetInto(key, &actions, retrieveActionList)
	if err != nil {
		err = errorx.Wrapf(err, CacheLayer, "ListActionBySystem",
			"ActionListCache.GetInto key=`%s` fail", key.Key())
	}

	return
}
