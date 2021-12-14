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
	"iam/pkg/database/dao"
	"iam/pkg/errorx"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

func convertToSubjectGroup(relation dao.SubjectRelation) types.SubjectGroup {
	return types.SubjectGroup{
		PK:              relation.ParentPK,
		Type:            relation.ParentType,
		ID:              relation.ParentID,
		PolicyExpiredAt: relation.PolicyExpiredAt,
		CreateAt:        relation.CreateAt,
	}
}

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

// ListSubjectEffectGroups 批量获取 subject 有效的 groups(未过期的)
func (l *subjectService) ListSubjectEffectGroups(pks []int64) (
	subjectGroups map[int64][]types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "ListSubjectEffectGroups")

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
	var relations []dao.SubjectRelation
	if beforeExpiredAt == 0 {
		relations, err = l.relationManager.ListRelation(_type, id)
	} else {
		relations, err = l.relationManager.ListRelationBeforeExpiredAt(_type, id, beforeExpiredAt)
	}

	if err != nil {
		return subjectGroups, errorWrapf(err, "ListSubjectGroups _type=`%s`, id=`%s` fail", _type, id)
	}

	subjectGroups = make([]types.SubjectGroup, 0, len(relations))

	for _, r := range relations {
		subjectGroups = append(subjectGroups, convertToSubjectGroup(r))
	}
	return subjectGroups, err
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

	existGroupIDs, err := l.relationManager.ListParentIDsBeforeExpiredAt(types.GroupType, groupIDs, expiredAt)
	if err != nil {
		return []types.Subject{}, errorWrapf(
			err, "ListParentIDsBeforeExpiredAt _type=`%s`, ids=`%+v`, expiredAt=`%d` fail",
			types.GroupType, groupIDs, expiredAt,
		)
	}
	if len(existGroupIDs) == 0 {
		return []types.Subject{}, nil
	}

	idSet := util.NewStringSetWithValues(existGroupIDs)
	existSubjects := make([]types.Subject, 0, len(existGroupIDs))
	for _, subject := range subjects {
		if subject.Type == types.GroupType && idSet.Has(subject.ID) {
			existSubjects = append(existSubjects, subject)
		}
	}

	return existSubjects, nil
}
