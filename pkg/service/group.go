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
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// ErrGroupMemberNotFound ...
var ErrGroupMemberNotFound = errors.New("group member not found")

// GroupService ...
type GroupService interface {
	// 鉴权
	ListEffectThinSubjectGroups(systemID string, subjectPKs []int64) (map[int64][]types.ThinSubjectGroup, error)

	// web api
	GetSubjectGroupCountBeforeExpiredAt(subjectPK int64, expiredAt int64) (int64, error)
	GetSubjectSystemGroupCountBeforeExpiredAt(subjectPK int64, systemID string, expiredAt int64) (int64, error)
	ListPagingSubjectGroups(subjectPK, beforeExpiredAt int64, limit, offset int64) ([]types.SubjectGroup, error)
	ListPagingSubjectSystemGroups(
		subjectPK int64, systemID string, beforeExpiredAt, limit, offset int64,
	) ([]types.SubjectGroup, error)
	ListEffectSubjectGroupsBySubjectPKGroupPKs(
		subjectPK int64,
		groupPKs []int64,
	) ([]types.SubjectGroupWithSource, error)
	FilterGroupPKsHasMemberBeforeExpiredAt(groupPKs []int64, expiredAt int64) ([]int64, error)

	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error
	BulkDeleteByGroupPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error

	GetGroupMemberCount(groupPK int64) (int64, error)
	GetGroupMemberCountBeforeExpiredAt(groupPK int64, expiredAt int64) (int64, error)
	ListPagingGroupMember(groupPK, limit, offset int64) ([]types.GroupMember, error)
	ListPagingGroupMemberBeforeExpiredAt(
		groupPK int64, expiredAt int64, limit, offset int64,
	) ([]types.GroupMember, error)
	ListGroupMember(groupPK int64) ([]types.GroupMember, error)
	GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (count int64, err error)
	ListPagingGroupSubjectBeforeExpiredAt(expiredAt int64, limit, offset int64) ([]types.GroupSubject, error)
	GetTemplateGroupMemberCount(groupPK, templateID int64) (int64, error)
	ListPagingTemplateGroupMember(groupPK, templateID, limit, offset int64) ([]types.GroupMember, error)

	UpdateGroupMembersExpiredAtWithTx(tx *sqlx.Tx, groupPK int64, members []types.SubjectTemplateGroup) error
	UpdateSubjectTemplateGroupExpiredAtWithTx(tx *sqlx.Tx, relations []types.SubjectTemplateGroup) error
	BulkDeleteGroupMembers(groupPK int64, userPKs, departmentPKs []int64) (map[string]int64, error)
	BulkCreateGroupMembersWithTx(tx *sqlx.Tx, groupPK int64, relations []types.SubjectTemplateGroup) error
	BulkCreateSubjectTemplateGroupWithTx(tx *sqlx.Tx, relations []types.SubjectTemplateGroup) error
	UpdateSubjectGroupExpiredAtWithTx(
		tx *sqlx.Tx,
		relations []types.SubjectTemplateGroup,
		updateSubjectRelation bool,
	) error
	BulkDeleteSubjectTemplateGroupWithTx(tx *sqlx.Tx, relations []types.SubjectTemplateGroup) error
	BulkUpdateSubjectSystemGroupBySubjectTemplateGroupWithTx(
		tx *sqlx.Tx,
		relations []types.SubjectTemplateGroup,
	) error

	// auth type
	GetGroupOneAuthSystem(groupPK int64) (string, error)
	ListGroupAuthSystemIDs(groupPK int64) ([]string, error)
	ListGroupAuthBySystemGroupPKs(systemID string, groupPKs []int64) ([]types.GroupAuthType, error)
	AlterGroupAuthType(tx *sqlx.Tx, systemID string, groupPK int64, authType int64) (changed bool, err error)

	// open api
	ListEffectThinSubjectGroupsBySubjectPKs(subjectPKs []int64) ([]types.ThinSubjectGroup, error)

	// task
	GetMaxExpiredAtBySubjectGroup(subjectPK, groupPK int64, excludeTemplateID int64) (int64, error)
}

type groupService struct {
	manager                     dao.SubjectGroupManager
	authTypeManger              dao.GroupSystemAuthTypeManager
	subjectSystemGroupManager   dao.SubjectSystemGroupManager
	subjectTemplateGroupManager dao.SubjectTemplateGroupManager
}

