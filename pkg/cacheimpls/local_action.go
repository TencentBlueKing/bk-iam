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

	"iam/pkg/cache"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

/*
for iam engine

get action from action_pk when list policy for iam engine sync policy data
*/

// ActionPKCacheKey ...
type ActionPKCacheKey struct {
	PK int64
}

// Key ...
func (k ActionPKCacheKey) Key() string {
	return strconv.FormatInt(k.PK, 10)
}

func retrieveAction(key cache.Key) (interface{}, error) {
	k := key.(ActionPKCacheKey)
	svc := service.NewActionService()
	return svc.GetThinActionByPK(k.PK)
}

// GetAction ...
func GetAction(pk int64) (action types.ThinAction, err error) {
	key := ActionPKCacheKey{
		PK: pk,
	}
	var value interface{}
	value, err = LocalActionCache.Get(key)
	if err != nil {
		return
	}

	var ok bool
	action, ok = value.(types.ThinAction)
	if !ok {
		err = errors.New("not svctypes.ThinAction in cache")
		return
	}
	return
}
