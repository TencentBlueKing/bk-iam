/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package expression

import (
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

type databaseRetriever struct {
	policyService service.PolicyService
}

func newDatabaseRetriever() *databaseRetriever {
	return &databaseRetriever{
		policyService: service.NewPolicyService(),
	}
}

func (r *databaseRetriever) retrieve(pks []int64) ([]types.AuthExpression, []int64, error) {
	expressions, err := r.policyService.ListExpressionByPKs(pks)
	if err != nil {
		return nil, nil, err
	}

	missingPKs := r.getMissingPKs(pks, expressions)

	return expressions, missingPKs, nil
}

func (r *databaseRetriever) getMissingPKs(fullPKs []int64, expressions []types.AuthExpression) []int64 {
	gotPKSet := util.NewFixedLengthInt64Set(len(expressions))
	for _, e := range expressions {
		gotPKSet.Add(e.PK)
	}

	missingPKs := make([]int64, 0, len(fullPKs)-len(expressions))
	for _, pk := range fullPKs {
		if !gotPKSet.Has(pk) {
			missingPKs = append(missingPKs, pk)
		}
	}

	return missingPKs
}

// setMissing for databaseRetriever this func is no logical
func (r *databaseRetriever) setMissing(expressions []types.AuthExpression, missingPKs []int64) error {
	return nil
}
