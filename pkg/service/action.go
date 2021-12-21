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

/*
service

types定义的数据结构的加载层
可以从db redis memory中加载
*/

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"database/sql"

	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/database/sdao"
	"iam/pkg/errorx"
	"iam/pkg/service/types"
)

// ActionSVC ...
const ActionSVC = "ActionSVC"

// ActionService action加载器
type ActionService interface {
	// in this file

	GetActionPK(system, id string) (int64, error)

	Get(system, id string) (types.Action, error)
	ListBySystem(system string) ([]types.Action, error)

	BulkCreate(system string, actions []types.Action) error
	Update(system, actionID string, action types.Action) error
	BulkDelete(system string, actionIDs []string) error

	GetThinActionByPK(pk int64) (types.ThinAction, error)

	// in action_thin.go

	ListThinActionByPKs(pks []int64) ([]types.ThinAction, error)
	ListThinActionBySystem(system string) ([]types.ThinAction, error)

	// in action_resource_type.go

	ListThinActionResourceTypes(system string, actionID string) ([]types.ThinActionResourceType, error)
	// ListActionResourceTypeIDByResourceTypeSystem 获取关联的资源类型 by 关联资源类型的系统id
	ListActionResourceTypeIDByResourceTypeSystem(resourceTypeSystem string) ([]types.ActionResourceTypeID, error)
	// ListActionResourceTypeIDByActionSystem 获取关联资源类型 by 操作的系统id
	ListActionResourceTypeIDByActionSystem(actionSystem string) ([]types.ActionResourceTypeID, error)

	// in action_instance_selection.go

	ListActionInstanceSelectionIDBySystem(system string) ([]types.ActionInstanceSelectionID, error)
}

type actionService struct {
	manager                       dao.ActionManager
	actionResourceTypeManager     dao.ActionResourceTypeManager
	saasManager                   sdao.SaaSActionManager
	saasActionResourceTypeManager sdao.SaaSActionResourceTypeManager
	saasInstanceSelectionManager  sdao.SaaSInstanceSelectionManager
}

// NewActionService ActionService 工厂
func NewActionService() ActionService {
	return &actionService{
		manager:                       dao.NewActionManager(),
		actionResourceTypeManager:     dao.NewActionResourceTypeManager(),
		saasManager:                   sdao.NewSaaSActionManager(),
		saasActionResourceTypeManager: sdao.NewSaaSActionResourceTypeManager(),
		saasInstanceSelectionManager:  sdao.NewSaaSInstanceSelectionManager(),
	}
}

// GetActionPK 获取action pk
func (l *actionService) GetActionPK(system, id string) (int64, error) {
	return l.manager.GetPK(system, id)
}

// GetThinActionByPK ...
func (l *actionService) GetThinActionByPK(pk int64) (sa types.ThinAction, err error) {
	action, err := l.manager.Get(pk)
	if err != nil {
		return
	}

	sa = types.ThinAction{
		PK:     action.PK,
		System: action.System,
		ID:     action.ID,
	}
	return
}

// Get 获取action详细信息
func (l *actionService) Get(system, actionID string) (types.Action, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "Get")

	action := types.Action{}
	dbActionResourceTypes, err := l.actionResourceTypeManager.ListResourceTypeByAction(system, actionID)
	if err != nil {
		return action, errorWrapf(err, "ListResourceTypeByAction(system=%s, actionID=%s) fail", system, actionID)
	}

	dbSaaSActionResourceTypes, err := l.saasActionResourceTypeManager.ListByActionID(system, actionID)
	if err != nil {
		return action, errorWrapf(err, "ListByActionID(system=%s, actionID=%s) fail", system, actionID)
	}

	dbAction, err := l.saasManager.Get(system, actionID)
	if err != nil {
		return action, errorWrapf(err, "saasManager.Get(system=%s, actionID=%s) fail", system, actionID)
	}

	action = types.Action{
		ID:      dbAction.ID,
		Name:    dbAction.Name,
		NameEn:  dbAction.NameEn,
		Type:    dbAction.Type,
		Version: dbAction.Version,
	}
	relatedResourceTypes := []types.ActionResourceType{}

	for idx := range dbActionResourceTypes {
		sart := &dbSaaSActionResourceTypes[idx]
		// art := &dbActionResourceTypes[idx]

		var actionResourceType types.ActionResourceType
		actionResourceType, err = l.toServiceActionResourceType(sart)
		if err != nil {
			return action, errorWrapf(err, "toServiceActionResourceType(system=%s, actionID=%s) fail", system, actionID)
		}

		relatedResourceTypes = append(relatedResourceTypes, actionResourceType)
	}
	action.RelatedResourceTypes = relatedResourceTypes
	return action, nil
}

