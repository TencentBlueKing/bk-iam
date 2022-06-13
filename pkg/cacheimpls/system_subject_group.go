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
	hitCacheResults, err := SubjectSystemGroupCache.BatchGet(keys)
	if err != nil {
		err = errorWrapf(err, "SubjectGroupCache.BatchGet keys=`%+v` fail", keys)
		return
	}

	for _, pk := range pks {
		key := SystemSubjectPKCacheKey{SystemID: systemID, SubjectPK: pk}
		if data, ok := hitCacheResults[key]; ok {
			// do unmarshal
			var sg []types.ThinSubjectGroup
			err = SubjectSystemGroupCache.Unmarshal(conv.StringToBytes(data), &sg)
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
	cacheMap := make(map[string][]types.ThinSubjectGroup, len(notCachedSubjectGroups)+len(notCachedSubjectGroups))

	hasGroupPKs := set.NewFixedLengthInt64Set(len(notCachedSubjectGroups))
	// 3. set to cache
	for pk, sgs := range notCachedSubjectGroups {
		hasGroupPKs.Add(pk)

		key := SystemSubjectPKCacheKey{
			SystemID:  systemID,
			SubjectPK: pk,
		}

		cacheMap[key.Key()] = sgs
	}

	// 4. set the no-groups key cache
	if len(missingPKs) != hasGroupPKs.Size() {
		for _, pk := range missingPKs {
			if !hasGroupPKs.Has(pk) {
				key := SystemSubjectPKCacheKey{
					SystemID:  systemID,
					SubjectPK: pk,
				}

				cacheMap[key.Key()] = []types.ThinSubjectGroup{}
			}
		}
	}

	err := batchSetSubjectSystemGroupCache(cacheMap)
	if err != nil {
		log.Errorf("batchSetSubjectSystemGroupCache to redis fail, kvs=`%+v`, err=`%s`", cacheMap, err)
	}
}

func batchSetSubjectSystemGroupCache(cacheMap map[string][]types.ThinSubjectGroup) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "batchSetSubjectSystemGroupCache")

	cacheKvs := make([]redis.KV, 0, len(cacheMap))
	// batch set the subject_groups at one time
	for key, value := range cacheMap {
		value, err := SubjectSystemGroupCache.Marshal(value)
		if err != nil {
			return errorWrapf(err, "SubjectSystemGroupCache.Marshal fail sg=`%+v`", value)
		}

		cacheKvs = append(cacheKvs, redis.KV{
			Key:   key,
			Value: conv.BytesToString(value),
		})
	}

	err := SubjectSystemGroupCache.BatchSetWithTx(cacheKvs, 0)
	if err != nil {
		return errorWrapf(err, "SubjectSystemGroupCache.BatchSetWithTx fail kvs=`%+v`", cacheKvs)
	}

	return nil
}

func batchDeleteSubjectSystemGroupCache(systemIDs []string, subjectPKs []int64) error {
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

	err := SubjectSystemGroupCache.BatchDelete(keys)
	if err != nil {
		err = errorWrapf(err, "SubjectSystemGroupCache.BatchDelete keys=`%+v` fail", keys)
		return err
	}

	return nil
}

// BatchDeleteSubjectAllSystemGroupCache 批量删除subject 所有系统的 group 缓存
func BatchDeleteSubjectAllSystemGroupCache(subjectPKs []int64) {
	systemSVC := service.NewSystemService()
	allSystems, err := systemSVC.ListAll()
	if err != nil {
		log.WithError(err).Errorf("BatchDeleteSubjectAllSystemGroupCache fail subjectPKs=`%v`", subjectPKs)
	} else {
		systemIDs := make([]string, 0, len(allSystems))
		for _, s := range allSystems {
			systemIDs = append(systemIDs, s.ID)
		}

		err = batchDeleteSubjectSystemGroupCache(systemIDs, subjectPKs)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// BatchDeleteSubjectAuthSystemGroupCache 批量删除subject group授权系统的 group 缓存
func BatchDeleteSubjectAuthSystemGroupCache(subjectPKs []int64, parentPK int64) {
	svc := service.NewGroupService()
	systems, err := svc.ListGroupAuthSystemIDs(parentPK)
	if err != nil {
		log.WithError(err).Errorf(
			"BatchDeleteSubjectAuthSystemGroupCache fail subjectPKs=`%v`, groupPK=`%d`", subjectPKs, parentPK,
		)
	} else {
		err = batchDeleteSubjectSystemGroupCache(systems, subjectPKs)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

// BatchDeleteGroupMemberSubjectSystemGroupCache 批量删除group的member的 group 缓存
func BatchDeleteGroupMemberSubjectSystemGroupCache(systemID string, parentPK int64) {
	svc := service.NewGroupService()
	members, err := svc.ListMember(parentPK)
	if err != nil {
		log.WithError(err).Errorf(
			"BatchDeleteGroupMemberSubjectSystemGroupCache fail systemID=`%s`, groupPK=`%d`", systemID, parentPK,
		)
	} else {
		subjectPKs := make([]int64, 0, len(members))
		for _, m := range members {
			subjectPKs = append(subjectPKs, m.SubjectPK)
		}

		err = batchDeleteSubjectSystemGroupCache([]string{systemID}, subjectPKs)
		if err != nil {
			log.Error(err.Error())
		}
	}
}
