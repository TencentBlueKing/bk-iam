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
	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

type keySubjectGroup struct {
	Key           cache.Key
	SubjectGroups []types.ThinSubjectGroup
}

// ListSystemSubjectEffectGroups ...
func ListSystemSubjectEffectGroups(systemID string, pks []int64) ([]types.ThinSubjectGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "ListSystemSubjectEffectGroups")

	// 1. get from cache
	subjectGroups := make([]types.ThinSubjectGroup, 0, len(pks))

	cachedSubjectGroups, notExistCachePKs, err := batchGetSystemSubjectGroups(systemID, pks)
	if err != nil {
		err = errorWrapf(err, "batchGetSystemSubjectGroups systemID=`%s`, pks=`%+v` fail", systemID, pks)
		return subjectGroups, err
	}
	subjectGroups = append(subjectGroups, cachedSubjectGroups...)

	// 2. all in cache, return
	if len(notExistCachePKs) == 0 {
		return subjectGroups, nil
	}
	// 3. ids of no cache, retrieve multiple
	svc := service.NewGroupService()
	// 按照时间过滤, 不应该查已过期的回来
	notCachedSubjectGroups, err := svc.ListEffectThinSubjectGroups(systemID, notExistCachePKs)
	if err != nil {
		err = errorWrapf(err, "SubjectService.ListEffectThinSubjectGroups pks=`%v` fail", notExistCachePKs)
		return nil, err
	}
	setMissingSystemSubjectGroup(systemID, notCachedSubjectGroups, notExistCachePKs)
	// append the notCachedSubjectGroups
	for _, sgs := range notCachedSubjectGroups {
		subjectGroups = append(subjectGroups, sgs...)
	}

	return subjectGroups, nil
}

func batchGetSystemSubjectGroups(
	systemID string,
	pks []int64,
) (subjectGroups []types.ThinSubjectGroup, notExistCachePKs []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "batchGetSystemSubjectGroups")

	// batch get the subject_groups at one time
	keys := make([]cache.Key, 0, len(pks))
	for _, pk := range pks {
		keys = append(keys, SystemSubjectPKCacheKey{
			SystemID:  systemID,
			SubjectPK: pk,
		})
	}
	hitCacheResults, err := SystemSubjectGroupCache.BatchGet(keys)
	if err != nil {
		err = errorWrapf(err, "SubjectGroupCache.BatchGet keys=`%+v` fail", keys)
		return
	}

	for _, pk := range pks {
		key := SystemSubjectPKCacheKey{SystemID: systemID, SubjectPK: pk}
		if data, ok := hitCacheResults[key]; ok {
			// do unmarshal
			var sg []types.ThinSubjectGroup
			err = SystemSubjectGroupCache.Unmarshal(conv.StringToBytes(data), &sg)
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

func setMissingSystemSubjectGroup(
	systemID string,
	notCachedSubjectGroups map[int64][]types.ThinSubjectGroup,
	missingPKs []int64,
) {
	kvs := make([]keySubjectGroup, 0, len(notCachedSubjectGroups)+len(notCachedSubjectGroups))

	hasGroupPKs := set.NewFixedLengthInt64Set(len(notCachedSubjectGroups))
	// 3. set to cache
	for pk, sgs := range notCachedSubjectGroups {
		hasGroupPKs.Add(pk)

		key := SystemSubjectPKCacheKey{
			SystemID:  systemID,
			SubjectPK: pk,
		}

		kvs = append(kvs, keySubjectGroup{
			Key:           key,
			SubjectGroups: sgs,
		})
	}

	// 4. set the no-groups key cache
	if len(missingPKs) != hasGroupPKs.Size() {
		for _, pk := range missingPKs {
			if !hasGroupPKs.Has(pk) {
				key := SystemSubjectPKCacheKey{
					SystemID:  systemID,
					SubjectPK: pk,
				}

				kvs = append(kvs, keySubjectGroup{
					Key:           key,
					SubjectGroups: []types.ThinSubjectGroup{},
				})
			}
		}
	}

	err := batchSetSystemSubjectGroupCache(kvs)
	if err != nil {
		log.Errorf("batchSetSystemSubjectGroupCache to redis fail, kvs=`%+v`, err=`%s`", kvs, err)
	}
}

func batchSetSystemSubjectGroupCache(kvs []keySubjectGroup) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "batchSetSystemSubjectGroup")

	cacheKvs := make([]redis.KV, 0, len(kvs))
	// batch set the subject_groups at one time
	for _, kv := range kvs {
		value, err := SystemSubjectGroupCache.Marshal(kv.SubjectGroups)
		if err != nil {
			return errorWrapf(err, "SystemSubjectGroupCache.Marshal fail sg=`%+v`", kv.SubjectGroups)
		}

		cacheKvs = append(cacheKvs, redis.KV{
			Key:   kv.Key.Key(),
			Value: conv.BytesToString(value),
		})
	}

	err := SystemSubjectGroupCache.BatchSetWithTx(cacheKvs, 0)
	if err != nil {
		return errorWrapf(err, "SystemSubjectGroupCache.BatchSetWithTx fail kvs=`%+v`", cacheKvs)
	}

	return nil
}

func BatchDeleteSystemSubjectGroupCache(systemIDs []string, subjectPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "batchGetSystemSubjectGroups")
	keys := make([]cache.Key, 0, len(subjectPKs)*len(systemIDs))
	for _, systemID := range systemIDs {
		for _, subjectPK := range subjectPKs {
			keys = append(keys, SystemSubjectPKCacheKey{
				SystemID:  systemID,
				SubjectPK: subjectPK,
			})
		}
	}

	if len(keys) == 0 {
		return nil
	}

	err := SystemSubjectGroupCache.BatchDelete(keys)
	if err != nil {
		err = errorWrapf(err, "SystemSubjectGroupCache.BatchDelete keys=`%+v` fail", keys)
		return err
	}

	return nil
}
