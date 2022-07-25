/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

// 查询policies
// input: subject + system_id + action_pk
// relation: pks = subject_pk + subject-department-pk, subject-group-pk, subject-department-group-pk
// database query:  system_id + action_pk + pks

// 思考:
// 组织架构变更, 影响到的是 relation, 此时获取一个subject的关系是`实时`算出来的, 所以是`实时生效`的
// 所以, 可以用  system_id + action_pk + subject_pk 为最小粒度存储策略列表

// 关键:
// 每个subject的策略变更时, 需要 删除掉对应策略缓存 (system_id + subject_pk + action_pk)

// 变数:
// 1. 某个subject的权限的增删改    => 暴力删除 system_id + subject_pk
// 2. 某个template变更后, 导致的一批subject的权限变更 => 目前限制了前端调用, 参数传递带了相关参数知道 system_id + subject_pk

// 方案:
// - 使用redis hash 存储
// - key = system_id + subject_pk
// - field = action_pk
// - value = []types.Policy{}

// 实现:
// - redis挂了不影响服务, 走db查询
// - 支持可以配置: policyCacheEnabled / policyCacheExpiration

// NOTE: 所有object - string 转换 必须使用cache.Marshal/Unmarshal

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/conv"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

const RedisLayer = "PolicyRedisLayer"

const RandExpireSeconds = 60

type redisRetriever struct {
	system              string
	actionPK            int64
	missingRetrieveFunc MissingRetrieveFunc

	keyPrefix string
}

func newRedisRetriever(system string, actionPK int64, retrieveFunc MissingRetrieveFunc) *redisRetriever {
	return &redisRetriever{
		system:              system,
		actionPK:            actionPK,
		missingRetrieveFunc: retrieveFunc,

		keyPrefix: system + ":",
	}
}

func (r *redisRetriever) genKey(subjectPK int64) cache.Key {
	return cache.NewStringKey(r.keyPrefix + strconv.FormatInt(subjectPK, 10))
}

func (r *redisRetriever) parseKey(key string) (subjectPK int64, err error) {
	subjectPKStr := strings.TrimPrefix(key, r.keyPrefix)

	subjectPK, err = strconv.ParseInt(subjectPKStr, 10, 64)
	if err != nil {
		log.WithError(err).Errorf("[%s] parseKey fail key=`%s`, keyPrefix=`%s`",
			RedisLayer, key, r.keyPrefix)
		return -1, err
	}
	return subjectPK, nil
}

func (r *redisRetriever) retrieve(subjectPKs []int64) ([]types.AuthPolicy, []int64, error) {
	nowUnix := time.Now().Unix()

	hitPolicies, missSubjectPKs, err := r.batchGet(subjectPKs)
	if err != nil {
		log.WithError(err).Errorf("[%s] batchHGet fail system=`%s`, actionPK=`%d`, subjectPKs=`%+v`",
			RedisLayer, r.system, r.actionPK, subjectPKs)

		// 从cache获取失败不影响主体功能, 走db查询
		missSubjectPKs = subjectPKs
		hitPolicies = nil
	}

	policies := make([]types.AuthPolicy, 0, len(hitPolicies))

	noPoliciesSubjectPKs := make([]int64, 0, len(subjectPKs))
	for subjectPK, policiesStr := range hitPolicies {
		var ps []types.AuthPolicy
		err = cacheimpls.PolicyCache.Unmarshal(conv.StringToBytes(policiesStr), &ps)
		if err != nil {
			log.WithError(err).
				Errorf("[%s] parse string to expression fail system=`%s`, actionPK=`%d`, subjectPKs=`%+v`",
					RedisLayer, r.system, r.actionPK, subjectPKs)

			// NOTE: 一条解析失败, 重新查/重新设置缓存
			missSubjectPKs = append(missSubjectPKs, subjectPK)

			continue
		}

		// empty policies
		if len(ps) == 0 {
			noPoliciesSubjectPKs = append(noPoliciesSubjectPKs, subjectPK)

			continue
		}

		// NOTE: check expired first!!!!!!!
		for _, p := range ps {
			if p.ExpiredAt > nowUnix {
				policies = append(policies, p)
			}
		}
	}

	if len(missSubjectPKs) == 0 {
		return policies, noPoliciesSubjectPKs, nil
	}

	// NOTE: missingPKs is missingSubjectPKs
	retrievedPolicies, missingPKs, err := r.missingRetrieveFunc(missSubjectPKs)
	if err != nil {
		return nil, nil, err
	}
	// set missing into cache
	r.setMissing(retrievedPolicies, missingPKs)
	// append the retrieved
	policies = append(policies, retrievedPolicies...)

	// NOTE: append the noPoliciesSubjectPKs into missingPKs, the upper cache layer should cache that too.
	missingPKs = append(missingPKs, noPoliciesSubjectPKs...)
	return policies, missingPKs, nil
}

func (r *redisRetriever) setMissing(policies []types.AuthPolicy, missingPKs []int64) error {
	// group policies by subjectPK
	groupedPolicies := map[int64][]types.AuthPolicy{}

	// some subject_pk has no policies, so init all as empty list first
	for _, subjectPK := range missingPKs {
		groupedPolicies[subjectPK] = []types.AuthPolicy{}
	}

	// append the exists policies into groupedPolicies
	for _, p := range policies {
		groupedPolicies[p.SubjectPK] = append(groupedPolicies[p.SubjectPK], p)
	}
	return r.batchSet(groupedPolicies)
}

