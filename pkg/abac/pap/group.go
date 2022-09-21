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

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// GroupCTL ...
const GroupCTL = "GroupCTL"

type GroupController interface {
	GetSubjectGroupCountBeforeExpiredAt(_type, id string, beforeExpiredAt int64) (int64, error)
	ListPagingSubjectGroups(_type, id string, beforeExpiredAt, limit, offset int64) ([]SubjectGroup, error)
	FilterGroupsHasMemberBeforeExpiredAt(subjects []Subject, expiredAt int64) ([]Subject, error)
	CheckSubjectEffectGroups(_type, id string, inherit bool, groupIDs []string) (map[string]bool, error)

	GetGroupMemberCount(_type, id string) (int64, error)
	ListPagingGroupMember(_type, id string, limit, offset int64) ([]GroupMember, error)
	GetGroupMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error)
	ListPagingGroupMemberBeforeExpiredAt(
		_type, id string, expiredAt int64, limit, offset int64,
	) ([]GroupMember, error)
	GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (count int64, err error)
	ListPagingGroupSubjectBeforeExpiredAt(expiredAt int64, limit, offset int64) ([]GroupSubject, error)

	CreateOrUpdateGroupMembers(_type, id string, members []GroupMember) (map[string]int64, error)
	UpdateGroupMembersExpiredAt(_type, id string, members []GroupMember) error
	DeleteGroupMembers(_type, id string, members []Subject) (map[string]int64, error)
}

type groupController struct {
	service service.GroupService

	subjectService         service.SubjectService
	groupAlterEventService service.GroupAlterEventService
}

func NewGroupController() GroupController {
	return &groupController{
		service:                service.NewGroupService(),
		subjectService:         service.NewSubjectService(),
		groupAlterEventService: service.NewGroupAlterEventService(),
	}
}

// GetSubjectGroupCountBeforeExpiredAt ...
func (c *groupController) GetSubjectGroupCountBeforeExpiredAt(
	_type, id string,
	expiredAt int64,
) (count int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetSubjectGroupCountBeforeExpiredAt")
	subjectPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err = c.service.GetSubjectGroupCountBeforeExpiredAt(subjectPK, expiredAt)
	if err != nil {
		return 0, errorWrapf(
			err,
			"service.GetSubjectGroupCountBeforeExpiredAt subjectPK=`%s`, expiredAt=`%d`",
			subjectPK,
			expiredAt,
		)
	}

	return count, nil
}

// GetGroupSubjectCountBeforeExpiredAt ...
func (c *groupController) GetGroupSubjectCountBeforeExpiredAt(expiredAt int64) (count int64, err error) {
	return c.service.GetGroupSubjectCountBeforeExpiredAt(expiredAt)
}

func (c *groupController) FilterGroupsHasMemberBeforeExpiredAt(subjects []Subject, expiredAt int64) ([]Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "FilterGroupsHasMemberBeforeExpiredAt")

	svcSubjects := convertToServiceSubjects(subjects)
	groupPKs, err := c.subjectService.ListPKsBySubjects(svcSubjects)
	if err != nil {
		return nil, errorWrapf(err, "service.ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	existGroupPKs, err := c.service.FilterGroupPKsHasMemberBeforeExpiredAt(groupPKs, expiredAt)
	if err != nil {
		return nil, errorWrapf(
			err, "service.FilterGroupPKsHasMemberBeforeExpiredAt groupPKs=`%+v`, expiredAt=`%d` fail",
			groupPKs, expiredAt,
		)
	}

	existGroups := make([]Subject, 0, len(existGroupPKs))
	for _, pk := range existGroupPKs {
		subject, err := cacheimpls.GetSubjectByPK(pk)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, errorWrapf(err, "cacheimpls.GetSubjectByPK pk=`%d` fail", pk)
		}

		existGroups = append(existGroups, Subject{
			Type: subject.Type,
			ID:   subject.ID,
			Name: subject.Name,
		})
	}

	return existGroups, nil
}

