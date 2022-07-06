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
	"time"

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
	ListEffectThinSubjectGroups(systemID string, subjectPKs []int64) (map[int64][]types.ThinSubjectGroup, error)

	// web api
	GetSubjectGroupCountBeforeExpiredAt(subjectPK int64, expiredAt int64) (int64, error)
	ListPagingSubjectGroups(subjectPK, beforeExpiredAt int64, limit, offset int64) ([]types.SubjectGroup, error)
	FilterExistEffectSubjectGroupPKs(subjectPKs []int64, groupPKs []int64) ([]int64, error)
	FilterGroupPKsHasMemberBeforeExpiredAt(groupPKs []int64, expiredAt int64) ([]int64, error)

	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error

	GetGroupMemberCount(groupPK int64) (int64, error)
	GetGroupMemberCountBeforeExpiredAt(groupPK int64, expiredAt int64) (int64, error)
	ListPagingGroupMember(groupPK, limit, offset int64) ([]types.GroupMember, error)
	ListPagingGroupMemberBeforeExpiredAt(
		groupPK int64, expiredAt int64, limit, offset int64,
	) ([]types.GroupMember, error)
	ListGroupMember(groupPK int64) ([]types.GroupMember, error)

	UpdateGroupMembersExpiredAtWithTx(tx *sqlx.Tx, groupPK int64, members []types.SubjectRelationForUpdate) error
	BulkDeleteGroupMembers(groupPK int64, userPKs, departmentPKs []int64) (map[string]int64, error)
	BulkCreateGroupMembersWithTx(tx *sqlx.Tx, groupPK int64, relations []types.SubjectRelationForCreate) error

	// auth type
	ListGroupAuthSystemIDs(groupPK int64) ([]string, error)
	ListGroupAuthBySystemGroupPKs(systemID string, groupPKs []int64) ([]types.GroupAuthType, error)
	AlterGroupAuthType(tx *sqlx.Tx, systemID string, groupPK int64, authType int64) (changed bool, err error)

	// open api
	ListEffectThinSubjectGroupsBySubjectPKs(subjectPKs []int64) ([]types.ThinSubjectGroup, error)
}

type groupService struct {
	manager                   dao.SubjectGroupManager
	authTypeManger            dao.GroupSystemAuthTypeManager
	subjectSystemGroupManager dao.SubjectSystemGroupManager
}

// NewGroupService GroupService工厂
func NewGroupService() GroupService {
	return &groupService{
		manager:                   dao.NewSubjectGroupManager(),
		authTypeManger:            dao.NewGroupSystemAuthTypeManager(),
		subjectSystemGroupManager: dao.NewSubjectSystemGroupManager(),
	}
}

// from subject_group.go

func convertToSubjectGroup(relation dao.SubjectRelation) types.SubjectGroup {
	return types.SubjectGroup{
		PK:        relation.PK,
		GroupPK:   relation.GroupPK,
		ExpiredAt: relation.ExpiredAt,
		CreateAt:  relation.CreateAt,
	}
}

func convertThinRelationToThinSubjectGroup(thinRelation dao.ThinSubjectRelation) types.ThinSubjectGroup {
	return types.ThinSubjectGroup{
		GroupPK:   thinRelation.GroupPK,
		ExpiredAt: thinRelation.ExpiredAt,
	}
}

// ListEffectThinSubjectGroupsBySubjectPKs 批量获取 subject 有效的 groups(未过期的)
func (l *groupService) ListEffectThinSubjectGroupsBySubjectPKs(
	subjectPKs []int64,
) (subjectGroups []types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListEffectThinSubjectGroupsBySubjectPKs")

	// 过期时间必须大于当前时间
	now := time.Now().Unix()

	relations, err := l.manager.ListThinRelationAfterExpiredAtBySubjectPKs(subjectPKs, now)
	if err != nil {
		return subjectGroups, errorWrapf(
			err,
			"manager.ListThinRelationAfterExpiredAtBySubjectPKs subjectPKs=`%+v` fail",
			subjectPKs,
		)
	}

	subjectGroups = make([]types.ThinSubjectGroup, 0, len(relations))
	for _, r := range relations {
		subjectGroups = append(subjectGroups, convertThinRelationToThinSubjectGroup(r))
	}
	return subjectGroups, nil
}

