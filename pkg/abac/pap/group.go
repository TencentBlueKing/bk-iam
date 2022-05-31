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

	"github.com/TencentBlueKing/gopkg/errorx"

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

	GetMemberCount(_type, id string) (int64, error)
	ListPagingMember(_type, id string, limit, offset int64) ([]SubjectMember, error)
	GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error)
	ListPagingMemberBeforeExpiredAt(
		_type, id string, expiredAt int64, limit, offset int64,
	) ([]SubjectMember, error)

	CreateOrUpdateSubjectMembers(_type, id string, members []SubjectMember) (map[string]int64, error)
	UpdateSubjectMembersExpiredAt(_type, id string, members []SubjectMember) error
	DeleteSubjectMembers(_type, id string, members []Subject) (map[string]int64, error)
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

func (c *groupController) ListPagingMember(_type, id string, limit, offset int64) ([]SubjectMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "GetMemberCount")
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

	members, err := convertToSubjectMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectMembers svcSubjectMembers=`%+v` fail", svcMembers)
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
) ([]SubjectMember, error) {
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

	members, err := convertToSubjectMembers(svcMembers)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectMembers svcSubjectMembers=`%+v` fail", svcMembers)
	}

	return members, nil
}

// CreateOrUpdateSubjectMembers ...
func (c *groupController) CreateOrUpdateSubjectMembers(
	_type, id string,
	members []SubjectMember,
) (typeCount map[string]int64, err error) {
	return c.alterSubjectMembers(_type, id, members, true)
}

func (c *groupController) alterSubjectMembers(
	_type, id string,
	members []SubjectMember,
	createIfNotExists bool,
) (typeCount map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "CreateSubjectMembers")
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
	memberMap := make(map[int64]types.SubjectMember, len(relations))
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
		err = c.service.UpdateMembersExpiredAtWithTx(tx, updateMembers)
		if err != nil {
			err = errorWrapf(err, "service.UpdateMembersExpiredAtWithTx members=`%+v`", updateMembers)
			return
		}
	}

	// 无成员可添加，直接返回
	if createIfNotExists && len(createMembers) != 0 {
		// 添加成员
		err = c.service.BulkCreateSubjectMembersWithTx(tx, createMembers)
		if err != nil {
			err = errorWrapf(err, "service.BulkCreateSubjectMembersWithTx relations=`%+v`", createMembers)
			return nil, err
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, errorWrapf(err, "tx commit error")
	}

	// 清理缓存
	cacheimpls.BatchDeleteSubjectCache(subjectPKs)

	return typeCount, nil
}

// UpdateSubjectMembersExpiredAt ...
func (c *groupController) UpdateSubjectMembersExpiredAt(_type, id string, members []SubjectMember) (err error) {
	_, err = c.alterSubjectMembers(_type, id, members, false)
	return
}

// DeleteSubjectMembers ...
func (c *groupController) DeleteSubjectMembers(
	_type, id string,
	members []Subject,
) (typeCount map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "DeleteSubjectMembers")

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

	parenPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	typeCount, err = c.service.BulkDeleteSubjectMembers(parenPK, userPKs, departmentPKs)
	if err != nil {
		return nil, errorWrapf(
			err, "service.BulkDeleteSubjectMembers parenPK=`%s`, userPKs=`%+v`, departmentPKs=`%+v` failed",
			parenPK, userPKs, departmentPKs,
		)
	}

	// 清理缓存
	cacheimpls.BatchDeleteSubjectCache(append(userPKs, departmentPKs...))

	return typeCount, nil
}

func convertToSubjectGroups(svcSubjectGroups []types.SubjectGroup) ([]SubjectGroup, error) {
	groups := make([]SubjectGroup, 0, len(svcSubjectGroups))
	for _, m := range svcSubjectGroups {
		subject, err := cacheimpls.GetSubjectByPK(m.PK)
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

func convertToSubjectMembers(svcSubjectMembers []types.SubjectMember) ([]SubjectMember, error) {
	members := make([]SubjectMember, 0, len(svcSubjectMembers))
	for _, m := range svcSubjectMembers {
		subject, err := cacheimpls.GetSubjectByPK(m.SubjectPK)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, err
		}

		members = append(members, SubjectMember{
			PK:              m.PK,
			Type:            subject.Type,
			ID:              subject.ID,
			PolicyExpiredAt: m.PolicyExpiredAt,
			CreateAt:        m.CreateAt,
		})
	}

	return members, nil
}
