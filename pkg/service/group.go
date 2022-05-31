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
	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// GroupSVC ...
const GroupSVC = "GroupSVC"

// GroupService ...
type GroupService interface {

	// 鉴权
	GetEffectThinSubjectGroups(pk int64) ([]types.ThinSubjectGroup, error)               // cache subject detail
	ListEffectThinSubjectGroups(pks []int64) (map[int64][]types.ThinSubjectGroup, error) // cache department groups

	// web api
	ListSubjectGroups(_type, id string, beforeExpiredAt int64) ([]types.SubjectGroup, error)
	GetMemberCount(_type, id string) (int64, error)
	GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error)
	ListPagingMember(_type, id string, limit, offset int64) ([]types.SubjectMember, error)
	ListPagingMemberBeforeExpiredAt(
		_type, id string, expiredAt int64, limit, offset int64,
	) ([]types.SubjectMember, error)
	ListExistSubjectsBeforeExpiredAt(subjects []types.Subject, expiredAt int64) ([]types.Subject, error)
	ListMember(_type, id string) ([]types.SubjectMember, error)

	UpdateMembersExpiredAtWithTx(tx *sqlx.Tx, members []types.SubjectRelationPKPolicyExpiredAt) error
	BulkDeleteSubjectMembers(_type, id string, members []types.Subject) (map[string]int64, error)
	BulkCreateSubjectMembersWithTx(tx *sqlx.Tx, relations []types.SubjectRelation) error

	// for pap
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, pks []int64) error
}

type groupService struct {
	manager dao.SubjectRelationManager
}

// NewGroupService GroupService工厂
func NewGroupService() GroupService {
	return &groupService{
		manager: dao.NewSubjectRelationManager(),
	}
}

// from subject_group.go

func convertToSubjectGroup(relation dao.SubjectRelation) types.SubjectGroup {
	return types.SubjectGroup{
		PK:              relation.ParentPK,
		Type:            relation.ParentType,
		ID:              relation.ParentID,
		PolicyExpiredAt: relation.PolicyExpiredAt,
		CreateAt:        relation.CreateAt,
	}
}

func convertToThinSubjectGroup(relation dao.ThinSubjectRelation) types.ThinSubjectGroup {
	return types.ThinSubjectGroup{
		PK:              relation.ParentPK,
		PolicyExpiredAt: relation.PolicyExpiredAt,
	}
}

func convertEffectiveRelationToThinSubjectGroup(effectRelation dao.EffectSubjectRelation) types.ThinSubjectGroup {
	return types.ThinSubjectGroup{
		PK:              effectRelation.ParentPK,
		PolicyExpiredAt: effectRelation.PolicyExpiredAt,
	}
}

// GetEffectThinSubjectGroups 获取授权对象的用户组(只返回groupPK/policyExpiredAt)
func (l *groupService) GetEffectThinSubjectGroups(pk int64) (thinSubjectGroup []types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "GetEffectThinSubjectGroups")

	relations, err := l.manager.ListEffectThinRelationBySubjectPK(pk)
	if err != nil {
		return thinSubjectGroup, errorWrapf(err, "ListEffectThinRelationBySubjectPK pk=`%d` fail", pk)
	}

	for _, r := range relations {
		thinSubjectGroup = append(thinSubjectGroup, convertToThinSubjectGroup(r))
	}
	return thinSubjectGroup, err
}

// ListEffectThinSubjectGroups 批量获取 subject 有效的 groups(未过期的)
func (l *groupService) ListEffectThinSubjectGroups(
	pks []int64,
) (subjectGroups map[int64][]types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListEffectThinSubjectGroups")

	subjectGroups = make(map[int64][]types.ThinSubjectGroup, len(pks))

	relations, err := l.manager.ListEffectRelationBySubjectPKs(pks)
	if err != nil {
		return subjectGroups, errorWrapf(err, "ListRelationByPKs pks=`%+v` fail", pks)
	}

	for _, r := range relations {
		subjectPK := r.SubjectPK
		subjectGroups[subjectPK] = append(subjectGroups[subjectPK], convertEffectiveRelationToThinSubjectGroup(r))
	}
	return subjectGroups, nil
}

