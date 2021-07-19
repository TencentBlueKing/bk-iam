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

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"iam/pkg/api/common"
	"iam/pkg/cache/impls"
	"iam/pkg/errorx"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"
)

// BatchCreateResourceTypes godoc
// @Summary batch resource type create
// @Description batch create resource_types
// @ID api-model-resource-type-create
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param body body []resourceTypeSerializer true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/resource-types [post]
func BatchCreateResourceTypes(c *gin.Context) {
	var body []resourceTypeSerializer
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

	// check resource type repeat
	if err := validateResourceTypesRepeat(body); err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	// check resource type exists
	err := checkAllResourceTypesQuotaAndUnique(systemID, body)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	resourceTypes := make([]svctypes.ResourceType, 0, len(body))

	for _, rt := range body {
		parents := make([]map[string]interface{}, 0, len(rt.Parents))
		for _, rrt := range rt.Parents {
			parents = append(parents, structs.Map(rrt))
		}
		resourceTypes = append(resourceTypes, svctypes.ResourceType{
			ID:             rt.ID,
			Name:           rt.Name,
			NameEn:         rt.NameEn,
			Description:    rt.Description,
			DescriptionEn:  rt.DescriptionEn,
			Parents:        parents,
			ProviderConfig: structs.Map(rt.ProviderConfig),
			Version:        rt.Version,
		})
	}
	svc := service.NewResourceTypeService()
	err = svc.BulkCreate(systemID, resourceTypes)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "BatchCreateResourceTypes",
			"BulkCreate systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	util.SuccessJSONResponse(c, "ok", nil)
}

// UpdateResourceType godoc
// @Summary resource type update
// @Description update resource_type
// @ID api-model-resource-type-update
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param resource_type_id path string true "Resource Type ID"
// @Param body body resourceTypeUpdateSerializer true "the request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/resource-types/{resource_type_id} [put]
func UpdateResourceType(c *gin.Context) {
	systemID := c.Param("system_id")

	// validate
	var body resourceTypeUpdateSerializer
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

	// check resource_type_id exists and name/name_en unique
	resourceTypeID := c.Param("resource_type_id")
	err = checkResourceTypeUpdateUnique(systemID, resourceTypeID, body.Name, body.NameEn)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	// var parents []map[string]interface{} => json.Marshal to null
	// json.Marshal to []
	parents := []map[string]interface{}{}
	if len(body.Parents) > 0 {
		for _, p := range body.Parents {
			parents = append(parents, structs.Map(p))
		}
	}
	var providerConfig map[string]interface{}
	if body.ProviderConfig != nil {
		providerConfig = structs.Map(body.ProviderConfig)
	}

	allowEmptyFields := svctypes.NewAllowEmptyFields()
	if _, ok := data["parents"]; ok {
		allowEmptyFields.AddKey("Parents")
	}
	if _, ok := data["description"]; ok {
		allowEmptyFields.AddKey("Description")
	}
	if _, ok := data["description_en"]; ok {
		allowEmptyFields.AddKey("DescriptionEn")
	}
	resourceType := svctypes.ResourceType{
		Name:           body.Name,
		NameEn:         body.NameEn,
		Description:    body.Description,
		DescriptionEn:  body.DescriptionEn,
		Version:        body.Version,
		Parents:        parents,
		ProviderConfig: providerConfig,

		AllowEmptyFields: allowEmptyFields,
	}

	svc := service.NewResourceTypeService()
	err = svc.Update(systemID, resourceTypeID, resourceType)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "UpdateResourceType",
			"Update systemID=`%s`, resourceTypeID=`%s` fail", systemID, resourceTypeID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// delete the cache
	impls.BatchDeleteResourceTypeCache(systemID, []string{resourceTypeID})

	util.SuccessJSONResponse(c, "ok", nil)
}

// DeleteResourceType godoc
// @Summary resource type delete
// @Description delete resource_type
// @ID api-model-resource-type-delete
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param resource_type_id path string true "Resource Type ID"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/resource-types/{resource_type_id} [delete]
func DeleteResourceType(c *gin.Context) {
	systemID := c.Param("system_id")
	resourceTypeID := c.Param("resource_type_id")

	ids := []string{resourceTypeID}
	batchDeleteResourceTypes(c, systemID, ids)
}

// BatchDeleteResourceTypes godoc
// @Summary resource type batch delete
// @Description batch delete resource_type
// @ID api-model-resource-type-batch-delete
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
// @Router /api/v1/systems/{system_id}/resource-types [delete]
func BatchDeleteResourceTypes(c *gin.Context) {
	systemID := c.Param("system_id")

	ids, err := validateDeleteViaID(c)
	if err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	batchDeleteResourceTypes(c, systemID, ids)
}

func batchDeleteResourceTypes(c *gin.Context, systemID string, ids []string) {
	checkExistence := c.Query("check_existence")
	if checkExistence != "false" {
		// check resource type exist
		err := checkResourceTypeIDsExist(systemID, ids)
		if err != nil {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}
	}

	// check related action
	actionSvc := service.NewActionService()
	actionResourceTypes, err := actionSvc.ListActionResourceTypeIDByResourceTypeSystem(systemID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "batchDeleteResourceTypes",
			"actionSvc.ListActionResourceTypeIDByResourceTypeSystem systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	eventSvc := service.NewModelChangeService()
	for _, art := range actionResourceTypes {
		for _, id := range ids {
			// NOTE: 只检查自己系统是否存在action关联了该resource type, 第三方系统依赖不影响本系统的删除
			if art.ActionSystem == systemID && art.ResourceTypeID == id {
				actionPK, err1 := impls.GetActionPK(systemID, art.ActionID)
				if err1 != nil {
					util.BadRequestErrorJSONResponse(c,
						fmt.Sprintf("query action pk fail, systemID=%s, id=%s", systemID, art.ActionID))
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
							systemID, art.ActionID, actionPK))
					return
				}
				if !eventExist {
					util.BadRequestErrorJSONResponse(c,
						fmt.Sprintf("resource type id[%s] related to action[system:%s, id:%s], please unbind action",
							id, art.ActionSystem, art.ActionID))
					return
				}
			}
		}
	}

	svc := service.NewResourceTypeService()
	err = svc.BulkDelete(systemID, ids)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "batchDeleteResourceTypes",
			"BulkDelete systemID=`%s`, ids=`%v` fail", systemID, ids)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// delete the cache
	impls.BatchDeleteResourceTypeCache(systemID, ids)

	util.SuccessJSONResponse(c, "ok", nil)
}
