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
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

const (
	MemoryLayer = "SubjectGroupsMemoryLayer"

	changeListTypeSubjectGroup = "subject_group"
	// local cache ttl should equals to changeList get by score (max=now, min=max-localCacheTTL)
	subjectGroupsLocalCacheTTL = 60

	// fetch the top 1000 in change list, protect the auth/query api performance,
	// if hit 1000, some local-cached(action -> expressionPK) updated event will not be notified,
	// will be expired in policyLocalCacheTTL, accepted by now
	maxChangeListCount = 1000
)

var changeList = common.NewChangeList(changeListTypeSubjectGroup, subjectGroupsLocalCacheTTL, maxChangeListCount)

type memoryRetriever struct {
	missingRetrieveFunc MissingRetrieveFunc

	changeListKey string
}

func newMemoryRetriever(subjectType string, retrieveFunc MissingRetrieveFunc) *memoryRetriever {
	return &memoryRetriever{
		changeListKey:       subjectType,
		missingRetrieveFunc: retrieveFunc,
	}
}

type cachedSubjectGroups struct {
	timestamp     int64
	subjectGroups []types.ThinSubjectGroup
}

func (r *memoryRetriever) genKey(subjectPK int64) string {
	return strconv.FormatInt(subjectPK, 10)
}

func (r *memoryRetriever) retrieve(pks []int64) (map[int64][]types.ThinSubjectGroup, []int64, error) {
	missSubjectPKs := make([]int64, 0, len(pks))
	hitSubjectGroups := make(map[int64][]types.ThinSubjectGroup, len(pks))

	changedTimestamps, err := changeList.FetchList(r.changeListKey)
	if err != nil {
		log.WithError(err).Errorf("[%s] batchFetchActionExpressionChangedList fail, will re-fetch all pks=`%v`",
			MemoryLayer, pks)
		// 全部重查, 不重查可能有脏数据
		missSubjectPKs = pks
	} else {
		for _, subjectPK := range pks {
			key := r.genKey(subjectPK)
			value, found := cacheimpls.LocalSubjectGroupsCache.Get(key)
			if !found {
				missSubjectPKs = append(missSubjectPKs, subjectPK)

				continue
			}

			cached, ok := value.(*cachedSubjectGroups)
			if !ok {
				log.Errorf("[%s] parse cachedExpression in memory cache fail, will do retrieve!", MemoryLayer)
				missSubjectPKs = append(missSubjectPKs, subjectPK)

				continue
			}

			// 如果 5min内有更新, 那么这个算missing
			if changedTS, ok := changedTimestamps[key]; ok {
				if cached.timestamp < changedTS {
					// not the newest
					// 1. append to missing
					missSubjectPKs = append(missSubjectPKs, subjectPK)
					// 2. delete from local. NOTE: you don't need to do here, setMissing will overwrite the cache key-value
					continue
				}
			}

			hitSubjectGroups[subjectPK] = cached.subjectGroups
		}
	}

	if len(missSubjectPKs) > 0 {
		retrievedSubjectGroups, missingPKs, err := r.missingRetrieveFunc(missSubjectPKs)
		if err != nil {
			return nil, nil, err
		}
		// set missing into cache
		r.setMissing(retrievedSubjectGroups, missingPKs)
		// append the retrieved
		for subjectPK, sgs := range retrievedSubjectGroups {
			hitSubjectGroups[subjectPK] = sgs
		}

		return hitSubjectGroups, missingPKs, nil
	}

	return hitSubjectGroups, nil, nil
}

// nolint:unparam
func (r *memoryRetriever) setMissing(subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64) error {
	// set into local cache
	nowTimestamp := time.Now().Unix()
	for subjectPK, sgs := range subjectGroups {
		key := r.genKey(subjectPK)

		cacheimpls.LocalSubjectGroupsCache.Set(key, &cachedSubjectGroups{
			timestamp:     nowTimestamp,
			subjectGroups: sgs,
		}, 0)
	}
	for _, pk := range missingPKs {
		key := r.genKey(pk)

		cacheimpls.LocalSubjectGroupsCache.Set(
			key,
			&cachedSubjectGroups{
				timestamp:     nowTimestamp,
				subjectGroups: []types.ThinSubjectGroup{},
			},
			subjectGroupsLocalCacheTTL*time.Second,
		)
	}

	return nil
}
