/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pap

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/abac/pap/event"
	"iam/pkg/abac/prp/expression"
	"iam/pkg/abac/prp/group"
	"iam/pkg/abac/prp/policy"
	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	"iam/pkg/database"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
)

const PolicyCTLV2 = "PolicyCTLV2"

type PolicyControllerV2 interface {
	Alter(
		systemID, subjectType, subjectID string, templateID int64,
		createPolicies, updatePolicies []types.Policy, deletePolicyIDs []int64,
		resourceChangedActions []types.ResourceChangedAction,
		groupAuthType int64,
	) (err error)
}

type policyControllerV2 struct {
	policyController

	// RBAC
	groupResourcePolicyService service.GroupResourcePolicyService
	groupService               service.GroupService

	resourceTypeService service.ResourceTypeService

	eventProducer event.PolicyEventProducer
}

func NewPolicyControllerV2() PolicyControllerV2 {
	return &policyControllerV2{
		policyController: policyController{
			subjectService:         service.NewSubjectService(),
			actionService:          service.NewActionService(),
			policyService:          service.NewPolicyService(),
			temporaryPolicyService: service.NewTemporaryPolicyService(),
		},

		groupResourcePolicyService: service.NewGroupResourcePolicyService(),
		groupService:               service.NewGroupService(),
		resourceTypeService:        service.NewResourceTypeService(),

		eventProducer: event.NewPolicyEventProducer(),
	}
}

func (c *policyControllerV2) Alter(
	systemID, subjectType, subjectID string, templateID int64,
	createPolicies, updatePolicies []types.Policy, deletePolicyIDs []int64,
	resourceChangedActions []types.ResourceChangedAction,
	groupAuthType int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTLV2, "Alter")
	// 0. 通用处理
	// 查询subject、action 相关的信息
	subjectPK, actionPKMap, actionPKWithResourceTypeSet, err := c.querySubjectActionForAlterPolicies(
		systemID, subjectType, subjectID,
	)
	if err != nil {
		err = errorWrapf(err, "c.querySubjectActionForAlterPolicies systemID=`%s` fail", systemID)
		return
	}
	// 生成统一的DB事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		err = errorWrapf(err, "define tx fail")
		return
	}

	// 1. 处理ABAC策略
	updatedActionPKExpressionPKs, err := c.alterABACPolicies(
		tx, subjectPK, templateID,
		createPolicies, updatePolicies, deletePolicyIDs,
		actionPKMap, actionPKWithResourceTypeSet,
	)
	if err != nil {
		err = errorWrapf(err, "c.alterABACPolicies systemID=`%s` subjectPK=`%d` fail", systemID, subjectPK)
		return
	}

	// 2. 处理RBAC策略处理
	resourceChangedContents, err := c.alterRBACPolicies(tx, subjectPK, templateID, systemID, resourceChangedActions)
	if err != nil {
		err = errorWrapf(err, "c.alterRBACPolicy systemID=`%s` subjectPK=`%d` fail", systemID, subjectPK)
		return
	}

	// 3. 处理GroupAuthType
	changed, err := c.groupService.AlterGroupAuthType(tx, systemID, subjectPK, groupAuthType)
	if err != nil {
		err = errorWrapf(err, "c.alterRBACPolicy systemID=`%s` subjectPK=`%d` fail", systemID, subjectPK)
		return
	}

	// DB事务提交
	err = tx.Commit()
	if err != nil {
		err = errorWrapf(err, "tx commit fail")
		return
	}

	// 4. 创建RBAC变更消息
	c.eventProducer.PublishRBACGroupAlterEvent(subjectPK, resourceChangedContents)

	// 5. 清理缓存
	// 5.1 ABAC相关缓存
	if len(createPolicies) > 0 || len(updatePolicies) > 0 || len(deletePolicyIDs) > 0 {
		policy.DeleteSystemSubjectPKsFromCache(systemID, []int64{subjectPK})
		// 只有自定义权限才需要清理Expression，模板权限不需要清理Expression，因为模板的Expression是复用的，有定时任务自动清理
		if templateID == 0 && len(updatedActionPKExpressionPKs) > 0 {
			expression.BatchDeleteExpressionsFromCache(updatedActionPKExpressionPKs)
		}
	}

	// 5.2 RBAC相关缓存
	for _, rcc := range resourceChangedContents {
		cacheimpls.DeleteResourceAuthorizedGroupPKsCache(
			systemID, rcc.ActionRelatedResourceTypePK, rcc.ResourceTypePK, rcc.ResourceID,
		)
	}

	// 5.3 GroupAuthType相关缓存
	// 只有AuthType被改变了才会进行缓存的清理
	if changed {
		group.DeleteGroupAuthTypeCache(systemID, subjectPK)
		cacheimpls.BatchDeleteGroupMemberSubjectSystemGroupCache(systemID, subjectPK)
	}

	return err
}