// ListSubjectGroups ...
func (l *groupService) ListSubjectGroups(
	_type, id string, beforeExpiredAt int64,
) (subjectGroups []types.SubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListSubjectGroups")
	var relations []dao.SubjectRelation
	if beforeExpiredAt == 0 {
		relations, err = l.manager.ListRelation(_type, id)
	} else {
		relations, err = l.manager.ListRelationBeforeExpiredAt(_type, id, beforeExpiredAt)
	}

	if err != nil {
		return subjectGroups, errorWrapf(err, "ListSubjectGroups _type=`%s`, id=`%s` fail", _type, id)
	}

	subjectGroups = make([]types.SubjectGroup, 0, len(relations))

	for _, r := range relations {
		subjectGroups = append(subjectGroups, convertToSubjectGroup(r))
	}
	return subjectGroups, err
}

// ListExistSubjectsBeforeExpiredAt filter the exists and not expired subjects
func (l *groupService) ListExistSubjectsBeforeExpiredAt(
	subjects []types.Subject, expiredAt int64,
) ([]types.Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "FilterSubjectsBeforeExpiredAt")

	groupIDs := make([]string, 0, len(subjects))
	for _, subject := range subjects {
		if subject.Type == types.GroupType {
			groupIDs = append(groupIDs, subject.ID)
		}
	}

	existGroupIDs, err := l.manager.ListParentIDsBeforeExpiredAt(types.GroupType, groupIDs, expiredAt)
	if err != nil {
		return []types.Subject{}, errorWrapf(
			err, "ListParentIDsBeforeExpiredAt _type=`%s`, ids=`%+v`, expiredAt=`%d` fail",
			types.GroupType, groupIDs, expiredAt,
		)
	}
	if len(existGroupIDs) == 0 {
		return []types.Subject{}, nil
	}

	idSet := set.NewStringSetWithValues(existGroupIDs)
	existSubjects := make([]types.Subject, 0, len(existGroupIDs))
	for _, subject := range subjects {
		if subject.Type == types.GroupType && idSet.Has(subject.ID) {
			existSubjects = append(existSubjects, subject)
		}
	}

	return existSubjects, nil
}

// from subject_member.go

func convertToSubjectMembers(daoRelations []dao.SubjectRelation) []types.SubjectMember {
	relations := make([]types.SubjectMember, 0, len(daoRelations))
	for _, r := range daoRelations {
		relations = append(relations, types.SubjectMember{
			PK:              r.PK,
			Type:            r.SubjectType,
			ID:              r.SubjectID,
			PolicyExpiredAt: r.PolicyExpiredAt,
			CreateAt:        r.CreateAt,
		})
	}
	return relations
}

// GetMemberCount ...
func (l *groupService) GetMemberCount(_type, id string) (int64, error) {
	cnt, err := l.manager.GetMemberCount(_type, id)
	if err != nil {
		err = errorx.Wrapf(err, GroupSVC, "GetMemberCount",
			"relationManager.GetMemberCount _type=`%s`, id=`%s` fail", _type, id)
		return 0, err
	}
	return cnt, nil
}

// ListPagingMember ...
func (l *groupService) ListPagingMember(_type, id string, limit, offset int64) ([]types.SubjectMember, error) {
	daoRelations, err := l.manager.ListPagingMember(_type, id, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListPagingMember", "relationManager.ListPagingMember _type=`%s`, id=`%s`, limit=`%d`, offset=`%d`",
			_type, id, limit, offset)
	}

	return convertToSubjectMembers(daoRelations), nil
}

// ListMember ...
func (l *groupService) ListMember(_type, id string) ([]types.SubjectMember, error) {
	daoRelations, err := l.manager.ListMember(_type, id)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListMember", "relationManager.ListMember _type=`%s`, id=`%s` fail", _type, id)
	}

	return convertToSubjectMembers(daoRelations), nil
}

// UpdateMembersExpiredAtWithTx ...
func (l *groupService) UpdateMembersExpiredAtWithTx(tx *sqlx.Tx, members []types.SubjectRelationPKPolicyExpiredAt) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteSubjectMember")

	relations := make([]dao.SubjectRelationPKPolicyExpiredAt, 0, len(members))
	for _, m := range members {
		relations = append(relations, dao.SubjectRelationPKPolicyExpiredAt{
			PK:              m.PK,
			PolicyExpiredAt: m.PolicyExpiredAt,
		})
	}

	err := l.manager.UpdateExpiredAtWithTx(tx, relations)
	if err != nil {
		err = errorWrapf(err,
			"relationManager.UpdateExpiredAtWithTx relations=`%+v` fail", relations)
		return err
	}

	return nil
}