// GetSubjectGroupCountBeforeExpiredAt ...
func (l *groupService) GetSubjectGroupCountBeforeExpiredAt(subjectPK int64, expiredAt int64) (count int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "GetSubjectGroupCountBeforeExpiredAt")

	if expiredAt == 0 {
		count, err = l.manager.GetSubjectGroupCount(subjectPK)
	} else {
		count, err = l.manager.GetSubjectGroupCountBeforeExpiredAt(subjectPK, expiredAt)
	}
	if err != nil {
		return 0, errorWrapf(
			err,
			"manager.GetSubjectGroupCountBeforeExpiredAt subjectPK=`%d, expiredAt=`%d` fail",
			subjectPK,
			expiredAt,
		)
	}
	return count, nil
}

// ListPagingSubjectGroups ...
func (l *groupService) ListPagingSubjectGroups(
	subjectPK, beforeExpiredAt, limit, offset int64,
) (subjectGroups []types.SubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListPagingSubjectGroups")

	var relations []dao.SubjectRelation
	if beforeExpiredAt == 0 {
		relations, err = l.manager.ListPagingSubjectGroups(subjectPK, limit, offset)
	} else {
		relations, err = l.manager.ListPagingSubjectGroupBeforeExpiredAt(subjectPK, beforeExpiredAt, limit, offset)
	}

	if err != nil {
		return subjectGroups, errorWrapf(
			err,
			"manager.ListPagingSubjectGroupBeforeExpiredAt subjectPK=`%d`, beforeExpiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			subjectPK,
			beforeExpiredAt,
			limit,
			offset,
		)
	}

	subjectGroups = make([]types.SubjectGroup, 0, len(relations))
	for _, r := range relations {
		subjectGroups = append(subjectGroups, convertToSubjectGroup(r))
	}
	return subjectGroups, err
}

// FilterGroupPKsHasMemberBeforeExpiredAt filter the exists and not expired subjects
func (l *groupService) FilterGroupPKsHasMemberBeforeExpiredAt(
	groupPKs []int64, expiredAt int64,
) ([]int64, error) {
	return l.manager.FilterGroupPKsHasMemberBeforeExpiredAt(groupPKs, expiredAt)
}

func (l *groupService) FilterExistEffectSubjectGroupPKs(
	subjectPKs []int64,
	groupPKs []int64,
) (existGroupPKs []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "FilterExistEffectSubjectGroupPKs")

	// 过期时间必须大于当前时间
	now := time.Now().Unix()

	existGroupPKs, err = l.manager.FilterSubjectPKsExistGroupPKsAfterExpiredAt(subjectPKs, groupPKs, now)
	if err != nil {
		return nil, errorWrapf(
			err,
			"manager.FilterSubjectPKsExistGroupPKsAfterExpiredAt subjectPK=`%+v`, parenPKs=`%+v`, now=`%d` fail",
			subjectPKs, groupPKs, now,
		)
	}
	return set.NewInt64SetWithValues(existGroupPKs).ToSlice(), nil
}

// from subject_member.go

func convertToGroupMembers(daoRelations []dao.SubjectRelation) []types.GroupMember {
	relations := make([]types.GroupMember, 0, len(daoRelations))
	for _, r := range daoRelations {
		relations = append(relations, types.GroupMember{
			PK:        r.PK,
			SubjectPK: r.SubjectPK,
			ExpiredAt: r.ExpiredAt,
			CreateAt:  r.CreateAt,
		})
	}
	return relations
}

