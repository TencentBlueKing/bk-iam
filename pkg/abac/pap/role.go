/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pap

import (
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cacheimpls"
	"iam/pkg/service"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// RoleCTL ...
const RoleCTL = "RoleCTL"

type RoleController interface {
	ListSubjectByRole(roleType, system string) ([]Subject, error)
	BulkCreateSubjectRoles(roleType, system string, subjects []Subject) error
	BulkDeleteSubjectRoles(roleType, system string, subjects []Subject) error
}

type roleController struct {
	service service.RoleService

	subjectService service.SubjectService
}

func NewRoleController() RoleController {
	return &roleController{
		service: service.NewRoleService(),

		subjectService: service.NewSubjectService(),
	}
}

// BulkCreateSubjectRoles ...
func (c *roleController) BulkCreateSubjectRoles(roleType, system string, subjects []Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RoleCTL, "BulkCreateSubjectRoles")

	svcSubjects := convertToServiceSubjects(subjects)
	subjectPKs, err := c.subjectService.ListPKsBySubjects(svcSubjects)
	if err != nil {
		return errorWrapf(err, "subjectService.ListPKsBySubjects subjects=`%+v` fail", svcSubjects)
	}

	err = c.service.BulkCreateSubjectRoles(roleType, system, subjectPKs)
	if err != nil {
		return errorWrapf(
			err, "service.BulkCreateSubjectRoles roleType=`%s` system=`%s` subjectPKs=`%+v` fail",
			roleType, system, subjectPKs,
		)
	}

	// clean cache
	for _, subject := range subjects {
		cacheimpls.DeleteSubjectRoleSystemID(subject.Type, subject.ID)
	}

	return nil
}

// BulkDeleteSubjectRoles ...
func (c *roleController) BulkDeleteSubjectRoles(roleType, system string, subjects []Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RoleCTL, "BulkDeleteSubjectRoles")

	svcSubjects := convertToServiceSubjects(subjects)
	subjectPKs, err := c.subjectService.ListPKsBySubjects(svcSubjects)
	if err != nil {
		return errorWrapf(err, "subjectService.ListPKsBySubjects subjects=`%+v` fail", svcSubjects)
	}

	err = c.service.BulkDeleteSubjectRoles(roleType, system, subjectPKs)
	if err != nil {
		return errorWrapf(
			err, "service.BulkDeleteSubjectRoles roleType=`%s` system=`%s` subjectPKs=`%+v` fail",
			roleType, system, subjectPKs,
		)
	}

	// clean cache
	for _, subject := range subjects {
		cacheimpls.DeleteSubjectRoleSystemID(subject.Type, subject.ID)
	}

	return nil
}

// ListSubjectByRole ...
func (c *roleController) ListSubjectByRole(roleType, system string) ([]Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(RoleCTL, "ListSubjectByRole")

	subjectPKs, err := c.service.ListSubjectPKByRole(roleType, system)
	if err != nil {
		return nil, errorWrapf(err, "service.ListSubjectByRole roleType=`%s` system=`%s` fail", roleType, system)
	}

	subjects := make([]Subject, 0, len(subjectPKs))
	for _, subjectPK := range subjectPKs {
		svcSubject, err := cacheimpls.GetSubjectByPK(subjectPK)
		if err != nil {
			log.WithError(err).Warningf("cacheimpls.GetSubjectByPK fail subjectPK=`%d`", subjectPK)
		}
		subjects = append(subjects, Subject{
			Type: svcSubject.Type,
			ID:   svcSubject.ID,
			Name: svcSubject.Name,
		})
	}

	return subjects, nil
}
