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
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

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
func (l *subjectService) GetEffectThinSubjectGroups(pk int64) (thinSubjectGroup []types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "GetEffectThinSubjectGroups")

	relations, err := l.relationManager.ListEffectThinRelationBySubjectPK(pk)
	if err != nil {
		return thinSubjectGroup, errorWrapf(err, "ListEffectThinRelationBySubjectPK pk=`%d` fail", pk)
	}

	for _, r := range relations {
		thinSubjectGroup = append(thinSubjectGroup, convertToThinSubjectGroup(r))
	}
	return thinSubjectGroup, err
}

// ListEffectThinSubjectGroups 批量获取 subject 有效的 groups(未过期的)
func (l *subjectService) ListEffectThinSubjectGroups(
	pks []int64,
) (subjectGroups map[int64][]types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListEffectThinSubjectGroups")

	subjectGroups = make(map[int64][]types.ThinSubjectGroup, len(pks))

	relations, err := l.relationManager.ListEffectRelationBySubjectPKs(pks)
	if err != nil {
		return subjectGroups, errorWrapf(err, "ListRelationByPKs pks=`%+v` fail", pks)
	}

	for _, r := range relations {
		subjectPK := r.SubjectPK
		subjectGroups[subjectPK] = append(subjectGroups[subjectPK], convertEffectiveRelationToThinSubjectGroup(r))
	}
	return subjectGroups, nil
}

// ListSubjectGroups ...
func (l *subjectService) ListSubjectGroups(
	_type, id string, beforeExpiredAt int64,
) (subjectGroups []types.SubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListSubjectGroups")
	// 查询subject PK
	pk, err := l.manager.GetPK(_type, id)
	if err != nil {
		return nil, errorWrapf(err, "manager.GetPK _type=`%s`, id=`%s` fail", _type, id)
	}

	var relations []dao.SubjectRelation
	if beforeExpiredAt == 0 {
		relations, err = l.relationManager.ListRelation(pk)
	} else {
		relations, err = l.relationManager.ListRelationBeforeExpiredAt(pk, beforeExpiredAt)
	}

	if err != nil {
		return subjectGroups, errorWrapf(err, "ListSubjectGroups _type=`%s`, id=`%s` fail", _type, id)
	}

	groups, err := l.convertToSubjectGroup(relations)
	if err != nil {
		return nil, errorWrapf(err, "convertToSubjectGroup relations=`%s`", relations)
	}

	return groups, nil
}

func (l *subjectService) convertToSubjectGroup(daoRelations []dao.SubjectRelation) ([]types.SubjectGroup, error) {
	if len(daoRelations) == 0 {
		return nil, nil
	}

	subjectPKs := make([]int64, 0, len(daoRelations))
	for _, r := range daoRelations {
		subjectPKs = append(subjectPKs, r.ParentPK)
	}

	subjects, err := l.manager.ListByPKs(subjectPKs)
	if err != nil {
		return nil, err
	}

	subjectMap := make(map[int64]dao.Subject, len(subjects))
	for _, s := range subjects {
		subjectMap[s.PK] = s
	}

	groups := make([]types.SubjectGroup, 0, len(daoRelations))
	for _, r := range daoRelations {
		var _type, id string
		subject, ok := subjectMap[r.SubjectPK]
		if ok {
			_type = subject.Type
			id = subject.ID
		}

		groups = append(groups, types.SubjectGroup{
			PK:              r.ParentPK,
			Type:            _type,
			ID:              id,
			PolicyExpiredAt: r.PolicyExpiredAt,
			CreateAt:        r.CreateAt,
		})
	}
	return groups, nil
}

// ListExistSubjectsBeforeExpiredAt filter the exists and not expired subjects
func (l *subjectService) ListExistSubjectsBeforeExpiredAt(
	subjects []types.Subject, expiredAt int64,
) ([]types.Subject, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "FilterSubjectsBeforeExpiredAt")

	groupIDs := make([]string, 0, len(subjects))
	for _, subject := range subjects {
		if subject.Type == types.GroupType {
			groupIDs = append(groupIDs, subject.ID)
		}
	}

	groups, err := l.manager.ListByIDs(types.GroupType, groupIDs)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListByIDs _type=`%s`, ids=`%+v` fail", types.GroupType, groupIDs)
	}
	parentPKs := make([]int64, 0, len(groups))
	for _, g := range groups {
		parentPKs = append(parentPKs, g.PK)
	}

	existGroupPKs, err := l.relationManager.ListParentPKsBeforeExpiredAt(parentPKs, expiredAt)
	if err != nil {
		return []types.Subject{}, errorWrapf(
			err, "ListParentPKsBeforeExpiredAt _type=`%s`, parentPKs=`%+v`, expiredAt=`%d` fail",
			parentPKs, expiredAt,
		)
	}
	if len(existGroupPKs) == 0 {
		return []types.Subject{}, nil
	}

	idSet := set.NewInt64SetWithValues(existGroupPKs)
	existSubjects := make([]types.Subject, 0, len(existGroupPKs))
	for _, group := range groups {
		if idSet.Has(group.PK) {
			existSubjects = append(existSubjects, types.Subject{
				Type: group.Type,
				ID:   group.ID,
				Name: group.Name,
			})
		}
	}

	return existSubjects, nil
}
