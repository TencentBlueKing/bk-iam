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
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service"
	"iam/pkg/service/types"
)

func retrieveSubjectDetail(key cache.Key) (interface{}, error) {
	k := key.(SubjectPKCacheKey)

	departmentSvc := service.NewDepartmentService()
	departments, err := departmentSvc.GetSubjectDepartmentPKs(k.PK)
	if err != nil {
		return nil, err
	}

	groupSvc := service.NewGroupService()
	// NOTE: 这里只获取当前有效的 subject-groups, 之后放入缓存; 使用的时候, 会再次过滤掉已过期的(入缓存时可能还没过期, 使用时过期)
	groups, err := groupSvc.GetEffectThinSubjectGroups(k.PK)
	if err != nil {
		return nil, err
	}

	thinSubjectGroups := make([]types.ThinSubjectGroup, 0, len(groups))
	for _, sg := range groups {
		thinSubjectGroups = append(thinSubjectGroups, types.ThinSubjectGroup{
			PK:              sg.PK,
			PolicyExpiredAt: sg.PolicyExpiredAt,
		})
	}

	// NOTE: you should not add new field in SubjectDetail, unless you know how to upgrade
	// 如果要加新成员, 必须变更cache名字, 防止从已有缓存数据拿不到对应的字段产生bug
	detail := &types.SubjectDetail{
		DepartmentPKs: departments,
		SubjectGroups: thinSubjectGroups,
	}

	return detail, nil
}

// GetSubjectDetail ...
func GetSubjectDetail(pk int64) (detail types.SubjectDetail, err error) {
	key := SubjectPKCacheKey{
		PK: pk,
	}

	err = SubjectDetailCache.GetInto(key, &detail, retrieveSubjectDetail)
	err = errorx.Wrapf(err, CacheLayer, "GetSubjectDetail",
		"SubjectDetailCache.Get key=`%s` fail", key.Key())
	return
}