func (c *policyControllerV2) alterABACPolicies(
	tx *sqlx.Tx,
	subjectPK, templateID int64,
	createPolicies []types.Policy,
	updatePolicies []types.Policy,
	deletePolicyIDs []int64,
	actionPKMap map[string]int64,
	actionPKWithResourceTypeSet *set.Int64Set,
) (updatedActionPKExpressionPKs map[int64][]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTLV2, "alterABACPolicies")

	// 避免无需变更情况，也进行各种数据查询
	if len(createPolicies) == 0 && len(updatePolicies) == 0 && len(deletePolicyIDs) == 0 {
		return
	}

	// 数据转换
	cps, err := convertToServicePolicies(subjectPK, createPolicies, actionPKMap)
	if err != nil {
		err = errorWrapf(
			err,
			"convertServicePolicies create policies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK,
			createPolicies,
			actionPKMap,
		)
		return
	}
	ups, err := convertToServicePolicies(subjectPK, updatePolicies, actionPKMap)
	if err != nil {
		err = errorWrapf(
			err,
			"convertServicePolicies update policies subjectPK=`%d`, policies=`%+v`, actionMap=`%+v` fail",
			subjectPK,
			updatePolicies,
			actionPKMap,
		)
		return
	}

	// 自定义权限
	if templateID == 0 {
		// service执行 create, update, delete
		updatedActionPKExpressionPKs, err = c.policyService.AlterCustomPoliciesWithTx(
			tx, subjectPK, cps, ups, deletePolicyIDs, actionPKWithResourceTypeSet,
		)
		if err != nil {
			err = errorWrapf(err, "policyService.AlterPolicies subjectPK=`%d` fail", subjectPK)
			return
		}
		// publish policy delete event
		c.eventProducer.PublishABACDeletePolicyEvent(deletePolicyIDs)
		// Note: 这里必须直接返回，否则会走到模板权限逻辑
		return
	}

	// 模板权限
	// 创建&删除
	if len(createPolicies) > 0 || len(deletePolicyIDs) > 0 {
		err = c.policyService.CreateAndDeleteTemplatePoliciesWithTx(
			tx, subjectPK, templateID, cps, deletePolicyIDs, actionPKWithResourceTypeSet)
		if err != nil {
			err = errorWrapf(err, "policyService.CreateAndDeleteTemplatePolicies subjectPK=`%d` fail", subjectPK)
			return
		}

		// publish policy delete event
		c.eventProducer.PublishABACDeletePolicyEvent(deletePolicyIDs)
	}
	// 更新
	if len(updatePolicies) > 0 {
		err = c.policyService.UpdateTemplatePoliciesWithTx(tx, subjectPK, ups, actionPKWithResourceTypeSet)
		if err != nil {
			err = errorWrapf(err, "policyService.UpdateTemplatePolicies subjectPK=`%d` fail", subjectPK)
			return
		}
	}

	return updatedActionPKExpressionPKs, nil
}

