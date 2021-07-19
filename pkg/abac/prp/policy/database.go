/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

type databaseRetriever struct {
	policyService service.PolicyService
	actionPK      int64
}

func newDatabaseRetriever(actionPK int64) *databaseRetriever {
	return &databaseRetriever{
		policyService: service.NewPolicyService(),
		actionPK:      actionPK,
	}
}

func (r *databaseRetriever) retrieve(subjectPKs []int64) ([]types.AuthPolicy, []int64, error) {
	policies, err := r.policyService.ListAuthBySubjectAction(subjectPKs, r.actionPK)
	if err != nil {
		return nil, nil, err
	}

	missingSubjectPKs := r.getMissingPKs(subjectPKs, policies)
	return policies, missingSubjectPKs, nil
}

func (r *databaseRetriever) getMissingPKs(subjectPKs []int64, policies []types.AuthPolicy) []int64 {
	gotSubjectPKSet := util.NewFixedLengthInt64Set(len(policies))
	for _, e := range policies {
		gotSubjectPKSet.Add(e.SubjectPK)
	}

	missingSubjectPKs := make([]int64, 0, len(subjectPKs))
	for _, pk := range subjectPKs {
		if !gotSubjectPKSet.Has(pk) {
			missingSubjectPKs = append(missingSubjectPKs, pk)
		}
	}

	return missingSubjectPKs
}

// setMissing for databaseRetriever this func is no logical
func (r *databaseRetriever) setMissing(policies []types.AuthPolicy, missingPKs []int64) error {
	return nil
}
