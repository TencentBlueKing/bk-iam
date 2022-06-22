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
	ListEffectThinSubjectGroups(systemID string, pks []int64) (map[int64][]types.ThinSubjectGroup, error)

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

	UpdateMembersExpiredAtWithTx(tx *sqlx.Tx, parentPK int64, members []types.SubjectRelationPKPolicyExpiredAt) error
	BulkDeleteGroupMembers(parentPK int64, userPKs, departmentPKs []int64) (map[string]int64, error)
	BulkCreateGroupMembersWithTx(tx *sqlx.Tx, parentPK int64, relations []types.SubjectRelation) error

	// auth type
	ListGroupAuthSystemIDs(groupPK int64) ([]string, error)
	ListGroupAuthBySystemGroupPKs(systemID string, groupPKs []int64) ([]types.GroupAuthType, error)
	AlterGroupAuthType(tx *sqlx.Tx, systemID string, groupPK int64, authType int64) (changed bool, err error)

	// open api
	ListEffectThinSubjectGroupsBySubjectPKs(pks []int64) ([]types.ThinSubjectGroup, error)
}

type groupService struct {
	manager                   dao.SubjectRelationManager
	authTypeManger            dao.GroupSystemAuthTypeManager
	subjectSystemGroupManager dao.SubjectSystemGroupManager
}

// NewGroupService GroupService工厂
func NewGroupService() GroupService {
	return &groupService{
		manager:                   dao.NewSubjectRelationManager(),
		authTypeManger:            dao.NewGroupSystemAuthTypeManager(),
		subjectSystemGroupManager: dao.NewSubjectSystemGroupManager(),
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

func convertRelationToThinSubjectGroup(effectRelation dao.EffectSubjectRelation) types.ThinSubjectGroup {
	return types.ThinSubjectGroup{
		GroupPK:         effectRelation.ParentPK,
		PolicyExpiredAt: effectRelation.PolicyExpiredAt,
	}
}

// ListEffectThinSubjectGroupsBySubjectPKs 批量获取 subject 有效的 groups(未过期的)
func (l *groupService) ListEffectThinSubjectGroupsBySubjectPKs(
	pks []int64,
) (subjectGroups []types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListEffectThinSubjectGroups")

	relations, err := l.manager.ListEffectRelationBySubjectPKs(pks)
	if err != nil {
		return subjectGroups, errorWrapf(err, "manager.ListEffectRelationBySubjectPKs pks=`%+v` fail", pks)
	}

	subjectGroups = make([]types.ThinSubjectGroup, 0, len(relations))
	for _, r := range relations {
		subjectGroups = append(subjectGroups, convertRelationToThinSubjectGroup(r))
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

func convertToGroupMembers(daoRelations []dao.SubjectRelation) []types.GroupMember {
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

	return convertToGroupMembers(daoRelations), nil
}

// ListMember ...
func (l *groupService) ListMember(parentPK int64) ([]types.GroupMember, error) {
	daoRelations, err := l.manager.ListMember(parentPK)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListMember", "manager.ListMember parentPK=`%d` fail", parentPK)
	}

	return convertToGroupMembers(daoRelations), nil
}

// UpdateMembersExpiredAtWithTx ...
func (l *groupService) UpdateMembersExpiredAtWithTx(
	tx *sqlx.Tx,
	parentPK int64,
	members []types.SubjectRelationPKPolicyExpiredAt,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "UpdateMembersExpiredAtWithTx")

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

	// 更新subject system group
	systemIDs, err := l.ListGroupAuthSystemIDs(parentPK)
	if err != nil {
		return errorWrapf(err, "listGroupAuthSystem parentPK=`%d` fail", parentPK)
	}

	for _, systemID := range systemIDs {
		for _, m := range members {
			err = l.addOrUpdateSubjectSystemGroup(tx, m.SubjectPK, systemID, parentPK, m.PolicyExpiredAt)
			if err != nil {
				return errorWrapf(
					err,
					"addOrUpdateSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, parentPK=`%d`, expiredAt=`%d`, fail",
					systemID,
					m.SubjectPK,
					parentPK,
					m.PolicyExpiredAt,
				)
			}
		}
	}

	return nil
}

// BulkDeleteGroupMembers ...
func (l *groupService) BulkDeleteGroupMembers(
	parentPK int64,
	userPKs, departmentPKs []int64,
) (map[string]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteGroupMember")

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

	// 更新subject system group
	systemIDs, err := l.ListGroupAuthSystemIDs(parentPK)
	if err != nil {
		return nil, errorWrapf(err, "listGroupAuthSystem parentPK=`%d` fail", parentPK)
	}

	subjectPKs := make([]int64, 0, len(userPKs)+len(departmentPKs))
	subjectPKs = append(subjectPKs, userPKs...)
	subjectPKs = append(subjectPKs, departmentPKs...)

	for _, systemID := range systemIDs {
		for _, subjectPK := range subjectPKs {
			err = l.removeSubjectSystemGroup(tx, subjectPK, systemID, parentPK)
			if err != nil {
				return nil, errorWrapf(
					err,
					"updateSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, parentPK=`%d`, fail",
					systemID,
					subjectPK,
					parentPK,
				)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, errorWrapf(err, "tx commit error")
	}
	return typeCount, err
}

// BulkCreateGroupMembersWithTx ...
func (l *groupService) BulkCreateGroupMembersWithTx(
	tx *sqlx.Tx,
	parentPK int64,
	relations []types.SubjectRelation,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkCreateGroupMembersWithTx")
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

	// 更新subject system group
	systemIDs, err := l.ListGroupAuthSystemIDs(parentPK)
	if err != nil {
		return errorWrapf(err, "listGroupAuthSystem parentPK=`%d` fail", parentPK)
	}

	for _, systemID := range systemIDs {
		for _, r := range relations {
			err = l.addOrUpdateSubjectSystemGroup(tx, r.SubjectPK, systemID, parentPK, r.PolicyExpiredAt)
			if err != nil {
				return errorWrapf(
					err,
					"addOrUpdateSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, parentPK=`%d`, expiredAt=`%d`, fail",
					systemID,
					r.SubjectPK,
					parentPK,
					r.PolicyExpiredAt,
				)
			}
		}
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

	return convertToGroupMembers(daoRelations), nil
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

	// 批量删除用户的subject system group
	err = l.subjectSystemGroupManager.DeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(err, "subjectSystemGroupManager.DeleteBySubjectPKsWithTx pks=`%+v` fail", pks)
	}

	return nil
}
