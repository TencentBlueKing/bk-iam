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
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// GroupCTL ...
const GroupCTL = "GroupCTL"

type GroupController interface {
	CreateOrUpdateSubjectMembers(_type, id string, members []SubjectMember) (map[string]int64, error)
	UpdateSubjectMembersExpiredAt(_type, id string, members []SubjectMember) error
	DeleteSubjectMembers(_type, id string, members []Subject) (map[string]int64, error)
}

type groupController struct {
	service service.GroupService
}

func NewGroupController() GroupController {
	return &groupController{
		service: service.NewGroupService(),
	}
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
	relations, err := c.service.ListMember(_type, id)
	if err != nil {
		err = errorWrapf(err, "service.ListMember type=`%s` id=`%s`", _type, id)
		return
	}

	// 重复和已经存在DB里的不需要
	memberMap := make(map[string]types.SubjectMember, len(relations))
	for _, m := range relations {
		memberMap[fmt.Sprintf("%s:%s", m.Type, m.ID)] = m
	}

	// 获取实际需要添加的member
	createMembers := make([]SubjectMember, 0, len(members))

	// 需要更新过期时间的member
	updateMembers := make([]types.SubjectRelationPKPolicyExpiredAt, 0, len(members))

	typeCount = map[string]int64{
		types.UserType:       0,
		types.DepartmentType: 0,
	}

	bodyMembers := set.NewStringSet() // 用于去重

	for _, m := range members {
		key := fmt.Sprintf("%s:%s", m.Type, m.ID)

		// 对Body Member参数去重
		if bodyMembers.Has(key) {
			continue
		}
		bodyMembers.Add(key)

		// member已存在则不再添加
		if oldMember, ok := memberMap[key]; ok {
			// 如果过期时间大于已有的时间, 则更新过期时间
			if m.PolicyExpiredAt > oldMember.PolicyExpiredAt {
				updateMembers = append(updateMembers, types.SubjectRelationPKPolicyExpiredAt{
					PK:              oldMember.PK,
					PolicyExpiredAt: m.PolicyExpiredAt,
				})
			}
			continue
		}

		if createIfNotExists {
			createMembers = append(createMembers, m)
			typeCount[m.Type]++
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
		err = c.bulkCreateSubjectMembers(tx, _type, id, createMembers)
		if err != nil {
			err = errorWrapf(
				err, "bulkCreateSubjectMembers type=`%s` id=`%s` members=`%+v` policy_expired_at=`%d`",
				_type, id, createMembers,
			)
			return
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, errorWrapf(err, "tx commit error")
	}

	// 清理缓存
	subjectPKs := make([]int64, 0, len(members))
	for _, m := range members {
		subjectPK, _ := cacheimpls.GetSubjectPK(m.Type, m.ID)
		subjectPKs = append(subjectPKs, subjectPK)
	}
	cacheimpls.BatchDeleteSubjectCache(subjectPKs)

	return typeCount, nil
}

func (c *groupController) bulkCreateSubjectMembers(tx *sqlx.Tx, _type, id string, members []SubjectMember) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupCTL, "bulkCreateSubjectMembers")
	// 查询subject PK
	parentPK, err := cacheimpls.GetSubjectPK(_type, id)
	if err != nil {
		return errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	// 需要type, id换成pk
	relations := make([]types.SubjectRelation, 0, len(members))
	for _, m := range members {
		// TODO 优化批量查询缓存
		subjectPK, err := cacheimpls.GetSubjectPK(m.Type, m.ID)
		if err != nil {
			return errorWrapf(err, "cacheimpls.GetSubjectPK _type=`%s`, id=`%s` fail", m.Type, m.ID)
		}

		relations = append(relations, types.SubjectRelation{
			SubjectPK:       subjectPK,
			SubjectID:       m.ID,
			SubjectType:     m.Type,
			ParentPK:        parentPK,
			ParentType:      _type,
			ParentID:        id,
			PolicyExpiredAt: m.PolicyExpiredAt,
		})
	}

	err = c.service.BulkCreateSubjectMembersWithTx(tx, relations)
	if err != nil {
		return errorWrapf(err, "service.BulkCreateSubjectMembersWithTx relations=`%+v`", relations)
	}

	return nil
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
	svcSubjects := convertToServiceSubjects(members)

	typeCount, err = c.service.BulkDeleteSubjectMembers(_type, id, svcSubjects)
	if err != nil {
		return nil, errorWrapf(
			err, "service.BulkDeleteSubjectMembers _type=`%s`, id=`%s`, subjects=`%+v` failed",
			_type, id, svcSubjects,
		)
	}

	// 清理缓存
	subjectPKs := make([]int64, 0, len(members))
	for _, m := range members {
		subjectPK, _ := cacheimpls.GetSubjectPK(m.Type, m.ID)
		subjectPKs = append(subjectPKs, subjectPK)
	}
	cacheimpls.BatchDeleteSubjectCache(subjectPKs)

	return typeCount, nil
}
