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
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
)

// ActionPIP ...
const ActionPIP = "ActionPIP"

// GetActionDetail ...
func GetActionDetail(system, id string) (pk int64, authType int64, arts []types.ActionResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionPIP, "GetActionDetail")

	detail, err := cacheimpls.GetActionDetail(system, id)
	if err != nil {
		err = errorWrapf(err,
			"cacheimpls.GetActionDetail system=`%s` actionID=`%s` fail", system, id)
		return
	}

	// 数据转换
	arts = make([]types.ActionResourceType, 0, len(detail.ResourceTypes))
	for _, art := range detail.ResourceTypes {
		arts = append(arts, types.ActionResourceType{
			System: art.System,
			Type:   art.ID,
		})
	}
	return detail.PK, detail.AuthType, arts, nil
}
