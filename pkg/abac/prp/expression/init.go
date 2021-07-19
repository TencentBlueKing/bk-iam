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
	"go.uber.org/multierr"

	"iam/pkg/service/types"
)

// TODO: 目前不支持debug, 将会导致不知道 memory - redis - database的所有行为

// GetExpressionsFromCache will retrieve expression from cache, the order is memory->redis->database
func GetExpressionsFromCache(actionPK int64, expressionPKs []int64) ([]types.AuthExpression, error) {
	l3 := newDatabaseRetriever()

	l2 := newRedisRetriever(l3.retrieve)

	l1 := newMemoryRetriever(actionPK, l2.retrieve)

	// NOTE: the missingPKs maybe nil
	expressions, _, err := l1.retrieve(expressionPKs)
	return expressions, err
}

// BatchDeleteExpressionsFromCache will delete cache from memory and redis
func BatchDeleteExpressionsFromCache(updatedActionPKExpressionPKs map[int64][]int64) error {
	err := multierr.Combine(
		// delete from redis
		batchDeleteExpressionsFromRedis(updatedActionPKExpressionPKs),
		// delete from memory
		batchDeleteExpressionsFromMemory(updatedActionPKExpressionPKs),
	)
	return err
}

// DebugRawGetExpressionFromCache for the /api/v1/debug to get the raw data in cache
func DebugRawGetExpressionFromCache(expressionPKs []int64) ([]types.AuthExpression, []int64, error) {
	f := func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error) {
		return nil, pks, nil
	}

	l2 := newRedisRetriever(f)
	return l2.retrieve(expressionPKs)
}
