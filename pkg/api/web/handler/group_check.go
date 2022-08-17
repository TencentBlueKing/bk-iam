/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"errors"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/pap"
	"iam/pkg/api/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
)

var errQuota = errors.New("quota error")

// checkSubjectGroupsQuota 检查用户组的成员数量是否超过配额
func checkSubjectGroupsQuota(_type, id string, subjects []pap.GroupMember) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "checkSubjectGroupsQuota")

	// 1. 查询groupPK
	groupPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s`", _type, id)
	}

	service := service.NewGroupService()

	// 2. 查询group所授权的所有system
	systems, err := service.ListGroupAuthSystemIDs(groupPK)
	if err != nil {
		return errorWrapf(err, "service.ListGroupAuthSystemIDs groupPK=`%d`", groupPK)
	}

	if len(systems) == 0 {
		return nil
	}

	// 3. 遍历system, 查询subject已加入的group数量
	for _, system := range systems {
		limit := common.GetMaxSubjectGroupsLimit(system)

		for _, subject := range subjects {
			subjectPK, err := cacheimpls.GetLocalSubjectPK(subject.Type, subject.ID)
			if err != nil {
				return errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s`", subject.Type, subject.ID)
			}

			// 查询subject 系统下 已授权的用户组数量
			subjectGroups, err := cacheimpls.ListSystemSubjectEffectGroups(system, []int64{subjectPK})
			if err != nil {
				return errorWrapf(
					err,
					"cacheimpls.ListSystemSubjectEffectGroups system=`%s`, subject=`%v`",
					system,
					subject,
				)
			}

			groupPKSet := set.NewFixedLengthInt64Set(len(subjectGroups))
			for _, group := range subjectGroups {
				groupPKSet.Add(group.GroupPK)
			}

			// 已授权的用户组, 只更新过期时间, 数量不会增加, 不需要检查
			if groupPKSet.Has(groupPK) {
				continue
			}

			// 校验数量是否超限
			if len(subjectGroups)+1 > limit {
				return errorWrapf(
					errQuota,
					"subject %v can only have %d groups in system %s.[current %d]",
					subject,
					limit,
					system,
					len(subjectGroups),
				)
			}
		}
	}

	return nil
}
