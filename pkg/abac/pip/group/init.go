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
	"errors"

	"go.uber.org/multierr"

	"iam/pkg/service/types"
	svctypes "iam/pkg/service/types"
)

// currently, only support `department-groups`, so we add a condition here, incase someone use it in a wrong way
// NOTE: memory and redis cache all operations like set/get should only show in `group`

// TODO: 目前不支持debug, 将会导致不知道 memory - redis - database的所有行为
var ErrSubjectTypeNotSupported = errors.New("subject type not supported, current only support `department`")

// GetSubjectGroupsFromCache will retrieve subject groups from cache, the order is memory->redis->database
func GetSubjectGroupsFromCache(subjectType string, subjectPKs []int64) (map[int64][]types.ThinSubjectGroup, error) {
	// NOTE: if we modify here, should modify the BatchDeleteSubjectGroupsFromCache too
	if subjectType != svctypes.DepartmentType {
		return nil, ErrSubjectTypeNotSupported
	}

	l3 := newDatabaseRetriever()

	l2 := newRedisRetriever(l3.retrieve)

	l1 := newMemoryRetriever(subjectType, l2.retrieve)

	// NOTE: the missingPKs maybe nil
	subjectGroups, _, err := l1.retrieve(subjectPKs)
	return subjectGroups, err
}

// BatchDeleteSubjectGroupsFromCache will delete cache from memory and redis
func BatchDeleteSubjectGroupsFromCache(subjectType string, subjectPKs []int64) error {
	// NOTE: if we modify here, should modify the GetSubjectGroupsFromCache too
	if subjectType != svctypes.DepartmentType {
		return nil
	}

	err := multierr.Combine(
		// delete from redis
		batchDeleteSubjectGroupsFromRedis(subjectPKs),
		// delete from memory
		batchDeleteSubjectGroupsFromMemory(subjectType, subjectPKs),
	)
	return err
}
