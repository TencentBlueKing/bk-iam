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
	"fmt"
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

// SystemResourceCacheKey ...
type SystemResourceCacheKey struct {
	SystemID             string
	ActionResourceTypePK int64
	ResourceTypePK       int64
	ResourceID           string
}

func (k SystemResourceCacheKey) Key() string {
	return fmt.Sprintf("%s:%d:%d:%s", k.SystemID, k.ActionResourceTypePK, k.ResourceTypePK, k.ResourceID)
}

// GetResourceActionAuthorizedGroupPKs 使用操作与资源信息拿rbac授权的用户组
// Get Resource Action Authorized Group PKs from redis Hash
// Key: system_id:action_resource_type_pk:resource_type_pk:resource_id
// Hash Field: action_pk
// Hash Value: group_pks []int64
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

	groupPKs := actionGroupPKs[actionPK]
	err = setActionGroupPKs(key, actionPK, groupPKs)
	if err != nil {
		// 注意, 不能返回err, 失败就失败了, 不影响正常返回
		// return nil, err
		log.WithError(err).
			Errorf("setActionGroupPKs error, key=%s, actionPK=%d, groupPKs=%v", key.Key(), actionPK, groupPKs)
	}

	return actionGroupPKs[actionPK], nil
}

func setActionGroupPKs(key cache.Key, actionPK int64, groupPKs []int64) error {
	hashKeyField := redis.HashKeyField{
		Key:   key.Key(),
		Field: strconv.FormatInt(actionPK, 10),
	}
	_bytes, err := GroupResourcePolicyCache.Marshal(groupPKs)
	if err != nil {
		return err
	}

	err = GroupResourcePolicyCache.HSet(hashKeyField, conv.BytesToString(_bytes))
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
