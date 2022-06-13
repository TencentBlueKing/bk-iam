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
	gocache "github.com/wklken/go-cache"
	"go.uber.org/multierr"

	"iam/pkg/abac/prp/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

const (
	memoryLayer = "GroupMemoryLayer"

	changeListTypeGroupAuthType = "group_auth_type"
	// local cache ttl should equals to changeList get by score (max=now, min=max-localCacheTTL)
	groupAuthTypeLocalCacheTTL = 60 * 10

	// fetch the top 1000 in change list, protect the auth/query api performance,
	// if hit 1000, some local-cached(system-action -> subjectPK) updated event will not be notified,
	// will be expired in policyLocalCacheTTL, accepted by now
	maxChangeListCount = 1000
)

var changeList = common.NewChangeList(changeListTypeGroupAuthType, groupAuthTypeLocalCacheTTL, maxChangeListCount)

type localCacheGroupAuthTypeRetriever struct {
	systemID         string
	cache            *gocache.Cache
	missingRetriever GroupAuthTypeRetriever
}

func NewLocalCacheGroupAuthTypeRetriever(
	systemID string,
	missingRetriever GroupAuthTypeRetriever,
) GroupAuthTypeRetriever {
	return &localCacheGroupAuthTypeRetriever{
		systemID:         systemID,
		cache:            cacheimpls.LocalGroupSystemAuthTypeCache,
		missingRetriever: missingRetriever,
	}
}

func (r *localCacheGroupAuthTypeRetriever) Retrieve(
	groupPKs []int64,
) (groupAuthTypes []types.GroupAuthType, err error) {
	groupAuthTypes, missPKs := r.batchGetGroupAuthType(groupPKs)

	// fetchMissPKs
	if len(missPKs) > 0 {
		// 设置默认未授权的group authType 0,
		missGroupAuthTypes, err := r.missingRetriever.Retrieve(missPKs)
		if err != nil {
			return nil, err
		}

		// setMissing
		r.batchSetGroupAuthTypeCache(missGroupAuthTypes)

		groupAuthTypes = append(groupAuthTypes, missGroupAuthTypes...)
	}

	return groupAuthTypes, nil
}

func (r *localCacheGroupAuthTypeRetriever) genKey(groupPK int64) string {
	return r.systemID + ":" + strconv.FormatInt(groupPK, 10)
}

func (r *localCacheGroupAuthTypeRetriever) batchGetGroupAuthType(
	groupPKs []int64,
) (groupAuthTypes []types.GroupAuthType, missPKs []int64) {
	groupAuthTypes = make([]types.GroupAuthType, 0, len(groupPKs))
	missPKs = make([]int64, 0, len(groupPKs))

	changedTimestamps, err := changeList.FetchList(r.systemID)
	if err != nil {
		log.WithError(err).Errorf("[%s] batchGetGroupAuthType fail, will re-fetch all groupPKs=`%v`",
			memoryLayer, groupPKs)
		// 全部重查, 不重查可能有脏数据
		return nil, groupPKs
	}

	// batchGet
	timestampNano := time.Now().UnixNano()
	for _, groupPK := range groupPKs {
		key := r.genKey(groupPK)

		item, ok := r.cache.GetItemAfterExpirationAnchor(key, timestampNano)
		if !ok {
			missPKs = append(missPKs, groupPK)
			continue
		}

		// 判断value的过期时间与changeList的过期时间是否大于
		groupPKStr := strconv.FormatInt(groupPK, 10)
		if changedTS, ok := changedTimestamps[groupPKStr]; ok {
			cacheSetTS := (item.Expiration / 1e9) - groupAuthTypeLocalCacheTTL
			if cacheSetTS < changedTS {
				// not the newest
				// 1. append to missing
				missPKs = append(missPKs, groupPK)
				// 2. delete from local. NOTE: you don't need to do here, setMissing will overwrite the cache key-value
				continue
			}
		}

		authType, ok := item.Object.(int64)
		if !ok {
			missPKs = append(missPKs, groupPK)
			continue
		}

		groupAuthTypes = append(groupAuthTypes, types.GroupAuthType{
			GroupPK:  groupPK,
			AuthType: authType,
		})
	}

	return groupAuthTypes, missPKs
}

func (r *localCacheGroupAuthTypeRetriever) batchSetGroupAuthTypeCache(groupAuthTypes []types.GroupAuthType) {
	for _, groupAuthType := range groupAuthTypes {
		key := r.genKey(groupAuthType.GroupPK)
		r.cache.Set(key, groupAuthType.AuthType, groupAuthTypeLocalCacheTTL*time.Second)
	}
}

func DeleteGroupAuthTypeCache(systemID string, groupPK int64) error {
	groupPKStr := strconv.FormatInt(groupPK, 10)
	key := systemID + ":" + groupPKStr
	cacheimpls.LocalGroupSystemAuthTypeCache.Delete(key)

	keyMembers := map[string][]string{
		systemID: {groupPKStr},
	}

	err := multierr.Combine(
		changeList.AddToChangeList(keyMembers),
		changeList.Truncate([]string{systemID}),
	)

	return err
}