func (c *groupController) CheckSubjectEffectGroups(
	_type, id string,
	inherit bool,
	groupIDs []string,
) (map[string]bool, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "CheckSubjectExistGroups")

	// subject Type+ID to PK
	subjectPK, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	subjectPKs := []int64{subjectPK}
	if inherit && _type == types.UserType {
		departmentPKs, err := cacheimpls.GetSubjectDepartmentPKs(subjectPK)
		if err != nil {
			return nil, errorWrapf(err, "cacheimpls.GetSubjectDepartmentPKs subjectPK=`%d` fail", subjectPK)
		}

		subjectPKs = append(subjectPKs, departmentPKs...)
	}

	groupIDToGroupPK := make(map[string]int64, len(groupIDs))
	groupPKs := make([]int64, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		// if groupID is empty, skip
		if groupID == "" {
			continue
		}

		// get the groupPK via groupID
		groupPK, err := cacheimpls.GetLocalSubjectPK(types.GroupType, groupID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.WithError(err).Debugf("cacheimpls.GetSubjectPK type=`group`, id=`%s` fail", groupID)
				continue
			}

			return nil, errorWrapf(
				err,
				"cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail",
				types.GroupType,
				groupID,
			)
		}

		groupPKs = append(groupPKs, groupPK)
		groupIDToGroupPK[groupID] = groupPK
	}

	// NOTE: if the performance is a problem, change this to a local cache, key: subjectPK, value int64Set
	effectGroupPKs, err := c.service.FilterExistEffectSubjectGroupPKs(subjectPKs, groupPKs)
	if err != nil {
		return nil, errorWrapf(
			err,
			"service.FilterExistEffectSubjectGroupPKs subjectPKs=`%+v`, groupPKs=`%+v` fail",
			subjectPKs,
			groupPKs,
		)
	}
	existGroupPKSet := set.NewInt64SetWithValues(effectGroupPKs)

	// the result
	groupIDBelong := make(map[string]bool, len(groupIDs))
	for _, groupID := range groupIDs {
		groupPK, ok := groupIDToGroupPK[groupID]
		if !ok {
			groupIDBelong[groupID] = false
			continue
		}
		groupIDBelong[groupID] = existGroupPKSet.Has(groupPK)
	}

	return groupIDBelong, nil
}

// ListPagingSubjectGroups ...
func (c *groupController) ListPagingSubjectGroups(
	_type, id string,
	beforeExpiredAt, limit, offset int64,
) ([]SubjectGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingSubjectGroups")
	subjectPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcSubjectGroups, err := c.service.ListPagingSubjectGroups(subjectPK, beforeExpiredAt, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingSubjectGroups subjectPK=`%s`, beforeExpiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			subjectPK, beforeExpiredAt, limit, offset,
		)
	}

	groups, err := convertToSubjectGroups(svcSubjectGroups)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectGroups svcSubjectGroups=`%+v` fail", svcSubjectGroups)
	}

	return groups, nil
}

// GetGroupMemberCount ...
func (c *groupController) GetGroupMemberCount(_type, id string) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetGroupMemberCount")
	groupPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := c.service.GetGroupMemberCount(groupPK)
	if err != nil {
		return 0, errorWrapf(err, "service.GetGroupMemberCount groupPK=`%d`", groupPK)
	}

	return count, nil
}

// ListPagingGroupMember ...
func (c *groupController) ListPagingGroupMember(_type, id string, limit, offset int64) ([]GroupMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingGroupMember")
	groupPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcMembers, err := c.service.ListPagingGroupMember(groupPK, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingGroupMember groupPK=`%d`, limit=`%d`, offset=`%d` fail",
			groupPK, limit, offset,
		)
	}

	members, err := convertToGroupMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupMembers svcMembers=`%+v` fail", svcMembers)
	}

	return members, nil
}

// ListPagingGroupSubjectBeforeExpiredAt ...
func (c *groupController) ListPagingGroupSubjectBeforeExpiredAt(
	expiredAt int64,
	limit, offset int64,
) ([]GroupSubject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingGroupSubjectBeforeExpiredAt")

	svcRelations, err := c.service.ListPagingGroupSubjectBeforeExpiredAt(expiredAt, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingGroupSubjectBeforeExpiredAt expiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			expiredAt, limit, offset,
		)
	}

	relations, err := convertToGroupSubjects(svcRelations)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupSubjects svcRelations=`%+v` fail", svcRelations)
	}

	return relations, nil
}

// GetGroupMemberCountBeforeExpiredAt ...
func (c *groupController) GetGroupMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetGroupMemberCountBeforeExpiredAt")
	groupPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := c.service.GetGroupMemberCountBeforeExpiredAt(groupPK, expiredAt)
	if err != nil {
		return 0, errorWrapf(
			err, "service.GetGroupMemberCountBeforeExpiredAt groupPK=`%d`, expiredAt=`%d`",
			groupPK, expiredAt,
		)
	}

	return count, nil
}

