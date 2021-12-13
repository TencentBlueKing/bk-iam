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
	"strings"

	"iam/pkg/cache"
	"iam/pkg/errorx"
	"iam/pkg/service"
)

/*
 * > system clients 每次鉴权都会去请求cache, 实际上, 基本不会变更(除了上线阶段更新支持更多的clients, 目前基本没有系统会这么做)
 *
 * 1. model注册等接口, 使用 redis-cache
 * 2. policy接口, 使用 local-cache
 *
 * 即, 不影响现在的模型注册逻辑
 *
 * 但是带来的影响
 *
 * 1. `无 到 有`, 有5s的时间
 * 2. `有 到 无/变更`, 缓存时间之内, 对应client调用会失败.
 *
 * 当前设置的缓存时间: 1min
 */

func retrieveSystemClients(k cache.Key) (interface{}, error) {
	k1 := k.(cache.StringKey)

	systemID := k1.Key()

	svc := service.NewSystemService()
	system, err := svc.Get(systemID)
	if err != nil {
		return nil, err
	}

	return strings.Split(system.Clients, ","), nil
}

// GetSystemClients ...
func GetSystemClients(systemID string) (clients []string, err error) {
	key := cache.NewStringKey(systemID)

	var value interface{}
	value, err = LocalSystemClientsCache.Get(key)
	if err != nil {
		err = errorx.Wrapf(err, CacheLayer, "GetSystemClients",
			"LocalSystemClientsCache.Get key=`%s` fail", key.Key())
		return
	}

	var ok bool
	clients, ok = value.([]string)
	if !ok {
		err = errors.New("not []string in cache")
		err = errorx.Wrapf(err, CacheLayer, "GetSystemClients",
			"LocalSystemClientsCache.Get systemID=`%s` fail", systemID)
		return
	}

	return clients, nil
}
