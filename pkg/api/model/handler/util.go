/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	svctypes "iam/pkg/service/types"

	"github.com/fatih/structs"
)

// SelectionModeAll ...
const (
	SelectionModeAll       = "all"
	SelectionModeInstance  = "instance"
	SelectionModeAttribute = "attribute"
)

func convertActionDefaultUsage(usage string) string {
	if usage == "" {
		return svctypes.ActionUsageAll
	}
	return usage
}

func convertToRelatedResourceTypes(rrts []relatedResourceType) []svctypes.ActionResourceType {
	arts := make([]svctypes.ActionResourceType, 0, len(rrts))
	for _, rrt := range rrts {
		var risList []map[string]interface{}
		if len(rrt.RelatedInstanceSelections) > 0 {
			for _, is := range rrt.RelatedInstanceSelections {
				risList = append(risList, structs.Map(is))
			}
		}

		// set default SelectionMode to instance
		selectionMode := rrt.SelectionMode
		if selectionMode == "" {
			selectionMode = SelectionModeInstance
		}

		arts = append(arts, svctypes.ActionResourceType{
			System:                    rrt.SystemID,
			ID:                        rrt.ID,
			NameAlias:                 rrt.NameAlias,
			NameAliasEn:               rrt.NameAliasEn,
			SelectionMode:             selectionMode,
			RelatedInstanceSelections: risList,
		})
	}
	return arts
}

func convertToRelatedEnvironments(res []relatedEnvironment) []svctypes.ActionEnvironment {
	aes := make([]svctypes.ActionEnvironment, 0, len(res))
	for _, re := range res {
		aes = append(aes, svctypes.ActionEnvironment{
			Type: re.Type,
		})
	}
	return aes
}
