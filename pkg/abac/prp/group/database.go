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

type serviceGroupAuthTypeRetriever struct {
	systemID string
	service  service.GroupService
}

func NewServiceGroupAuthTypeRetriever(systemID string) GroupAuthTypeRetriever {
	return &serviceGroupAuthTypeRetriever{
		systemID: systemID,
		service:  service.NewGroupService(),
	}
}

func (r *serviceGroupAuthTypeRetriever) Retrieve(
	groupPKs []int64,
) (groupAuthTypes []types.GroupAuthType, err error) {
	groupAuthTypes, err = r.service.ListGroupAuthBySystemGroupPKs(r.systemID, groupPKs)
	if err != nil {
		return nil, err
	}

	// 设置默认未授权的group authType 0
	if len(groupPKs) > len(groupAuthTypes) {
		groupPKSet := set.NewInt64Set()
		for _, groupAuthType := range groupAuthTypes {
			groupPKSet.Add(groupAuthType.GroupPK)
		}

		for _, groupPK := range groupPKs {
			if !groupPKSet.Has(groupPK) {
				groupAuthTypes = append(groupAuthTypes, types.GroupAuthType{
					GroupPK:  groupPK,
					AuthType: types.AuthTypeNone,
				})
			}
		}
	}

	return groupAuthTypes, nil
}
