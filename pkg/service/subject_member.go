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
func (l *subjectService) GetMemberCount(_type, id string) (int64, error) {
	cnt, err := l.relationManager.GetMemberCount(_type, id)
	if err != nil {
		err = errorx.Wrapf(err, SubjectSVC, "GetMemberCount",
			"relationManager.GetMemberCount _type=`%s`, id=`%s` fail", _type, id)
		return 0, err
	}
	return cnt, nil
}

// ListPagingMember ...
func (l *subjectService) ListPagingMember(_type, id string, limit, offset int64) ([]types.SubjectMember, error) {
	daoRelations, err := l.relationManager.ListPagingMember(_type, id, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, SubjectSVC,
			"ListPagingMember", "relationManager.ListPagingMember _type=`%s`, id=`%s`, limit=`%d`, offset=`%d`",
			_type, id, limit, offset)
	}

	return convertToSubjectMembers(daoRelations), nil
}

// ListMember ...
func (l *subjectService) ListMember(_type, id string) ([]types.SubjectMember, error) {
	daoRelations, err := l.relationManager.ListMember(_type, id)
	if err != nil {
		return nil, errorx.Wrapf(err, SubjectSVC,
			"ListMember", "relationManager.ListMember _type=`%s`, id=`%s` fail", _type, id)
	}

	return convertToSubjectMembers(daoRelations), nil
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

	return nil
}

// BulkDeleteSubjectMembers ...
func (l *subjectService) BulkDeleteSubjectMembers(_type, id string, members []types.Subject) (map[string]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkDeleteSubjectMember")

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
		count, err = l.relationManager.BulkDeleteByMembersWithTx(tx, _type, id, types.UserType, userIDs)
		if err != nil {
			return nil, errorWrapf(err,
				"relationManager.BulkDeleteByMembersWithTx _type=`%s`, id=`%s`, subjectType=`%s`, subjectIDs=`%+v` fail",
				_type, id, types.UserType, userIDs)
		}
		typeCount[types.UserType] = count
	}

	if len(departmentIDs) != 0 {
		count, err = l.relationManager.BulkDeleteByMembersWithTx(tx, _type, id, types.DepartmentType, departmentIDs)
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
func (l *subjectService) BulkCreateSubjectMembers(
	_type, id string,
	members []types.Subject,
	policyExpiredAt int64,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "BulkCreateSubjectMembers")
	// 查询subject PK
	pk, err := l.manager.GetPK(_type, id)
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
			SubjectType:     m.Type,
			SubjectID:       m.ID,
			ParentPK:        pk,
			ParentType:      _type,
			ParentID:        id,
			PolicyExpiredAt: policyExpiredAt,
			CreateAt:        now,
		})
	}

	err = l.relationManager.BulkCreate(relations)
	if err != nil {
		return errorWrapf(err, "relationManager.BulkCreate relations=`%+v` fail", relations)
	}
	return nil
}

// GetMemberCountBeforeExpiredAt ...
func (l *subjectService) GetMemberCountBeforeExpiredAt(_type, id string, expiredAt int64) (int64, error) {
	cnt, err := l.relationManager.GetMemberCountBeforeExpiredAt(_type, id, expiredAt)
	if err != nil {
		err = errorx.Wrapf(err, SubjectSVC, "GetMemberCountBeforeExpiredAt",
			"relationManager.GetMemberCountBeforeExpiredAt _type=`%s`, id=`%s`, expiredAt=`%d` fail",
			_type, id, expiredAt)
		return 0, err
	}
	return cnt, nil
}

// ListPagingMemberBeforeExpiredAt ...
func (l *subjectService) ListPagingMemberBeforeExpiredAt(
	_type, id string, expiredAt int64, limit, offset int64,
) ([]types.SubjectMember, error) {
	daoRelations, err := l.relationManager.ListPagingMemberBeforeExpiredAt(
		_type, id, expiredAt, limit, offset)
	if err != nil {
		return nil, errorx.Wrapf(err, SubjectSVC,
			"ListPagingMemberBeforeExpiredAt", "_type=`%s`, id=`%s`, expiredAt=`%d`, limit=`%d`, offset=`%d`",
			_type, id, expiredAt, limit, offset)
	}

	return convertToSubjectMembers(daoRelations), nil
}
