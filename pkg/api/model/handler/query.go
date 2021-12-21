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

	"iam/pkg/service"
	"iam/pkg/util"
)

// SystemQueryFieldBaseInfo ...
const (
	SystemQueryFieldBaseInfo               = "base_info"
	SystemQueryFieldResourceTypes          = "resource_types"
	SystemQueryFieldActions                = "actions"
	SystemQueryFieldInstanceSelections     = "instance_selections"
	SystemQueryFieldActionGroups           = "action_groups"
	SystemQueryFieldResourceCreatorActions = "resource_creator_actions"
	SystemQueryFieldCommonActions          = "common_actions"
	SystemQueryFieldFeatureShieldRules     = "feature_shield_rules"
)

// SystemInfoQuery godoc
// @Summary system info query
// @Description query the system
// @ID api-model-system-query
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param body body querySerializer true "the policy request"
// @Success 200 {object} util.Response{data=gin.H}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/systems/{system_id}/query [get]
//nolint:gocognit
func SystemInfoQuery(c *gin.Context) {
	var query querySerializer
	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	systemID := c.Param("system_id")
	fields := query.Fields
	if fields == "" {
		fields = "base_info,resource_types,actions,action_groups,instance_selections,resource_creator_actions," +
			"common_actions,feature_shield_rules"
	}
	fieldSet := set.SplitStringToSet(fields, ",")

	BuildSystemInfoQueryResponse(c, systemID, fieldSet)
}

//nolint:gocognit
// BuildSystemInfoQueryResponse will only the data requested
func BuildSystemInfoQueryResponse(c *gin.Context, systemID string, fieldSet *set.StringSet) {
	// make the return data
	data := gin.H{}

	if fieldSet.Has(SystemQueryFieldBaseInfo) {
		systemSvc := service.NewSystemService()
		systemInfo, err := systemSvc.Get(systemID)
		if err != nil {
			err = errorx.Wrapf(err, "Handler", "SystemInfoQuery",
				"systemSvc.Get system_id=`%s` fail", systemID)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		// delete the token from provider_config
		_, ok := systemInfo.ProviderConfig["token"]
		if ok {
			delete(systemInfo.ProviderConfig, "token")
		}

		data[SystemQueryFieldBaseInfo] = systemInfo
	}

	if fieldSet.Has(SystemQueryFieldResourceTypes) {
		rtSvc := service.NewResourceTypeService()
		resourceTypes, err := rtSvc.ListBySystem(systemID)
		if err != nil {
			err = errorx.Wrapf(err, "Handler", "SystemInfoQuery",
				"rtSvc.ListBySystem system_id=`%s` fail", systemID)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		data[SystemQueryFieldResourceTypes] = resourceTypes
	}

	// field: action => actions
	if fieldSet.Has(SystemQueryFieldActions) {
		acSvc := service.NewActionService()
		actions, err := acSvc.ListBySystem(systemID)
		if err != nil {
			err = errorx.Wrapf(err, "Handler", "SystemInfoQuery",
				"acSvc.ListBySystem system_id=`%s` fail", systemID)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		data[SystemQueryFieldActions] = actions
	}

	if fieldSet.Has(SystemQueryFieldInstanceSelections) {
		isSvc := service.NewInstanceSelectionService()
		instanceSelections, err := isSvc.ListBySystem(systemID)
		if err != nil {
			err = errorx.Wrapf(err, "Handler", "SystemInfoQuery",
				"isSvc.ListBySystem system_id=`%s` fail", systemID)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		data[SystemQueryFieldInstanceSelections] = instanceSelections
	}

	if fieldSet.Has(SystemQueryFieldActionGroups) ||
		fieldSet.Has(SystemQueryFieldResourceCreatorActions) ||
		fieldSet.Has(SystemQueryFieldCommonActions) ||
		fieldSet.Has(SystemQueryFieldFeatureShieldRules) {
		svc := service.NewSystemConfigService()

		if fieldSet.Has(SystemQueryFieldActionGroups) {
			ag, err := svc.GetActionGroups(systemID)
			if err != nil {
				data[SystemQueryFieldActionGroups] = map[string]interface{}{}
			}
			data[SystemQueryFieldActionGroups] = ag
		}

		if fieldSet.Has(SystemQueryFieldResourceCreatorActions) {
			rca, err := svc.GetResourceCreatorActions(systemID)
			if err != nil {
				data[SystemQueryFieldResourceCreatorActions] = map[string]interface{}{}
			}
			data[SystemQueryFieldResourceCreatorActions] = rca
		}

		if fieldSet.Has(SystemQueryFieldCommonActions) {
			ac, err := svc.GetCommonActions(systemID)
			if err != nil {
				data[SystemQueryFieldCommonActions] = map[string]interface{}{}
			}
			data[SystemQueryFieldCommonActions] = ac
		}
		if fieldSet.Has(SystemQueryFieldFeatureShieldRules) {
			fsrs, err := svc.GetFeatureShieldRules(systemID)
			if err != nil {
				data[SystemQueryFieldFeatureShieldRules] = map[string]interface{}{}
			}
			data[SystemQueryFieldFeatureShieldRules] = fsrs
		}
	}

	util.SuccessJSONResponse(c, "ok", data)
}
