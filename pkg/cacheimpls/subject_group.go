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
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache"
	"iam/pkg/errorx"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

func retrieveSubjectGroups(key cache.Key) (interface{}, error) {
	k := key.(SubjectPKCacheKey)

	svc := service.NewSubjectService()
	return svc.GetThinSubjectGroups(k.PK)
}

// TODO: remove this cache? => if we can know a department add or remove from a group?

// GetSubjectGroups get groups by subject pk
func GetSubjectGroups(pk int64) (subjectGroups []types.ThinSubjectGroup, err error) {
	key := SubjectPKCacheKey{
		PK: pk,
	}

	err = SubjectGroupCache.GetInto(key, &subjectGroups, retrieveSubjectGroups)
	err = errorx.Wrapf(err, CacheLayer, "GetSubjectGroups",
		"SubjectGroupCache.Get key=`%s` fail", key.Key())
	return
}

// ListSubjectEffectGroups ...
func ListSubjectEffectGroups(pks []int64) ([]types.ThinSubjectGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "ListSubjectEffectGroups")

	// 1. get from cache
	subjectGroups := make([]types.ThinSubjectGroup, 0, len(pks))

	cachedSubjectGroups, notExistCachePKs, err := batchGetSubjectGroups(pks)
	if err != nil {
		err = errorWrapf(err, "batchGetSubjectGroups pks=`%+v` fail", pks)
		return subjectGroups, err
	}
	subjectGroups = append(subjectGroups, cachedSubjectGroups...)

	// 2. all in cache, return
	if len(notExistCachePKs) == 0 {
		return subjectGroups, nil
	}
	// 3. ids of no cache, retrieve multiple
	svc := service.NewSubjectService()
	// 按照时间过滤, 不应该查已过期的回来
	notCachedSubjectGroups, err := svc.ListSubjectEffectGroups(notExistCachePKs)
	if err != nil {
		err = errorWrapf(err, "SubjectService.ListSubjectEffectGroups pks=`%v` fail", notExistCachePKs)
		return nil, err
	}
	setMissing(notCachedSubjectGroups, notExistCachePKs)
	// append the notCachedSubjectGroups
	for _, sgs := range notCachedSubjectGroups {
		subjectGroups = append(subjectGroups, sgs...)
	}

	return subjectGroups, nil
}

func batchGetSubjectGroups(pks []int64) (subjectGroups []types.ThinSubjectGroup, notExistCachePKs []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "batchGetSubjectGroups")

	// batch get the subject_groups at one time
	keys := make([]cache.Key, 0, len(pks))
	for _, pk := range pks {
		keys = append(keys, SubjectPKCacheKey{
			PK: pk,
		})
	}
	hitCacheResults, err := SubjectGroupCache.BatchGet(keys)
	if err != nil {
		err = errorWrapf(err, "SubjectGroupCache.BatchGet keys=`%+v` fail", keys)
		return
	}

	for _, pk := range pks {
		key := SubjectPKCacheKey{PK: pk}
		if data, ok := hitCacheResults[key]; ok {
			// do unmarshal
			var sg []types.ThinSubjectGroup
			err = SubjectGroupCache.Unmarshal(util.StringToBytes(data), &sg)
			if err != nil {
				err = errorWrapf(err, "unmarshal text in cache into SubjectGroup fail", "")
				return
			}
			subjectGroups = append(subjectGroups, sg...)
		} else {
			notExistCachePKs = append(notExistCachePKs, pk)
		}
	}

	return subjectGroups, notExistCachePKs, err
}

func setMissing(notCachedSubjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64) {
	hasGroupPKs := util.NewFixedLengthInt64Set(len(notCachedSubjectGroups))
	// 3. set to cache
	for pk, sgs := range notCachedSubjectGroups {
		hasGroupPKs.Add(pk)

		key := SubjectPKCacheKey{
			PK: pk,
		}

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
				key := SubjectPKCacheKey{
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
}

func setSubjectGroupCache(key cache.Key, subjectGroups []types.ThinSubjectGroup) error {
	return SubjectGroupCache.Set(key, subjectGroups, 0)
}
