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

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database/dao"
)

// RoleSVC ...
const RoleSVC = "RoleSVC"

// RoleService ...
type RoleService interface {
	// 鉴权
	ListRoleSystemIDBySubjectPK(pk int64) ([]string, error) // cache subject role system

	// web api
	ListSubjectPKByRole(roleType, system string) ([]int64, error)
	BulkCreateSubjectRoles(roleType, system string, subjectPKs []int64) error
	BulkDeleteSubjectRoles(roleType, system string, subjectPKs []int64) error
}

type roleService struct {
	manager dao.SubjectRoleManager
}

// NewRoleService ...
func NewRoleService() RoleService {
	return &roleService{
		manager: dao.NewSubjectRoleManager(),
	}
}

// ListRoleSystemIDBySubjectPK ...
func (l *roleService) ListRoleSystemIDBySubjectPK(pk int64) ([]string, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RoleSVC, "ListRoleSystemIDBySubjectPK")

	systemIDs, err := l.manager.ListSystemIDBySubjectPK(pk)
	if err != nil {
		return nil, errorWrapf(err, "roleManager.ListSystemIDBySubjectPK pk=`%d` fail", pk)
	}

	return systemIDs, err
}

// ListSubjectPKByRole ...
func (l *roleService) ListSubjectPKByRole(roleType, system string) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RoleSVC, "ListSubjectPKByRole")
	subjectPKs, err := l.manager.ListSubjectPKByRole(roleType, system)
	if err != nil {
		err = errorWrapf(
			err, "roleManager.ListSubjectPKByRole roleType=`%s`, system=`%s` fail", roleType, system,
		)
		return subjectPKs, err
	}
	return subjectPKs, err
}

// BulkCreateSubjectRoles ...
func (l *roleService) BulkCreateSubjectRoles(roleType, system string, subjectPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RoleSVC, "BulkCreateSubjectRoles")

	// 查询角色已有的subjectPK
	oldSubjectPKs, err := l.manager.ListSubjectPKByRole(roleType, system)
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
	err = l.manager.BulkCreate(roles)
	if err != nil {
		err = errorWrapf(err, "roleManager.BulkCreate roles=`%+v` fail", roles)
		return err
	}

	return nil
}

// BulkDeleteSubjectRoles ...
func (l *roleService) BulkDeleteSubjectRoles(roleType, system string, subjectPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RoleSVC, "BulkDeleteSubjectRoles")

	if len(subjectPKs) == 0 {
		return nil
	}

	// 创建SubjectRole
	err := l.manager.BulkDelete(roleType, system, subjectPKs)
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
