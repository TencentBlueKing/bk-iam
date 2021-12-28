/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// ListSubjectPKByRole ...
func (l *subjectService) ListSubjectPKByRole(roleType, system string) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListSubjectPKByRole")
	subjectPKs, err := l.roleManager.ListSubjectPKByRole(roleType, system)
	if err != nil {
		err = errorWrapf(
			err, "roleManager.ListSubjectPKByRole roleType=`%s`, system=`%s` fail", roleType, system,
		)
		return subjectPKs, err
	}
	return subjectPKs, err
}

// BulkCreateSubjectRoles ...
func (l *subjectService) BulkCreateSubjectRoles(roleType, system string, subjects []types.Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkCreateSubjectRoles")

	// 查询用户的subjectPK
	subjectPKs, err := l.listSubjectPKs(subjects)
	if err != nil {
		err = errorWrapf(err, "listSubjectPKs subjects=`%+v` fail", subjects)
		return err
	}

	// 查询角色已有的subjectPK
	oldSubjectPKs, err := l.roleManager.ListSubjectPKByRole(roleType, system)
	if err != nil {
		err = errorWrapf(err, "roleManager.ListSubjectPKByRole roleType=`%s`, system=`%s` fail", roleType, system)
		return err
	}

	// 对比出需要创建subjectPK
	oldPKs := set.NewInt64SetWithValues(oldSubjectPKs)

	roles := make([]dao.SubjectRole, 0, len(subjectPKs))

	for _, pk := range subjectPKs {
		if !oldPKs.Has(pk) {
			roles = append(roles, dao.SubjectRole{
				RoleType:  roleType,
				System:    system,
				SubjectPK: pk,
			})
		}
	}

	// 创建SubjectRole
	err = l.roleManager.BulkCreate(roles)
	if err != nil {
		err = errorWrapf(err, "roleManager.BulkCreate roles=`%+v` fail", roles)
		return err
	}

	return nil
}

// BulkDeleteSubjectRoles ...
func (l *subjectService) BulkDeleteSubjectRoles(roleType, system string, subjects []types.Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkDeleteSubjectRoles")

	// 查询用户的subjectPK
	subjectPKs, err := l.listSubjectPKs(subjects)
	if err != nil {
		err = errorWrapf(err, "listSubjectPKs subjects`%+v` fail", subjects)
		return err
	}

	if len(subjectPKs) == 0 {
		return nil
	}

	// 创建SubjectRole
	err = l.roleManager.BulkDelete(roleType, system, subjectPKs)
	if err != nil {
		err = errorWrapf(
			err,
			"roleManager.BulkDelete roleType=`%s`, system=`%s`, subjectPKs=`%+v` fail",
			roleType,
			system,
			subjectPKs,
		)
		return err
	}

	return nil
}

func (l *subjectService) listSubjectPKs(subjects []types.Subject) ([]int64, error) {
	// 查询用户的subjectPK
	subjectIDs := make([]string, 0, len(subjects))
	for _, s := range subjects {
		subjectIDs = append(subjectIDs, s.ID)
	}

	users, err := l.manager.ListByIDs(types.UserType, subjectIDs)
	if err != nil {
		return nil, err
	}

	subjectPKs := make([]int64, 0, len(subjectIDs))
	for _, u := range users {
		subjectPKs = append(subjectPKs, u.PK)
	}

	return subjectPKs, err
}
