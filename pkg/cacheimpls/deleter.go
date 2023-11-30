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
	"go.uber.org/multierr"
)

// NOTE: action
// handler/action.go => batchDeleteActions =>  BatchDeleteActionPK(systemID, actionIDs)
//                                         |=> BatchDeleteActionResourceTypes(systemID, actionIDs)
//                   => UpdateAction =>    =>  DeleteActionResourceType(systemID, actionID)
// systemID + actionID, batch support

type actionCacheDeleter struct{}

// Execute ...
func (d actionCacheDeleter) Execute(key cache.Key) (err error) {
	err = multierr.Combine(
		ActionPKCache.Delete(key),
		ActionDetailCache.Delete(key),
	)
	return
}

type actionListCacheDeleter struct{}

// Execute ...
func (d actionListCacheDeleter) Execute(key cache.Key) (err error) {
	err = multierr.Combine(
		ActionListCache.Delete(key),
	)
	return
}

// NOTE: resource_type
// handler/resource_type.go => UpdateResourceType => DeleteResourceType(systemID, resourceTypeID)
//                          => batchDeleteResourceTypes => BatchDeleteResourceTypeCache(systemID, resourceTypeIDs)
// systemID + resourceTypeID, batch

type resourceTypeCacheDeleter struct{}

// Execute ...
func (d resourceTypeCacheDeleter) Execute(key cache.Key) (err error) {
	err = multierr.Combine(
		ResourceTypeCache.Delete(key),
		ResourceTypePKCache.Delete(key),
		LocalResourceTypePKCache.Delete(key),
	)
	return
}

// NOTE: subject
// handler/subject.go => BatchDeleteSubjects  =>      for DeleteSubjectPK(s.Type, s.ID)
//                                           |=>          BatchDeleteSubjectGroups(pks)
//                                           |=>          BatchDeleteSubjectDepartments(pks)
//                    => DeleteGroupMembers =>      for DeleteSubjectGroup(pk)
//                    => BatchAddGroupMembers =>    for DeleteSubjectGroup(pk)
//                    => BatchDeleteSubjectDepartments => BatchDeleteSubjectDepartments(pks)
//                    => BatchUpdateSubjectDepartments => BatchDeleteSubjectDepartments(pks)
// subject => 一个subject更新, 批量刷掉其所有缓存, 不考虑范围?  Delete SubjectPK/SubjectGroups/SubjectDepartments, batch support

type subjectDepartmentCacheDeleter struct{}

// Execute ...
func (d subjectDepartmentCacheDeleter) Execute(key cache.Key) (err error) {
	return SubjectDepartmentCache.Delete(key)
}

type systemCacheDeleter struct{}

// Execute ...
func (d systemCacheDeleter) Execute(key cache.Key) (err error) {
	err = multierr.Combine(
		SystemCache.Delete(key),
		LocalSystemClientsCache.Delete(key),
	)
	return
}

// PolicyCacheDeleter ...
type PolicyCacheDeleter struct{}

// BatchDeleteActionCache ...
func BatchDeleteActionCache(systemID string, actionIDs []string) error {
	keys := make([]cache.Key, 0, len(actionIDs))
	for _, actionID := range actionIDs {
		key := ActionIDCacheKey{
			SystemID: systemID,
			ActionID: actionID,
		}
		keys = append(keys, key)
	}

	ActionCacheCleaner.BatchDelete(keys)
	return nil
}

// DeleteActionListCache ...
func DeleteActionListCache(systemID string) error {
	key := cache.NewStringKey(systemID)
	ActionListCacheCleaner.Delete(key)
	return nil
}

// BatchDeleteResourceTypeCache ...
func BatchDeleteResourceTypeCache(systemID string, resourceTypeIDs []string) error {
	keys := make([]cache.Key, 0, len(resourceTypeIDs))
	for _, resourceTypeID := range resourceTypeIDs {
		key := ResourceTypeCacheKey{
			SystemID:       systemID,
			ResourceTypeID: resourceTypeID,
		}
		keys = append(keys, key)
	}

	ResourceTypeCacheCleaner.BatchDelete(keys)
	return nil
}

// BatchDeleteSubjectDepartmentCache ...
func BatchDeleteSubjectDepartmentCache(pks []int64) error {
	keys := make([]cache.Key, 0, len(pks))
	for _, pk := range pks {
		key := SubjectPKCacheKey{
			PK: pk,
		}
		keys = append(keys, key)
	}

	SubjectDepartmentCacheCleaner.BatchDelete(keys)
	return nil
}

// DeleteSystemCache ...
func DeleteSystemCache(systemID string) error {
	key := cache.NewStringKey(systemID)
	SystemCacheCleaner.Delete(key)
	return nil
}
