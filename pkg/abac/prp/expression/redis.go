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
	"math/rand"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/conv"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cache/redis"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

const RedisLayer = "ExpressionRedisLayer"

const RandExpireSeconds = 60

type redisRetriever struct {
	missingRetrieveFunc MissingRetrieveFunc
}

func newRedisRetriever(retrieveFunc MissingRetrieveFunc) *redisRetriever {
	return &redisRetriever{
		missingRetrieveFunc: retrieveFunc,
	}
}

func (r *redisRetriever) retrieve(pks []int64) ([]types.AuthExpression, []int64, error) {
	hitExpressions, missExpressionPKs, err := r.batchGet(pks)
	// 1. if retrieve from redis fail, will fall through to retrieve from database
	if err != nil {
		log.WithError(err).Errorf("[%s] batchGet fail expressionPKs=`%+v`, will fallthrough to database",
			RedisLayer, pks)
		missExpressionPKs = pks
		hitExpressions = nil
	}

	expressions := make([]types.AuthExpression, 0, len(pks))
	emptyExpressionPKs := make([]int64, 0, len(pks))
	for exprPK, exprStr := range hitExpressions {
		var expression types.AuthExpression
		err = cacheimpls.ExpressionCache.Unmarshal(conv.StringToBytes(exprStr), &expression)
		if err != nil {
			log.WithError(err).Errorf("[%s] parse string to expression fail expressionPKs=`%+v`",
				RedisLayer, pks)

			// NOTE: 一条解析失败, 重新查/重新设置缓存
			missExpressionPKs = append(missExpressionPKs, exprPK)

			continue
		}

		// check here
		if expression.IsEmpty() {
			emptyExpressionPKs = append(emptyExpressionPKs, exprPK)

			continue
		}
		expressions = append(expressions, expression)
	}

	if len(missExpressionPKs) == 0 {
		return expressions, emptyExpressionPKs, nil
	}

	retrievedExpressions, missingPKs, err := r.missingRetrieveFunc(missExpressionPKs)
	if err != nil {
		return nil, nil, err
	}
	// set missing into cache
	r.setMissing(retrievedExpressions, missingPKs)
	// append the retrieved
	expressions = append(expressions, retrievedExpressions...)

	// NOTE: append the emptyExpressionPKs into missingPKs, the upper cache layer should cache that too.
	missingPKs = append(missingPKs, emptyExpressionPKs...)
	return expressions, missingPKs, nil
}

func (r *redisRetriever) setMissing(expressions []types.AuthExpression, missingPKs []int64) error {
	groupedExpressions := map[int64]types.AuthExpression{}
	for _, expression := range expressions {
		groupedExpressions[expression.PK] = expression
	}
	for _, pk := range missingPKs {
		groupedExpressions[pk] = types.AuthExpression{}
	}

	// set into the cache with pipeline
	return r.batchSet(groupedExpressions)
}

func (r *redisRetriever) batchGet(expressionPKs []int64) (
	hitExpressions map[int64]string,
	missExpressionPKs []int64,
	err error,
) {
	keys := make([]cache.Key, 0, len(expressionPKs))
	for _, i := range expressionPKs {
		keys = append(keys, cache.NewInt64Key(i))
	}

	hitStrings, err := cacheimpls.ExpressionCache.BatchGet(keys)
	if err != nil {
		return
	}

	hitExpressions = make(map[int64]string, len(hitStrings))
	for _, i := range expressionPKs {
		key := cache.NewInt64Key(i)
		if _, ok := hitStrings[key]; ok {
			hitExpressions[i] = hitStrings[key]
		} else {
			missExpressionPKs = append(missExpressionPKs, i)
		}
	}
	return
}

func (r *redisRetriever) batchSet(authExpressions map[int64]types.AuthExpression) error {
	// set into cache
	kvs := make([]redis.KV, 0, len(authExpressions))
	for pk, expression := range authExpressions {
		key := cache.NewInt64Key(pk)

		exprBytes, err := cacheimpls.ExpressionCache.Marshal(expression)
		if err != nil {
			return err
		}

		kvs = append(kvs, redis.KV{
			Key:   key.Key(),
			Value: conv.BytesToString(exprBytes),
		})
	}

	// keep cache for 7 days
	err := cacheimpls.ExpressionCache.BatchSetWithTx(
		kvs,
		cacheimpls.PolicyCacheExpiration+time.Duration(rand.Intn(RandExpireSeconds))*time.Second,
	)
	if err != nil {
		log.WithError(err).Errorf("[%s] cacheimpls.ExpressionCache.BatchSetWithTx fail kvs=`%+v`", RedisLayer, kvs)
		return err
	}

	return nil
}

func (r *redisRetriever) batchDelete(expressionPKs []int64) error {
	if len(expressionPKs) == 0 {
		return nil
	}

	keys := make([]cache.Key, 0, len(expressionPKs))
	for _, pk := range expressionPKs {
		keys = append(keys, cache.NewInt64Key(pk))
	}

	err := cacheimpls.ExpressionCache.BatchDelete(keys)
	if err != nil {
		log.WithError(err).Errorf("[%s] cacheimpls.ExpressionCache.BatchDelete fail keys=`%+v`", RedisLayer, keys)

		// report to sentry
		util.ReportToSentry("redis cache: expression cache delete fail",
			map[string]interface{}{
				"expressionPKs": expressionPKs,
				"keys":          keys,
				"error":         err.Error(),
			})

		return err
	}

	return nil
}

func batchDeleteExpressionsFromRedis(updatedActionPKExpressionPKs map[int64][]int64) error {
	if len(updatedActionPKExpressionPKs) == 0 {
		return nil
	}

	expressionPKs := make([]int64, 0, len(updatedActionPKExpressionPKs))
	for _, ePKs := range updatedActionPKExpressionPKs {
		expressionPKs = append(expressionPKs, ePKs...)
	}

	r := &redisRetriever{}
	return r.batchDelete(expressionPKs)
}
