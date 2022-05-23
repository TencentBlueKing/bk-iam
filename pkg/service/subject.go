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

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// SubjectSVC ...
const SubjectSVC = "SubjectSVC"

// SubjectService subject加载器
type SubjectService interface {
	// in this file
	// Subject

	Get(pk int64) (types.Subject, error)
	GetPK(_type, id string) (int64, error)
	GetCount(_type string) (int64, error)
	ListPaging(_type string, limit, offset int64) ([]types.Subject, error)
	ListPKsBySubjects(subjects []types.Subject) ([]int64, error)
	ListByPKs(pks []int64) ([]types.Subject, error)
	BulkCreate(subjects []types.Subject) error
	BulkDelete(subjects []types.Subject) ([]int64, error)
	BulkUpdateName(subjects []types.Subject) error

	// in subject_group.go

	GetEffectThinSubjectGroups(pk int64) ([]types.ThinSubjectGroup, error)
	ListEffectThinSubjectGroups(pks []int64) (map[int64][]types.ThinSubjectGroup, error)
	ListSubjectGroups(_type, id string, beforeExpiredAt int64) ([]types.SubjectGroup, error)

	// in subject_member.go
	// Member:

	GetMemberCount(_type, id string) (int64, error)
	GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error)
	ListPagingMember(_type, id string, limit, offset int64) ([]types.SubjectMember, error)
	ListPagingMemberBeforeExpiredAt(
		_type, id string, expiredAt int64, limit, offset int64,
	) ([]types.SubjectMember, error)
	ListExistSubjectsBeforeExpiredAt(subjects []types.Subject, expiredAt int64) ([]types.Subject, error)
	ListMember(_type, id string) ([]types.SubjectMember, error)
	UpdateMembersExpiredAt(members []types.SubjectMember) error
	BulkDeleteSubjectMembers(_type, id string, members []types.Subject) (map[string]int64, error)
	BulkCreateSubjectMembers(_type, id string, members []types.Subject, policyExpiredAt int64) error

	// in subject_department.go
	// Department

	GetSubjectDepartmentPKs(subjectPK int64) ([]int64, error)
	GetSubjectDepartmentCount() (int64, error)
	ListPagingSubjectDepartment(limit, offset int64) ([]types.SubjectDepartment, error)
	BulkCreateSubjectDepartments(subjectDepartments []types.SubjectDepartment) error
	BulkUpdateSubjectDepartments(subjectDepartments []types.SubjectDepartment) ([]int64, error)
	BulkDeleteSubjectDepartments(subjectIDs []string) ([]int64, error)

	// in subject_role.go
	// Role

	ListSubjectPKByRole(roleType, system string) ([]int64, error)
	ListRoleSystemIDBySubjectPK(pk int64) ([]string, error)
	BulkCreateSubjectRoles(roleType, system string, subjects []types.Subject) error
	BulkDeleteSubjectRoles(roleType, system string, subjects []types.Subject) error
}

type subjectService struct {
	manager           dao.SubjectManager
	policyManager     dao.PolicyManager
	expressionManager dao.ExpressionManager

	relationManager   dao.SubjectRelationManager
	departmentManager dao.SubjectDepartmentManager
	roleManager       dao.SubjectRoleManager

	groupSystemAuthTypeManager dao.GroupSystemAuthTypeManager
	subjectSystemGroupManager  dao.SubjectSystemGroupManager
}

// NewSubjectService SubjectService工厂
func NewSubjectService() SubjectService {
	return &subjectService{
		manager:           dao.NewSubjectManager(),
		policyManager:     dao.NewPolicyManager(),
		expressionManager: dao.NewExpressionManager(),

		relationManager:   dao.NewSubjectRelationManager(),
		departmentManager: dao.NewSubjectDepartmentManager(),
		roleManager:       dao.NewSubjectRoleManager(),

		groupSystemAuthTypeManager: dao.NewGroupSystemAuthTypeManager(),
		subjectSystemGroupManager:  dao.NewSubjectSystemGroupManager(),
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
	count, err := l.manager.GetCount(_type)
	if err != nil {
		err = errorx.Wrapf(err, SubjectSVC, "GetCount", "manager.GetCount _type=`%s` fail", _type)
		return 0, err
	}
	return count, nil
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

// ListByPKs ...
func (l *subjectService) ListByPKs(pks []int64) ([]types.Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListByPKs")

	daoSubjects, err := l.manager.ListByPKs(pks)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListByPKs pks=`%v` fail", pks)
	}
	subjects := convertToSubjects(daoSubjects)
	return subjects, nil
}

// BulkDelete ...
func (l *subjectService) BulkDelete(subjects []types.Subject) (pks []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkDelete")
	pks, err = l.ListPKsBySubjects(subjects)
	if err != nil {
		return pks, errorWrapf(err, "subjectService.ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	// 查询Policy里的Subject单独的Expression
	expressionPKs, err := l.policyManager.ListExpressionBySubjectsTemplate(pks, 0)
	if err != nil {
		return pks, errorWrapf(err, "policyManager.ListExpressionBySubjectsTemplate subjectPKs=`%+v` fail", pks)
	}

	// 按照PK删除Subject所有相关的
	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return pks, errorWrapf(err, "define tx error")
	}

	// 删除策略 policy
	err = l.policyManager.BulkDeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return pks, errorWrapf(
			err, "policyManager.BulkDeleteBySubjectPKsWithTx subject_pks=`%+v` fail", pks)
	}

	// 删除策略对应的非来着权限模板的Expression
	_, err = l.expressionManager.BulkDeleteByPKsWithTx(tx, expressionPKs)
	if err != nil {
		return pks, errorWrapf(
			err, "expressionManager.BulkDeleteByPKsWithTx pks=`%+v` fail", expressionPKs)
	}

	// 批量用户组删除成员关系 subjectRelation
	err = l.relationManager.BulkDeleteByParentPKs(tx, pks)
	if err != nil {
		return pks, errorWrapf(
			err, "relationManager.BulkDeleteByParentPKs parent_pks=`%+v` fail", pks)
	}
	// 批量其加入的用户组关系 subjectRelation
	err = l.relationManager.BulkDeleteBySubjectPKs(tx, pks)
	if err != nil {
		return pks, errorWrapf(
			err, "relationManager.BulkDeleteBySubjectPKs subject_pks=`%+v` fail", pks)
	}

	// 对于用户，需要删除subject department
	err = l.departmentManager.BulkDeleteWithTx(tx, pks)
	if err != nil {
		return pks, errorWrapf(
			err, "departmentManager.BulkDeleteWithTx subject_pks=`%+v` fail", pks)
	}

	// 删除对象 subject
	err = l.manager.BulkDeleteByPKsWithTx(tx, pks)
	if err != nil {
		return pks, errorWrapf(
			err, "manager.BulkDeleteByPKsWithTx pks=`%+v` fail", pks)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return pks, errorWrapf(err, "tx commit error")
	}
	return pks, err
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

// ListRoleSystemIDBySubjectPK ...
func (l *subjectService) ListRoleSystemIDBySubjectPK(pk int64) ([]string, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListRoleSystemIDBySubjectPK")

	systemIDs, err := l.roleManager.ListSystemIDBySubjectPK(pk)
	if err != nil {
		return nil, errorWrapf(err, "roleManager.ListSystemIDBySubjectPK pk=`%d` fail", pk)
	}

	return systemIDs, err
}