// GetGroupMemberCount ...
func (l *groupService) GetGroupMemberCount(groupPK int64) (int64, error) {
	return l.manager.GetGroupMemberCount(groupPK)
}

// ListPagingGroupMember ...
func (l *groupService) ListPagingGroupMember(groupPK, limit, offset int64) ([]types.GroupMember, error) {
	daoRelations, err := l.manager.ListPagingGroupMember(groupPK, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListPagingGroupMember", "manager.ListPagingGroupMember groupPK=`%d`, limit=`%d`, offset=`%d`",
			groupPK, limit, offset)
	}

	return convertToGroupMembers(daoRelations), nil
}

// ListGroupMember ...
func (l *groupService) ListGroupMember(groupPK int64) ([]types.GroupMember, error) {
	daoRelations, err := l.manager.ListGroupMember(groupPK)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListGroupMember", "manager.ListGroupMember groupPK=`%d` fail", groupPK)
	}

	return convertToGroupMembers(daoRelations), nil
}

// UpdateGroupMembersExpiredAtWithTx ...
func (l *groupService) UpdateGroupMembersExpiredAtWithTx(
	tx *sqlx.Tx,
	groupPK int64,
	members []types.SubjectRelationForUpdate,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "UpdateGroupMembersExpiredAtWithTx")

	relations := make([]dao.SubjectRelationForUpdateExpiredAt, 0, len(members))
	for _, m := range members {
		relations = append(relations, dao.SubjectRelationForUpdateExpiredAt{
			PK:        m.PK,
			ExpiredAt: m.ExpiredAt,
		})
	}

	err := l.manager.UpdateExpiredAtWithTx(tx, relations)
	if err != nil {
		err = errorWrapf(err, "manager.UpdateExpiredAtWithTx relations=`%+v` fail", relations)
		return err
	}

	// 更新subject system group
	systemIDs, err := l.ListGroupAuthSystemIDs(groupPK)
	if err != nil {
		return errorWrapf(err, "listGroupAuthSystem groupPK=`%d` fail", groupPK)
	}

	for _, systemID := range systemIDs {
		for _, m := range members {
			err = l.addOrUpdateSubjectSystemGroup(tx, m.SubjectPK, systemID, groupPK, m.ExpiredAt)
			if err != nil {
				return errorWrapf(
					err,
					"addOrUpdateSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, groupPK=`%d`, expiredAt=`%d`, fail",
					systemID,
					m.SubjectPK,
					groupPK,
					m.ExpiredAt,
				)
			}
		}
	}

	return nil
}

