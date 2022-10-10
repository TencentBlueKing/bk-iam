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

type actionHasAnyPolicyChecker struct {
	policyService                     service.PolicyService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
	modelChangeService                service.ModelChangeEventService
}

func newActionHasAnyPolicyChecker() *actionHasAnyPolicyChecker {
	return &actionHasAnyPolicyChecker{
		policyService:                     service.NewPolicyService(),
		subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
		modelChangeService:                service.NewModelChangeService(),
	}
}

func (c *actionHasAnyPolicyChecker) hasAnyPolicy(actionPK int64) (bool, error) {
	// 1. check abac policy exists
	exist, err := c.policyService.HasAnyByActionPK(actionPK)
	if err != nil {
		return false, fmt.Errorf("query action abac policies fail, actionPK=%d",
			actionPK)
	}

	if exist {
		return true, nil
	}

	// 2. check rbac policy exists
	exist, err = c.subjectActionGroupResourceService.HasAnyByActionPK(actionPK)
	if err != nil {
		return false, fmt.Errorf("query action rbac policies fail, actionPK=%d",
			actionPK)
	}

	return exist, nil
}

func (c *actionHasAnyPolicyChecker) CanAlter(systemID, actionID string) (bool, error) {
	actionPK, err := cacheimpls.GetActionPK(systemID, actionID)
	if err != nil {
		return false, fmt.Errorf("query action pk fail, systemID=%s, id=%s", systemID, actionID)
	}

	policyExist, err := c.hasAnyPolicy(actionPK)
	if err != nil {
		return false, fmt.Errorf("query action policy fail, actionPK=%d, error=%v", actionPK, err)
	}

	if !policyExist {
		return true, nil
	}

	// 如果策略存在，需要再检查是否已经发起异步删除策略的事件
	// TODO: 可以重构，提供一个定制部分参数的ExistByTypeModel方法，因为该方法使用的地方挺多，但是status/modelType是一样的
	eventExist, err := c.modelChangeService.ExistByTypeModel(
		service.ModelChangeEventTypeActionPolicyDeleted,
		service.ModelChangeEventStatusPending,
		service.ModelChangeEventModelTypeAction,
		actionPK,
	)
	if err != nil {
		return false, fmt.Errorf("query action model event fail, actionPK=%d",
			actionPK)
	}

	if !eventExist {
		return false, fmt.Errorf("action has releated policies, "+
			"you can't delete it or update the related_resource_types unless delete all the related policies. "+
			"please contact administrator. [systemID=%s, id=%s, actionPK=%d]",
			systemID, actionID, actionPK)
	}

	return true, nil
}

func (c *actionHasAnyPolicyChecker) FilterActionWithPolicy(
	systemID string,
	actionIDs []string,
) (actionIDsWithAnyPolicy []string, err error) {
	actionIDsWithAnyPolicy = make([]string, 0, len(actionIDs))
	for _, id := range actionIDs {
		canAlter, err := c.CanAlter(systemID, id)
		if err != nil {
			return nil, err
		}
		if !canAlter {
			actionIDsWithAnyPolicy = append(actionIDsWithAnyPolicy, id)
		}
	}
	return actionIDsWithAnyPolicy, nil
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
	canAlter, err := newActionHasAnyPolicyChecker().CanAlter(systemID, actionID)
	// TODO: 目前删除Action策略的事件只能用于删除Action模型，其他都暂时不可用，所以这里调整Action关联的资源类型还是必须保证DB里真正无策略
	if err == nil && canAlter {
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

func checkUpdatedActionAuthType(systemID, actionID, authType string, relatedResourceTypes []relatedResourceType) error {
	// if not change the auth_type and relatedResourceTypes
	if authType == "" && len(relatedResourceTypes) == 0 {
		return nil
	}

	// if authType != "" || len(relatedResourceTypes) > 0

	oldAction, err := service.NewActionService().Get(systemID, actionID)
	if err != nil {
		return fmt.Errorf("actionService get systemID=%s, actionID=%s fail, %w", systemID, actionID, err)
	}

	// 1. if auth_type want to change, should has no policies
	if authType != "" && authType != oldAction.AuthType {
		canAlter, err := newActionHasAnyPolicyChecker().CanAlter(systemID, actionID)
		if err != nil {
			return fmt.Errorf("checkActionIDsHashAnyPolicies systemID=%s, actionID=%s: %w", systemID, actionID, err)
		}
		if !canAlter {
			return fmt.Errorf("systemID=%s, actionID=%s has related policies, you cant't update the auth_type",
				systemID, actionID)
		}
	}

	// 2. new auth_type/related_resource_types should be valid
	var newAuthType string
	var newRelatedResourceTypes []relatedResourceType

	if authType != "" {
		newAuthType = authType
	} else {
		newAuthType = oldAction.AuthType
	}

	if len(relatedResourceTypes) > 0 {
		newRelatedResourceTypes = relatedResourceTypes
	} else {
		for _, rrt := range oldAction.RelatedResourceTypes {
			relatedInstanceSelections := make([]referenceInstanceSelection, 0)
			for _, rrtis := range rrt.RelatedInstanceSelections {
				relatedInstanceSelections = append(relatedInstanceSelections, referenceInstanceSelection{
					IgnoreIAMPath: rrtis["ignore_iam_path"].(bool),
				})
			}

			newRelatedResourceTypes = append(newRelatedResourceTypes, relatedResourceType{
				SelectionMode:             rrt.SelectionMode,
				RelatedInstanceSelections: relatedInstanceSelections,
			})
		}
	}
	valid, message := validateActionAuthType(newAuthType, newRelatedResourceTypes)
	if !valid {
		return fmt.Errorf("validateActionAuthType fail:%s", message)
	}
	return nil
}
