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
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"
)

// BatchCreateActions godoc
// @Summary batch actions create
// @Description batch create actions
// @ID api-model-action-create
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param body body []actionSerializer true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/actions [post]
func BatchCreateActions(c *gin.Context) {
	var body []actionSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if len(body) == 0 {
		util.BadRequestErrorJSONResponse(c, "the array should contain at least 1 item")
		return
	}
	// NOTE: 这里会校验 related_resource_types 中 id 唯一
	if valid, message := validateAction(body); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	systemID := c.Param("system_id")

	// check action repeat
	if err := validateActionsRepeat(body); err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	// check action id/name/name_en exists
	err := checkActionsQuotaAndAllUnique(systemID, body)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	// check related resource type exist
	err = checkActionCreateResourceTypeAllExists(body)
	if err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	svc := service.NewActionService()
	actions := make([]svctypes.Action, 0, len(body))
	for _, ac := range body {
		action := svctypes.Action{
			ID:            ac.ID,
			Name:          ac.Name,
			NameEn:        ac.NameEn,
			Description:   ac.Description,
			DescriptionEn: ac.DescriptionEn,
			Type:          ac.Type,
			Version:       ac.Version,

			RelatedActions: ac.RelatedActions,
		}
		action.RelatedResourceTypes = convertToRelatedResourceTypes(ac.RelatedResourceTypes)

		actions = append(actions, action)
	}
	err = svc.BulkCreate(systemID, actions)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "BatchCreateActions", "systemID=`%s`", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	util.SuccessJSONResponse(c, "ok", nil)
}

// UpdateAction godoc
// @Summary action update
// @Description update action
// @ID api-model-action-update
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param action_id path string true "Action ID"
// @Param body body actionUpdateSerializer true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/action/{action_id} [put]
func UpdateAction(c *gin.Context) {
	systemID := c.Param("system_id")

	// validate
	var body actionUpdateSerializer
	err := c.ShouldBindBodyWith(&body, binding.JSON)
	if err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	var data map[string]interface{}
	err = c.ShouldBindBodyWith(&data, binding.JSON)
	if err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if len(data) == 0 {
		util.BadRequestErrorJSONResponse(c, "fields required, should not be empty json")
		return
	}
	valid, message := body.validate(data)
	if !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	// check action id exists, and name/name_en unique
	actionID := c.Param("action_id")
	err = checkActionUpdateUnique(systemID, actionID, body.Name, body.NameEn)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	// check related resource type exist
	if len(body.RelatedResourceTypes) > 0 {
		err = checkActionUpdateResourceTypeAllExists(actionID, body.RelatedResourceTypes)
		if err != nil {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}
	}

	if _, ok := data["related_resource_types"]; ok {
		// NOTE: the action related_resource_types should not be changed if action has any policies!!!!!!
		err = checkUpdateActionRelatedResourceTypeNotChanged(systemID, actionID,
			body.RelatedResourceTypes)
		if err != nil {
			util.ConflictJSONResponse(c, err.Error())
			return
		}
	}

	// build the data
	allowEmptyFields := svctypes.NewAllowEmptyFields()
	if _, ok := data["type"]; ok {
		allowEmptyFields.AddKey("Type")
	}
	if _, ok := data["related_resource_types"]; ok {
		allowEmptyFields.AddKey("RelatedResourceTypes")
	}
	if _, ok := data["related_actions"]; ok {
		allowEmptyFields.AddKey("RelatedActions")
	}
	if _, ok := data["description"]; ok {
		allowEmptyFields.AddKey("Description")
	}
	if _, ok := data["description_en"]; ok {
		allowEmptyFields.AddKey("DescriptionEn")
	}

	action := svctypes.Action{
		Name:                 body.Name,
		NameEn:               body.NameEn,
		Description:          body.Description,
		DescriptionEn:        body.DescriptionEn,
		Version:              body.Version,
		Type:                 body.Type,
		RelatedResourceTypes: convertToRelatedResourceTypes(body.RelatedResourceTypes),
		RelatedActions:       body.RelatedActions,

		AllowEmptyFields: allowEmptyFields,
	}

	svc := service.NewActionService()
	err = svc.Update(systemID, actionID, action)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "UpdateAction",
			"systemID=`%s`, actionID=`%s`", systemID, actionID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// delete from cache
	cacheimpls.BatchDeleteActionCache(systemID, []string{actionID})

	util.SuccessJSONResponse(c, "ok", nil)
}

