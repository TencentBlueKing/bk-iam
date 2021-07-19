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
	"iam/pkg/errorx"
	"iam/pkg/service/types"
)

// ListThinActionResourceTypes 获取操作关联的资源类型
func (l *actionService) ListThinActionResourceTypes(
	system, actionID string,
) (actionResourceTypes []types.ThinActionResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListThinActionResourceTypes")

	arts, err := l.actionResourceTypeManager.ListResourceTypeByAction(system, actionID)
	if err != nil {
		return nil, errorWrapf(err,
			"actionResourceTypeManager.ListResourceTypeByAction system=`%s`, actionID=`%s` fail",
			system, actionID)
	}

	if len(arts) == 0 {
		return
	}
	actionResourceTypes = make([]types.ThinActionResourceType, 0, len(arts))
	for _, art := range arts {
		actionResourceTypes = append(actionResourceTypes, types.ThinActionResourceType{
			System: art.ResourceTypeSystem,
			ID:     art.ResourceTypeID,
		})
	}
	return actionResourceTypes, nil
}

// ListActionResourceTypeIDByResourceTypeSystem ...
func (l *actionService) ListActionResourceTypeIDByResourceTypeSystem(resourceTypeSystem string) (
	actionResourceTypeIDs []types.ActionResourceTypeID, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListActionResourceTypeIDByResourceTypeSystem")

	actionResourceTypes, err := l.actionResourceTypeManager.ListByResourceTypeSystem(resourceTypeSystem)
	if err != nil {
		err = errorWrapf(err, "actionResourceTypeManager.ListByResourceTypeSystem resourceTypeSystem=`%s` fail",
			resourceTypeSystem)
		return actionResourceTypeIDs, err
	}
	for _, art := range actionResourceTypes {
		actionResourceTypeIDs = append(actionResourceTypeIDs, types.ActionResourceTypeID{
			ActionSystem:       art.ActionSystem,
			ActionID:           art.ActionID,
			ResourceTypeSystem: art.ResourceTypeSystem,
			ResourceTypeID:     art.ResourceTypeID,
		})
	}
	return
}

// ListActionResourceTypeIDByActionSystem ...
func (l *actionService) ListActionResourceTypeIDByActionSystem(actionSystem string) (
	actionResourceTypeIDs []types.ActionResourceTypeID, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListActionResourceTypeIDByActionSystem")

	actionResourceTypes, err := l.actionResourceTypeManager.ListByActionSystem(actionSystem)
	if err != nil {
		err = errorWrapf(err, "actionResourceTypeManager.ListByActionSystem actionSystem=`%s` fail",
			actionSystem)
		return actionResourceTypeIDs, err
	}
	for _, art := range actionResourceTypes {
		actionResourceTypeIDs = append(actionResourceTypeIDs, types.ActionResourceTypeID{
			ActionSystem:       art.ActionSystem,
			ActionID:           art.ActionID,
			ResourceTypeSystem: art.ResourceTypeSystem,
			ResourceTypeID:     art.ResourceTypeID,
		})
	}
	return
}
