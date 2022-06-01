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
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
	"iam/pkg/util"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"
)

// DepartmentSVC ...
const DepartmentSVC = "DepartmentSVC"

// DepartmentService ...
type DepartmentService interface {

	// 鉴权
	GetSubjectDepartmentPKs(subjectPK int64) ([]int64, error) // cache subject detail

	// web api
	GetCount() (int64, error)
	ListPaging(limit, offset int64) ([]types.SubjectDepartment, error)
	BulkCreate(subjectDepartments []types.SubjectDepartment) error
	BulkUpdate(subjectDepartments []types.SubjectDepartment) error
	BulkDelete(subjectPKs []int64) error

	// for pap
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, pks []int64) error
}

type departmentService struct {
	manager dao.SubjectDepartmentManager
}

// NewDepartmentService ...
func NewDepartmentService() DepartmentService {
	return &departmentService{
		manager: dao.NewSubjectDepartmentManager(),
	}
}

// GetSubjectDepartmentPKs ...
func (l *departmentService) GetSubjectDepartmentPKs(subjectPK int64) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentSVC, "GetSubjectDepartment")
	departmentPKStr, err := l.manager.Get(subjectPK)
	if err != nil {
		return nil, errorWrapf(err, "manager.Get subjectPK=`%d` fail", subjectPK)
	}

	departmentPKs, err := util.StringToInt64Slice(departmentPKStr, ",")
	if err != nil {
		return nil, errorWrapf(err, "util.StringToInt64Slice s=`%s` fail", departmentPKStr)
	}
	return departmentPKs, nil
}

// BulkCreate 批量创建用户部门关系
func (l *departmentService) BulkCreate(subjectDepartments []types.SubjectDepartment) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentSVC, "BulkCreate")
	daoSubjectDepartments := make([]dao.SubjectDepartment, 0, len(subjectDepartments))
	for _, subjectDepartment := range subjectDepartments {
		daoSubjectDepartment := dao.SubjectDepartment{
			SubjectPK:     subjectDepartment.SubjectPK,
			DepartmentPKs: util.Int64SliceToString(subjectDepartment.DepartmentPKs, ","),
		}
		daoSubjectDepartments = append(daoSubjectDepartments, daoSubjectDepartment)
	}

	if len(daoSubjectDepartments) == 0 {
		return nil
	}

	err := l.manager.BulkCreate(daoSubjectDepartments)
	if err != nil {
		return errorWrapf(err, "manager.BulkCreate subjectDepartments=`%+v` fail", daoSubjectDepartments)
	}
	return nil
}

// BulkDelete ...
func (l *departmentService) BulkDelete(subjectPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentSVC, "BulkDeleteSubjectDepartments")
	err := l.manager.BulkDelete(subjectPKs)
	if err != nil {
		return errorWrapf(err, "manager.BulkDelete pks=`%+v` fail", subjectPKs)
	}
	return err
}

// BulkUpdate ...
func (l *departmentService) BulkUpdate(subjectDepartments []types.SubjectDepartment) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentSVC, "BulkUpdate")
	daoSubjectDepartments := make([]dao.SubjectDepartment, 0, len(subjectDepartments))
	for _, subjectDepartment := range subjectDepartments {
		daoSubjectDepartment := dao.SubjectDepartment{
			SubjectPK:     subjectDepartment.SubjectPK,
			DepartmentPKs: util.Int64SliceToString(subjectDepartment.DepartmentPKs, ","),
		}
		daoSubjectDepartments = append(daoSubjectDepartments, daoSubjectDepartment)
	}

	if len(daoSubjectDepartments) == 0 {
		return nil
	}

	err := l.manager.BulkUpdate(daoSubjectDepartments)
	if err != nil {
		return errorWrapf(err, "manager.BulkUpdate subjectDepartments=`%+v` fail", daoSubjectDepartments)
	}
	return nil
}

// ListPaging ...
func (l *departmentService) ListPaging(limit, offset int64) ([]types.SubjectDepartment, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentSVC, "ListPaging")
	daoSubjectDepartments, err := l.manager.ListPaging(limit, offset)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListPaging limit=`%d`, offset=`%d` fail", limit, offset)
	}

	if len(daoSubjectDepartments) == 0 {
		return nil, nil
	}

	subjectDepartments := make([]types.SubjectDepartment, 0, len(daoSubjectDepartments))
	var departmentPKs []int64
	for _, sd := range daoSubjectDepartments {
		departmentPKs, err = util.StringToInt64Slice(sd.DepartmentPKs, ",")
		if err != nil {
			return nil, errorWrapf(err, "util.StringToInt64Slice s=`%s` fail", sd.DepartmentPKs)
		}
		subjectDepartments = append(subjectDepartments, types.SubjectDepartment{
			SubjectPK:     sd.SubjectPK,
			DepartmentPKs: departmentPKs,
		})
	}

	return subjectDepartments, nil
}

// GetCount ...
func (l *departmentService) GetCount() (int64, error) {
	count, err := l.manager.GetCount()
	if err != nil {
		return count, errorx.Wrapf(err, DepartmentSVC, "GetCount", "manager.GetCount fail")
	}
	return count, err
}

// BulkDeleteBySubjectPKs ...
func (l *departmentService) BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(DepartmentSVC, "BulkDeleteBySubjectPKs")
	// 对于用户，需要删除subject department
	err := l.manager.BulkDeleteWithTx(tx, pks)
	if err != nil {
		return errorWrapf(
			err, "manager.BulkDeleteWithTx subject_pks=`%+v` fail", pks)
	}
	return nil
}