// BulkDeleteSubjectMembers ...
func (l *groupService) BulkDeleteSubjectMembers(_type, id string, members []types.Subject) (map[string]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteSubjectMember")

	// 按类型分组
	userIDs, departmentIDs, _ := groupBySubjectType(members)

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return nil, errorWrapf(err, "define tx error")
	}

	typeCount := map[string]int64{
		types.UserType:       0,
		types.DepartmentType: 0,
	}

	var count int64
	if len(userIDs) != 0 {
		count, err = l.manager.BulkDeleteByMembersWithTx(tx, _type, id, types.UserType, userIDs)
		if err != nil {
			return nil, errorWrapf(err,
				"relationManager.BulkDeleteByMembersWithTx _type=`%s`, id=`%s`, subjectType=`%s`, subjectIDs=`%+v` fail",
				_type, id, types.UserType, userIDs)
		}
		typeCount[types.UserType] = count
	}

	if len(departmentIDs) != 0 {
		count, err = l.manager.BulkDeleteByMembersWithTx(tx, _type, id, types.DepartmentType, departmentIDs)
		if err != nil {
			return nil, errorWrapf(
				err, "relationManager.BulkDeleteByMembersWithTx _type=`%s`, id=`%s`, subjectType=`%s`, subjectIDs=`%+v` fail",
				_type, id, types.DepartmentType, departmentIDs)
		}
		typeCount[types.DepartmentType] = count
	}

	err = tx.Commit()
	if err != nil {
		return nil, errorWrapf(err, "tx commit error")
	}
	return typeCount, err
}

// BulkCreateSubjectMembers ...
func (l *groupService) BulkCreateSubjectMembersWithTx(tx *sqlx.Tx, relations []types.SubjectRelation) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkCreateSubjectMembers")
	// 组装需要创建的Subject关系
	daoRelations := make([]dao.SubjectRelation, 0, len(relations))
	for _, r := range relations {
		daoRelations = append(daoRelations, dao.SubjectRelation{
			SubjectPK:       r.SubjectPK,
			SubjectType:     r.SubjectType,
			SubjectID:       r.SubjectID,
			ParentPK:        r.ParentPK,
			ParentType:      r.ParentType,
			ParentID:        r.ParentID,
			PolicyExpiredAt: r.PolicyExpiredAt,
		})
	}

	err := l.manager.BulkCreateWithTx(tx, daoRelations)
	if err != nil {
		return errorWrapf(err, "relationManager.BulkCreateWithTx relations=`%+v` fail", daoRelations)
	}
	return nil
}

// GetMemberCountBeforeExpiredAt ...
func (l *groupService) GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error) {
	cnt, err := l.manager.GetMemberCountBeforeExpiredAt(_type, id, expiredAt)
	if err != nil {
		err = errorx.Wrapf(err, GroupSVC, "GetMemberCountBeforeExpiredAt",
			"relationManager.GetMemberCountBeforeExpiredAt _type=`%s`, id=`%s`, expiredAt=`%d` fail",
			_type, id, expiredAt)
		return 0, err
	}
	return cnt, nil
}

// ListPagingMemberBeforeExpiredAt ...
func (l *groupService) ListPagingMemberBeforeExpiredAt(
	_type, id string, expiredAt int64, limit, offset int64,
) ([]types.SubjectMember, error) {
	daoRelations, err := l.manager.ListPagingMemberBeforeExpiredAt(
		_type, id, expiredAt, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListPagingMemberBeforeExpiredAt", "_type=`%s`, id=`%s`, expiredAt=`%d`, limit=`%d`, offset=`%d`",
			_type, id, expiredAt, limit, offset)
	}

	return convertToSubjectMembers(daoRelations), nil
}

// BulkDeleteBySubjectPKs ...
func (l *groupService) BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteBySubjectPKs")

	// 批量用户组删除成员关系 subjectRelation
	err := l.manager.BulkDeleteByParentPKs(tx, pks)
	if err != nil {
		return errorWrapf(
			err, "manager.BulkDeleteByParentPKs parent_pks=`%+v` fail", pks)
	}
	// 批量其加入的用户组关系 subjectRelation
	err = l.manager.BulkDeleteBySubjectPKs(tx, pks)
	if err != nil {
		return errorWrapf(
			err, "manager.BulkDeleteBySubjectPKs subject_pks=`%+v` fail", pks)
	}

	return nil
}
