/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package group

import (
	"github.com/TencentBlueKing/gopkg/collection/set"

	"iam/pkg/service"
	"iam/pkg/service/types"
)

type databaseRetriever struct {
	subjectService service.SubjectService
}

func newDatabaseRetriever() *databaseRetriever {
	return &databaseRetriever{
		subjectService: service.NewSubjectService(),
	}
}

func (r *databaseRetriever) retrieve(pks []int64) (map[int64][]types.ThinSubjectGroup, []int64, error) {
	subjectGroups, err := r.subjectService.ListEffectThinSubjectGroups(pks)
	if err != nil {
		return nil, nil, err
	}

	missingPKs := r.getMissingPKs(pks, subjectGroups)

	return subjectGroups, missingPKs, nil
}

func (r *databaseRetriever) getMissingPKs(fullPKs []int64, subjectGroups map[int64][]types.ThinSubjectGroup) []int64 {
	gotPKSet := set.NewFixedLengthInt64Set(len(subjectGroups))
	for pk := range subjectGroups {
		gotPKSet.Add(pk)
	}

	missingPKs := make([]int64, 0, len(fullPKs)-len(subjectGroups))
	for _, pk := range fullPKs {
		if !gotPKSet.Has(pk) {
			missingPKs = append(missingPKs, pk)
		}
	}

	return missingPKs
}

// setMissing for databaseRetriever this func is no logical
func (r *databaseRetriever) setMissing(subjectGroups map[int64][]types.ThinSubjectGroup, missingPKs []int64) error {
	return nil
}
