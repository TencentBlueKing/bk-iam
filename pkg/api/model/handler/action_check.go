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

	"iam/pkg/api/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
)

// 需要db操作的校验, 统一叫 checkXXXX

type AllActions struct {
	AllBaseInfo

	Actions []svctypes.ActionBaseInfo
}

func (a *AllActions) Size() int {
	return len(a.Actions)
}

// NewAllActions ...
func NewAllActions(actions []svctypes.ActionBaseInfo) *AllActions {
	idSet := map[string]string{}
	nameSet := map[string]string{}
	nameEnSet := map[string]string{}

	for _, ac := range actions {
		idSet[ac.ID] = ac.ID
		nameSet[ac.Name] = ac.ID
		nameEnSet[ac.NameEn] = ac.ID
	}

	return &AllActions{
		AllBaseInfo: AllBaseInfo{
			IDSet:     idSet,
			NameSet:   nameSet,
			NameEnSet: nameEnSet,
		},
		Actions: actions,
	}
}

func BuildAllActions(systemID string) (*AllActions, error) {
	svc := service.NewActionService()
	actions, err := svc.ListBaseInfoBySystem(systemID)
	if err != nil {
		return nil, errors.New("query all action fail")
	}

	return NewAllActions(actions), nil
}

func checkActionsQuotaAndAllUnique(systemID string, inActions []actionSerializer) error {
	allActions, err := BuildAllActions(systemID)
	if err != nil {
		return err
	}

	for _, ac := range inActions {
		if allActions.ContainsID(ac.ID) {
			return fmt.Errorf("action id[%s] already exists", ac.ID)
		}
		if allActions.ContainsName(ac.Name) {
			return fmt.Errorf("action name[%s] already exists", ac.Name)
		}
		if allActions.ContainsNameEn(ac.NameEn) {
			return fmt.Errorf("action name_en[%s] already exists", ac.NameEn)
		}
	}

	// quota
	if allActions.Size()+len(inActions) > common.GetMaxActionsLimit(systemID) {
		return fmt.Errorf("quota error: system %s can only have %d actions.[current %d, want to create %d]",
			systemID, common.GetMaxActionsLimit(systemID), allActions.Size(), len(inActions))
	}

	return nil
}

func checkResourceTypeAllExists(arts map[string][]relatedResourceType) error {
	rtSvc := service.NewResourceTypeService()
	systemSet := map[string]*AllResourceTypes{}

	for actionID, rrts := range arts {
		for _, rrt := range rrts {
			allResourceTypes, ok := systemSet[rrt.SystemID]
			if !ok {
				// 一般关联的系统不会超过2个
				var err error
				allRts, err := rtSvc.ListBySystem(rrt.SystemID)

				allResourceTypes = NewAllResourceTypes(allRts)

				if err != nil {
					return fmt.Errorf("query system[%s] all resource type fail", rrt.SystemID)
				}
				systemSet[rrt.SystemID] = allResourceTypes
			}

			if !allResourceTypes.ContainsID(rrt.ID) {
				return fmt.Errorf("action id[%s] related resource type[%s] not exists", actionID, rrt.ID)
			}
		}
	}

	return nil
}

func checkActionCreateResourceTypeAllExists(actions []actionSerializer) error {
	actionResourceTypes := map[string][]relatedResourceType{}
	for _, ac := range actions {
		actionResourceTypes[ac.ID] = ac.RelatedResourceTypes
	}
	return checkResourceTypeAllExists(actionResourceTypes)
}

func checkActionUpdateResourceTypeAllExists(actionID string, resourceTypes []relatedResourceType) error {
	actionResourceTypes := map[string][]relatedResourceType{}
	actionResourceTypes[actionID] = resourceTypes
	return checkResourceTypeAllExists(actionResourceTypes)
}

func checkActionUpdateUnique(systemID, actionID, name, nameEn string) error {
	allActions, err := BuildAllActions(systemID)
	if err != nil {
		return err
	}

	if !allActions.ContainsID(actionID) {
		return fmt.Errorf("action id[%s] not exists", actionID)
	}

	// check name / name_en should be unique
	if name != "" && allActions.ContainsNameExcludeSelf(name, actionID) {
		return fmt.Errorf("action name[%s] already exists", name)
	}
	if nameEn != "" && allActions.ContainsNameEnExcludeSelf(nameEn, actionID) {
		return fmt.Errorf("action name_en[%s] already exists", nameEn)
	}
	return nil
}

