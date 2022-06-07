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
	ListSubjectGroups(subjectPK, beforeExpiredAt int64) ([]types.SubjectGroup, error)
	ListExistSubjectsBeforeExpiredAt(parentPKs []int64, expiredAt int64) ([]int64, error)

	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, pks []int64) error

	GetMemberCount(parentPK int64) (int64, error)
	GetMemberCountBeforeExpiredAt(parentPK int64, expiredAt int64) (int64, error)
	ListPagingMember(parentPK, limit, offset int64) ([]types.GroupMember, error)
	ListPagingMemberBeforeExpiredAt(
		parentPK int64, expiredAt int64, limit, offset int64,
	) ([]types.GroupMember, error)
	ListMember(parentPK int64) ([]types.GroupMember, error)

	UpdateMembersExpiredAtWithTx(tx *sqlx.Tx, members []types.SubjectRelationPKPolicyExpiredAt) error
	BulkDeleteSubjectMembers(parentPK int64, userPKs, departmentPKs []int64) (map[string]int64, error)
	BulkCreateSubjectMembersWithTx(tx *sqlx.Tx, relations []types.SubjectRelation) error
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
		PK:              relation.PK,
		ParentPK:        relation.ParentPK,
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
		return thinSubjectGroup, errorWrapf(err, "manager.ListEffectThinRelationBySubjectPK pk=`%d` fail", pk)
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
		return subjectGroups, errorWrapf(err, "manager.ListRelationByPKs pks=`%+v` fail", pks)
	}

	for _, r := range relations {
		subjectPK := r.SubjectPK
		subjectGroups[subjectPK] = append(subjectGroups[subjectPK], convertEffectiveRelationToThinSubjectGroup(r))
	}
	return subjectGroups, nil
}

// ListSubjectGroups ...
func (l *groupService) ListSubjectGroups(
	subjectPK, beforeExpiredAt int64,
) (subjectGroups []types.SubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListSubjectGroups")
	var relations []dao.SubjectRelation
	if beforeExpiredAt == 0 {
		relations, err = l.manager.ListRelation(subjectPK)
	} else {
		relations, err = l.manager.ListRelationBeforeExpiredAt(subjectPK, beforeExpiredAt)
	}
	if err != nil {
		return subjectGroups, errorWrapf(err, "manager.ListSubjectGroups subjectPK=`%d` fail", subjectPK)
	}

	subjectGroups = make([]types.SubjectGroup, 0, len(relations))
	for _, r := range relations {
		subjectGroups = append(subjectGroups, convertToSubjectGroup(r))
	}
	return subjectGroups, err
}

// ListExistSubjectsBeforeExpiredAt filter the exists and not expired subjects
func (l *groupService) ListExistSubjectsBeforeExpiredAt(
	parentPKs []int64, expiredAt int64,
) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "FilterSubjectsBeforeExpiredAt")

	existGroupPKs, err := l.manager.ListParentPKsBeforeExpiredAt(parentPKs, expiredAt)
	if err != nil {
		return []int64{}, errorWrapf(
			err, "manager.ListParentPKsBeforeExpiredAt parentPKs=`%+v`, expiredAt=`%d` fail",
			parentPKs, expiredAt,
		)
	}
	if len(existGroupPKs) == 0 {
		return []int64{}, nil
	}

	return existGroupPKs, err
}

// from subject_member.go

func convertToSubjectMembers(daoRelations []dao.SubjectRelation) []types.GroupMember {
	relations := make([]types.GroupMember, 0, len(daoRelations))
	for _, r := range daoRelations {
		relations = append(relations, types.GroupMember{
			PK:              r.PK,
			SubjectPK:       r.SubjectPK,
			PolicyExpiredAt: r.PolicyExpiredAt,
			CreateAt:        r.CreateAt,
		})
	}
	return relations
}

