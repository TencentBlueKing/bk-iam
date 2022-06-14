/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

import (
	"iam/pkg/abac/prp/group"
	"iam/pkg/service/types"
)

// splitGroupPKsToAuthTypeGroupPKs 将传入的system groupPKs拆分成ABAC和RBAC groupPKs
func splitGroupPKsToAuthTypeGroupPKs(
	systemID string,
	groupPKs []int64,
) (abacGroupPks []int64, rbacGroupPKs []int64, err error) {
	l3 := group.NewGroupAuthTypeDatabaseRetriever(systemID)

	l2 := group.NewGroupAuthTypeRedisRetriever(systemID, l3)

	l1 := group.NewGroupAuthTypeMemoryRetriever(systemID, l2)

	groupAuthTypes, err := l1.Retrieve(groupPKs)
	if err != nil {
		return nil, nil, err
	}

	abacGroupPks = make([]int64, 0, len(groupAuthTypes))
	rbacGroupPKs = make([]int64, 0, len(groupAuthTypes))

	for _, groupAuthType := range groupAuthTypes {
		switch groupAuthType.AuthType {
		case types.AuthTypeABAC:
			abacGroupPks = append(abacGroupPks, groupAuthType.GroupPK)
		case types.AuthTypeRBAC:
			rbacGroupPKs = append(rbacGroupPKs, groupAuthType.GroupPK)
		case types.AuthTypeAll:
			abacGroupPks = append(abacGroupPks, groupAuthType.GroupPK)
			rbacGroupPKs = append(rbacGroupPKs, groupAuthType.GroupPK)
		}
	}

	return abacGroupPks, rbacGroupPKs, nil
}
