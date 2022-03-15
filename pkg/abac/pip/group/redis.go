/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package group

import (
	"math/rand"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

const (
	RedisLayer        = "SubjectGroupsRedisLayer"
	RandExpireSeconds = 60
	cacheTTL          = 1 * time.Hour
)

type redisRetriever struct {
	missingRetrieveFunc MissingRetrieveFunc
}

func newRedisRetriever(retrieveFunc MissingRetrieveFunc) *redisRetriever {
	return &redisRetriever{
		missingRetrieveFunc: retrieveFunc,
	}
}

func (r *redisRetriever) retrieve(pks []int64) (map[int64][]types.ThinSubjectGroup, []int64, error) {
	hitSubjectGroups, missSubjectPKs, err := r.batchGet(pks)
	// 1. if retrieve from redis fail, will fall through to retrieve from database
	if err != nil {
		log.WithError(err).Errorf("[%s] batchGet fail subjectPKs=`%+v`, will fallthrough to database",
			RedisLayer, pks)
		missSubjectPKs = pks
		hitSubjectGroups = nil
	}

	// if all hit, return
	if len(missSubjectPKs) == 0 {
		return hitSubjectGroups, missSubjectPKs, nil
	}

	// if got miss, do retrieve
	retrievedSubjectGroups, missingPKs, err := r.missingRetrieveFunc(missSubjectPKs)
	if err != nil {
		return nil, nil, err
	}

	// set missing into cache
	r.setMissing(retrievedSubjectGroups, missingPKs)
	// append the retrieved
	for pk, sgs := range retrievedSubjectGroups {
		hitSubjectGroups[pk] = sgs
	}

	return hitSubjectGroups, missingPKs, nil
}

func (r *redisRetriever) setMissing(
	notCachedSubjectGroups map[int64][]types.ThinSubjectGroup,
	missingPKs []int64,
) error {
	// init with exists
	subjectGroups := notCachedSubjectGroups

	hasGroupPKs := set.NewFixedLengthInt64Set(len(notCachedSubjectGroups))
	for pk := range notCachedSubjectGroups {
		hasGroupPKs.Add(pk)
	}

	for _, pk := range missingPKs {
		if !hasGroupPKs.Has(pk) {
			subjectGroups[pk] = []types.ThinSubjectGroup{}
		}
	}
	return r.batchSet(subjectGroups)
}

func (r *redisRetriever) batchGet(pks []int64) (map[int64][]types.ThinSubjectGroup, []int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RedisLayer, "batchGet")

	hitSubjectGroups := make(map[int64][]types.ThinSubjectGroup, len(pks))
	missSubjectPKs := make([]int64, 0, len(pks))

	// batch get the subject_groups at one time
	keys := make([]cache.Key, 0, len(pks))
	for _, pk := range pks {
		keys = append(keys, cacheimpls.SubjectPKCacheKey{
			PK: pk,
		})
	}
	hitCacheResults, err := cacheimpls.SubjectGroupCache.BatchGet(keys)
	if err != nil {
		err = errorWrapf(err, "SubjectGroupCache.BatchGet keys=`%+v` fail", keys)
		return nil, nil, err
	}

	for _, pk := range pks {
		key := cacheimpls.SubjectPKCacheKey{PK: pk}
		if data, ok := hitCacheResults[key]; ok {
			// do unmarshal
			var sg []types.ThinSubjectGroup
			err = cacheimpls.SubjectGroupCache.Unmarshal(conv.StringToBytes(data), &sg)
			if err != nil {
				err = errorWrapf(err, "unmarshal text in cache into SubjectGroup fail", "")
				return nil, nil, err
			}
			hitSubjectGroups[pk] = append(hitSubjectGroups[pk], sg...)
		} else {
			missSubjectPKs = append(missSubjectPKs, pk)
		}
	}

	return hitSubjectGroups, missSubjectPKs, err
}

func (r *redisRetriever) batchSet(subjectGroups map[int64][]types.ThinSubjectGroup) error {
	// set into cache
	kvs := make([]redis.KV, 0, len(subjectGroups))
	for subjectPK, sgs := range subjectGroups {
		key := cacheimpls.SubjectPKCacheKey{
			PK: subjectPK,
		}

		sgsBytes, err := cacheimpls.SubjectGroupCache.Marshal(sgs)
		if err != nil {
			return err
		}

		kvs = append(kvs, redis.KV{
			Key:   key.Key(),
			Value: sgsBytes,
		})
	}

	// keep cache for 1 hour
	err := cacheimpls.SubjectGroupCache.BatchSetWithTx(
		kvs,
		cacheTTL+time.Duration(rand.Intn(RandExpireSeconds))*time.Second,
	)
	if err != nil {
		log.WithError(err).Errorf("[%s] cacheimpls.SubjectGroupCache.BatchSetWithTx fail kvs=`%+v`", RedisLayer, kvs)
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
		key := cacheimpls.SubjectPKCacheKey{
			PK: subjectPK,
		}
		keys = append(keys, key)
	}

	err := cacheimpls.SubjectGroupCache.BatchDelete(keys)
	if err != nil {
		log.WithError(err).Errorf("[%s] cacheimpls.SubjectGroupCache.BatchDelete fail keys=`%+v`", RedisLayer, keys)

		// report to sentry
		util.ReportToSentry("redis cache: subject_group cache delete fail",
			map[string]interface{}{
				"subjectPKs": subjectPKs,
				"keys":       keys,
				"error":      err.Error(),
			})

		return err
	}

	return nil
}

func batchDeleteSubjectGroupsFromRedis(subjectPKs []int64) error {
	if len(subjectPKs) == 0 {
		return nil
	}

	r := &redisRetriever{}
	return r.batchDelete(subjectPKs)
}
