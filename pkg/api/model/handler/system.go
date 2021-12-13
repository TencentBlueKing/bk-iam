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
	"iam/pkg/cacheimpls"
	"iam/pkg/errorx"

	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"iam/pkg/api/common"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"
)

// add the clientID into body.Clients if not exists, 注册这个系统的client一定是其合法client!
func defaultValidClients(c *gin.Context, originClients string) string {
	clients := originClients

	clientID := util.GetClientID(c)
	bodyClients := util.SplitStringToSet(clients, ",")
	if !bodyClients.Has(clientID) {
		bodyClients.Add(clientID)
		clients = bodyClients.ToString(",")
	}

	return clients
}

// CreateSystem godoc
// @Summary system create
// @Description register a system to iam
// @ID api-model-system-create
// @Tags model
// @Accept json
// @Produce json
// @Param body body systemSerializer true "the policy request"
// @Success 200 {object} util.Response{data=systemCreateResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems [post]
func CreateSystem(c *gin.Context) {
	// validate the body
	var body systemSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// validate the ID
	if !common.ValidIDRegex.MatchString(body.ID) {
		util.BadRequestErrorJSONResponse(c, common.ErrInvalidID.Error())
		return
	}

	// 如果关闭这个限制, 不校验
	if !common.GetSwitchDisableCreateSystemClientValidation() {
		// 创建的系统 system_id == clientID => 限制一个app只能创建一个system
		if body.ID != util.GetClientID(c) {
			util.BadRequestErrorJSONResponse(c, "system_id should be the app_code!")
			return
		}
	}

	// TODO: check the provider_config[healthz] is a valid url path

	err := checkSystemCreateUnique(body.ID, body.Name, body.NameEn)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	// 注册这个系统的client一定是其合法client!
	clients := defaultValidClients(c, body.Clients)

	// add the logical here
	system := svctypes.System{
		ID:             body.ID,
		Name:           body.Name,
		NameEn:         body.NameEn,
		Description:    body.Description,
		DescriptionEn:  body.DescriptionEn,
		Clients:        clients,
		ProviderConfig: structs.Map(body.ProviderConfig),
	}

	svc := service.NewSystemService()
	err = svc.Create(system)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "CreateSystem", "Create system=`%s` fail", system)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", systemCreateResponse{
		ID: body.ID,
	})
}

// UpdateSystem godoc
// @Summary system update
// @Description update a system
// @ID api-model-system-update
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param body body systemUpdateSerializer true "the policy request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id} [put]
func UpdateSystem(c *gin.Context) {
	systemID := c.Param("system_id")

	// validate
	var body systemUpdateSerializer
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	var data map[string]interface{}
	if err := c.ShouldBindBodyWith(&data, binding.JSON); err != nil {
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

	// check system name and name_en is unique
	err := checkSystemUpdateUnique(systemID, body.Name, body.NameEn)
	if err != nil {
		util.ConflictJSONResponse(c, err.Error())
		return
	}

	var providerConfig map[string]interface{}
	if body.ProviderConfig != nil {
		providerConfig = structs.Map(body.ProviderConfig)
	}
	allowEmptyFields := svctypes.NewAllowEmptyFields()
	if _, ok := data["description"]; ok {
		allowEmptyFields.AddKey("Description")
	}
	if _, ok := data["description_en"]; ok {
		allowEmptyFields.AddKey("DescriptionEn")
	}

	// if body.Clients
	clients := body.Clients
	if body.Clients != "" {
		// 更新这个系统的client一定是其合法client!
		clients = defaultValidClients(c, body.Clients)
	}

	system := svctypes.System{
		Name:           body.Name,
		NameEn:         body.NameEn,
		Description:    body.Description,
		DescriptionEn:  body.DescriptionEn,
		Clients:        clients,
		ProviderConfig: providerConfig,

		AllowEmptyFields: allowEmptyFields,
	}

	svc := service.NewSystemService()
	err = svc.Update(systemID, system)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "UpdateSystem", "Update system=`%s` fail", system)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// delete the cache
	cacheimpls.DeleteSystemCache(systemID)

	util.SuccessJSONResponse(c, "ok", nil)
}

// GetSystem godoc
// @Summary system get
// @Description get a system
// @ID api-model-system-get
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Success 200 {object} util.Response{data=systemResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id} [get]
func GetSystem(c *gin.Context) {
	// validate the body
	systemID := c.Param("system_id")

	// get info via system_id
	svc := service.NewSystemService()
	system, err := svc.Get(systemID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "GetSystem", "Get systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// delete the token from provider_config
	_, ok := system.ProviderConfig["token"]
	if ok {
		delete(system.ProviderConfig, "token")
	}

	util.SuccessJSONResponse(c, "ok", systemResponse{
		ID:             systemID,
		Name:           system.Name,
		NameEn:         system.NameEn,
		Description:    system.Description,
		DescriptionEn:  system.DescriptionEn,
		Clients:        system.Clients,
		ProviderConfig: system.ProviderConfig,
	})
}

// GetSystemClients godoc
// @Summary get system clients
// @Description get clients of the system
// @ID api-model-system-clients-get
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Success 200 {object} util.Response{data=systemClientsResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/clients [get]
func GetSystemClients(c *gin.Context) {
	// validate the body
	systemID := c.Param("system_id")

	// get info via system_id
	svc := service.NewSystemService()
	system, err := svc.Get(systemID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "GetSystem", "Get systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// NOTE: v1版本的api, clients是string, 原因是为了保持和 get system返回结构体中字段一致

	// TODO: v2版, 两个接口都需要升级到 array => Clients: util.SplitStringToSet(system.Clients, ",").ToSlice(),

	util.SuccessJSONResponse(c, "ok", systemClientsResponse{
		Clients: system.Clients,
	})
}
