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
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// GroupCTL ...
const GroupCTL = "GroupCTL"

type GroupController interface {
	ListSubjectGroups(_type, id string, beforeExpiredAt int64) ([]SubjectGroup, error)
	ListExistSubjectsBeforeExpiredAt(subjects []Subject, expiredAt int64) ([]Subject, error)
	CheckSubjectEffectGroups(_type, id string, inherit bool, groupIDs []string) (map[string]bool, error)

	GetMemberCount(_type, id string) (int64, error)
	ListPagingMember(_type, id string, limit, offset int64) ([]GroupMember, error)
	GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error)
	ListPagingMemberBeforeExpiredAt(
		_type, id string, expiredAt int64, limit, offset int64,
	) ([]GroupMember, error)

	CreateOrUpdateGroupMembers(_type, id string, members []GroupMember) (map[string]int64, error)
	UpdateGroupMembersExpiredAt(_type, id string, members []GroupMember) error
	DeleteGroupMembers(_type, id string, members []Subject) (map[string]int64, error)
}

type groupController struct {
	service service.GroupService

	subjectService service.SubjectService
}

func NewGroupController() GroupController {
	return &groupController{
		service:        service.NewGroupService(),
		subjectService: service.NewSubjectService(),
	}
}

func (c *groupController) ListExistSubjectsBeforeExpiredAt(subjects []Subject, expiredAt int64) ([]Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListExistSubjectsBeforeExpiredAt")

	svcSubjects := convertToServiceSubjects(subjects)
	subjectPKs, err := c.subjectService.ListPKsBySubjects(svcSubjects)
	if err != nil {
		return nil, errorWrapf(err, "service.ListPKsBySubjects subjects=`%+v` fail", subjects)
	}

	existSubjectPKs, err := c.service.ListExistSubjectsBeforeExpiredAt(subjectPKs, expiredAt)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListExistSubjectsBeforeExpiredAt subjectPKs=`%+v`, expiredAt=`%d` fail",
			subjectPKs, expiredAt,
		)
	}

	existSubjects := make([]Subject, 0, len(existSubjectPKs))
	for _, pk := range existSubjectPKs {
		subject, err := cacheimpls.GetSubjectByPK(pk)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, errorWrapf(err, "cacheimpls.GetSubjectByPK pk=`%d` fail", pk)
		}

		existSubjects = append(existSubjects, Subject{
			Type: subject.Type,
			ID:   subject.ID,
			Name: subject.Name,
		})
	}

	return existSubjects, nil
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
	effectGroupPKs, err := c.service.ListExistEffectSubjectGroupPKs(subjectPKs, groupPKs)
	if err != nil {
		return nil, errorWrapf(
			err,
			"service.ListExistEffectSubjectGroupPKs subjectPKs=`%+v`, groupPKs=`%+v` fail",
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
		}
		groupIDBelong[groupID] = existGroupPKSet.Has(groupPK)
	}

	return groupIDBelong, nil
}

// ListSubjectGroups ...
func (c *groupController) ListSubjectGroups(_type, id string, beforeExpiredAt int64) ([]SubjectGroup, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListSubjectGroups")
	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcSubjectGroups, err := c.service.ListSubjectGroups(parentPK, beforeExpiredAt)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListSubjectGroups parentPK=`%s`, beforeExpiredAt=`%d` fail",
			parentPK, beforeExpiredAt,
		)
	}

	groups, err := convertToSubjectGroups(svcSubjectGroups)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectGroups svcSubjectGroups=`%+v` fail", svcSubjectGroups)
	}

	return groups, nil
}

// GetMemberCount ...
func (c *groupController) GetMemberCount(_type, id string) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetMemberCount")
	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := c.service.GetMemberCount(parentPK)
	if err != nil {
		return 0, errorWrapf(err, "service.GetMemberCount parentPK=`%s`", parentPK)
	}

	return count, nil
}

func (c *groupController) ListPagingMember(_type, id string, limit, offset int64) ([]GroupMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingMember")
	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcMembers, err := c.service.ListPagingMember(parentPK, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingMember parentPK=`%d`, limit=`%d`, offset=`%d` fail",
			parentPK, limit, offset,
		)
	}

	members, err := convertToGroupMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToGroupMembers svcMembers=`%+v` fail", svcMembers)
	}

	return members, nil
}

