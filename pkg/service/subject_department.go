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
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"

	"iam/pkg/database/dao"
	"iam/pkg/errorx"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// GetSubjectDepartmentPKs ...
func (l *subjectService) GetSubjectDepartmentPKs(subjectPK int64) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "GetSubjectDepartment")
	departmentPKStr, err := l.departmentManager.Get(subjectPK)
	if err != nil {
		return nil, errorWrapf(err, "departmentManager.Get subjectPK=`%d` fail", subjectPK)
	}

	departmentPKs, err := util.StringToInt64Slice(departmentPKStr, ",")
	if err != nil {
		return nil, errorWrapf(err, "util.StringToInt64Slice s=`%s` fail", departmentPKStr)
	}
	return departmentPKs, nil
}

// BulkCreateSubjectDepartments 批量创建用户部门关系
func (l *subjectService) BulkCreateSubjectDepartments(subjectDepartments []types.SubjectDepartment) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkCreateSubjectDepartments")
	daoSubjectDepartments, err := l.convertSubjectDepartments(subjectDepartments)
	if err != nil {
		return errorWrapf(err, "convertSubjectDepartments subjectDepartments=`%+v` fail", subjectDepartments)
	}

	if len(daoSubjectDepartments) == 0 {
		return nil
	}

	err = l.departmentManager.BulkCreate(daoSubjectDepartments)
	if err != nil {
		return errorWrapf(err, "departmentManager.BulkCreate subjectDepartments=`%+v` fail", daoSubjectDepartments)
	}
	return nil
}

// BulkDeleteSubjectDepartments ...
func (l *subjectService) BulkDeleteSubjectDepartments(subjectIDs []string) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkDeleteSubjectDepartments")

	subjects := make([]types.Subject, 0, len(subjectIDs))
	for _, id := range subjectIDs {
		subjects = append(subjects, types.Subject{
			Type: types.UserType,
			ID:   id,
		})
	}

	pks, err := l.ListPKsBySubjects(subjects)
	if err != nil {
		return pks, errorWrapf(err, "ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	if len(pks) == 0 {
		return pks, nil
	}

	err = l.departmentManager.BulkDelete(pks)
	if err != nil {
		return pks, errorWrapf(err, "departmentManager.BulkDelete pks=`%+v` fail", pks)
	}
	return pks, err
}

// BulkUpdateSubjectDepartments ...
func (l *subjectService) BulkUpdateSubjectDepartments(subjectDepartments []types.SubjectDepartment) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkUpdateSubjectDepartments")
	daoSubjectDepartments, err := l.convertSubjectDepartments(subjectDepartments)
	if err != nil {
		return nil, errorWrapf(err, "convertSubjectDepartments subjectDepartments=`%+v` fail", subjectDepartments)
	}

	if len(daoSubjectDepartments) == 0 {
		return nil, nil
	}

	err = l.departmentManager.BulkUpdate(daoSubjectDepartments)
	if err != nil {
		return nil, errorWrapf(err, "departmentManager.BulkUpdate subjectDepartments=`%+v` fail", daoSubjectDepartments)
	}

	pks := make([]int64, 0, len(daoSubjectDepartments))
	for _, sd := range daoSubjectDepartments {
		pks = append(pks, sd.SubjectPK)
	}
	return pks, nil
}

