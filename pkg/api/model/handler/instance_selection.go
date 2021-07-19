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
	"fmt"

	"iam/pkg/cache/impls"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"iam/pkg/api/common"
	"iam/pkg/errorx"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"
)

// BatchCreateInstanceSelections godoc
// @Summary batch instance selection create
// @Description batch create instance_selections
// @ID api-model-instance-selection-create
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param body body []instanceSelectionSerializer true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/instance-selections [post]
func BatchCreateInstanceSelections(c *gin.Context) {
	var body []instanceSelectionSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if valid, message := common.ValidateArray(body); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	for index, data := range body {
		if !common.ValidIDRegex.MatchString(data.ID) {
			message := fmt.Sprintf("data in array[%d] id=%s, %s", index, data.ID, common.ErrInvalidID)
			util.BadRequestErrorJSONResponse(c, message)
			return
		}
	}

	systemID := c.Param("system_id")

	// check instance selection repeat
	if err := validateInstanceSelectionsRepeat(body); err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	// check instance selection exists
	err := checkAllInstanceSelectionsQuotaAndUnique(systemID, body)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	instanceSelections := make([]svctypes.InstanceSelection, 0, len(body))

	for _, is := range body {
		resourceTypeChain := make([]map[string]interface{}, 0, len(is.ResourceTypeChain))
		for _, c := range is.ResourceTypeChain {
			resourceTypeChain = append(resourceTypeChain, structs.Map(c))
		}
		instanceSelections = append(instanceSelections, svctypes.InstanceSelection{
			ID:                is.ID,
			Name:              is.Name,
			NameEn:            is.NameEn,
			IsDynamic:         is.IsDynamic,
			ResourceTypeChain: resourceTypeChain,
		})
	}

	svc := service.NewInstanceSelectionService()
	err = svc.BulkCreate(systemID, instanceSelections)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "BatchCreateInstanceSelections",
			"BulkCreate systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	util.SuccessJSONResponse(c, "ok", nil)
}

// UpdateInstanceSelection godoc
// @Summary instance selection update
// @Description update instance_selection
// @ID api-model-instance-selection-update
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param instance_selection_id path string true "Instance Selection ID"
// @Param body body instanceSelectionUpdateSerializer true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/instance-selections/{instance_selection_id} [put]
func UpdateInstanceSelection(c *gin.Context) {
	systemID := c.Param("system_id")

	// validate
	var body instanceSelectionUpdateSerializer
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

	// check instance_selection_id exists and name/name_en unique
	instanceSelectionID := c.Param("instance_selection_id")
	err = checkInstanceSelectionUpdateUnique(systemID, instanceSelectionID, body.Name, body.NameEn)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	// var parents []map[string]interface{} => json.Marshal to null
	// json.Marshal to []
	chain := []map[string]interface{}{}
	if len(body.ResourceTypeChain) > 0 {
		for _, p := range body.ResourceTypeChain {
			chain = append(chain, structs.Map(p))
		}
	}

	allowEmptyFields := svctypes.NewAllowEmptyFields()
	if _, ok := data["is_dynamic"]; ok {
		allowEmptyFields.AddKey("IsDynamic")
	}

	instanceSelection := svctypes.InstanceSelection{
		Name:              body.Name,
		NameEn:            body.NameEn,
		IsDynamic:         body.IsDynamic,
		ResourceTypeChain: chain,

		AllowEmptyFields: allowEmptyFields,
	}

	svc := service.NewInstanceSelectionService()
	err = svc.Update(systemID, instanceSelectionID, instanceSelection)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "UpdateInstanceSelection",
			"Update systemID=`%s`, instanceSelectionID=`%s` fail", systemID, instanceSelectionID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

// DeleteInstanceSelection godoc
// @Summary instance selection delete
// @Description delete instance_selection
// @ID api-model-instance-selection-delete
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param instance_selection_id path string true "Instance Selection ID"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/instance-selections/{instance_selection_id} [delete]
func DeleteInstanceSelection(c *gin.Context) {
	systemID := c.Param("system_id")
	instanceSelectionID := c.Param("instance_selection_id")

	ids := []string{instanceSelectionID}
	batchDeleteInstanceSelections(c, systemID, ids)
}

// BatchDeleteInstanceSelections godoc
// @Summary instance selection batch delete
// @Description batch delete instance_selection
// @ID api-model-instance-selection-batch-delete
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param instance_selection_id path string true "Instance Selection ID"
// @Param body body []deleteViaID true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/instance-selections [delete]
func BatchDeleteInstanceSelections(c *gin.Context) {
	systemID := c.Param("system_id")

	ids, err := validateDeleteViaID(c)
	if err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	batchDeleteInstanceSelections(c, systemID, ids)
}

func batchDeleteInstanceSelections(c *gin.Context, systemID string, ids []string) {
	checkExistence := c.Query("check_existence")
	if checkExistence != "false" {
		// check instance selection exist
		err := checkInstanceSelectionIDsExist(systemID, ids)
		if err != nil {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}
	}

	// check related action
	actionSvc := service.NewActionService()
	actionInstanceSelectionIDs, err := actionSvc.ListActionInstanceSelectionIDBySystem(systemID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "batchDeleteInstanceSelections",
			"actionSvc.ListActionInstanceSelectionIDByInstanceSelectionSystem systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	eventSvc := service.NewModelChangeService()
	for _, ais := range actionInstanceSelectionIDs {
		for _, id := range ids {
			// NOTE: 只检查本系统的action是否关联了对应的实例视图
			if ais.ActionSystem == systemID && ais.InstanceSelectionSystem == systemID && ais.InstanceSelectionID == id {
				actionPK, err1 := impls.GetActionPK(systemID, ais.ActionID)
				if err1 != nil {
					util.BadRequestErrorJSONResponse(c,
						fmt.Sprintf("query action pk fail, systemID=%s, id=%s", systemID, ais.ActionID))
					return
				}
				// 如果Action关联了该实例视图，则再检查是否已经有删除Action的事件
				eventExist, err1 := eventSvc.ExistByTypeModel(
					ModelChangeEventTypeActionDeleted,
					ModelChangeEventStatusPending,
					ModelChangeEventModelTypeAction,
					actionPK,
				)
				if err1 != nil {
					util.BadRequestErrorJSONResponse(c,
						fmt.Sprintf("query action model event fail, systemID=%s, id=%s, actionPK=%d",
							systemID, ais.ActionID, actionPK))
					return
				}
				if !eventExist {
					util.BadRequestErrorJSONResponse(c,
						fmt.Sprintf("instance selection id[%s] related to action[system:%s, id:%s], please unbind action",
							id, ais.ActionSystem, ais.ActionID))
					return
				}
			}
		}
	}

	svc := service.NewInstanceSelectionService()
	err = svc.BulkDelete(systemID, ids)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "batchDeleteInstanceSelections",
			"BulkDelete systemID=`%s`, ids=`%v` fail", systemID, ids)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}
