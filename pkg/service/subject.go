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
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// SubjectSVC ...
const SubjectSVC = "SubjectSVC"

// SubjectService subject加载器
type SubjectService interface {
	// 鉴权
	Get(pk int64) (types.Subject, error)
	GetPK(_type, id string) (int64, error)

	// web api
	GetCount(_type string) (int64, error)
	ListPaging(_type string, limit, offset int64) ([]types.Subject, error)
	ListPKsBySubjects(subjects []types.Subject) ([]int64, error)
	BulkCreate(subjects []types.Subject) error
	BulkUpdateName(subjects []types.Subject) error

	// for pap
	BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error
}

type subjectService struct {
	manager dao.SubjectManager
}

// NewSubjectService SubjectService工厂
func NewSubjectService() SubjectService {
	return &subjectService{
		manager: dao.NewSubjectManager(),
	}
}

// Get ...
func (l *subjectService) Get(pk int64) (subject types.Subject, err error) {
	var s dao.Subject
	s, err = l.manager.Get(pk)
	if err != nil {
		err = errorx.Wrapf(err, SubjectSVC, "Get", "Get pk=`%d` fail", pk)
		return
	}

	subject = types.Subject{
		Type: s.Type,
		ID:   s.ID,
		Name: s.Name,
	}
	return
}

// GetPK ...
func (l *subjectService) GetPK(_type, id string) (pk int64, err error) {
	pk, err = l.manager.GetPK(_type, id)
	if err != nil {
		return pk, errorx.Wrapf(err, SubjectSVC, "GetPK", "GetPK _type=`%s`, id=`%s` fail", _type, id)
	}
	return pk, err
}

// GetCount ...
func (l *subjectService) GetCount(_type string) (int64, error) {
	cnt, err := l.manager.GetCount(_type)
	if err != nil {
		err = errorx.Wrapf(err, SubjectSVC, "GetCount", "manager.GetCount _type=`%s` fail", _type)
		return 0, err
	}
	return cnt, nil
}

func convertToSubjects(daoSubjects []dao.Subject) []types.Subject {
	subjects := make([]types.Subject, 0, len(daoSubjects))
	for _, s := range daoSubjects {
		subjects = append(subjects, types.Subject{
			Type: s.Type,
			ID:   s.ID,
			Name: s.Name,
		})
	}
	return subjects
}

// ListPaging ...
func (l *subjectService) ListPaging(_type string, limit, offset int64) ([]types.Subject, error) {
	daoSubjects, err := l.manager.ListPaging(_type, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, SubjectSVC,
			"ListPaging", "manager.ListPaging _type=`%s`, limit=`%d`, offset=`%d`",
			_type, limit, offset)
	}

	subjects := convertToSubjects(daoSubjects)
	return subjects, nil
}

// BulkCreate ...
func (l *subjectService) BulkCreate(subjects []types.Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkCreate")

	daoSubjects := make([]dao.Subject, 0, len(subjects))
	for _, s := range subjects {
		daoSubjects = append(daoSubjects, dao.Subject{
			Type: s.Type,
			ID:   s.ID,
			Name: s.Name,
		})
	}

	err := l.manager.BulkCreate(daoSubjects)
	if err != nil {
		return errorWrapf(err, "manager.BulkCreate subjects=`%+v`", daoSubjects)
	}
	return err
}

func groupBySubjectType(subjects []types.Subject) (userIDs []string, departmentIDs []string, groupIDs []string) {
	// 分组获取Subject PK
	userIDs = make([]string, 0, len(subjects))
	departmentIDs = make([]string, 0, len(subjects))
	groupIDs = make([]string, 0, len(subjects))
	for _, s := range subjects {
		switch s.Type {
		case types.UserType:
			userIDs = append(userIDs, s.ID)
		case types.DepartmentType:
			departmentIDs = append(departmentIDs, s.ID)
		case types.GroupType:
			groupIDs = append(groupIDs, s.ID)
		}
	}
	return
}

// ListPKsBySubjects ...
func (l *subjectService) ListPKsBySubjects(subjects []types.Subject) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListPKsBySubjects")

	// 分组获取Subject PK
	userIDs, departmentIDs, groupIDs := groupBySubjectType(subjects)

	pks := []int64{}
	if len(userIDs) > 0 {
		users, newErr := l.manager.ListByIDs(types.UserType, userIDs)
		if newErr != nil {
			return nil, errorWrapf(newErr, "manager.ListByIDs _type=`%s`, ids=`%+v` fail", types.UserType, userIDs)
		}
		for _, u := range users {
			pks = append(pks, u.PK)
		}
	}
	if len(departmentIDs) > 0 {
		departments, newErr := l.manager.ListByIDs(types.DepartmentType, departmentIDs)
		if newErr != nil {
			return nil, errorWrapf(newErr, "manager.ListByIDs _type=`%s`, ids=`%+v` fail",
				types.DepartmentType, departmentIDs)
		}
		for _, d := range departments {
			pks = append(pks, d.PK)
		}
	}
	if len(groupIDs) > 0 {
		groups, newErr := l.manager.ListByIDs(types.GroupType, groupIDs)
		if newErr != nil {
			return nil, errorWrapf(newErr, "manager.ListByIDs _type=`%s`, ids=`%+v` fail", types.GroupType, groupIDs)
		}
		for _, g := range groups {
			pks = append(pks, g.PK)
		}
	}
	return pks, nil
}

// BulkDeleteByPKsWithTx ...
func (l *subjectService) BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkDeleteByPKsWithTx")

	err := l.manager.BulkDeleteByPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "manager.BulkDeleteByPKsWithTx pks=`%+v` fail", pks)
	}
	return err
}

// BulkUpdateName ...
func (l *subjectService) BulkUpdateName(subjects []types.Subject) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkUpdateName")

	daoSubjects := make([]dao.Subject, 0, len(subjects))
	for _, s := range subjects {
		daoSubjects = append(daoSubjects, dao.Subject{
			Type: s.Type,
			ID:   s.ID,
			Name: s.Name,
		})
	}

	err := l.manager.BulkUpdate(daoSubjects)
	if err != nil {
		return errorWrapf(err, "manager.BulkUpdate subjects=`%+v`", daoSubjects)
	}
	return err
}