// GetMemberCount ...
func (l *groupService) GetMemberCount(parentPK int64) (int64, error) {
	cnt, err := l.manager.GetMemberCount(parentPK)
	if err != nil {
		err = errorx.Wrapf(err, GroupSVC, "GetMemberCount",
			"manager.GetMemberCount parentPK=`%d` fail", parentPK)
		return 0, err
	}
	return cnt, nil
}

// ListPagingMember ...
func (l *groupService) ListPagingMember(parentPK, limit, offset int64) ([]types.GroupMember, error) {
	daoRelations, err := l.manager.ListPagingMember(parentPK, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListPagingMember", "manager.ListPagingMember parentPK=`%d`, limit=`%d`, offset=`%d`",
			parentPK, limit, offset)
	}

	return convertToSubjectMembers(daoRelations), nil
}

// ListMember ...
func (l *groupService) ListMember(parentPK int64) ([]types.GroupMember, error) {
	daoRelations, err := l.manager.ListMember(parentPK)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListMember", "manager.ListMember parentPK=`%d` fail", parentPK)
	}

	return convertToSubjectMembers(daoRelations), nil
}

// UpdateMembersExpiredAtWithTx ...
func (l *groupService) UpdateMembersExpiredAtWithTx(
	tx *sqlx.Tx,
	members []types.SubjectRelationPKPolicyExpiredAt,
) error {
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
		err = errorWrapf(err, "manager.UpdateExpiredAtWithTx relations=`%+v` fail", relations)
		return err
	}

	return nil
}

// BulkDeleteSubjectMembers ...
func (l *groupService) BulkDeleteSubjectMembers(
	parentPK int64,
	userPKs, departmentPKs []int64,
) (map[string]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteSubjectMember")

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
	if len(userPKs) != 0 {
		count, err = l.manager.BulkDeleteByMembersWithTx(tx, parentPK, userPKs)
		if err != nil {
			return nil, errorWrapf(err,
				"manager.BulkDeleteByMembersWithTx parentPK=`%d`, userPKs=`%+v` fail",
				parentPK, userPKs)
		}
		typeCount[types.UserType] = count
	}

	if len(departmentPKs) != 0 {
		count, err = l.manager.BulkDeleteByMembersWithTx(tx, parentPK, departmentPKs)
		if err != nil {
			return nil, errorWrapf(
				err, "manager.BulkDeleteByMembersWithTx parentPK=`%d`, departmentPKs=`%+v` fail",
				parentPK, departmentPKs)
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
			ParentPK:        r.ParentPK,
			PolicyExpiredAt: r.PolicyExpiredAt,
		})
	}

	err := l.manager.BulkCreateWithTx(tx, daoRelations)
	if err != nil {
		return errorWrapf(err, "manager.BulkCreateWithTx relations=`%+v` fail", daoRelations)
	}
	return nil
}

// GetMemberCountBeforeExpiredAt ...
func (l *groupService) GetMemberCountBeforeExpiredAt(parentPK int64, expiredAt int64) (int64, error) {
	cnt, err := l.manager.GetMemberCountBeforeExpiredAt(parentPK, expiredAt)
	if err != nil {
		err = errorx.Wrapf(err, GroupSVC, "GetMemberCountBeforeExpiredAt",
			"manager.GetMemberCountBeforeExpiredAt parentPK=`%d`, expiredAt=`%d` fail",
			parentPK, expiredAt)
		return 0, err
	}
	return cnt, nil
}

// ListPagingMemberBeforeExpiredAt ...
func (l *groupService) ListPagingMemberBeforeExpiredAt(
	parentPK int64, expiredAt int64, limit, offset int64,
) ([]types.GroupMember, error) {
	daoRelations, err := l.manager.ListPagingMemberBeforeExpiredAt(
		parentPK, expiredAt, limit, offset,
	)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListPagingMemberBeforeExpiredAt", "parentPK=`%d`, expiredAt=`%d`, limit=`%d`, offset=`%d`",
			parentPK, expiredAt, limit, offset)
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
