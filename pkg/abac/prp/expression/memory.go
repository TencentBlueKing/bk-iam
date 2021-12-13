/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package expression

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"go.uber.org/multierr"

	"iam/pkg/abac/prp/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

const (
	MemoryLayer = "ExpressionMemoryLayer"

	changeListTypeExpression = "expression"
	// local cache ttl should equals to changeList get by score (max=now, min=max-localCacheTTL)
	expressionLocalCacheTTL = 60

	// fetch the top 1000 in change list, protect the auth/query api performance,
	// if hit 1000, some local-cached(action -> expressionPK) updated event will not be notified,
	// will be expired in policyLocalCacheTTL, accepted by now
	maxChangeListCount = 1000
)

// expression cache的失效规则:
// 1. 只有用户自定义的(template_id=0)的, 才会更新(通过alterPolicies)
// 2. 来自于模板的(template_id!=0), 不会更新, 只会新增和删除

var changeList = common.NewChangeList(changeListTypeExpression, expressionLocalCacheTTL, maxChangeListCount)

// TODO: 如何加入debug? 感知两层缓存+database的结果?

type memoryRetriever struct {
	actionPK            int64
	missingRetrieveFunc MissingRetrieveFunc

	changeListKey string
}

func newMemoryRetriever(actionPK int64, retrieveFunc MissingRetrieveFunc) *memoryRetriever {
	return &memoryRetriever{
		actionPK: actionPK,

		changeListKey:       strconv.FormatInt(actionPK, 10),
		missingRetrieveFunc: retrieveFunc,
	}
}

type cachedExpression struct {
	timestamp  int64
	expression types.AuthExpression
}

func (r *memoryRetriever) genKey(expressionPK int64) string {
	return strconv.FormatInt(expressionPK, 10)
}

func (r *memoryRetriever) retrieve(pks []int64) ([]types.AuthExpression, []int64, error) {
	missExpressionPKs := make([]int64, 0, len(pks))
	expressions := make([]types.AuthExpression, 0, len(pks))

	changedTimestamps, err := changeList.FetchList(r.changeListKey)
	if err != nil {
		log.WithError(err).Errorf("[%s] batchFetchActionExpressionChangedList fail, will re-fetch all pks=`%v`",
			MemoryLayer, pks)
		// 全部重查, 不重查可能有脏数据
		missExpressionPKs = pks
	} else {
		for _, expressionPK := range pks {
			key := r.genKey(expressionPK)
			value, found := cacheimpls.LocalExpressionCache.Get(key)
			if !found {
				missExpressionPKs = append(missExpressionPKs, expressionPK)
				continue
			}

			cached, ok := value.(*cachedExpression)
			if !ok {
				log.Errorf("[%s] parse cachedExpression in memory cache fail, will do retrieve!", MemoryLayer)
				missExpressionPKs = append(missExpressionPKs, expressionPK)
				continue
			}

			// 如果 5min内有更新, 那么这个算missing
			if changedTS, ok := changedTimestamps[key]; ok {
				if cached.timestamp < changedTS {
					// not the newest
					// 1. append to missing
					missExpressionPKs = append(missExpressionPKs, expressionPK)
					// 2. delete from local. NOTE: you don't need to do here, setMissing will overwrite the cache key-value
					continue
				}
			}

			// skip empty
			if cached.expression.IsEmpty() {
				continue
			}

			expressions = append(expressions, cached.expression)
		}
	}

	if len(missExpressionPKs) > 0 {
		retrievedExpressions, missingPKs, err := r.missingRetrieveFunc(missExpressionPKs)
		if err != nil {
			return nil, nil, err
		}
		// set missing into cache
		r.setMissing(retrievedExpressions, missingPKs)
		// append the retrieved
		expressions = append(expressions, retrievedExpressions...)

		return expressions, missingPKs, nil
	}

	return expressions, nil, nil
}

// nolint:unparam
func (r *memoryRetriever) setMissing(expressions []types.AuthExpression, missingPKs []int64) error {
	// set into local cache
	nowTimestamp := time.Now().Unix()
	for _, expr := range expressions {
		key := r.genKey(expr.PK)

		cacheimpls.LocalExpressionCache.Set(key, &cachedExpression{
			timestamp:  nowTimestamp,
			expression: expr,
		}, 0)
	}
	for _, pk := range missingPKs {
		key := r.genKey(pk)

		cacheimpls.LocalExpressionCache.Set(
			key,
			&cachedExpression{
				timestamp:  nowTimestamp,
				expression: types.AuthExpression{},
			},
			expressionLocalCacheTTL*time.Second,
		)
	}

	return nil
}

func batchDeleteExpressionsFromMemory(updatedActionPKExpressionPKs map[int64][]int64) error {
	if len(updatedActionPKExpressionPKs) == 0 {
		return nil
	}

	keyMembers := make(map[string][]string, len(updatedActionPKExpressionPKs))
	changeListKeys := make([]string, 0, len(updatedActionPKExpressionPKs))

	for actionPK, expressionPKs := range updatedActionPKExpressionPKs {
		members := make([]string, 0, len(expressionPKs))
		for _, expressionPK := range expressionPKs {
			members = append(members, strconv.FormatInt(expressionPK, 10))

			// delete from local cache
			cacheimpls.LocalExpressionCache.Delete(strconv.FormatInt(expressionPK, 10))
		}

		key := strconv.FormatInt(actionPK, 10)
		keyMembers[key] = members
		changeListKeys = append(changeListKeys, key)
	}

	err := multierr.Combine(
		changeList.AddToChangeList(keyMembers),
		changeList.Truncate(changeListKeys),
	)

	return err
}
