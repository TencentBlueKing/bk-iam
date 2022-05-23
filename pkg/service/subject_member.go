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

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// GetMemberCount ...
func (l *subjectService) GetMemberCount(_type, id string) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "GetMemberCount")
	// TODO 后续通过缓存提高性能
	pk, err := l.manager.GetPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := l.relationManager.GetMemberCount(pk)
	if err != nil {
		err = errorWrapf(err, "relationManager.GetMemberCount _type=`%s`, id=`%s` fail", _type, id)
		return 0, err
	}
	return count, nil
}

// ListPagingMember ...
func (l *subjectService) ListPagingMember(_type, id string, limit, offset int64) ([]types.SubjectMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListPagingMember")
	// 查询subject PK
	pk, err := l.manager.GetPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

	daoRelations, err := l.relationManager.ListPagingMember(pk, limit, offset)
	if err != nil {
		return nil, errorWrapf(err, "relationManager.ListPagingMember _type=`%s`, id=`%s`, limit=`%d`, offset=`%d`",
			_type, id, limit, offset)
	}

	members, err := l.convertToSubjectMembers(daoRelations)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectMembers relations=`%s`", members)
	}

	return members, nil
}

func (l *subjectService) getSubjectMapByPKs(pks []int64) (map[int64]dao.Subject, error) {
	if len(pks) == 0 {
		return nil, nil
	}

	subjects, err := l.manager.ListByPKs(pks)
	if err != nil {
		return nil, err
	}

	subjectMap := make(map[int64]dao.Subject, len(subjects))
	for _, s := range subjects {
		subjectMap[s.PK] = s
	}
	return subjectMap, nil
}

func (l *subjectService) convertToSubjectMembers(daoRelations []dao.SubjectRelation) ([]types.SubjectMember, error) {
	if len(daoRelations) == 0 {
		return nil, nil
	}

	subjectPKs := make([]int64, 0, len(daoRelations))
	for _, r := range daoRelations {
		subjectPKs = append(subjectPKs, r.SubjectPK)
	}

	// TODO 后续通过缓存提高性能
	subjectMap, err := l.getSubjectMapByPKs(subjectPKs)
	if err != nil {
		return nil, err
	}

	members := make([]types.SubjectMember, 0, len(daoRelations))
	for _, r := range daoRelations {
		var _type, id string
		subject, ok := subjectMap[r.SubjectPK]
		if ok {
			_type = subject.Type
			id = subject.ID
		}

		members = append(members, types.SubjectMember{
			PK:              r.PK,
			Type:            _type,
			ID:              id,
			PolicyExpiredAt: r.PolicyExpiredAt,
			CreateAt:        r.CreateAt,
		})
	}
	return members, nil
}

// ListMember ...
func (l *subjectService) ListMember(_type, id string) ([]types.SubjectMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListMember")
	// 查询subject PK
	pk, err := l.manager.GetPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

	daoRelations, err := l.relationManager.ListMember(pk)
	if err != nil {
		return nil, errorx.Wrapf(err, SubjectSVC,
			"ListMember", "relationManager.ListMember _type=`%s`, id=`%s` fail", _type, id)
	}

	members, err := l.convertToSubjectMembers(daoRelations)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectMembers relations=`%s`", members)
	}

	return members, nil
}

// UpdateMembersExpiredAt ...
func (l *subjectService) UpdateMembersExpiredAt(members []types.SubjectMember) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkDeleteSubjectMember")

	relations := make([]dao.SubjectRelationPKPolicyExpiredAt, 0, len(members))
	for _, m := range members {
		relations = append(relations, dao.SubjectRelationPKPolicyExpiredAt{
			PK:              m.PK,
			PolicyExpiredAt: m.PolicyExpiredAt,
		})
	}

	err := l.relationManager.UpdateExpiredAt(relations)
	if err != nil {
		err = errorWrapf(err,
			"relationManager.UpdateExpiredAt relations=`%+v` fail", relations)
		return err
	}

	// TODO 更新subject_system_groups表的groups字段
	// 1. 更新一个subject的group的过期时间

	return nil
}

