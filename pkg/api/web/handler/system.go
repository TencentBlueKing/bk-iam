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
	"github.com/gin-gonic/gin"

	"iam/pkg/errorx"
	"iam/pkg/service"
	"iam/pkg/util"
)

// ListSystem 获取system列表
func ListSystem(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListSystem")

	var query systemQuerySerializer

	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if valid, message := query.validate(); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	svc := service.NewSystemService()
	allSystems, err := svc.ListAll()
	if err != nil {
		err = errorWrapf(err, "fields=`%s`", query.Fields)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	set := set.SplitStringToSet(query.Fields, ",")
	systems := make([]map[string]interface{}, 0, len(allSystems))
	for _, system := range allSystems {
		systemInfo, err := filterFields(set, system)
		if err != nil {
			err = errorWrapf(err, "filterFields set=`%+v`,system=`%s` fail", set, system)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		systems = append(systems, systemInfo)
	}

	util.SuccessJSONResponse(c, "ok", systems)
}

// GetSystem 获取系统基础信息
func GetSystem(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "GetSystem")

	var query systemQuerySerializer

	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if valid, message := query.validate(); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	systemID := c.Param("system_id")

	svc := service.NewSystemService()
	systemInfo, err := svc.Get(systemID)
	if err != nil {
		err = errorWrapf(err, "systemID=`%s`", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	set := set.SplitStringToSet(query.Fields, ",")
	data, err := filterFields(set, systemInfo)
	if err != nil {
		err = errorWrapf(err, "filterFields systemID=`%s`, set=`%+v` fail", systemID, set)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", data)
}