func (c *policyControllerV2) alterRBACPolicies(
	tx *sqlx.Tx,
	groupPK, templateID int64,
	systemID string,
	resourceChangedActions []types.ResourceChangedAction,
) (resourceChangedContents []svctypes.ResourceChangedContent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTLV2, "alterRBACPolicy")

	// 避免无需变更情况，也进行各种数据查询
	if len(resourceChangedActions) == 0 {
		return
	}

	// 1. 将ResourceChangedAction数据转换为service层函数需要的ResourceChangedContent数据
	resourceChangedContents, err = c.convertToResourceChangedContent(systemID, resourceChangedActions)
	if err != nil {
		return resourceChangedContents, errorWrapf(
			err,
			"convertToResourceChangedContent systemID=`%s` resourceChangedActions=`%v` fail",
			systemID, resourceChangedActions,
		)
	}

	// 2. 变更RBAC策略
	var deletedPolicyPKs []int64
	deletedPolicyPKs, err = c.groupResourcePolicyService.Alter(
		tx,
		groupPK,
		templateID,
		systemID,
		resourceChangedContents,
	)
	if err != nil {
		return resourceChangedContents, errorWrapf(
			err,
			"groupResourcePolicyService.Alter "+
				"groupPK=`%d` templateID=`%d` system=`%s` resourceChangedContents=`%v` fail",
			groupPK, templateID, systemID, resourceChangedContents,
		)
	}

	// publish the rbac delete pks to the engine redis queue
	c.eventProducer.PublishRBACDeletePolicyEvent(deletedPolicyPKs)

	return resourceChangedContents, nil
}

func (c *policyControllerV2) queryResourceTypePK(
	resourceChangedActions *[]types.ResourceChangedAction,
) (resourceTypePKMap map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTLV2, "queryResourceTypePK")

	resourceTypePKMap = make(map[string]int64, len(*resourceChangedActions))
	for _, rca := range *resourceChangedActions {
		pk, err := cacheimpls.GetLocalResourceTypePK(rca.Resource.System, rca.Resource.Type)
		if err != nil {
			return nil, errorWrapf(
				err,
				"cacheimpls.GetLocalResourceTypePK system=`%s` type=`%s` fail",
				rca.Resource.System,
				rca.Resource.Type,
			)
		}
		resourceTypePKMap[rca.Resource.System+":"+rca.Resource.Type] = pk
	}

	return resourceTypePKMap, nil
}

func (c *policyControllerV2) queryActionDetail(
	systemID string, resourceChangedActions *[]types.ResourceChangedAction,
) (actionDetailMap map[string]svctypes.ActionDetail, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTLV2, "queryActionDetail")

	// 1. 只查询有需要的Action
	actionIDSet := set.NewStringSet()
	for _, rca := range *resourceChangedActions {
		actionIDSet.Append(rca.CreatedActionIDs...)
		actionIDSet.Append(rca.DeletedActionIDs...)
	}

	// 2. 遍历Action，从缓存里查询每个Action的Detail
	actionDetailMap = make(map[string]svctypes.ActionDetail, actionIDSet.Size())
	for _, actionID := range actionIDSet.ToSlice() {
		detail, err := cacheimpls.GetActionDetail(systemID, actionID)
		if err != nil {
			return actionDetailMap, errorWrapf(
				err, "cacheimpls.GetActionDetail system=`%s`, actionID=`%s` fail", systemID, actionID,
			)
		}
		actionDetailMap[actionID] = detail
	}

	return
}

type changedAction struct {
	CreatedActionPKs []int64
	DeletedActionPKs []int64
}

