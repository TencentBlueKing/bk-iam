/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package impls

import (
	"database/sql"
	"errors"

	"iam/pkg/cache"
	"iam/pkg/errorx"
	"iam/pkg/service"
)

/*
 * > 当前在 SaaS 上勾选后, 某个人就拥有某个系统的所有权限; 这个变更次数很少, 但是变成了权限链路上的关键路径; 并且, 大多数人是不会命中的
 *
 * 处理: 变成local-cache, 1分钟生效
 *
 * 1. `无 到 有`, 有 5s 的时间差, 没有权限
 * 2. `有 到 无/变更`, 缓存时间内, 对应的身份没有变
 *
 * 当前设置的缓存时间: 1min
 */

// SubjectRoleCacheKey ...
type SubjectRoleCacheKey struct {
	SubjectType string
	SubjectID   string
}

// Key ...
func (k SubjectRoleCacheKey) Key() string {
	return k.SubjectType + ":" + k.SubjectID
}

func retrieveSubjectRole(key cache.Key) (interface{}, error) {
	k := key.(SubjectRoleCacheKey)

	pk, err := GetSubjectPK(k.SubjectType, k.SubjectID)

	// 如果用户不存在, 表现为没有任何一个系统的特殊角色
	if errors.Is(err, sql.ErrNoRows) {
		return []string{}, nil
	}

	if err != nil {
		return nil, err
	}

	svc := service.NewSubjectService()
	return svc.ListRoleSystemIDBySubjectPK(pk)
}

// ListSubjectRoleSystemID ...
func ListSubjectRoleSystemID(subjectType, subjectID string) (systemIDs []string, err error) {
	key := SubjectRoleCacheKey{
		SubjectType: subjectType,
		SubjectID:   subjectID,
	}

	var value interface{}
	value, err = LocalSubjectRoleCache.Get(key)
	if err != nil {
		err = errorx.Wrapf(err, CacheLayer, "ListSubjectRoleSystemID",
			"LocalSubjectRoleCache.Get subjectType=`%s`, subjectID=`%s` fail", subjectType, subjectID)
		return
	}

	var ok bool
	systemIDs, ok = value.([]string)
	if !ok {
		err = errors.New("not []string in cache")
		err = errorx.Wrapf(err, CacheLayer, "ListSubjectRoleSystemID",
			"LocalSubjectRoleCache.Get subjectType=`%s`, subjectID=`%s` fail", subjectType, subjectID)
		return
	}
	return systemIDs, nil
}

// DeleteSubjectRoleSystemID ...
func DeleteSubjectRoleSystemID(subjectType, subjectID string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "DeleteSubjectRole")

	key := SubjectRoleCacheKey{
		SubjectType: subjectType,
		SubjectID:   subjectID,
	}

	err := LocalSubjectRoleCache.Delete(key)
	if err != nil {
		err = errorWrapf(err, "LocalSubjectRoleCache.Delete key=%v", key)
	}
	return err
}