func (r *redisRetriever) batchGet(subjectPKs []int64) (
	hitPolicies map[int64]string,
	missSubjectPKs []int64,
	err error,
) {
	// build for batch HGet
	hashKeyFields := make([]redis.HashKeyField, 0, len(subjectPKs))
	for _, subjectPK := range subjectPKs {
		key := r.genKey(subjectPK)
		field := strconv.FormatInt(r.actionPK, 10)

		hashKeyFields = append(hashKeyFields, redis.HashKeyField{
			Key:   key.Key(),
			Field: field,
		})
	}

	// HGet in pipeline
	hitValues, err := cacheimpls.PolicyCache.BatchHGet(hashKeyFields)
	if err != nil {
		return
	}

	// the key can identify the hit or miss, here we only need system + subjectPK
	hitPolicies = make(map[int64]string, len(hitValues))
	hitIDs := make(map[string]bool, len(hitValues))
	for hkf, policy := range hitValues {
		subjectPK, err := r.parseKey(hkf.Key)
		if err != nil {
			// skip the hit, if key parse fail
			continue
		}
		hitPolicies[subjectPK] = policy
		hitIDs[hkf.Key] = true
	}

	// the missing subjectPKs
	for _, subjectPK := range subjectPKs {
		key := r.genKey(subjectPK)
		if _, ok := hitIDs[key.Key()]; !ok {
			missSubjectPKs = append(missSubjectPKs, subjectPK)
		}
	}

	return hitPolicies, missSubjectPKs, nil
}

func (r *redisRetriever) batchSet(subjectPKPolicies map[int64][]types.AuthPolicy) error {
	// 特征: system + actionPK 是固定的, subject不固定
	// 但是: key=system:subject, field=actionPK
	// 所以: 用不了HMSet, 只能用 HSet with Pipeline
	hashes := make([]redis.Hash, 0, len(subjectPKPolicies))
	keys := make([]cache.Key, 0, len(subjectPKPolicies))

	for subjectPK, policies := range subjectPKPolicies {
		key := r.genKey(subjectPK)
		field := strconv.FormatInt(r.actionPK, 10)
		policiesBytes, err := cacheimpls.PolicyCache.Marshal(policies)
		if err != nil {
			return err
		}

		// make a hash
		hashes = append(hashes, redis.Hash{
			HashKeyField: redis.HashKeyField{
				Key:   key.Key(),
				Field: field,
			},
			Value: conv.BytesToString(policiesBytes),
		})

		// collect keys to set expire
		keys = append(keys, key)
	}

	// HSet, in a pipeline, with tx
	err := cacheimpls.PolicyCache.BatchHSetWithTx(hashes)
	if err != nil {
		log.WithError(err).Errorf(
			"[%s] cacheimpls.PolicyCache.BatchHSetWithTx fail system=`%s`, actionPK=`%d`, keys=`%+v`",
			RedisLayer, r.system, r.actionPK, keys)
		return err
	}

	// keep policy cache for 7 days
	err = cacheimpls.PolicyCache.BatchExpireWithTx(
		keys,
		cacheimpls.PolicyCacheExpiration+time.Duration(rand.Intn(RandExpireSeconds))*time.Second,
	)
	if err != nil {
		log.WithError(err).Errorf(
			"[%s] cacheimpls.PolicyCache.BatchExpireWithTx fail system=`%s`, actionPK=`%d`, keys=`%+v`",
			RedisLayer, r.system, r.actionPK, keys)
		return err
	}

	return nil
}

func (r *redisRetriever) batchDelete(subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}

	keys := make([]cache.Key, 0, len(subjectPKs))
	for _, subjectPK := range subjectPKs {
		keys = append(keys, r.genKey(subjectPK))
	}

	err := cacheimpls.PolicyCache.BatchDelete(keys)
	if err != nil {
		log.WithError(err).Errorf(
			"[%s] cacheimpls.PolicyCache.BatchDelete fail system=`%s`, actionPK=`%d`, subjectPKs=`%+v`, keys=`%+v`",
			RedisLayer, r.system, r.actionPK, subjectPKs, keys)

		// report to sentry
		util.ReportToSentry(
			"redis cache: policy cache delete fail",
			map[string]interface{}{
				"system":     r.system,
				"actionPK":   r.actionPK,
				"subjectPKs": subjectPKs,
				"keys":       keys,
				"error":      err.Error(),
			})

		return err
	}
	return nil
}

func deleteSystemSubjectPKsFromRedis(system string, subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}

	c := newRedisRetriever(system, -1, nil)
	return c.batchDelete(subjectPKs)
}

func batchDeleteSystemSubjectPKsFromRedis(systems []string, subjectPKs []int64) error {
	keys := make([]cache.Key, 0, 2*len(subjectPKs))

	for _, system := range systems {
		c := newRedisRetriever(system, -1, nil)
		for _, subjectPK := range subjectPKs {
			keys = append(keys, c.genKey(subjectPK))
		}
	}

	if len(keys) == 0 {
		return nil
	}

	err := cacheimpls.PolicyCache.BatchDelete(keys)
	if err != nil {
		log.WithError(err).Errorf("cacheimpls.PolicyCache.BatchDelete fail systems=`%+v`, subjectPKs=`%+v`, keys=`%+v`",
			systems, subjectPKs, keys)
		return err
	}
	return nil
}
