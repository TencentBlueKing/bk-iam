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

	// 查询groupPK
	groupPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s`", _type, id)
	}

	service := service.NewGroupService()
	// 1. 查询用户组所有的成员, 并移除已存在的成员
	members, err := service.ListGroupMember(groupPK)
	if err != nil {
		return errorWrapf(err, "service.ListGroupMember groupPK=`%d`", groupPK)
	}

	existsSubjectPKs := set.NewFixedLengthInt64Set(len(members))
	for _, member := range members {
		existsSubjectPKs.Add(member.SubjectPK)
	}

	// 待检查的subjects
	subjectMap := make(map[int64]pap.GroupMember, len(subjects))
	for _, subject := range subjects {
		subjectPK, err := cacheimpls.GetSubjectPK(subject.Type, subject.ID)
		if err != nil {
			return errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s`", subject.Type, subject.ID)
		}

		if existsSubjectPKs.Has(subjectPK) {
			continue
		}

		subjectMap[subjectPK] = subject
	}

	if len(subjectMap) == 0 {
		return nil
	}

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

		for subjectPK, subject := range subjectMap {
			// 查询subject 系统下 已授权的用户组数量
			// NOTE: 这里系统授权的用户组数量不是绝对准确, 可能有已过期的用户组不在这个数组中
			subjectGroups, err := cacheimpls.ListSystemSubjectEffectGroups(system, []int64{subjectPK})
			if err != nil {
				return errorWrapf(
					err,
					"cacheimpls.ListSystemSubjectEffectGroups system=`%s`, subject=`%v`",
					system,
					subject,
				)
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
