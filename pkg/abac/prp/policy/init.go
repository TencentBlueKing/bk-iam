/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"go.uber.org/multierr"

	"iam/pkg/service/types"
)

// GetPoliciesFromCache will retrieve policies from cache, the order is memory->redis->database
func GetPoliciesFromCache(system string, actionPK int64, subjectPKs []int64) ([]types.AuthPolicy, error) {
	l3 := newDatabaseRetriever(actionPK)

	l2 := newRedisRetriever(system, actionPK, l3.retrieve)

	l1 := newMemoryRetriever(system, actionPK, l2.retrieve)

	policies, _, err := l1.retrieve(subjectPKs)
	return policies, err
}

// DeleteSystemSubjectPKsFromCache will delete cache from memory and redis
func DeleteSystemSubjectPKsFromCache(system string, subjectPKs []int64) error {
	err := multierr.Combine(
		deleteSystemSubjectPKsFromRedis(system, subjectPKs),
		deleteSystemSubjectPKsFromMemory(system, subjectPKs),
	)

	return err
}

// BatchDeleteSystemSubjectPKsFromCache will delete cache from memory and redis
func BatchDeleteSystemSubjectPKsFromCache(systems []string, subjectPKs []int64) error {
	// NOTE: 两个调用点
	//  policy crud的可以拿到actionPK列表 => subjectPK  [-> 可以细化]
	//  delete subject 拿不到actionPK列表 => group subjectPKs

	err := multierr.Combine(
		batchDeleteSystemSubjectPKsFromRedis(systems, subjectPKs),
		batchDeleteSystemSubjectPKsFromMemory(systems, subjectPKs),
	)

	return err
}

// DebugRawGetPolicyFromCache ...
func DebugRawGetPolicyFromCache(system string, actionPK int64, subjectPKs []int64) (
	[]types.AuthPolicy,
	[]int64,
	error,
) {
	f := func(pks []int64) (policies []types.AuthPolicy, missingSubjectPKs []int64, err error) {
		return nil, pks, nil
	}
	l2 := newRedisRetriever(system, actionPK, f)
	return l2.retrieve(subjectPKs)
}