// NewGroupService GroupService工厂
func NewGroupService() GroupService {
	return &groupService{
		manager:                     dao.NewSubjectGroupManager(),
		authTypeManger:              dao.NewGroupSystemAuthTypeManager(),
		subjectSystemGroupManager:   dao.NewSubjectSystemGroupManager(),
		subjectTemplateGroupManager: dao.NewSubjectTemplateGroupManager(),
	}
}

// from subject_group.go

func convertToSubjectGroup(relation dao.SubjectRelation) types.SubjectGroup {
	return types.SubjectGroup{
		PK:        relation.PK,
		GroupPK:   relation.GroupPK,
		ExpiredAt: relation.ExpiredAt,
		CreatedAt: relation.CreatedAt,
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

// GetSubjectSystemGroupCountBeforeExpiredAt ...
func (l *groupService) GetSubjectSystemGroupCountBeforeExpiredAt(
	subjectPK int64,
	systemID string,
	expiredAt int64,
) (count int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "GetSubjectSystemGroupCountBeforeExpiredAt")

	if expiredAt == 0 {
		count, err = l.manager.GetSubjectSystemGroupCount(subjectPK, systemID)
	} else {
		count, err = l.manager.GetSubjectSystemGroupCountBeforeExpiredAt(subjectPK, systemID, expiredAt)
	}
	if err != nil {
		return 0, errorWrapf(
			err,
			"manager.GetSubjectSystemGroupCountBeforeExpiredAt subjectPK=`%d, systemID=`%s`, expiredAt=`%d` fail",
			subjectPK,
			systemID,
			expiredAt,
		)
	}
	return count, nil
}

// GetGroupSubjectCountBeforeExpiredAt ...
func (l *groupService) GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (count int64, err error) {
	return l.manager.GetGroupSubjectCountBeforeExpiredAt(expiredAt)
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

// ListPagingSubjectSystemGroups ...
func (l *groupService) ListPagingSubjectSystemGroups(
	subjectPK int64, systemID string, beforeExpiredAt, limit, offset int64,
) (subjectGroups []types.SubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListPagingSubjectGroups")

	var relations []dao.SubjectRelation
	if beforeExpiredAt == 0 {
		relations, err = l.manager.ListPagingSubjectSystemGroups(subjectPK, systemID, limit, offset)
	} else {
		relations, err = l.manager.ListPagingSubjectSystemGroupBeforeExpiredAt(
			subjectPK, systemID, beforeExpiredAt, limit, offset,
		)
	}

	if err != nil {
		return subjectGroups, errorWrapf(
			err,
			"manager.ListPagingSubjectSystemGroupBeforeExpiredAt subjectPK=`%d`, "+
				"systemID=`%s`, beforeExpiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			subjectPK,
			systemID,
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

func (l *groupService) ListEffectSubjectGroupsBySubjectPKGroupPKs(
	subjectPK int64,
	groupPKs []int64,
) (subjectGroups []types.SubjectGroupWithSource, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListEffectSubjectGroupsBySubjectPKGroupPKs")

	relations, err := l.manager.ListRelationBySubjectPKGroupPKs(subjectPK, groupPKs)
	if err != nil {
		return nil, errorWrapf(
			err,
			"manager.ListRelationBySubjectPKGroupPKs subjectPK=`%d`, parenPKs=`%+v` fail",
			subjectPK, groupPKs,
		)
	}

	templateRelations, err := l.subjectTemplateGroupManager.ListRelationBySubjectPKGroupPKs(subjectPK, groupPKs)
	if err != nil {
		return nil, errorWrapf(
			err,
			"subjectTemplateGroupManager.ListRelationBySubjectPKGroupPKs subjectPK=`%d`, parenPKs=`%+v` fail",
			subjectPK, groupPKs,
		)
	}

	groupPKset := set.NewInt64Set()

	subjectGroups = make([]types.SubjectGroupWithSource, 0, len(relations)+len(templateRelations))
	for _, r := range relations {
		subjectGroups = append(subjectGroups, types.SubjectGroupWithSource{
			PK:            r.PK,
			GroupPK:       r.GroupPK,
			ExpiredAt:     r.ExpiredAt,
			CreatedAt:     r.CreatedAt,
			IsDirectAdded: true,
		})

		groupPKset.Add(r.GroupPK)
	}

	for _, r := range templateRelations {
		if groupPKset.Has(r.GroupPK) {
			continue
		}

		subjectGroups = append(subjectGroups, types.SubjectGroupWithSource{
			PK:        r.PK,
			GroupPK:   r.GroupPK,
			ExpiredAt: r.ExpiredAt,
			CreatedAt: r.CreatedAt,
		})

		groupPKset.Add(r.GroupPK)
	}
	return subjectGroups, nil
}

// from subject_member.go

func convertToGroupMembers(daoRelations []dao.SubjectRelation) []types.GroupMember {
	relations := make([]types.GroupMember, 0, len(daoRelations))
	for _, r := range daoRelations {
		relations = append(relations, types.GroupMember{
			PK:        r.PK,
			SubjectPK: r.SubjectPK,
			ExpiredAt: r.ExpiredAt,
			CreatedAt: r.CreatedAt,
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

// GetTemplateGroupMemberCount ...
func (l *groupService) GetTemplateGroupMemberCount(groupPK, templateID int64) (int64, error) {
	return l.subjectTemplateGroupManager.GetTemplateGroupMemberCount(groupPK, templateID)
}

// ListPagingTemplateGroupMember ...
func (l *groupService) ListPagingTemplateGroupMember(
	groupPK, templateID, limit, offset int64,
) ([]types.GroupMember, error) {
	templateMembers, err := l.subjectTemplateGroupManager.ListPagingTemplateGroupMember(
		groupPK,
		templateID,
		limit,
		offset,
	)
	if err != nil {
		return nil, errorx.Wrapf(
			err,
			GroupSVC,
			"ListPagingTemplateGroupMember",
			"subjectTemplateGroupManager.ListPagingTemplateGroupMember groupPK=`%d`, limit=`%d`, offset=`%d`",
			groupPK,
			limit,
			offset,
		)
	}

	members := make([]types.GroupMember, 0, len(templateMembers))
	for _, m := range templateMembers {
		members = append(members, types.GroupMember{
			PK:        m.PK,
			SubjectPK: m.SubjectPK,
			ExpiredAt: m.ExpiredAt,
			CreatedAt: m.CreatedAt,
		})
	}
	return members, nil
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
	members []types.SubjectTemplateGroup,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "UpdateGroupMembersExpiredAtWithTx")

	relations := make([]dao.SubjectRelation, 0, len(members))
	for _, m := range members {
		relations = append(relations, dao.SubjectRelation{
			SubjectPK: m.SubjectPK,
			GroupPK:   groupPK,
			ExpiredAt: m.ExpiredAt,
		})
	}

	err := l.manager.BulkUpdateExpiredAtWithTx(tx, relations)
	if err != nil {
		err = errorWrapf(err, "manager.BulkUpdateExpiredAtWithTx relations=`%+v` fail", relations)
		return err
	}

	// 更新subject system group
	systemIDs, err := l.ListGroupAuthSystemIDs(groupPK)
	if err != nil {
		return errorWrapf(err, "listGroupAuthSystem groupPK=`%d` fail", groupPK)
	}

	for _, m := range members {
		if !m.NeedUpdate {
			continue
		}

		for _, systemID := range systemIDs {
			err = l.addOrUpdateSubjectSystemGroup(tx, m.SubjectPK, systemID, map[int64]int64{groupPK: m.ExpiredAt})
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

	now := time.Now().Unix()
	// 需要判断还有没有其他的数据再来删除
	for _, subjectPK := range subjectPKs {
		// 如果还有其它的未过期的, 不需要删除
		expiredAt, err := l.subjectTemplateGroupManager.GetMaxExpiredAtBySubjectGroup(subjectPK, groupPK, 0)
		if err != nil && err != ErrGroupMemberNotFound {
			return nil, errorWrapf(
				err,
				"getMaxExpiredAtBySubjectGroup subjectPK=`%d`, groupPK=`%d` fail",
				subjectPK,
				groupPK,
			)
		}

		if err == nil && expiredAt > now {
			continue
		}

		for _, systemID := range systemIDs {
			err = l.removeSubjectSystemGroup(tx, subjectPK, systemID, map[int64]int64{groupPK: 0})
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
	relations []types.SubjectTemplateGroup,
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

	for _, r := range relations {
		if !r.NeedUpdate {
			continue
		}

		for _, systemID := range systemIDs {
			err = l.addOrUpdateSubjectSystemGroup(tx, r.SubjectPK, systemID, map[int64]int64{groupPK: r.ExpiredAt})
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

// BulkCreateSubjectTemplateGroupWithTx ...
func (l *groupService) BulkCreateSubjectTemplateGroupWithTx(
	tx *sqlx.Tx,
	relations []types.SubjectTemplateGroup,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkCreateSubjectTemplateGroupWithTx")
	// 组装需要创建的Subject关系
	daoRelations := convertToDaoSubjectTemplateGroup(relations)

	err := l.subjectTemplateGroupManager.BulkCreateWithTx(tx, daoRelations)
	if err != nil {
		return errorWrapf(err, "subjectTemplateGroupManager.BulkCreateWithTx relations=`%+v` fail", daoRelations)
	}

	// 更新subject system group
	err = l.BulkUpdateSubjectSystemGroupBySubjectTemplateGroupWithTx(tx, relations)
	if err != nil {
		return err
	}
	return nil
}

func (l *groupService) BulkUpdateSubjectSystemGroupBySubjectTemplateGroupWithTx(
	tx *sqlx.Tx,
	relations []types.SubjectTemplateGroup,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(
		GroupSVC,
		"BulkUpdateSubjectSystemGroupBySubjectTemplateGroupWithTx",
	)

	subjectSystemGroup := newSubjectSystemGroupMerger()
	groupSystemIDCache := make(map[int64][]string)
	for _, relation := range relations {
		if !relation.NeedUpdate {
			continue
		}

		systemIDs, ok := groupSystemIDCache[relation.GroupPK]
		if !ok {
			var err error
			systemIDs, err = l.ListGroupAuthSystemIDs(relation.GroupPK)
			if err != nil {
				return errorWrapf(err, "listGroupAuthSystem groupPK=`%d` fail", relation.GroupPK)
			}

			groupSystemIDCache[relation.GroupPK] = systemIDs
		}

		for _, systemID := range systemIDs {
			subjectSystemGroup.Add(
				relation.SubjectPK,
				systemID,
				relation.GroupPK,
				relation.ExpiredAt,
			)
		}
	}

	for key, groups := range subjectSystemGroup.subjectSystemGroup {
		subjectPK, systemID, err := subjectSystemGroup.ParseKey(key)
		if err != nil {
			return errorWrapf(err, "parseKey key=`%s` fail", key)
		}

		err = l.addOrUpdateSubjectSystemGroup(
			tx,
			subjectPK,
			systemID,
			groups,
		)
		if err != nil {
			return errorWrapf(
				err,
				"addOrUpdateSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, groups=`%v`, fail",
				systemID,
				subjectPK,
				groups,
			)
		}
	}

	return nil
}

func (l *groupService) UpdateSubjectGroupExpiredAtWithTx(
	tx *sqlx.Tx,
	relations []types.SubjectTemplateGroup,
	updateSubjectRelation bool,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "UpdateSubjectGroupExpiredAtWithTx")

	// 变更subject group的过期时间, 包括 subject relation 与 subject template group
	daoRelations := make([]dao.SubjectRelation, 0, len(relations))
	for _, r := range relations {
		daoRelations = append(daoRelations, dao.SubjectRelation{
			SubjectPK: r.SubjectPK,
			GroupPK:   r.GroupPK,
			ExpiredAt: r.ExpiredAt,
		})
	}

	if updateSubjectRelation {
		err := l.manager.BulkUpdateExpiredAtWithTx(tx, daoRelations)
		if err != nil {
			return errorWrapf(err, "manager.BulkUpdateExpiredAtWithTx relations=`%+v` fail", daoRelations)
		}
	}

	err := l.subjectTemplateGroupManager.BulkUpdateExpiredAtByRelationWithTx(tx, daoRelations)
	if err != nil {
		return errorWrapf(
			err,
			"subjectTemplateGroupManager.BulkUpdateExpiredAtByRelationWithTx relations=`%+v` fail",
			daoRelations,
		)
	}
	return nil
}

func (l *groupService) UpdateSubjectTemplateGroupExpiredAtWithTx(
	tx *sqlx.Tx,
	relations []types.SubjectTemplateGroup,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "UpdateSubjectTemplateGroupExpiredAtWithTx")

	// 变更subject group的过期时间, 包括 subject relation 与 subject template group
	daoRelations := convertToDaoSubjectTemplateGroup(relations)

	err := l.subjectTemplateGroupManager.BulkUpdateExpiredAtWithTx(tx, daoRelations)
	if err != nil {
		return errorWrapf(
			err,
			"subjectTemplateGroupManager.BulkUpdateExpiredAtWithTx relations=`%+v` fail",
			daoRelations,
		)
	}
	return nil
}

// BulkDeleteSubjectTemplateGroupWithTx ...
func (l *groupService) BulkDeleteSubjectTemplateGroupWithTx(
	tx *sqlx.Tx,
	relations []types.SubjectTemplateGroup,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteSubjectTemplateGroupWithTx")
	// 组装需要创建的Subject关系
	daoRelations := convertToDaoSubjectTemplateGroup(relations)

	err := l.subjectTemplateGroupManager.BulkDeleteWithTx(tx, daoRelations)
	if err != nil {
		return errorWrapf(err, "subjectTemplateGroupManager.BulkDeleteWithTx relations=`%+v` fail", daoRelations)
	}

	subjectSystemGroup := newSubjectSystemGroupMerger()
	groupSystemIDCache := make(map[int64][]string)
	for _, relation := range relations {
		if !relation.NeedUpdate {
			continue
		}

		// 更新subject system group
		systemIDs, ok := groupSystemIDCache[relation.GroupPK]
		if !ok {
			systemIDs, err = l.ListGroupAuthSystemIDs(relation.GroupPK)
			if err != nil {
				return errorWrapf(err, "listGroupAuthSystem groupPK=`%d` fail", relation.GroupPK)
			}

			groupSystemIDCache[relation.GroupPK] = systemIDs
		}

		for _, systemID := range systemIDs {
			subjectSystemGroup.Add(relation.SubjectPK, systemID, relation.GroupPK, 0)
		}
	}

	for key, groups := range subjectSystemGroup.subjectSystemGroup {
		subjectPK, systemID, err := subjectSystemGroup.ParseKey(key)
		if err != nil {
			return errorWrapf(err, "parseKey key=`%s` fail", key)
		}

		err = l.removeSubjectSystemGroup(
			tx,
			subjectPK,
			systemID,
			groups,
		)
		if err != nil {
			return errorWrapf(
				err,
				"removeSubjectSystemGroup systemID=`%s`, subjectPK=`%d`, groups=`%v`, fail",
				systemID,
				subjectPK,
				groups,
			)
		}
	}
	return nil
}

func convertToDaoSubjectTemplateGroup(relations []types.SubjectTemplateGroup) []dao.SubjectTemplateGroup {
	daoRelations := make([]dao.SubjectTemplateGroup, 0, len(relations))
	for _, r := range relations {
		daoRelations = append(daoRelations, dao.SubjectTemplateGroup{
			SubjectPK:  r.SubjectPK,
			TemplateID: r.TemplateID,
			GroupPK:    r.GroupPK,
			ExpiredAt:  r.ExpiredAt,
		})
	}
	return daoRelations
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

// ListPagingGroupSubjectBeforeExpiredAt ...
func (l *groupService) ListPagingGroupSubjectBeforeExpiredAt(
	expiredAt int64, limit, offset int64,
) ([]types.GroupSubject, error) {
	daoRelations, err := l.manager.ListPagingGroupSubjectBeforeExpiredAt(
		expiredAt, limit, offset,
	)
	if err != nil {
		return nil, errorx.Wrapf(err, GroupSVC,
			"ListPagingGroupSubjectBeforeExpiredAt", "expiredAt=`%d`, limit=`%d`, offset=`%d`",
			expiredAt, limit, offset)
	}

	return convertToGroupSubjects(daoRelations), nil
}

func convertToGroupSubjects(daoRelations []dao.ThinSubjectRelation) []types.GroupSubject {
	relations := make([]types.GroupSubject, 0, len(daoRelations))
	for _, r := range daoRelations {
		relations = append(relations, types.GroupSubject{
			SubjectPK: r.SubjectPK,
			GroupPK:   r.GroupPK,
			ExpiredAt: r.ExpiredAt,
		})
	}
	return relations
}

// BulkDeleteByGroupPKsWithTx ...
func (l *groupService) BulkDeleteByGroupPKsWithTx(tx *sqlx.Tx, groupPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteByGroupPKsWithTx")

	// 批量用户组删除成员关系 subjectRelation
	err := l.manager.BulkDeleteByGroupPKsWithTx(tx, groupPKs)
	if err != nil {
		return errorWrapf(
			err, "manager.BulkDeleteByGroupPKs group_pks=`%+v` fail", groupPKs)
	}

	return nil
}

// BulkDeleteBySubjectPKsWithTx ...
func (l *groupService) BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, subjectPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "BulkDeleteBySubjectPKsWithTx")

	// 批量其加入的用户组关系 subjectRelation
	err := l.manager.BulkDeleteBySubjectPKsWithTx(tx, subjectPKs)
	if err != nil {
		return errorWrapf(
			err, "manager.BulkDeleteBySubjectPKs subject_pks=`%+v` fail", subjectPKs)
	}

	// 批量删除subject template group
	err = l.subjectTemplateGroupManager.BulkDeleteBySubjectPKsWithTx(tx, subjectPKs)
	if err != nil {
		return errorWrapf(
			err,
			"subjectTemplateGroupManager.BulkDeleteBySubjectPKsWithTx subjectPKs=`%+v` fail",
			subjectPKs,
		)
	}

	// 批量删除用户的subject system group
	err = l.subjectSystemGroupManager.DeleteBySubjectPKsWithTx(tx, subjectPKs)
	if err != nil {
		return errorWrapf(err, "subjectSystemGroupManager.DeleteBySubjectPKsWithTx subjectPKs=`%+v` fail", subjectPKs)
	}

	return nil
}

// GetMaxExpiredAtBySubjectGroup ...
func (l *groupService) GetMaxExpiredAtBySubjectGroup(subjectPK, groupPK int64, excludeTemplateID int64) (int64, error) {
	// 同时查询 subject relation 与 subject template group, 取最大的过期时间
	subjectRelationExpiredAt, err1 := l.manager.GetExpiredAtBySubjectGroup(subjectPK, groupPK)
	if err1 != nil && !errors.Is(err1, sql.ErrNoRows) {
		err1 = errorx.Wrapf(
			err1, GroupSVC, "manager.GetExpiredAtBySubjectGroup", "subjectPK=`%d`, groupPK=`%d`", subjectPK, groupPK,
		)
		return 0, err1
	}

	subjectTemplateGroupExpiredAt, err2 := l.subjectTemplateGroupManager.GetMaxExpiredAtBySubjectGroup(
		subjectPK,
		groupPK,
		excludeTemplateID,
	)
	if err2 != nil && !errors.Is(err2, sql.ErrNoRows) {
		err2 = errorx.Wrapf(
			err2,
			GroupSVC,
			"subjectTemplateGroupManager.GetMaxExpiredAtBySubjectGroup",
			"subjectPK=`%d`, groupPK=`%d`, excludeTemplateID=`%d`",
			subjectPK,
			groupPK,
			excludeTemplateID,
		)
		return 0, err2
	}

	if errors.Is(err1, sql.ErrNoRows) && errors.Is(err2, sql.ErrNoRows) {
		return 0, ErrGroupMemberNotFound
	}

	// 取最大的过期时间
	if subjectRelationExpiredAt > subjectTemplateGroupExpiredAt {
		return subjectRelationExpiredAt, nil
	}

	return subjectTemplateGroupExpiredAt, nil
}

// subjectSystemGroupMerger 合并相同subject system的多个group同时变更, 用于subject template group的成员变更场景
type subjectSystemGroupMerger struct {
	subjectSystemGroup map[string]map[int64]int64 // key: subjectPK:systemID, map: groupPK-expiredAt
}

func newSubjectSystemGroupMerger() *subjectSystemGroupMerger {
	return &subjectSystemGroupMerger{
		subjectSystemGroup: make(map[string]map[int64]int64),
	}
}

// Add adds a group to the subjectSystemGroup map
func (h *subjectSystemGroupMerger) Add(subjectPK int64, systemID string, groupPK int64, expiredAt int64) {
	key := h.generateKey(subjectPK, systemID)
	if _, ok := h.subjectSystemGroup[key]; !ok {
		h.subjectSystemGroup[key] = make(map[int64]int64)
	}

	h.subjectSystemGroup[key][groupPK] = expiredAt
}

// generateKey generates a key based on subjectPK and systemID
func (h *subjectSystemGroupMerger) generateKey(subjectPK int64, systemID string) string {
	return fmt.Sprintf("%d:%s", subjectPK, systemID)
}

// ParseKey parses a key into subjectPK and systemID
func (h *subjectSystemGroupMerger) ParseKey(key string) (subjectPK int64, systemID string, err error) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid key format")
	}

	subjectPK, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", err
	}

	systemID = parts[1]
	return subjectPK, systemID, nil
}