// DeleteAction godoc
// @Summary action delete
// @Description delete action
// @ID api-model-action-delete
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param action_id path string true "Action ID"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/action/{action_id} [delete]
func DeleteAction(c *gin.Context) {
	systemID := c.Param("system_id")
	actionID := c.Param("action_id")

	ids := []string{actionID}

	batchDeleteActions(c, systemID, ids)
}

// BatchDeleteActions godoc
// @Summary actions batch delete
// @Description batch delete actions
// @ID api-model-action-batch-delete
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param resource_type_id path string true "Resource Type ID"
// @Param body body []deleteViaID true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/actions [delete]
func BatchDeleteActions(c *gin.Context) {
	systemID := c.Param("system_id")

	ids, err := validateDeleteViaID(c)
	if err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	batchDeleteActions(c, systemID, ids)
}

func batchDeleteActions(c *gin.Context, systemID string, ids []string) {
	// NOTE: migration will ignore the check
	checkExistence := c.Query("check_existence")
	if checkExistence != "false" {
		// check action exist
		err := checkActionIDsExist(systemID, ids)
		if err != nil {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}
	}

	// NOTE: the action should not be deleted if action has any policies!!!!!!
	needAsyncDeletedActionIDs, err := checkActionIDsHasAnyPolicies(systemID, ids)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	// 如果存在需要异步删除Action模型的，则添加对应事件
	if len(needAsyncDeletedActionIDs) > 0 {
		// 创建异步删除Action的事件
		eventSvc := service.NewModelChangeService()
		events := make([]svctypes.ModelChangeEvent, 0, len(needAsyncDeletedActionIDs))
		for _, id := range needAsyncDeletedActionIDs {
			actionPK, err1 := cacheimpls.GetActionPK(systemID, id)
			if err1 != nil {
				err1 = errorx.Wrapf(err1, "Handler", "batchDeleteActions",
					"query action pk fail, systemID=`%s`, ids=`%v`", systemID, ids)
				util.SystemErrorJSONResponse(c, err1)
				return
			}
			exist, err1 := eventSvc.ExistByTypeModel(ModelChangeEventTypeActionDeleted, ModelChangeEventStatusPending,
				ModelChangeEventModelTypeAction, actionPK)
			if err1 != nil {
				err1 = errorx.Wrapf(err1, "Handler", "batchDeleteActions",
					"eventSvc.ExistByTypeModel fail, systemID=`%s`, ids=`%v`", systemID, ids)
				util.SystemErrorJSONResponse(c, err1)
				return
			}
			if exist {
				continue
			}
			events = append(events, svctypes.ModelChangeEvent{
				Type:      ModelChangeEventTypeActionDeleted,
				Status:    ModelChangeEventStatusPending,
				SystemID:  systemID,
				ModelType: ModelChangeEventModelTypeAction,
				ModelID:   id,
				ModelPK:   actionPK,
			})
		}
		if len(events) != 0 {
			err = eventSvc.BulkCreate(events)
			if err != nil {
				err = errorx.Wrapf(err, "Handler", "batchDeleteActions",
					"eventSvc.BulkCreate events=`%+v` fail", events)
				util.SystemErrorJSONResponse(c, err)
				return
			}
		}
	}

	// 从同步删除的操作里剔除掉需要异步删除的
	asyncDeletedIDs := set.NewStringSetWithValues(needAsyncDeletedActionIDs)
	newIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		if !asyncDeletedIDs.Has(id) {
			newIDs = append(newIDs, id)
		}
	}
	if len(newIDs) > 0 {
		svc := service.NewActionService()
		err = svc.BulkDelete(systemID, newIDs)
		if err != nil {
			err = errorx.Wrapf(err, "Handler", "batchDeleteActions",
				"systemID=`%s`, ids=`%v`", systemID, newIDs)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		// delete from cache
		cacheimpls.BatchDeleteActionCache(systemID, newIDs)
	}

	util.SuccessJSONResponse(c, "ok", nil)
}