// GetMemberCountBeforeExpiredAt ...
func (c *groupController) GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetMemberCountBeforeExpiredAt")
	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := c.service.GetMemberCountBeforeExpiredAt(parentPK, expiredAt)
	if err != nil {
		return 0, errorWrapf(
			err, "service.GetMemberCountBeforeExpiredAt parentPK=`%s`, expiredAt=`%d`",
			parentPK, expiredAt,
		)
	}

	return count, nil
}

// ListPagingMemberBeforeExpiredAt ...
func (c *groupController) ListPagingMemberBeforeExpiredAt(
	_type, id string, expiredAt int64, limit, offset int64,
) ([]GroupMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "ListPagingMemberBeforeExpiredAt")
	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	svcMembers, err := c.service.ListPagingMemberBeforeExpiredAt(parentPK, expiredAt, limit, offset)
	if err != nil {
		return nil, errorWrapf(
			err, "service.ListPagingMemberBeforeExpiredAt parentPK=`%d`, expiredAt=`%d`, limit=`%d`, offset=`%d` fail",
			parentPK, expiredAt, limit, offset,
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
	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	relations, err := c.service.ListMember(parentPK)
	if err != nil {
		err = errorWrapf(err, "service.ListMember type=`%s` id=`%s`", _type, id)
		return
	}

	// 重复和已经存在DB里的不需要
	memberMap := make(map[int64]types.GroupMember, len(relations))
	for _, m := range relations {
		memberMap[m.SubjectPK] = m
	}

	// 获取实际需要添加的member
	createMembers := make([]types.SubjectRelation, 0, len(members))

	// 需要更新过期时间的member
	updateMembers := make([]types.SubjectRelationPKPolicyExpiredAt, 0, len(members))

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
			if m.PolicyExpiredAt > oldMember.PolicyExpiredAt {
				updateMembers = append(updateMembers, types.SubjectRelationPKPolicyExpiredAt{
					PK:              oldMember.PK,
					SubjectPK:       subjectPK,
					PolicyExpiredAt: m.PolicyExpiredAt,
				})

				subjectPKs = append(subjectPKs, subjectPK)
			}
			continue
		}

		if createIfNotExists {
			createMembers = append(createMembers, types.SubjectRelation{
				SubjectPK:       subjectPK,
				ParentPK:        parentPK,
				PolicyExpiredAt: m.PolicyExpiredAt,
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
		err = c.service.UpdateMembersExpiredAtWithTx(tx, parentPK, updateMembers)
		if err != nil {
			err = errorWrapf(err, "service.UpdateMembersExpiredAtWithTx members=`%+v`", updateMembers)
			return
		}
	}

	// 无成员可添加，直接返回
	if createIfNotExists && len(createMembers) != 0 {
		// 添加成员
		err = c.service.BulkCreateGroupMembersWithTx(tx, parentPK, createMembers)
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

	// 清理缓存
	cacheimpls.BatchDeleteSubjectGroupCache(subjectPKs)

	// 清理subject system group 缓存
	cacheimpls.BatchDeleteSubjectAuthSystemGroupCache(subjectPKs, parentPK)

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

	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	typeCount, err = c.service.BulkDeleteGroupMembers(parentPK, userPKs, departmentPKs)
	if err != nil {
		return nil, errorWrapf(
			err, "service.BulkDeleteGroupMembers parenPK=`%s`, userPKs=`%+v`, departmentPKs=`%+v` failed",
			parentPK, userPKs, departmentPKs,
		)
	}

	// 清理缓存
	subjectPKs := make([]int64, 0, len(members))
	subjectPKs = append(subjectPKs, userPKs...)
	subjectPKs = append(subjectPKs, departmentPKs...)

	cacheimpls.BatchDeleteSubjectGroupCache(subjectPKs)

	// group auth system
	cacheimpls.BatchDeleteSubjectAuthSystemGroupCache(subjectPKs, parentPK)

	return typeCount, nil
}

func convertToSubjectGroups(svcSubjectGroups []types.SubjectGroup) ([]SubjectGroup, error) {
	groups := make([]SubjectGroup, 0, len(svcSubjectGroups))
	for _, m := range svcSubjectGroups {
		subject, err := cacheimpls.GetSubjectByPK(m.ParentPK)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, err
		}

		groups = append(groups, SubjectGroup{
			PK:              m.PK,
			Type:            subject.Type,
			ID:              subject.ID,
			PolicyExpiredAt: m.PolicyExpiredAt,
			CreateAt:        m.CreateAt,
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
			PK:              m.PK,
			Type:            subject.Type,
			ID:              subject.ID,
			PolicyExpiredAt: m.PolicyExpiredAt,
			CreateAt:        m.CreateAt,
		})
	}

	return members, nil
}