// ListBySystem ...
func (l *actionService) ListBySystem(system string) ([]types.Action, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListBySystem")

	actions := []types.Action{}
	dbActionResourceTypes, err := l.actionResourceTypeManager.ListByActionSystem(system)
	if err != nil {
		return nil, errorWrapf(err, "actionResourceTypeManager.ListByActionSystem system=`%s` fail", system)
	}
	actionResourceTypeMap := map[string][]dao.ActionResourceType{}
	for _, art := range dbActionResourceTypes {
		_, ok := actionResourceTypeMap[art.ActionID]
		if !ok {
			actionResourceTypeMap[art.ActionID] = []dao.ActionResourceType{}
		}
		actionResourceTypeMap[art.ActionID] = append(actionResourceTypeMap[art.ActionID], art)
	}

	dbSaaSActionResourceTypes, err := l.saasActionResourceTypeManager.ListByActionSystem(system)
	if err != nil {
		return nil, errorWrapf(err, "saasActionResourceTypeManager.ListByActionSystem system=`%s` fail", system)
	}
	saasActionResourceTypeMap := map[string][]sdao.SaaSActionResourceType{}
	for _, sart := range dbSaaSActionResourceTypes {
		_, ok := saasActionResourceTypeMap[sart.ActionID]
		if !ok {
			saasActionResourceTypeMap[sart.ActionID] = []sdao.SaaSActionResourceType{}
		}
		saasActionResourceTypeMap[sart.ActionID] = append(saasActionResourceTypeMap[sart.ActionID], sart)
	}

	dbActions, err := l.saasManager.ListBySystem(system)
	if err != nil {
		return nil, errorWrapf(err, "saasManager.ListBySystem system=`%s` fail", system)
	}

	for _, ac := range dbActions {
		action := types.Action{
			ID:            ac.ID,
			Name:          ac.Name,
			NameEn:        ac.NameEn,
			Description:   ac.Description,
			DescriptionEn: ac.DescriptionEn,
			Type:          ac.Type,
			Version:       ac.Version,
		}
		if ac.RelatedActions != "" {
			err = jsoniter.UnmarshalFromString(ac.RelatedActions, &action.RelatedActions)
			if err != nil {
				return nil, errorWrapf(err, "unmarshal action.RelatedActions=`%+v` fail", ac.RelatedActions)
			}
		}

		relatedResourceTypes := []types.ActionResourceType{}
		_, ok := actionResourceTypeMap[ac.ID]
		if !ok {
			action.RelatedResourceTypes = relatedResourceTypes
			actions = append(actions, action)

			continue
		}

		for idx := range actionResourceTypeMap[ac.ID] {
			sart := &saasActionResourceTypeMap[ac.ID][idx]
			// art := &actionResourceTypeMap[ac.ID][idx]

			var actionResourceType types.ActionResourceType
			actionResourceType, err = l.toServiceActionResourceType(sart)
			if err != nil {
				// return nil, errorWrapf(err, "toServiceActionResourceType art=`%+v`, sart=`%+v` fail", art, sart)
				return nil, errorWrapf(err, "toServiceActionResourceType art=`%+v`, sart=`%+v` fail", sart)
			}

			relatedResourceTypes = append(relatedResourceTypes, actionResourceType)
		}
		action.RelatedResourceTypes = relatedResourceTypes
		actions = append(actions, action)
	}
	return actions, err
}

