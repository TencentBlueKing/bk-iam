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

// TemporaryPolicyListService ...
type TemporaryPolicyListService interface {
	ListThinBySubjectAction(subjectPK, actionPK int64) ([]types.ThinTemporaryPolicy, error)
	ListByPKs(pks []int64) ([]types.TemporaryPolicy, error)
}

type temporaryPolicyCache struct {
	*temporaryPolicyRedisCache
	*temporaryPolicyLocalCache
}

func newTemporaryPolicyCache(system string) *temporaryPolicyCache {
	return &temporaryPolicyCache{
		temporaryPolicyRedisCache: newTemporaryPolicyRedisCache(system),
		temporaryPolicyLocalCache: newTemporaryPolicyLocalCache(),
	}
}

type temporaryPolicyRedisCache struct {
	system                 string
	keyPrefix              string
	temporaryPolicyService service.TemporaryPolicyService
}

func newTemporaryPolicyRedisCache(system string) *temporaryPolicyRedisCache {
	return &temporaryPolicyRedisCache{
		system:                 system,
		keyPrefix:              system + ":",
		temporaryPolicyService: service.NewTemporaryPolicyService(),
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
	ps, err = c.getThinTemporaryPoliciesFromCache(subjectPK, actionPK)
	if err == nil {
		return
	}

	ps, err = c.temporaryPolicyService.ListThinBySubjectAction(subjectPK, actionPK)
	if err != nil {
		return nil, err
	}

	c.setThinTemporaryPoliciesToCache(subjectPK, actionPK, ps)
	return ps, nil
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
	err = cacheimpls.TemporaryPolicyCache.HSet(hashKeyField, conv.BytesToString(valueByets), 0)
	if err != nil {
		log.WithError(err).Errorf("[%s] HSet fail system=`%s`, actionPK=`%d`, subjectPK=`%d`, policies=`%+v`",
			TemporaryPolicyRedisLayer, c.system, actionPK, subjectPK, ps)
	}
}

func (c *temporaryPolicyRedisCache) getThinTemporaryPoliciesFromCache(
	subjectPK, actionPK int64,
) (ps []types.ThinTemporaryPolicy, err error) {
	hashKeyField := c.genHashKeyField(subjectPK, actionPK)
	value, err := cacheimpls.TemporaryPolicyCache.HGet(hashKeyField)
	if err != nil {
		log.WithError(err).Errorf("[%s] HGet fail system=`%s`, actionPK=`%d`, subjectPK=`%d`",
			TemporaryPolicyRedisLayer, c.system, actionPK, subjectPK)
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

func newTemporaryPolicyLocalCache() *temporaryPolicyLocalCache {
	return &temporaryPolicyLocalCache{
		temporaryPolicyService: service.NewTemporaryPolicyService(),
	}
}

// ListByPKs ...
func (c *temporaryPolicyLocalCache) ListByPKs(pks []int64) ([]types.TemporaryPolicy, error) {
	policies, missPKs := c.batchGet(pks)
	if len(missPKs) == 0 {
		return policies, nil
	}

	retrievedPolicies, err := c.temporaryPolicyService.ListByPKs(missPKs)
	if err != nil {
		return nil, err
	}

	if len(retrievedPolicies) != 0 {
		c.setMissing(retrievedPolicies)
		policies = append(policies, retrievedPolicies...)
	}
	return policies, nil
}

func (c *temporaryPolicyLocalCache) batchGet(pks []int64) ([]types.TemporaryPolicy, []int64) {
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
			log.Errorf("[%s] parse cachedTemporaryPolicy in memory cache fail, will do retrieve!", TemporaryPolicyMemoryLayer)
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
		p := &policies[i]
		key := strconv.FormatInt(p.PK, 10)
		ttl := p.ExpiredAt - nowTimestamp + TemporaryPolicyCacheDelaySeconds

		cacheimpls.LocalTemporayPolicyCache.Set(key, p, time.Duration(ttl))
	}
}