// ListPagingSubjectDepartment ...
func (l *subjectService) ListPagingSubjectDepartment(limit, offset int64) ([]types.SubjectDepartment, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListPagingSubjectDepartment")
	daoSubjectDepartments, err := l.departmentManager.ListPaging(limit, offset)
	if err != nil {
		return nil, errorWrapf(err, "departmentManager.ListPaging limit=`%d`, offset=`%d` fail", limit, offset)
	}

	if len(daoSubjectDepartments) == 0 {
		return []types.SubjectDepartment{}, nil
	}

	pks := make([]int64, 0, len(daoSubjectDepartments)*5) // 预估每个人大概会有4个部门, 加上用户本身, 所以乘以5
	subjectPKDepartmentPKs := make(map[int64][]int64, len(daoSubjectDepartments))
	var departmentPKs []int64
	for _, sd := range daoSubjectDepartments {
		pks = append(pks, sd.SubjectPK)
		departmentPKs, err = util.StringToInt64Slice(sd.DepartmentPKs, ",")
		if err != nil {
			return nil, errorWrapf(err, "util.StringToInt64Slice s=`%s` fail", sd.DepartmentPKs)
		}
		subjectPKDepartmentPKs[sd.SubjectPK] = departmentPKs
		pks = append(pks, departmentPKs...)
	}

	subjects, err := l.manager.ListByPKs(pks)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListByPKs pks=`%v` fail", pks)
	}
	subjectMap := make(map[int64]dao.Subject, len(subjects))
	for _, s := range subjects {
		subjectMap[s.PK] = s
	}

	subjectDepartments := make([]types.SubjectDepartment, 0, len(subjectPKDepartmentPKs))
	for pk, departmentPKs := range subjectPKDepartmentPKs {
		subject, ok := subjectMap[pk]
		if !ok {
			err = fmt.Errorf("subject pk: `%d` not exists", pk)
			return nil, errorWrapf(err, "")
		}
		deptIDs := make([]string, 0, len(departmentPKs))
		for _, deptPK := range departmentPKs {
			department, ok := subjectMap[deptPK]
			if !ok {
				err = fmt.Errorf("department deptPK: `%d` not exists", deptPK)
				return nil, errorWrapf(err, "")
			}
			deptIDs = append(deptIDs, department.ID)
		}
		subjectDepartment := types.SubjectDepartment{
			SubjectID:     subject.ID,
			DepartmentIDs: deptIDs,
		}
		subjectDepartments = append(subjectDepartments, subjectDepartment)
	}
	return subjectDepartments, nil
}

// GetSubjectDepartmentCount ...
func (l *subjectService) GetSubjectDepartmentCount() (int64, error) {
	count, err := l.departmentManager.GetCount()
	if err != nil {
		return count, errorx.Wrapf(err, SubjectSVC, "GetSubjectDepartmentCount", "departmentManager.GetCount fail")
	}
	return count, err
}

func (l *subjectService) convertSubjectDepartments(
	subjectDepartments []types.SubjectDepartment) ([]dao.SubjectDepartment, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "convertSubjectDepartments")

	subjectIDs := make([]string, 0, len(subjectDepartments))
	departmentIDSet := set.NewStringSet()
	for _, sd := range subjectDepartments {
		subjectIDs = append(subjectIDs, sd.SubjectID)
		for _, did := range sd.DepartmentIDs {
			departmentIDSet.Add(did)
		}
	}

	departmentIDs := departmentIDSet.ToSlice()

	subjects, err := l.manager.ListByIDs(types.UserType, subjectIDs)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListByIDs type=`%s`, ids=`%+v` fail", types.UserType, subjectIDs)
	}
	subjectMap := convertSubjectsToMap(subjects)

	departments, err := l.manager.ListByIDs(types.DepartmentType, departmentIDs)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListByIDs type=`%s`, ids=`%+v` fail", types.DepartmentType, departmentIDs)
	}
	departmentMap := convertSubjectsToMap(departments)

	daoSubjectDepartment := make([]dao.SubjectDepartment, 0, len(subjects))
	for _, sd := range subjectDepartments {
		pk, ok := subjectMap.Get(types.UserType, sd.SubjectID)
		if !ok {
			continue
		}
		departmentPKs := make([]int64, 0, len(sd.DepartmentIDs))
		for _, did := range sd.DepartmentIDs {
			departmentPK, ok := departmentMap.Get(types.DepartmentType, did)
			if !ok {
				continue
			}
			departmentPKs = append(departmentPKs, departmentPK)
		}
		daoSubjectDepartment = append(daoSubjectDepartment, dao.SubjectDepartment{
			SubjectPK:     pk,
			DepartmentPKs: util.Int64SliceToString(departmentPKs, ","),
		})
	}
	return daoSubjectDepartment, nil
}

type subjectPKMap map[string]int64

// Get ...
func (m subjectPKMap) Get(_type, id string) (pk int64, ok bool) {
	key := fmt.Sprintf("%s:%s", _type, id)
	pk, ok = m[key]
	return
}

// Add ...
func (m subjectPKMap) Add(_type, id string, pk int64) {
	key := fmt.Sprintf("%s:%s", _type, id)
	m[key] = pk
}

func convertSubjectsToMap(subjects []dao.Subject) subjectPKMap {
	subjectMap := make(subjectPKMap, len(subjects))
	for _, s := range subjects {
		subjectMap.Add(s.Type, s.ID, s.PK)
	}
	return subjectMap
}
