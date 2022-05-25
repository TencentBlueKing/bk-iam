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
	"errors"
	"fmt"

	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

// ActionIDResourceTypeID ...
type ActionIDResourceTypeID struct {
	ActionID       string
	ResourceTypeID string
}

func checkResourceCreatorActionsRelateResourceType(systemID string, rcas resourceCreatorActionSerializer) error {
	actionResourceTypes := rcas.getAllActionIDResourceTypeIDFromConfig()
	actions, err := cacheimpls.ListActionBySystem(systemID)
	if err != nil {
		return errors.New("query all action fail")
	}

	actionMap := map[string]types.Action{}
	for _, action := range actions {
		actionMap[action.ID] = action
	}

	// 检查每个action-resourceType是否OK
	// 0. action 必须存在
	// 1. 不能关联的Action的RelatedResourceType > 1，但是可以关联`与资源实例无关`的Action
	// 2. 必须与注册的模型一致
	for _, art := range actionResourceTypes {
		action, ok := actionMap[art.ActionID]
		if !ok {
			return fmt.Errorf("action id[%s] not exists", art.ActionID)
		}

		// 不能关联的Action的RelatedResourceType > 1
		if len(action.RelatedResourceTypes) > 1 {
			return fmt.Errorf(
				"action id[%s] can not be associated to resource type[%s], "+
					"because the length of related resource types should be greater than one",
				art.ActionID, art.ResourceTypeID)
		}

		// 允许关联`与资源实例无关`的Action
		if len(action.RelatedResourceTypes) == 0 {
			continue
		}

		// 关联的ResourceType是否一致
		if action.RelatedResourceTypes[0].ID != art.ResourceTypeID {
			return fmt.Errorf("in registered perm model, action id[%s] is not related to resource type[%s]",
				art.ActionID, art.ResourceTypeID)
		}
	}

	return nil
}
