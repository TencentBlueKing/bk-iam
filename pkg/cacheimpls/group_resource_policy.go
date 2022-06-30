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
	"math/rand"
	"strconv"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/conv"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/service"
)

const RandExpireSeconds = 60

// GetResourceActionAuthorizedGroupPKs 使用操作与资源信息拿rbac授权的用户组
func GetResourceActionAuthorizedGroupPKs(
	systemID string,
	actionPK, actionResourceTypePK, resourceTypePK int64,
	resourceID string,
) ([]int64, error) {
	key := SystemResourceCacheKey{
		SystemID:             systemID,
		ActionResourceTypePK: actionResourceTypePK,
		ResourceTypePK:       resourceTypePK,
		ResourceID:           resourceID,
	}

	groupPKs, err := getResourceActionAuthorizedGroupPKsFromCache(key, actionPK)
	if err == nil {
		return groupPKs, nil
	}

	return retrieveResourceActionAuthorizedGroupPKs(key, actionPK)
}

func retrieveResourceActionAuthorizedGroupPKs(key SystemResourceCacheKey, actionPK int64) ([]int64, error) {
	svc := service.NewGroupResourcePolicyService()
	actionGroupPKs, err := svc.GetAuthorizedActionGroupMap(
		key.SystemID,
		key.ActionResourceTypePK,
		key.ResourceTypePK,
		key.ResourceID,
	)
	if err != nil {
		return nil, err
	}

	// 填充缓存空值
	if _, ok := actionGroupPKs[actionPK]; !ok {
		actionGroupPKs[actionPK] = []int64{}
	}

	err = batchSetActionGroupPKs(key, actionGroupPKs)
	if err != nil {
		return nil, err
	}

	return actionGroupPKs[actionPK], nil
}

// batchSetActionGroupPKs 批量设置action group pks缓存
func batchSetActionGroupPKs(key cache.Key, actionGroupPKs map[int64][]int64) error {
	hashes := make([]redis.Hash, 0, len(actionGroupPKs))
	for aPK, groupPKs := range actionGroupPKs {
		field := strconv.FormatInt(aPK, 10)
		_bytes, err := GroupResourcePolicyCache.Marshal(groupPKs)
		if err != nil {
			return err
		}

		hashes = append(hashes, redis.Hash{
			HashKeyField: redis.HashKeyField{
				Key:   key.Key(),
				Field: field,
			},
			Value: conv.BytesToString(_bytes),
		})
	}

	err := GroupResourcePolicyCache.BatchHSetWithTx(hashes)
	if err != nil {
		return err
	}

	err = GroupResourcePolicyCache.Expire(
		key,
		GroupResourcePolicyCacheExpiration+time.Duration(rand.Intn(RandExpireSeconds))*time.Second,
	)

	return err
}

// getResourceActionAuthorizedGroupPKsFromCache 从缓存拿操作与资源信息拿rbac授权的用户组
func getResourceActionAuthorizedGroupPKsFromCache(key cache.Key, actionPK int64) ([]int64, error) {
	hashKeyField := redis.HashKeyField{
		Key:   key.Key(),
		Field: strconv.FormatInt(actionPK, 10),
	}

	value, err := GroupResourcePolicyCache.HGet(hashKeyField)
	if err != nil {
		log.Errorf("GetResourceActionAuthorizedGroupPKs error: %s", err.Error())
		return nil, err
	}

	var groupPKs []int64
	err = GroupResourcePolicyCache.Unmarshal(conv.StringToBytes(value), &groupPKs)
	if err != nil {
		log.Errorf("GetResourceActionAuthorizedGroupPKs error: %s", err.Error())
		return nil, err
	}

	return groupPKs, nil
}

// DeleteResourceAuthorizedGroupPKsCache 删除资源授权的group pks缓存
func DeleteResourceAuthorizedGroupPKsCache(
	systemID string,
	actionResourceTypePK, resourceTypePK int64,
	resourceID string,
) error {
	key := SystemResourceCacheKey{
		SystemID:             systemID,
		ActionResourceTypePK: actionResourceTypePK,
		ResourceTypePK:       resourceTypePK,
		ResourceID:           resourceID,
	}

	return GroupResourcePolicyCache.Delete(key)
}