func (l *actionService) convertToDBRelatedResourceTypes(
	system string,
	action types.Action,
) (dbActionResourceTypes []dao.ActionResourceType, dbSaaSActionResourceTypes []sdao.SaaSActionResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "convertToDbRelatedResourceTypes")

	for _, rt := range action.RelatedResourceTypes {
		var relatedInstanceSelections string
		if len(rt.RelatedInstanceSelections) > 0 {
			relatedInstanceSelections, err = jsoniter.MarshalToString(rt.RelatedInstanceSelections)
			if err != nil {
				err = errorWrapf(err, "marshal rt.RelatedInstanceSelections=`%+v` fail", rt.RelatedInstanceSelections)
				return nil, nil, err
			}
		}
		dbActionResourceTypes = append(dbActionResourceTypes, dao.ActionResourceType{
			ActionSystem:       system,
			ActionID:           action.ID,
			ResourceTypeSystem: rt.System,
			ResourceTypeID:     rt.ID,
		})
		dbSaaSActionResourceTypes = append(dbSaaSActionResourceTypes, sdao.SaaSActionResourceType{
			ActionSystem:              system,
			ActionID:                  action.ID,
			ResourceTypeSystem:        rt.System,
			ResourceTypeID:            rt.ID,
			NameAlias:                 rt.NameAlias,
			NameAliasEn:               rt.NameAliasEn,
			SelectionMode:             rt.SelectionMode,
			RelatedInstanceSelections: relatedInstanceSelections,
		})
	}

	return dbActionResourceTypes, dbSaaSActionResourceTypes, nil
}

// BulkCreate ...
func (l *actionService) BulkCreate(system string, actions []types.Action) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "BulkCreate")

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error%s", "")
	}

	// 数据转换
	dbActions := make([]dao.Action, 0, len(actions))
	dbActionResourceTypes := []dao.ActionResourceType{}
	dbSaaSActions := make([]sdao.SaaSAction, 0, len(actions))
	dbSaaSActionResourceTypes := []sdao.SaaSActionResourceType{}
	for _, ac := range actions {
		dbActions = append(dbActions, dao.Action{
			System: system,
			ID:     ac.ID,
		})

		relatedActions, err1 := jsoniter.MarshalToString(ac.RelatedActions)
		if err1 != nil {
			return errorWrapf(err1, "marshal action.RelatedActions=`%+v` fail", ac.RelatedActions)
		}

		dbSaaSActions = append(dbSaaSActions, sdao.SaaSAction{
			System:         system,
			ID:             ac.ID,
			Name:           ac.Name,
			NameEn:         ac.NameEn,
			Description:    ac.Description,
			DescriptionEn:  ac.DescriptionEn,
			RelatedActions: relatedActions,
			Type:           ac.Type,
			Version:        ac.Version,
		})

		singleDBActionResourceTypes, singleDBSaaSActionResourceTypes, err1 := l.convertToDBRelatedResourceTypes(system, ac)
		if err1 != nil {
			return errorWrapf(err1, "convertToDbRelatedResourceTypes system=`%s`, action=`%+v`", system, ac)
		}

		dbActionResourceTypes = append(dbActionResourceTypes, singleDBActionResourceTypes...)
		dbSaaSActionResourceTypes = append(dbSaaSActionResourceTypes, singleDBSaaSActionResourceTypes...)
	}

	// 执行插入
	err = l.manager.BulkCreateWithTx(tx, dbActions)
	if err != nil {
		return errorWrapf(err, "manager.BulkCreateWithTx dbActions=`%+v` fail", dbActions)
	}
	if len(dbActionResourceTypes) > 0 {
		err = l.actionResourceTypeManager.BulkCreateWithTx(tx, dbActionResourceTypes)
		if err != nil {
			return errorWrapf(err,
				"actionResourceTypeManager.BulkCreateWithTx dbAct6ionResourceTypes=`%s` fail",
				dbActionResourceTypes)
		}
	}

	err = l.saasManager.BulkCreateWithTx(tx, dbSaaSActions)
	if err != nil {
		return errorWrapf(err, "saasManager.BulkCreateWithTx dbSaaSActions=`%+v` fail", dbSaaSActions)
	}
	if len(dbSaaSActionResourceTypes) > 0 {
		err = l.saasActionResourceTypeManager.BulkCreateWithTx(tx, dbSaaSActionResourceTypes)
		if err != nil {
			return errorWrapf(err,
				"saasActionResourceTypeManager.BulkCreateWithTx dbSaaSActionResourceTypes=`%+v` fail",
				dbSaaSActionResourceTypes)
		}
	}

	return tx.Commit()
}