func (c *policyControllerV2) groupByActionRelatedResourceTypePK(
	createdActionIDs, deletedActionIDs []string,
	actionDetailMap *map[string]svctypes.ActionDetail,
) (relatedResourceTypePKToChangedActionMap map[int64]changedAction, err error) {
	// 记录每个relatedResourceTypePK对应的changedAction
	changedActions := make([]changedAction, 0, len(createdActionIDs)+len(deletedActionIDs))
	// Note: relateResourceTypePKToIndex用于记录其对应ChangedAction在changedActions数组里的位置
	relateResourceTypePKToIndex := map[int64]int{}

	for _, actionID := range createdActionIDs {
		detail := (*actionDetailMap)[actionID]
		// Note: 由于只能关联一个资源类型的操作才可配置RBAC权限，所以这里直接取第一个关联的资源类型
		// pk := detail.ResourceTypes[0].PK
		resourceType := detail.ResourceTypes[0]
		pk, err := cacheimpls.GetLocalResourceTypePK(resourceType.System, resourceType.ID)
		if err != nil {
			return nil, err
		}

		if _, ok := relateResourceTypePKToIndex[pk]; !ok {
			changedActions = append(changedActions, changedAction{
				CreatedActionPKs: []int64{},
				DeletedActionPKs: []int64{},
			})
			relateResourceTypePKToIndex[pk] = len(changedActions) - 1
		}

		idx := relateResourceTypePKToIndex[pk]
		changedActions[idx].CreatedActionPKs = append(changedActions[idx].CreatedActionPKs, detail.PK)
	}

	for _, actionID := range deletedActionIDs {
		detail := (*actionDetailMap)[actionID]
		// Note: 由于只能关联一个资源类型的操作才可配置RBAC权限，所以这里直接取第一个关联的资源类型
		resourceType := detail.ResourceTypes[0]
		pk, err := cacheimpls.GetLocalResourceTypePK(resourceType.System, resourceType.ID)
		if err != nil {
			return nil, err
		}

		if _, ok := relateResourceTypePKToIndex[pk]; !ok {
			changedActions = append(changedActions, changedAction{
				CreatedActionPKs: []int64{},
				DeletedActionPKs: []int64{},
			})
			relateResourceTypePKToIndex[pk] = len(changedActions) - 1
		}

		idx := relateResourceTypePKToIndex[pk]
		changedActions[idx].DeletedActionPKs = append(changedActions[idx].DeletedActionPKs, detail.PK)
	}

	relatedResourceTypePKToChangedActionMap = make(map[int64]changedAction, len(relateResourceTypePKToIndex))
	for pk, idx := range relateResourceTypePKToIndex {
		relatedResourceTypePKToChangedActionMap[pk] = changedActions[idx]
	}

	return relatedResourceTypePKToChangedActionMap, nil
}

func (c *policyControllerV2) convertToResourceChangedContent(
	systemID string, resourceChangedActions []types.ResourceChangedAction,
) (resourceChangedContents []svctypes.ResourceChangedContent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicyCTLV2, "convertToResourceChangedContent")

	// 1. 查询每个操作的详情
	actionDetailMap, err := c.queryActionDetail(systemID, &resourceChangedActions)
	if err != nil {
		return resourceChangedContents, errorWrapf(
			err, "queryActionDetail systemID=`%s` resourceChangedActions=`%v` fail", systemID, resourceChangedActions,
		)
	}

	// 2. 查询每个资源类型的PK
	resourceTypePKMap, err := c.queryResourceTypePK(&resourceChangedActions)
	if err != nil {
		return resourceChangedContents, errorWrapf(
			err, "queryResourceTypePK resourceChangedActions=`%v` fail", resourceChangedActions,
		)
	}

	// 3. 组装数据
	resourceChangedContents = make([]svctypes.ResourceChangedContent, 0, 3*len(resourceChangedActions))
	for _, rca := range resourceChangedActions {
		// 根据ActionRelatedResourceTypePK对Action进行分组
		relatedResourceTypePKToChangedActionMap, err := c.groupByActionRelatedResourceTypePK(
			rca.CreatedActionIDs, rca.DeletedActionIDs, &actionDetailMap,
		)
		if err != nil {
			return nil, errorWrapf(
				err, "groupByActionRelatedResourceTypePK rca=`%v` fail", rca,
			)
		}

		// 组织最终数据
		resourceTypePK := resourceTypePKMap[rca.Resource.System+":"+rca.Resource.Type]
		for relatedResourceTypePK, ca := range relatedResourceTypePKToChangedActionMap {
			resourceChangedContents = append(resourceChangedContents, svctypes.ResourceChangedContent{
				ResourceTypePK:              resourceTypePK,
				ResourceID:                  rca.Resource.ID,
				ActionRelatedResourceTypePK: relatedResourceTypePK,
				CreatedActionPKs:            ca.CreatedActionPKs,
				DeletedActionPKs:            ca.DeletedActionPKs,
			})
		}
	}

	return resourceChangedContents, nil
}