// BulkDeleteGroupMembers ...
func (l *groupService) BulkDeleteGroupMembers(
	groupPK int64,
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
		count, err = l.manager.BulkDeleteByGroupMembersWithTx(tx, groupPK, userPKs)
		if err != nil {
			return nil, errorWrapf(err,
				"manager.BulkDeleteByGroupMembersWithTx groupPK=`%d`, userPKs=`%+v` fail",
				groupPK, userPKs)
		}
		typeCount[types.UserType] = count
	}

	if len(departmentPKs) != 0 {
		count, err = l.manager.BulkDeleteByGroupMembersWithTx(tx, groupPK, departmentPKs)
		if err != nil {
			return nil, errorWrapf(
				err, "manager.BulkDeleteByGroupMembersWithTx groupPK=`%d`, departmentPKs=`%+v` fail",
				groupPK, departmentPKs)
		}
		typeCount[types.DepartmentType] = count
	}

	// 更新subject system group
	systemIDs, err := l.ListGroupAuthSystemIDs(groupPK)
	if err != nil {
		return nil, errorWrapf(err, "listGroupAuthSystem groupPK=`%d` fail", groupPK)
	}

	subjectPKs := make([]int64, 0, len(userPKs)+len(departmentPKs))
	subjectPKs = append(subjectPKs, userPKs...)
	subjectPKs = append(subjectPKs, departmentPKs...)

	for _, systemID := range systemIDs {
		for _, subjectPK := range subjectPKs {
			err = l.removeSubjectSystemGroup(tx, subjectPK, systemID, groupPK)
			if err != nil {
				return nil, errorWrapf(
					err,
					"updateSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, groupPK=`%d`, fail",
					systemID,
					subjectPK,
					groupPK,
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
	groupPK int64,
	relations []types.SubjectRelationForCreate,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkCreateGroupMembersWithTx")
	// 组装需要创建的Subject关系
	daoRelations := make([]dao.SubjectRelation, 0, len(relations))
	for _, r := range relations {
		daoRelations = append(daoRelations, dao.SubjectRelation{
			SubjectPK: r.SubjectPK,
			GroupPK:   r.GroupPK,
			ExpiredAt: r.ExpiredAt,
		})
	}

	err := l.manager.BulkCreateWithTx(tx, daoRelations)
	if err != nil {
		return errorWrapf(err, "manager.BulkCreateWithTx relations=`%+v` fail", daoRelations)
	}

	// 更新subject system group
	systemIDs, err := l.ListGroupAuthSystemIDs(groupPK)
	if err != nil {
		return errorWrapf(err, "listGroupAuthSystem groupPK=`%d` fail", groupPK)
	}

	for _, systemID := range systemIDs {
		for _, r := range relations {
			err = l.addOrUpdateSubjectSystemGroup(tx, r.SubjectPK, systemID, groupPK, r.ExpiredAt)
			if err != nil {
				return errorWrapf(
					err,
					"addOrUpdateSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, groupPK=`%d`, expiredAt=`%d`, fail",
					systemID,
					r.SubjectPK,
					groupPK,
					r.ExpiredAt,
				)
			}
		}
	}
	return nil
}

// GetGroupMemberCountBeforeExpiredAt ...
func (l *groupService) GetGroupMemberCountBeforeExpiredAt(groupPK int64, expiredAt int64) (int64, error) {
	return l.manager.GetGroupMemberCountBeforeExpiredAt(groupPK, expiredAt)
}

// ListPagingGroupMemberBeforeExpiredAt ...
func (l *groupService) ListPagingGroupMemberBeforeExpiredAt(
	groupPK int64, expiredAt int64, limit, offset int64,
) ([]types.GroupMember, error) {
	daoRelations, err := l.manager.ListPagingGroupMemberBeforeExpiredAt(
		groupPK, expiredAt, limit, offset,
	)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListPagingGroupMemberBeforeExpiredAt", "groupPK=`%d`, expiredAt=`%d`, limit=`%d`, offset=`%d`",
			groupPK, expiredAt, limit, offset)
	}

	return convertToGroupMembers(daoRelations), nil
}

// BulkDeleteBySubjectPKsWithTx ...
func (l *groupService) BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteBySubjectPKsWithTx")

	// 批量用户组删除成员关系 subjectRelation
	err := l.manager.BulkDeleteByGroupPKs(tx, subjectPKs)
	if err != nil {
		return errorWrapf(
			err, "manager.BulkDeleteByGroupPKs group_pks=`%+v` fail", subjectPKs)
	}
	// 批量其加入的用户组关系 subjectRelation
	err = l.manager.BulkDeleteBySubjectPKs(tx, subjectPKs)
	if err != nil {
		return errorWrapf(
			err, "manager.BulkDeleteBySubjectPKs subject_pks=`%+v` fail", subjectPKs)
	}

	// 批量删除用户的subject system group
	err = l.subjectSystemGroupManager.DeleteBySubjectPKsWithTx(tx, subjectPKs)
	if err != nil {
		return errorWrapf(err, "subjectSystemGroupManager.DeleteBySubjectPKsWithTx subjectPKs=`%+v` fail", subjectPKs)
	}

	return nil
}
