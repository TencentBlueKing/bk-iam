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

// GetAction 获取操作资源类型信息
func GetAction(c *gin.Context) {
	systemID := c.Param("system_id")
	actionID := c.Param("action_id")

	// 获取resource type信息
	svc := service.NewActionService()
	action, err := svc.Get(systemID, actionID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "GetAction", "systemID=`%s`, actionID=`%s`",
			systemID, actionID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// 返回资源类
	util.SuccessJSONResponse(c, "ok", action)
}

// ListAction 获取action列表
func ListAction(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "ListAction")

	var query actionQuerySerializer

	if err := c.ShouldBindQuery(&query); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if valid, message := query.validate(); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	systemID := c.Param("system_id")

	// 获取action信息
	svc := service.NewActionService()
	allActions, err := svc.ListBySystem(systemID)
	if err != nil {
		err = errorWrapf(err, "systemID=`%s`", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	set := set.SplitStringToSet(query.Fields, ",")
	actions := make([]map[string]interface{}, 0, len(allActions))
	for _, action := range allActions {
		ac, err := filterFields(set, action)
		if err != nil {
			err = errorWrapf(err, "filterFields set=`%+v`, systemID=`%s`, actionID=`%s`",
				set, systemID, action.ID)
			util.SystemErrorJSONResponse(c, err)
			return
		}

		actions = append(actions, ac)
	}

	util.SuccessJSONResponse(c, "ok", actions)
}
