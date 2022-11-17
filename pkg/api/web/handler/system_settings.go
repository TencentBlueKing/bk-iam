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
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"iam/pkg/service"
	"iam/pkg/util"
)

// SystemActionGroups ...
const (
	SystemActionGroups           = "action-groups"
	SystemResourceCreatorActions = "resource-creator-actions"
	SystemCommonActions          = "common-actions"
	SystemFeatureShieldRules     = "feature-shield-rules"
	SystemMangers                = "system-managers"
	SystemCustomFrontendSettings = "custom-frontend-settings"
)

// GetSystemSettings ...
func GetSystemSettings(c *gin.Context) {
	name := c.Param("name")
	switch name {
	case SystemActionGroups:
		GetActionGroup(c)
		return
	case SystemResourceCreatorActions:
		GetResourceCreatorAction(c)
		return
	case SystemCommonActions:
		GetCommonAction(c)
		return
	case SystemFeatureShieldRules:
		GetFeatureShieldRule(c)
		return
	case SystemMangers:
		GetSystemManger(c)
		return
	case SystemCustomFrontendSettings:
		GetSystemCustomFrontendSettings(c)
		return
	default:
		util.BadRequestErrorJSONResponse(c, fmt.Sprintf("unsupported settings name %s", name))
	}
}

// GetActionGroup ...
func GetActionGroup(c *gin.Context) {
	systemID := c.Param("system_id")

	svc := service.NewSystemConfigService()
	ag, err := svc.GetActionGroups(systemID)
	if errors.Is(err, sql.ErrNoRows) {
		util.SuccessJSONResponse(c, "ok", []interface{}{})
		return
	}
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", ag)
}

// GetResourceCreatorAction ...
func GetResourceCreatorAction(c *gin.Context) {
	systemID := c.Param("system_id")

	svc := service.NewSystemConfigService()
	rca, err := svc.GetResourceCreatorActions(systemID)
	if errors.Is(err, sql.ErrNoRows) {
		util.SuccessJSONResponse(c, "ok", map[string]interface{}{})
		return
	}
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", rca)
}

// GetCommonAction ...
func GetCommonAction(c *gin.Context) {
	systemID := c.Param("system_id")

	svc := service.NewSystemConfigService()
	ca, err := svc.GetCommonActions(systemID)
	if errors.Is(err, sql.ErrNoRows) {
		util.SuccessJSONResponse(c, "ok", []interface{}{})
		return
	}
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", ca)
}

// GetFeatureShieldRule ...
func GetFeatureShieldRule(c *gin.Context) {
	systemID := c.Param("system_id")

	svc := service.NewSystemConfigService()
	fsrs, err := svc.GetFeatureShieldRules(systemID)
	if errors.Is(err, sql.ErrNoRows) {
		util.SuccessJSONResponse(c, "ok", []interface{}{})
		return
	}
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", fsrs)
}

// GetSystemManger ...
func GetSystemManger(c *gin.Context) {
	systemID := c.Param("system_id")

	svc := service.NewSystemConfigService()
	sm, err := svc.GetSystemManagers(systemID)
	if errors.Is(err, sql.ErrNoRows) {
		util.SuccessJSONResponse(c, "ok", []interface{}{})
		return
	}
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", sm)
}

// GetSystemCustomFrontendSettings ...
func GetSystemCustomFrontendSettings(c *gin.Context) {
	systemID := c.Param("system_id")

	svc := service.NewSystemConfigService()
	settings, err := svc.GetCustomFrontendSettings(systemID)
	if errors.Is(err, sql.ErrNoRows) {
		util.SuccessJSONResponse(c, "ok", []interface{}{})
		return
	}
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", settings)
}
