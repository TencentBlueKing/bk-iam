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

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/gin-gonic/gin"

	"iam/pkg/errorx"
	"iam/pkg/service"
	"iam/pkg/util"
)

// AllowConfigNames ...
const (
	AllowConfigNames = "action_groups,resource_creator_actions,common_actions,feature_shield_rules"

	ConfigNameActionGroups           = "action_groups"
	ConfigNameResourceCreatorActions = "resource_creator_actions"
	ConfigCommonActions              = "common_actions"
	ConfigNameFeatureShieldRules     = "feature_shield_rules"
)

// CreateOrUpdateConfigDispatch godoc
// @Summary system config create
// @Description create system config
// @ID api-model-system-config-create
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param name path string true "Config Name" Enums("action_groups")
// @Param body body []actionGroupSerializer true "the policy request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/configs/{name} [POST]
func CreateOrUpdateConfigDispatch(c *gin.Context) {
	systemID := c.Param("system_id")

	name := c.Param("name")
	set := set.SplitStringToSet(AllowConfigNames, ",")
	if !set.Has(name) {
		util.BadRequestErrorJSONResponse(c, fmt.Sprintf("config `%s` is not supported yet", name))
		return
	}

	// do dispatch here
	switch name {
	case ConfigNameActionGroups:
		actionGroupHandler(systemID, c)
		return
	case ConfigNameResourceCreatorActions:
		resourceCreatorActionHandler(systemID, c)
		return
	case ConfigCommonActions:
		commonActionHandler(systemID, c)
		return
	case ConfigNameFeatureShieldRules:
		featureShieldRuleHandler(systemID, c)
		return
	default:
		util.SystemErrorJSONResponse(c, errors.New("should not be here"))
		return
	}
}

func actionGroupHandler(systemID string, c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "configActionGroupHandler")
	var body []actionGroupSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if len(body) == 0 {
		util.BadRequestErrorJSONResponse(c, "the array should contain at least 1 item")
		return
	}
	if valid, message := validateActionGroup(body, ""); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	// 一个操作只能属于一个组(挂载到某个位置), 不允许挂载到多个位置, 即 全局uniq
	actionIDs := getAllFromActionGroupsActionIDs(body)
	uniqActionIDs := set.NewStringSetWithValues(actionIDs)

	if len(actionIDs) > uniqActionIDs.Size() {
		util.BadRequestErrorJSONResponse(c, "one action can only belong to one group")
		return
	}

	// 所有action id合法
	if err := checkActionIDsExist(systemID, actionIDs); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// do create
	ags := make([]interface{}, 0, len(body))
	for _, ag := range body {
		ags = append(ags, ag)
	}
	svc := service.NewSystemConfigService()
	err := svc.CreateOrUpdateActionGroups(systemID, ags)
	if err != nil {
		err = errorWrapf(err, "svc.CreateOrUpdateActionGroups systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

func resourceCreatorActionHandler(systemID string, c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "configResourceCreatorActionHandler")
	var body resourceCreatorActionSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if err := body.validate(); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// 资源类型对应的操作是否允许被授权，是否符合注册的权限模型
	if err := checkResourceCreatorActionsRelateResourceType(systemID, body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// 设置默认值：请求数据，有部分参数无的话需要设置默认
	body.setDefaultValue()

	// do create
	svc := service.NewSystemConfigService()
	err := svc.CreateOrUpdateResourceCreatorActions(systemID, body.toMapInterface())
	if err != nil {
		err = errorWrapf(err, "svc.CreateOrUpdateResourceCreatorActions systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

func commonActionHandler(systemID string, c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "commonActionHandler")
	var body []commonActionSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}
	if len(body) == 0 {
		util.BadRequestErrorJSONResponse(c, "the array should contain at least 1 item")
		return
	}

	// 一个操作只能属于一个组(挂载到某个位置), 不允许挂载到多个位置, 即 全局uniq
	actionIDs := getAllFromCommonActions(body)

	// 所有action id合法
	if err := checkActionIDsExist(systemID, actionIDs); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// do create
	cas := make([]interface{}, 0, len(body))
	for _, ag := range body {
		cas = append(cas, ag)
	}
	svc := service.NewSystemConfigService()
	err := svc.CreateOrUpdateCommonActions(systemID, cas)
	if err != nil {
		err = errorWrapf(err, "svc.CreateOrUpdateCommonActions systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}

func featureShieldRuleHandler(systemID string, c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "featureShieldRuleHandler")
	var body []featureShieldRuleSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if valid, message := validateFeatureShieldRules(body); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	// do create
	fsrs := make([]interface{}, 0, len(body))
	for _, fsr := range body {
		fsrs = append(fsrs, fsr)
	}
	svc := service.NewSystemConfigService()
	err := svc.CreateOrUpdateFeatureShieldRules(systemID, fsrs)
	if err != nil {
		err = errorWrapf(err, "svc.CreateOrUpdateFeatureShieldRules systemID=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}