// Update ...
func (l *actionService) Update(system, actionID string, action types.Action) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "Update")

	// use tx
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error%s", "")
	}

	// FIXME: for the bug below
	action.ID = actionID

	if action.AllowEmptyFields.HasKey("RelatedResourceTypes") {
		// 1. delete from saas_action_resource_type where action_system_id=xxx and action_id=xxxxx
		//    delete from action_resource_type where action_system_id=xxx and action_id=xxxxx
		err = l.saasActionResourceTypeManager.BulkDeleteWithTx(tx, system, []string{actionID})
		if err != nil {
			return errorWrapf(err, "saasActionResourceTypeManager.BulkDeleteWithTx system=`%s`, actionID=`%s`",
				system, actionID)
		}
		err = l.actionResourceTypeManager.BulkDeleteWithTx(tx, system, []string{actionID})
		if err != nil {
			return errorWrapf(err, "actionResourceTypeManager.BulkDeleteWithTx system=`%s`, actionID=`%s`",
				system, actionID)
		}

		// 2. insert into saas_action_resource_type ()
		//    insert into action_resource_type
		// TODO: 与创建逻辑一样，代码需复用
		// NOTE: bug here maybe, 这里如果action.ID没有传, 会更新空到db
		dbActionResourceTypes, sarts, err1 := l.convertToDBRelatedResourceTypes(system, action)
		if err1 != nil {
			return errorWrapf(err, "convertToDbRelatedResourceTypes system=`%s`, action=`%+v`", system, action)
		}

		err = l.saasActionResourceTypeManager.BulkCreateWithTx(tx, sarts)
		if err != nil {
			return errorWrapf(err, "saasActionResourceTypeManager.BulkCreateWithTx sarts=`%+v`", sarts)
		}
		err = l.actionResourceTypeManager.BulkCreateWithTx(tx, dbActionResourceTypes)
		if err != nil {
			return errorWrapf(err, "actionResourceTypeManager.BulkCreateWithTx actionResourceTypes=`%+v`", dbActionResourceTypes)
		}
	}

	allowBlank := database.NewAllowBlankFields()
	// action.Type is set => set saasAction.Type
	if action.AllowEmptyFields.HasKey("Type") {
		allowBlank.AddKey("Type")
	}
	if action.AllowEmptyFields.HasKey("Description") {
		allowBlank.AddKey("Description")
	}
	if action.AllowEmptyFields.HasKey("DescriptionEn") {
		allowBlank.AddKey("DescriptionEn")
	}

	var relatedActions string
	if action.AllowEmptyFields.HasKey("RelatedActions") {
		allowBlank.AddKey("RelatedActions")

		var err1 error
		relatedActions, err1 = jsoniter.MarshalToString(action.RelatedActions)
		if err1 != nil {
			return errorWrapf(err, "unmarshal action.RelatedActions=`%+v` fail", action.RelatedActions)
		}
	}

	// 4. update saas action
	data := sdao.SaaSAction{
		Name:           action.Name,
		NameEn:         action.NameEn,
		Description:    action.Description,
		DescriptionEn:  action.DescriptionEn,
		Type:           action.Type,
		Version:        action.Version,
		RelatedActions: relatedActions,

		AllowBlankFields: allowBlank,
	}

	// use tx here
	err = l.saasManager.Update(tx, system, actionID, data)
	if err != nil {
		return errorWrapf(err, "saasManager.Update system=`%s`, actionID=`%s`, data=`%+v`",
			system, actionID, data)
	}
	return tx.Commit()
}

