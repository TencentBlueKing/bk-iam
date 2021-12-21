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
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service/types"
)

// ListActionInstanceSelectionIDBySystem ...
func (l *actionService) ListActionInstanceSelectionIDBySystem(system string) (
	instanceSelections []types.ActionInstanceSelectionID, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListActionInstanceSelectionIDBySystem")

	allActions, err := l.ListBySystem(system)
	if err != nil {
		err = errorWrapf(err, "actionService.ListBySystem system=`%s` fail",
			system)
		return instanceSelections, err
	}
	for _, action := range allActions {
		for _, rrt := range action.RelatedResourceTypes {
			for _, is := range rrt.InstanceSelections {
				instanceSelections = append(instanceSelections, types.ActionInstanceSelectionID{
					ActionSystem:            system,
					ActionID:                action.ID,
					InstanceSelectionSystem: is["system_id"].(string),
					InstanceSelectionID:     is["id"].(string),
				})
			}
		}
	}

	return instanceSelections, nil
}
