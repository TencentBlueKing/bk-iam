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
	system        string
	keyPrefix     string
	policyService service.TemporaryPolicyService
}

func newTemporaryPolicyCache(system string) *temporaryPolicyCache {
	return &temporaryPolicyCache{
		system:        system,
		keyPrefix:     system + ":",
		policyService: service.NewTemporaryPolicyService(),
	}
}

func (c *temporaryPolicyCache) genKey(subjectPK int64) cache.Key {
	return cache.NewStringKey(c.keyPrefix + strconv.FormatInt(subjectPK, 10))
}

// ListThinBySubjectAction ...
func (c *temporaryPolicyCache) ListThinBySubjectAction(
	subjectPK, actionPK int64,
) (ps []types.ThinTemporaryPolicy, err error) {
	key := c.genKey(subjectPK)
	field := strconv.FormatInt(actionPK, 10)

	hashKeyField := redis.HashKeyField{
		Key:   key.Key(),
		Field: field,
	}

	err = cacheimpls.TemporaryPolicyCache.HGet(hashKeyField, &ps)
	if err != nil {
		log.WithError(err).Errorf("[%s] HGet fail system=`%s`, actionPK=`%d`, subjectPK=`%d`",
			TemporaryPolicyRedisLayer, c.system, actionPK, subjectPK)

		ps, err = c.policyService.ListThinBySubjectAction(subjectPK, actionPK)
		if err != nil {
			return nil, err
		}

		err = cacheimpls.TemporaryPolicyCache.HSet(hashKeyField, ps, 0)
		if err != nil {
			log.WithError(err).Errorf("[%s] HSet fail system=`%s`, actionPK=`%d`, subjectPK=`%d`, policies=`%+v`",
				TemporaryPolicyRedisLayer, c.system, actionPK, subjectPK, ps)
		}
	}
	return ps, nil
}

// DeleteBySubject ...
func (c *temporaryPolicyCache) DeleteBySubject(subjectPK int64) error {
	key := c.genKey(subjectPK)
	return cacheimpls.TemporaryPolicyCache.Delete(key)
}

// ListByPKs ...
func (c *temporaryPolicyCache) ListByPKs(pks []int64) ([]types.TemporaryPolicy, error) {
	policies, missPKs := c.batchGet(pks)
	if len(missPKs) == 0 {
		return policies, nil
	}

	retrievedPolicies, err := c.policyService.ListByPKs(missPKs)
	if err != nil {
		return nil, err
	}

	if len(retrievedPolicies) != 0 {
		c.setMissing(retrievedPolicies)
		policies = append(policies, retrievedPolicies...)
	}
	return policies, nil
}

func (c *temporaryPolicyCache) batchGet(pks []int64) ([]types.TemporaryPolicy, []int64) {
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

func (c *temporaryPolicyCache) setMissing(policies []types.TemporaryPolicy) {
	nowTimestamp := time.Now().Unix()
	for i := range policies {
		p := &policies[i]
		key := strconv.FormatInt(p.PK, 10)
		ttl := p.ExpiredAt - nowTimestamp + TemporaryPolicyCacheDelaySeconds

		cacheimpls.LocalTemporayPolicyCache.Set(key, p, time.Duration(ttl))
	}
}