// BulkDelete ...
func (l *actionService) BulkDelete(system string, actionIDs []string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "BulkDelete")

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error%s", "")
	}

	err = l.manager.BulkDeleteWithTx(tx, system, actionIDs)
	if err != nil {
		return errorWrapf(err, "manager.BulkDeleteWithTx system=`%s`, actionIDs=`%+v` fail", system, actionIDs)
	}

	err = l.actionResourceTypeManager.BulkDeleteWithTx(tx, system, actionIDs)
	if err != nil {
		return errorWrapf(err, "actionResourceTypeManager.BulkDeleteWithTx system=`%s`, actionIDs=`%+v` fail",
			system, actionIDs)
	}

	err = l.saasManager.BulkDeleteWithTx(tx, system, actionIDs)
	if err != nil {
		return errorWrapf(err, "saasManager.BulkDeleteWithTx system=`%s`, actionIDs=`%+v` fail",
			system, actionIDs)
	}

	err = l.saasActionResourceTypeManager.BulkDeleteWithTx(tx, system, actionIDs)
	if err != nil {
		return errorWrapf(err, "saasActionResourceTypeManager.BulkDeleteWithTx system=`%s`, actionIDs=`%+v` fail",
			system, actionIDs)
	}

	return tx.Commit()
}

func (l *actionService) toServiceActionResourceType(
	sart *sdao.SaaSActionResourceType,
) (types.ActionResourceType, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "toServiceActionResourceType")

	// 转换类型
	actionResourceType := types.ActionResourceType{
		System:        sart.ResourceTypeSystem,
		ID:            sart.ResourceTypeID,
		NameAlias:     sart.NameAlias,
		NameAliasEn:   sart.NameAliasEn,
		SelectionMode: sart.SelectionMode,
	}

	// NOTE: 底层存储实例视图的引用: rawInstanceSelections is {"system_id": a, "id": b}
	//       给前端SaaS的时候, 查到真实的 实例视图 返回
	// TODO: 对于模型的查询接口，不应该填充，这整个函数应该拆成两个，一个是SaaS Web API的，一个是模型API的
	instanceSelections, err := l.fillRelatedInstanceSelections(sart.RelatedInstanceSelections)
	if err != nil {
		return actionResourceType, errorWrapf(err, "fillRelatedInstanceSelections rawString=`%+v` fail",
			sart.RelatedInstanceSelections)
	}
	actionResourceType.InstanceSelections = instanceSelections

	return actionResourceType, nil
}

func (l *actionService) fillRelatedInstanceSelections(rawRelatedInstanceSelections string) (
	instanceSelections []map[string]interface{}, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "fillRelatedInstanceSelections")
	// rawRelatedInstanceSelections is {"system_id": a, "id": b}
	if rawRelatedInstanceSelections == "" {
		return
	}

	relatedInstanceSelections := []types.ReferenceInstanceSelection{}
	err = jsoniter.UnmarshalFromString(rawRelatedInstanceSelections, &relatedInstanceSelections)
	if err != nil {
		err = errorWrapf(err, "unmarshal rawRelatedInstanceSelections=`%s` fail", rawRelatedInstanceSelections)
		return
	}

	// NOTE: here query one by one
	for _, r := range relatedInstanceSelections {
		is, err1 := l.saasInstanceSelectionManager.Get(r.System, r.ID)
		if err1 != nil {
			// NOTE: if not exists, continue
			if err1 == sql.ErrNoRows {
				continue
			}

			err = errorWrapf(err1, "saasInstanceSelectionManager.Get system=`%s`, id=`%s` fail", r.System, r.ID)
			return
		}

		chain := []map[string]interface{}{}
		err = jsoniter.UnmarshalFromString(is.ResourceTypeChain, &chain)
		if err != nil {
			err = errorWrapf(err, "unmarshal instanceSelection.ResourceTypeChain=`%s` fail", is.ResourceTypeChain)
			return
		}

		instanceSelections = append(instanceSelections, map[string]interface{}{
			"id":                  is.ID,
			"system_id":           is.System,
			"ignore_iam_path":     r.IgnoreIAMPath,
			"is_dynamic":          is.IsDynamic,
			"name":                is.Name,
			"name_en":             is.NameEn,
			"resource_type_chain": chain,
		})
	}
	return instanceSelections, nil
}