// ListPagingGroupMemberBeforeExpiredAt ...
func (c *groupController) ListPagingGroupMemberBeforeExpiredAt(
	_type, id string, expiredAt int64, limit, offset int64,
) ([]GroupMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingGroupMemberBeforeExpiredAt")
	groupPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcMembers, err := c.service.ListPagingGroupMemberBeforeExpiredAt(groupPK, expiredAt, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err,
			"service.ListPagingGroupMemberBeforeExpiredAt groupPK=`%d`, expiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			groupPK,
			expiredAt,
			limit,
			offset,
		)
	}

	members, err := convertToGroupMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupMembers svcMembers=`%+v` fail", svcMembers)
	}

	return members, nil
}

// CreateOrUpdateGroupMembers ...
func (c *groupController) CreateOrUpdateGroupMembers(
	_type, id string,
	members []GroupMember,
) (typeCount map[string]int64, err error) {
	return c.alterGroupMembers(_type, id, members, true)
}

func (c *groupController) alterGroupMembers(
	_type, id string,
	members []GroupMember,
	createIfNotExists bool,
) (typeCount map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "alterGroupMembers")
	groupPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	relations, err := c.service.ListGroupMember(groupPK)
	if err != nil {
		err = errorWrapf(err, "service.ListGroupMember type=`%s` id=`%s`", _type, id)
		return
	}

	// 重复和已经存在DB里的不需要
	memberMap := make(map[int64]types.GroupMember, len(relations))
	for _, m := range relations {
		memberMap[m.SubjectPK] = m
	}

	// 获取实际需要添加的member
	createMembers := make([]types.SubjectRelationForCreate, 0, len(members))

	// 需要更新过期时间的member
	updateMembers := make([]types.SubjectRelationForUpdate, 0, len(members))

	// 用于清理缓存
	subjectPKs := make([]int64, 0, len(members))

	typeCount = map[string]int64{
		types.UserType:       0,
		types.DepartmentType: 0,
	}

	for _, m := range members {
		subjectPK, err := cacheimpls.GetSubjectPK(m.Type, m.ID)
		if err != nil {
			return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", m.Type, m.ID)
		}

		// member已存在则不再添加
		if oldMember, ok := memberMap[subjectPK]; ok {
			// 如果过期时间大于已有的时间, 则更新过期时间
			if m.ExpiredAt > oldMember.ExpiredAt {
				updateMembers = append(updateMembers, types.SubjectRelationForUpdate{
					PK:        oldMember.PK,
					SubjectPK: subjectPK,
					ExpiredAt: m.ExpiredAt,
				})

				subjectPKs = append(subjectPKs, subjectPK)
			}
			continue
		}

		if createIfNotExists {
			createMembers = append(createMembers, types.SubjectRelationForCreate{
				SubjectPK: subjectPK,
				GroupPK:   groupPK,
				ExpiredAt: m.ExpiredAt,
			})
			typeCount[m.Type]++
			subjectPKs = append(subjectPKs, subjectPK)
		}
	}

	// 按照PK删除Subject所有相关的
	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return nil, errorWrapf(err, "define tx error")
	}

	if len(updateMembers) != 0 {
		// 更新成员过期时间
		err = c.service.UpdateGroupMembersExpiredAtWithTx(tx, groupPK, updateMembers)
		if err != nil {
			err = errorWrapf(err, "service.UpdateGroupMembersExpiredAtWithTx members=`%+v`", updateMembers)
			return
		}
	}

	// 无成员可添加，直接返回
	if createIfNotExists && len(createMembers) != 0 {
		// 添加成员
		err = c.service.BulkCreateGroupMembersWithTx(tx, groupPK, createMembers)
		if err != nil {
			err = errorWrapf(err, "service.BulkCreateGroupMembersWithTx relations=`%+v`", createMembers)
			return nil, err
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, errorWrapf(err, "tx commit error")
	}

	// 创建group_alter_event
	c.createGroupAlterEvent(groupPK, subjectPKs)

	// 清理缓存
	cacheimpls.BatchDeleteSubjectGroupCache(subjectPKs)

	// 清理subject system group 缓存
	cacheimpls.BatchDeleteSubjectAuthSystemGroupCache(subjectPKs, groupPK)

	return typeCount, nil
}

// UpdateGroupMembersExpiredAt ...
func (c *groupController) UpdateGroupMembersExpiredAt(_type, id string, members []GroupMember) (err error) {
	_, err = c.alterGroupMembers(_type, id, members, false)
	return
}