func checkActionIDsExist(systemID string, ids []string) error {
	allActions, err := BuildAllActions(systemID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if !allActions.ContainsID(id) {
			return fmt.Errorf("action id[%s] not exists", id)
		}
	}
	return nil
}

func checkActionIDsHasAnyPolicies(systemID string, ids []string) ([]string, error) {
	// TODO: 需要重构，基于单一原则，规范里check方法返回只有error，不返回其他数据
	svc := service.NewPolicyService()
	eventSvc := service.NewModelChangeService()
	// 记录需要异步删除的Action
	needAsyncDeletedActionIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		actionPK, err := cacheimpls.GetActionPK(systemID, id)
		if err != nil {
			return []string{}, fmt.Errorf("query action pk fail, systemID=%s, id=%s", systemID, id)
		}
		exist, err := svc.HasAnyByActionPK(actionPK)
		if err != nil {
			return []string{}, fmt.Errorf("query action policies fail, systemID=%s, id=%s, actionPK=%d",
				systemID, id, actionPK)
		}
		if exist {
			// 如果策略存在，需要再检查是否已经发起异步删除策略的事件
			// TODO: 可以重构，提供一个定制部分参数的ExistByTypeModel方法，因为该方法使用的地方挺多，但是status/modelType是一样的
			eventExist, err := eventSvc.ExistByTypeModel(
				service.ModelChangeEventTypeActionPolicyDeleted,
				service.ModelChangeEventStatusPending,
				service.ModelChangeEventModelTypeAction,
				actionPK,
			)
			if err != nil {
				return []string{}, fmt.Errorf("query action model event fail, systemID=%s, id=%s, actionPK=%d",
					systemID, id,
					actionPK)
			}
			// 若删除Action策略时间不存在，则Action不可删除
			if !eventExist {
				return []string{}, fmt.Errorf("action has releated policies, "+
					"you can't delete it or update the related_resource_types unless delete all the related policies. "+
					"please contact administrator. [systemID=%s, id=%s, actionPK=%d]",
					systemID, id, actionPK)
			}
			needAsyncDeletedActionIDs = append(needAsyncDeletedActionIDs, id)
		}
	}
	return needAsyncDeletedActionIDs, nil
}

func checkUpdateActionRelatedResourceTypeNotChanged(
	systemID, actionID string,
	inputActionResourceTypes []relatedResourceType,
) error {
	svc := service.NewActionService()
	// get old action
	oldAction, err := svc.Get(systemID, actionID)
	if err != nil {
		return fmt.Errorf("get action from db fail, %w", err)
	}
	oldActionResourceTypes := oldAction.RelatedResourceTypes

	// if input has related_resource_types or the old action has related_resource_types
	if len(inputActionResourceTypes) == 0 && len(oldActionResourceTypes) == 0 {
		return nil
	}

	// if not policies, no need to check
	needAsyncDeletedActionIDs, err := checkActionIDsHasAnyPolicies(systemID, []string{actionID})
	// TODO: 目前删除Action策略的事件只能用于删除Action模型，其他都暂时不可用，所以这里调整Action关联的资源类型还是必须保证DB里真正无策略
	if err == nil && len(needAsyncDeletedActionIDs) == 0 {
		return nil
	}

	// get the old action
	// make sure the related
	if len(inputActionResourceTypes) != len(oldActionResourceTypes) {
		err = fmt.Errorf("%w, input related_resource_types len=%d, while the exist related_resource_types len=%d",
			err, len(inputActionResourceTypes), len(oldActionResourceTypes))
		return err
	}

	// the order and the content should be the same
	for i := 0; i < len(oldActionResourceTypes); i++ {
		currentRT := inputActionResourceTypes[i]
		oldRT := oldActionResourceTypes[i]
		if currentRT.SystemID != oldRT.System || currentRT.ID != oldRT.ID {
			err = fmt.Errorf(
				"%w, related_resource_types[%d](system=%s, id=%s) different to exist related_resource_types[%d](system=%s, id=%s)",
				err,
				i,
				currentRT.SystemID,
				currentRT.ID,
				i,
				oldRT.System,
				oldRT.ID,
			)
			return err
		}
	}
	return nil
}