// BulkDeleteSubjectMembers ...
func (l *subjectService) BulkDeleteSubjectMembers(_type, id string, members []types.Subject) (map[string]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkDeleteSubjectMember")

	// 查询subject PK
	parentPK, err := l.manager.GetPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

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
		users, newErr := l.manager.ListByIDs(types.UserType, userIDs)
		if newErr != nil {
			return nil, errorWrapf(newErr, "manager.ListByIDs _type=`%s`, ids=`%+v` fail", types.UserType, userIDs)
		}
		subjectPKs := make([]int64, 0, len(users))
		for _, u := range users {
			subjectPKs = append(subjectPKs, u.PK)
		}

		count, err = l.relationManager.BulkDeleteByMembersWithTx(tx, parentPK, subjectPKs)
		if err != nil {
			return nil, errorWrapf(err,
				"relationManager.BulkDeleteByMembersWithTx _type=`%s`, id=`%s`, subjectType=`%s`, subjectIDs=`%+v` fail",
				_type, id, types.UserType, userIDs)
		}
		typeCount[types.UserType] = count
	}

	if len(departmentIDs) != 0 {
		departments, newErr := l.manager.ListByIDs(types.DepartmentType, departmentIDs)
		if newErr != nil {
			return nil, errorWrapf(newErr, "manager.ListByIDs _type=`%s`, ids=`%+v` fail", types.DepartmentType, departmentIDs)
		}
		subjectPKs := make([]int64, 0, len(departments))
		for _, u := range departments {
			subjectPKs = append(subjectPKs, u.PK)
		}

		count, err = l.relationManager.BulkDeleteByMembersWithTx(tx, parentPK, subjectPKs)
		if err != nil {
			return nil, errorWrapf(
				err, "relationManager.BulkDeleteByMembersWithTx _type=`%s`, id=`%s`, subjectType=`%s`, subjectIDs=`%+v` fail",
				_type, id, types.DepartmentType, departmentIDs)
		}
		typeCount[types.DepartmentType] = count
	}

	// TODO 更新subject_system_groups表的groups字段
	// 提供subject删除group关系的方法

	err = tx.Commit()
	if err != nil {
		return nil, errorWrapf(err, "tx commit error")
	}
	return typeCount, err
}

// BulkCreateSubjectMembers ...
func (l *subjectService) BulkCreateSubjectMembers(
	_type, id string,
	members []types.Subject,
	policyExpiredAt int64,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkCreateSubjectMembers")
	// 查询subject PK
	parentPK, err := l.manager.GetPK(_type, id)
	if err != nil {
		return errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

	// 分组查询members PK
	memberPKMap := subjectPKMap{}
	// 按类型分组
	userIDs, departmentIDs, _ := groupBySubjectType(members)

	if len(userIDs) > 0 {
		users, newErr := l.manager.ListByIDs(types.UserType, userIDs)
		if newErr != nil {
			return errorWrapf(newErr, "manager.ListByIDs _type=`%s`, ids=`%+v` fail", types.UserType, userIDs)
		}
		for _, u := range users {
			memberPKMap.Add(u.Type, u.ID, u.PK)
		}
	}
	if len(departmentIDs) > 0 {
		departments, newErr := l.manager.ListByIDs(types.DepartmentType, departmentIDs)
		if newErr != nil {
			return errorWrapf(newErr, "manager.ListByIDs _type=`%s`, ids=`%+v` fail", types.DepartmentType, departmentIDs)
		}
		for _, d := range departments {
			memberPKMap.Add(d.Type, d.ID, d.PK)
		}
	}

	now := time.Now()
	// 组装需要创建的Subject关系
	relations := make([]dao.SubjectRelation, 0, len(members))
	for _, m := range members {
		mPK, ok := memberPKMap.Get(m.Type, m.ID)
		if !ok {
			return errorWrapf(errors.New("member don't exists pk"), "memberPKMap type=`%s`, id=`%s` fail",
				m.Type, m.ID)
		}
		relations = append(relations, dao.SubjectRelation{
			SubjectPK:       mPK,
			ParentPK:        parentPK,
			PolicyExpiredAt: policyExpiredAt,
			CreateAt:        now,
		})
	}

	err = l.relationManager.BulkCreate(relations)
	if err != nil {
		return errorWrapf(err, "relationManager.BulkCreate relations=`%+v` fail", relations)
	}

	// TODO 更新或创建 subject_system_groups 数据, 乐观锁
	// 向subject增加一个group关系
	return nil
}

// GetMemberCountBeforeExpiredAt ...
func (l *subjectService) GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "GetMemberCountBeforeExpiredAt")
	// 查询subject PK
	parentPK, err := l.manager.GetPK(_type, id)
	if err != nil {
		return 0, errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

	count, err := l.relationManager.GetMemberCountBeforeExpiredAt(parentPK, expiredAt)
	if err != nil {
		err = errorx.Wrapf(err, SubjectSVC, "GetMemberCountBeforeExpiredAt",
			"relationManager.GetMemberCountBeforeExpiredAt _type=`%s`, id=`%s`, expiredAt=`%d` fail",
			_type, id, expiredAt)
		return 0, err
	}
	return count, nil
}

// ListPagingMemberBeforeExpiredAt ...
func (l *subjectService) ListPagingMemberBeforeExpiredAt(
	_type, id string, expiredAt int64, limit, offset int64,
) ([]types.SubjectMember, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListPagingMemberBeforeExpiredAt")
	// 查询subject PK
	parentPK, err := l.manager.GetPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

	daoRelations, err := l.relationManager.ListPagingMemberBeforeExpiredAt(
		parentPK, expiredAt, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, SubjectSVC,
			"ListPagingMemberBeforeExpiredAt", "_type=`%s`, id=`%s`, expiredAt=`%d`, limit=`%d`, offset=`%d`",
			_type, id, expiredAt, limit, offset)
	}
	members, err := l.convertToSubjectMembers(daoRelations)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectMembers relations=`%s`", members)
	}

	return members, nil
}
