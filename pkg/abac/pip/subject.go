/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pip

import (
	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/errorx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pip/group"
	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	svctypes "iam/pkg/service/types"
)

// SubjectPIP ...
const SubjectPIP = "SubjectPIP"

func convertSubjectGroups(subjectGroups []svctypes.ThinSubjectGroup) []types.SubjectGroup {
	sgs := make([]types.SubjectGroup, 0, len(subjectGroups))
	for _, i := range subjectGroups {
		sgs = append(sgs, types.SubjectGroup{
			// PK is the subject_pk of group
			PK:              i.PK,
			PolicyExpiredAt: i.PolicyExpiredAt,
		})
	}
	return sgs
}

// GetSubjectPK 获取subject的PK, note this will cache in local for 1 minutes
func GetSubjectPK(_type, id string) (int64, error) {
	// pk, err := cacheimpls.GetSubjectPK(_type, id)
	pk, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return pk, errorx.Wrapf(err, SubjectPIP, "GetSubjectPK",
			"cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	return pk, err
}

// GetSubjectDetail ...
func GetSubjectDetail(pk int64) (departments []int64, groups []types.SubjectGroup, err error) {
	detail, err := cacheimpls.GetSubjectDetail(pk)
	if err != nil {
		err = errorx.Wrapf(err, SubjectPIP, "GetSubjectDetail",
			"cacheimpls.GetSubjectDetail pk=`%d` fail", pk)
		return
	}

	departments = detail.DepartmentPKs
	groups = convertSubjectGroups(detail.SubjectGroups)
	return departments, groups, nil
}

func BatchDeleteSubjectCache(pks []int64) error {
	keys := make([]cache.Key, 0, len(pks))
	subjectTypePKs := make(map[string][]int64, 2)
	for _, pk := range pks {
		key := cacheimpls.SubjectPKCacheKey{
			PK: pk,
		}
		keys = append(keys, key)

		subject, err := cacheimpls.GetSubjectByPK(pk)
		if err != nil {
			log.WithError(err).Errorf("failed to get subject by pk %d", pk)
			continue
		}
		subjectTypePKs[subject.Type] = append(subjectTypePKs[subject.Type], pk)
	}

	// delete subject detail cache
	cacheimpls.SubjectCacheCleaner.BatchDelete(keys)
	// delete subject groups
	for subjectType, pks := range subjectTypePKs {
		err := group.BatchDeleteSubjectGroupsFromCache(subjectType, pks)
		if err != nil {
			log.WithError(err).Errorf("group.BatchDeleteSubjectGroupsFromCache subjectType=`%s`, pks=`%v` fail",
				subjectType, pks)
			continue
		}
	}

	return nil
}
