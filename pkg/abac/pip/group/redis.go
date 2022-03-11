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
	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

const RedisLayer = "SubjectGroupsRedisLayer"

const RandExpireSeconds = 60

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
	hasGroupPKs := set.NewFixedLengthInt64Set(len(notCachedSubjectGroups))
	// 3. set to cache
	for pk, sgs := range notCachedSubjectGroups {
		hasGroupPKs.Add(pk)

		key := cacheimpls.SubjectPKCacheKey{
			PK: pk,
		}

		// TODO: should collect all and do set in pipeline
		// TODO: pipeline batch set
		err := setSubjectGroupCache(key, sgs)
		if err != nil {
			log.Errorf("set subject_group to redis fail, key=%s, err=%s", key.Key(), err)
		}
	}

	// 4. set the no-groups key cache
	if len(missingPKs) != hasGroupPKs.Size() {
		for _, pk := range missingPKs {
			if !hasGroupPKs.Has(pk) {
				key := cacheimpls.SubjectPKCacheKey{
					PK: pk,
				}
				// TODO: pipeline batch set
				err := setSubjectGroupCache(key, []types.ThinSubjectGroup{})
				if err != nil {
					log.Errorf("set empty subject_group to redis fail, key=%s, err=%s", key.Key(), err)
				}
			}
		}
	}
	return nil
}

func setSubjectGroupCache(key cache.Key, subjectGroups []types.ThinSubjectGroup) error {
	return cacheimpls.SubjectGroupCache.Set(key, subjectGroups, 0)
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