// DeleteGroupMembers ...
func (c *groupController) DeleteGroupMembers(
	_type, id string,
	members []Subject,
) (typeCount map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "DeleteGroupMembers")

	userPKs := make([]int64, 0, len(members))
	departmentPKs := make([]int64, 0, len(members))
	for _, m := range members {
		pk, err := cacheimpls.GetSubjectPK(m.Type, m.ID)
		if err != nil {
			return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", m.Type, m.ID)
		}

		if m.Type == types.UserType {
			userPKs = append(userPKs, pk)
		} else if m.Type == types.DepartmentType {
			departmentPKs = append(departmentPKs, pk)
		}
	}

	groupPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	typeCount, err = c.service.BulkDeleteGroupMembers(groupPK, userPKs, departmentPKs)
	if err != nil {
		return nil, errorWrapf(
			err, "service.BulkDeleteGroupMembers groupPK=`%d`, userPKs=`%+v`, departmentPKs=`%+v` failed",
			groupPK, userPKs, departmentPKs,
		)
	}

	// 清理缓存
	subjectPKs := make([]int64, 0, len(members))
	subjectPKs = append(subjectPKs, userPKs...)
	subjectPKs = append(subjectPKs, departmentPKs...)

	// 创建group_alter_event
	c.createGroupAlterEvent(groupPK, subjectPKs)

	cacheimpls.BatchDeleteSubjectGroupCache(subjectPKs)

	// group auth system
	cacheimpls.BatchDeleteSubjectAuthSystemGroupCache(subjectPKs, groupPK)

	return typeCount, nil
}

func (c *groupController) createGroupAlterEvent(groupPK int64, subjectPKs []int64) {
	err := c.groupAlterEventService.CreateByGroupSubject(groupPK, subjectPKs)
	if err != nil {
		log.WithError(err).
			Errorf("groupAlterEventService.CreateByGroupSubject groupPK=%d subjectPKs=%v fail", groupPK, subjectPKs)

		// report to sentry
		util.ReportToSentry("createGroupAlterEvent groupAlterEventService.CreateByGroupSubject fail",
			map[string]interface{}{
				"layer":      GroupCTL,
				"groupPK":    groupPK,
				"subjectPKs": subjectPKs,
				"error":      err.Error(),
			},
		)
	}
}

func convertToSubjectGroups(svcSubjectGroups []types.SubjectGroup) ([]SubjectGroup, error) {
	groups := make([]SubjectGroup, 0, len(svcSubjectGroups))
	for _, m := range svcSubjectGroups {
		subject, err := cacheimpls.GetSubjectByPK(m.GroupPK)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, err
		}

		groups = append(groups, SubjectGroup{
			PK:        m.PK,
			Type:      subject.Type,
			ID:        subject.ID,
			ExpiredAt: m.ExpiredAt,
			CreatedAt: m.CreatedAt,
		})
	}

	return groups, nil
}

func convertToGroupMembers(svcGroupMembers []types.GroupMember) ([]GroupMember, error) {
	members := make([]GroupMember, 0, len(svcGroupMembers))
	for _, m := range svcGroupMembers {
		subject, err := cacheimpls.GetSubjectByPK(m.SubjectPK)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, err
		}

		members = append(members, GroupMember{
			PK:        m.PK,
			Type:      subject.Type,
			ID:        subject.ID,
			ExpiredAt: m.ExpiredAt,
			CreatedAt: m.CreatedAt,
		})
	}

	return members, nil
}

func convertToGroupSubjects(svcGroupSubjects []types.GroupSubject) ([]GroupSubject, error) {
	groupSubjects := make([]GroupSubject, 0, len(svcGroupSubjects))
	for _, m := range svcGroupSubjects {
		subject, err := cacheimpls.GetSubjectByPK(m.SubjectPK)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, err
		}

		group, err := cacheimpls.GetSubjectByPK(m.GroupPK)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, err
		}

		groupSubjects = append(groupSubjects, GroupSubject{
			PK: m.PK,
			Subject: Subject{
				Type: subject.Type,
				ID:   subject.ID,
				Name: subject.Name,
			},
			Group: Subject{
				Type: group.Type,
				ID:   group.ID,
				Name: group.Name,
			},
			ExpiredAt: m.ExpiredAt,
			CreatedAt: m.CreatedAt,
		})
	}

	return groupSubjects, nil
}
