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
	"database/sql"
	"errors"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// DepartmentCTL ...
const DepartmentCTL = "DepartmentCTL"

type DepartmentController interface {
	ListPaging(limit, offset int64) ([]SubjectDepartment, error)
	BulkCreate(subjectDepartments []SubjectDepartment) error
	BulkUpdate(subjectDepartments []SubjectDepartment) error
	BulkDelete(subjectIDs []string) error
}

type departmentController struct {
	service service.DepartmentService

	subjectService service.SubjectService
}

func NewDepartmentController() DepartmentController {
	return &departmentController{
		service: service.NewDepartmentService(),

		subjectService: service.NewSubjectService(),
	}
}

// ListPaging ...
func (c *departmentController) ListPaging(limit, offset int64) ([]SubjectDepartment, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentCTL, "ListPaging")
	svcSubjectDepartments, err := c.service.ListPaging(limit, offset)
	if err != nil {
		return nil, errorWrapf(err, "service.ListPaging limit=`%d` offset=`%d` fail", limit, offset)
	}

	pks := make([]int64, 0, len(svcSubjectDepartments)*5)
	for _, svcSubjectDepartment := range svcSubjectDepartments {
		pks = append(pks, svcSubjectDepartment.SubjectPK)
		pks = append(pks, svcSubjectDepartment.DepartmentPKs...)
	}

	subjects, err := cacheimpls.BatchGetSubjectByPKs(pks)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.BatchGetSubjectByPKs pks=`%v` fail", pks)
	}

	subjectMap := make(map[int64]types.Subject, len(pks))
	for _, subject := range subjects {
		subjectMap[subject.PK] = subject
	}

	subjectDepartments := make([]SubjectDepartment, 0, len(svcSubjectDepartments))
	for _, svcSubjectDepartment := range svcSubjectDepartments {
		if _, ok := subjectMap[svcSubjectDepartment.SubjectPK]; !ok {
			continue
		}

		subjectID := subjectMap[svcSubjectDepartment.SubjectPK].ID
		departmentIDs := make([]string, 0, len(svcSubjectDepartment.DepartmentPKs))
		for _, depPK := range svcSubjectDepartment.DepartmentPKs {
			if _, ok := subjectMap[depPK]; !ok {
				continue
			}

			departmentIDs = append(departmentIDs, subjectMap[depPK].ID)
		}

		subjectDepartments = append(subjectDepartments, SubjectDepartment{
			SubjectID:     subjectID,
			DepartmentIDs: departmentIDs,
		})
	}

	return subjectDepartments, nil
}

// BulkCreate ...
func (c *departmentController) BulkCreate(subjectDepartments []SubjectDepartment) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentCTL, "BulkCreate")
	serviceSubjectDepartments, err := convertToServiceSubjectDepartments(subjectDepartments)
	if err != nil {
		return errorWrapf(err, "convertToServiceSubjectDepartments subjectDepartments=`%+v` fail", subjectDepartments)
	}

	err = c.service.BulkCreate(serviceSubjectDepartments)
	if err != nil {
		return errorWrapf(err, "service.BulkCreate subjectDepartments=`%+v` fail", subjectDepartments)
	}

	return nil
}

// BulkUpdate ...
func (c *departmentController) BulkUpdate(subjectDepartments []SubjectDepartment) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentCTL, "BulkUpdate")
	serviceSubjectDepartments, err := convertToServiceSubjectDepartments(subjectDepartments)
	if err != nil {
		return errorWrapf(err, "convertToServiceSubjectDepartments subjectDepartments=`%+v` fail", subjectDepartments)
	}

	err = c.service.BulkUpdate(serviceSubjectDepartments)
	if err != nil {
		return errorWrapf(err, "service.BulkUpdate subjectDepartments=`%+v` fail", subjectDepartments)
	}

	subjectPKs := make([]int64, 0, len(serviceSubjectDepartments))
	for _, svcSubjectDepartment := range serviceSubjectDepartments {
		subjectPKs = append(subjectPKs, svcSubjectDepartment.SubjectPK)
	}

	// delete from cache
	cacheimpls.BatchDeleteSubjectDepartmentCache(subjectPKs)

	return nil
}

// BulkDelete ...
func (c *departmentController) BulkDelete(subjectIDs []string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentCTL, "BulkDelete")
	subjects := make([]types.Subject, 0, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		subjects = append(subjects, types.Subject{
			Type: types.UserType,
			ID:   subjectID,
		})
	}

	subjectPKs, err := c.subjectService.ListPKsBySubjects(subjects)
	if err != nil {
		return errorWrapf(err, "subjectService.ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	err = c.service.BulkDelete(subjectPKs)
	if err != nil {
		return errorWrapf(err, "service.BulkDelete subjectIDs=`%s` fail", subjectIDs)
	}

	// delete from cache
	cacheimpls.BatchDeleteSubjectDepartmentCache(subjectPKs)

	return nil
}

func convertToServiceSubjectDepartments(subjectDepartments []SubjectDepartment) ([]types.SubjectDepartment, error) {
	serviceSubjectDepartments := make([]types.SubjectDepartment, 0, len(subjectDepartments))
	for _, subjectDepartment := range subjectDepartments {
		subjectPK, err := cacheimpls.GetLocalSubjectPK(types.UserType, subjectDepartment.SubjectID)
		if err != nil {
			return nil, err
		}

		departmentPKs := make([]int64, 0, len(subjectDepartment.DepartmentIDs))
		for _, departmentID := range subjectDepartment.DepartmentIDs {
			departmentPK, err := cacheimpls.GetLocalSubjectPK(types.DepartmentType, departmentID)
			if err != nil {
				// 兼容不存在的情况
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}

				return nil, err
			}
			departmentPKs = append(departmentPKs, departmentPK)
		}

		serviceSubjectDepartments = append(serviceSubjectDepartments, types.SubjectDepartment{
			SubjectPK:     subjectPK,
			DepartmentPKs: departmentPKs,
		})
	}

	return serviceSubjectDepartments, nil
}
