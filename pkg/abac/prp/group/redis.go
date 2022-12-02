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
	"strconv"
	"strings"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

const RedisLayer = "GroupRedisLayer"

const RandExpireSeconds = 60

type groupAuthTypeRedisRetriever struct {
	systemID         string
	missingRetriever GroupAuthTypeRetriever
	keyPrefix        string
}

func NewGroupAuthTypeRedisRetriever(
	systemID string,
	missingRetriever GroupAuthTypeRetriever,
) GroupAuthTypeRetriever {
	return &groupAuthTypeRedisRetriever{
		systemID:         systemID,
		missingRetriever: missingRetriever,
		keyPrefix:        systemID + ":",
	}
}

func (r *groupAuthTypeRedisRetriever) genKey(groupPK int64) cache.Key {
	return cache.NewStringKey(r.keyPrefix + strconv.FormatInt(groupPK, 10))
}

func (r *groupAuthTypeRedisRetriever) parseKey(key string) (groupPK int64, err error) {
	groupPKStr := strings.TrimPrefix(key, r.keyPrefix)

	groupPK, err = strconv.ParseInt(groupPKStr, 10, 64)
	if err != nil {
		log.WithError(err).Errorf("[%s] parseKey fail key=`%s`, keyPrefix=`%s`",
			RedisLayer, key, r.keyPrefix)
		return -1, err
	}
	return groupPK, nil
}

func (r *groupAuthTypeRedisRetriever) Retrieve(
	groupPKs []int64,
) (groupAuthTypes []types.GroupAuthType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RedisLayer, "Retrieve")
	groupAuthTypes, missPKs, err := r.batchGetGroupAuthType(groupPKs)
	if err != nil {
		return nil, errorWrapf(err, "batchGetGroupAuthType fail groupPKs=`%+v`", groupPKs)
	}

	if len(missPKs) > 0 {
		// get the missing groupAuthTypes
		missGroupAuthTypes, err := r.missingRetriever.Retrieve(missPKs)
		if err != nil {
			return nil, errorWrapf(err, "missingRetriever.Retrieve fail missPKs=`%+v`", missPKs)
		}

		err = r.batchSetGroupAuthTypeCache(missGroupAuthTypes)
		if err != nil {
			return nil, errorWrapf(err, "batchSetGroupAuthTypeCache fail missGroupAuthTypes=`%+v`", missGroupAuthTypes)
		}

		// merge the missGroupAuthTypes to groupAuthTypes
		groupAuthTypes = append(groupAuthTypes, missGroupAuthTypes...)
	}

	return groupAuthTypes, nil
}

func (r *groupAuthTypeRedisRetriever) batchGetGroupAuthType(
	groupPKs []int64,
) (groupAuthTypes []types.GroupAuthType, missPKs []int64, err error) {
	// build for batch get
	keys := make([]cache.Key, 0, len(groupPKs))
	for _, groupPK := range groupPKs {
		keys = append(keys, r.genKey(groupPK))
	}

	// get in pipeline
	hitValues, err := cacheimpls.GroupSystemAuthTypeCache.BatchGet(keys)
	if err != nil {
		// redis 查询失败, 只记日志, fallback 到 db 查询
		log.WithError(err).Errorf("[%s] cacheimpls.GroupSystemAuthTypeCache.BatchGet keys=`%+v` fail", RedisLayer, keys)
		return nil, groupPKs, nil
	}

	// the key can identify the hit or miss, here we only need system + subjectPK
	groupAuthTypes = make([]types.GroupAuthType, 0, len(hitValues))
	hitPKs := set.NewInt64Set()
	for hkf, value := range hitValues {
		groupPK, err := r.parseKey(hkf.Key())
		if err != nil {
			// skip the hit, if key parse fail
			continue
		}

		authType, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			// skip the hit, if value parse fail
			continue
		}

		groupAuthTypes = append(groupAuthTypes, types.GroupAuthType{
			GroupPK:  groupPK,
			AuthType: authType,
		})
		hitPKs.Add(groupPK)
	}

	// the missing groupPKs
	for _, groupPK := range groupPKs {
		if !hitPKs.Has(groupPK) {
			missPKs = append(missPKs, groupPK)
		}
	}

	return groupAuthTypes, missPKs, nil
}

func (r *groupAuthTypeRedisRetriever) batchSetGroupAuthTypeCache(groupAuthTypes []types.GroupAuthType) error {
	cacheKvs := make([]redis.KV, 0, len(groupAuthTypes))
	// batch set the subject_groups at one time
	for _, groupAuthType := range groupAuthTypes {
		key := r.genKey(groupAuthType.GroupPK)

		cacheKvs = append(cacheKvs, redis.KV{
			Key:   key.Key(),
			Value: strconv.FormatInt(groupAuthType.AuthType, 10),
		})
	}

	err := cacheimpls.GroupSystemAuthTypeCache.BatchSetWithTx(
		cacheKvs,
		cacheimpls.GroupSystemAuthTypeCacheExpiration+time.Duration(rand.Intn(RandExpireSeconds))*time.Second,
	)
	if err != nil {
		// 缓存设置失败, 不影响正常鉴权, 只记日志
		log.WithError(err).Errorf("[%s] cacheimpls.GroupSystemAuthTypeCache.BatchSetWithTx keys=`%+v` fail", RedisLayer, cacheKvs)
	}

	return nil
}

func deleteRedisGroupAuthTypeCache(systemID string, groupPK int64) error {
	key := cache.NewStringKey(systemID + ":" + strconv.FormatInt(groupPK, 10))
	return cacheimpls.GroupSystemAuthTypeCache.Delete(key)
}
