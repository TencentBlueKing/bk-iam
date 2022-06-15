/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

/*
临时权限的查询分为2步:

1. 查询临时权限的pks
	Redis Hash

	key:   system + subjectPK
	filed: actionPK
	value: policy pk and expired_at

	HGet (hit) -> return pk and expired_at
	     (miss) -> DB -> HSet -> return pk and expired_at

2. 通过pks查询临时权限的策略列表
	Local Cache

	key:   policy pk
	value: policy

	Get From Cache (without miss pks) -> return policies
	               (miss pks) -> DB -> Set miss -> return policies
*/

import (
	"strconv"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/conv"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

const (
	TemporaryPolicyRedisLayer  = "TemporaryPolicyRedisLayer"
	TemporaryPolicyMemoryLayer = "TemporaryPolicyMemoryLayer"

	TemporaryPolicyCacheDelaySeconds = 10
)

type temporaryPolicyCache struct {
	*temporaryPolicyRedisCache
	*temporaryPolicyLocalCache
}

func newTemporaryPolicyCacheRetriever(
	system string,
	svc service.TemporaryPolicyService,
) TemporaryPolicyRetriever {
	return &temporaryPolicyCache{
		temporaryPolicyRedisCache: newTemporaryPolicyRedisCache(system, svc),
		temporaryPolicyLocalCache: newTemporaryPolicyLocalCache(svc),
	}
}

type temporaryPolicyRedisCache struct {
	keyPrefix              string
	temporaryPolicyService service.TemporaryPolicyService
}

func newTemporaryPolicyRedisCache(
	system string,
	svc service.TemporaryPolicyService,
) *temporaryPolicyRedisCache {
	return &temporaryPolicyRedisCache{
		keyPrefix:              system + ":",
		temporaryPolicyService: svc,
	}
}

func (c *temporaryPolicyRedisCache) genKey(subjectPK int64) cache.Key {
	return cache.NewStringKey(c.keyPrefix + strconv.FormatInt(subjectPK, 10))
}

func (c *temporaryPolicyRedisCache) genHashKeyField(subjectPK, actionPK int64) redis.HashKeyField {
	key := c.genKey(subjectPK)
	field := strconv.FormatInt(actionPK, 10)
	return redis.HashKeyField{
		Key:   key.Key(),
		Field: field,
	}
}

// ListThinBySubjectAction ...
func (c *temporaryPolicyRedisCache) ListThinBySubjectAction(
	subjectPK, actionPK int64,
) (ps []types.ThinTemporaryPolicy, err error) {
	// 从Redis中查询缓存
	ps, err = c.getThinTemporaryPoliciesFromCache(subjectPK, actionPK)
	if err == nil {
		return ps, nil
	}

	// Redis miss时, 从DB查询
	ps, err = c.temporaryPolicyService.ListThinBySubjectAction(subjectPK, actionPK)
	if err != nil {
		return
	}

	// set 数据到Redis中
	c.setThinTemporaryPoliciesToCache(subjectPK, actionPK, ps)
	return
}

func (c *temporaryPolicyRedisCache) setThinTemporaryPoliciesToCache(
	subjectPK, actionPK int64, ps []types.ThinTemporaryPolicy,
) {
	valueByets, err := cacheimpls.TemporaryPolicyCache.Marshal(ps)
	if err != nil {
		log.WithError(err).Errorf("[%s] Marshal fail ps=`%+v`",
			TemporaryPolicyRedisLayer, ps)
		return
	}

	hashKeyField := c.genHashKeyField(subjectPK, actionPK)
	err = cacheimpls.TemporaryPolicyCache.HSet(hashKeyField, conv.BytesToString(valueByets))
	if err != nil {
		log.WithError(err).Errorf("[%s] HSet fail keyPrefix=`%s`, actionPK=`%d`, subjectPK=`%d`, policies=`%+v`",
			TemporaryPolicyRedisLayer, c.keyPrefix, actionPK, subjectPK, ps)
	}
	err = cacheimpls.TemporaryPolicyCache.Expire(c.genKey(subjectPK), 0) // 0 means default duration
	if err != nil {
		log.WithError(err).Errorf("[%s] Expire fail keyPrefix=`%s`, subjectPK=`%d`",
			TemporaryPolicyRedisLayer, c.keyPrefix, subjectPK)
	}
}

func (c *temporaryPolicyRedisCache) getThinTemporaryPoliciesFromCache(
	subjectPK, actionPK int64,
) (ps []types.ThinTemporaryPolicy, err error) {
	hashKeyField := c.genHashKeyField(subjectPK, actionPK)
	value, err := cacheimpls.TemporaryPolicyCache.HGet(hashKeyField)
	if err != nil {
		log.WithError(err).Errorf("[%s] HGet fail keyPrefix=`%s`, actionPK=`%d`, subjectPK=`%d`",
			TemporaryPolicyRedisLayer, c.keyPrefix, actionPK, subjectPK)
		return
	}

	err = cacheimpls.TemporaryPolicyCache.Unmarshal(conv.StringToBytes(value), &ps)
	if err != nil {
		log.WithError(err).Errorf(
			"[%s] Unmarshal fail value=`%s`", TemporaryPolicyRedisLayer, value)
		return
	}
	return ps, nil
}

// DeleteBySubject ...
func (c *temporaryPolicyRedisCache) DeleteBySubject(subjectPK int64) error {
	key := c.genKey(subjectPK)
	return cacheimpls.TemporaryPolicyCache.Delete(key)
}

type temporaryPolicyLocalCache struct {
	temporaryPolicyService service.TemporaryPolicyService
}

func newTemporaryPolicyLocalCache(svc service.TemporaryPolicyService) *temporaryPolicyLocalCache {
	return &temporaryPolicyLocalCache{
		temporaryPolicyService: svc,
	}
}

// ListByPKs ...
func (c *temporaryPolicyLocalCache) ListByPKs(pks []int64) ([]types.TemporaryPolicy, error) {
	// 从本地缓存中批量查询
	policies, missPKs := c.batchGet(pks)
	if len(missPKs) == 0 {
		return policies, nil
	}

	// 没有查到的pks回落到db查询
	retrievedPolicies, err := c.temporaryPolicyService.ListByPKs(missPKs)
	if err != nil {
		return nil, err
	}

	// 从db中查询到的policies set到本地缓存
	if len(retrievedPolicies) != 0 {
		c.setMissing(retrievedPolicies)
		policies = append(policies, retrievedPolicies...)
	}
	return policies, nil
}

func (c *temporaryPolicyLocalCache) batchGet(pks []int64) ([]types.TemporaryPolicy, []int64) {
	// 本地缓存以policy pk为key, 批量查询循环get
	policies := make([]types.TemporaryPolicy, 0, len(pks))
	missPKs := make([]int64, 0, len(pks))
	for _, pk := range pks {
		key := strconv.FormatInt(pk, 10)
		value, found := cacheimpls.LocalTemporayPolicyCache.Get(key)
		if !found {
			missPKs = append(missPKs, pk)
			continue
		}
		policy, ok := value.(*types.TemporaryPolicy)
		if !ok {
			log.Errorf(
				"[%s] parse cachedTemporaryPolicy in memory cache fail, will do retrieve!",
				TemporaryPolicyMemoryLayer,
			)
			missPKs = append(missPKs, pk)
			continue
		}

		policies = append(policies, *policy)
	}
	return policies, missPKs
}

func (c *temporaryPolicyLocalCache) setMissing(policies []types.TemporaryPolicy) {
	nowTimestamp := time.Now().Unix()
	for i := range policies {
		p := &policies[i] // NOTE: 避免循环中对象复制, 取指针地址重复
		key := strconv.FormatInt(p.PK, 10)
		ttl := p.ExpiredAt - nowTimestamp + TemporaryPolicyCacheDelaySeconds

		cacheimpls.LocalTemporayPolicyCache.Set(key, p, time.Duration(ttl))
	}
}

func DeleteTemporaryPolicyBySystemSubjectFromCache(systemID string, subjectPK int64) error {
	tpRedisCache := newTemporaryPolicyRedisCache(systemID, service.NewTemporaryPolicyService())
	return tpRedisCache.DeleteBySubject(subjectPK)
}
