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
	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service"
	"iam/pkg/util"
)

// ListModelChangeEvent 查询变更事件列表
func ListModelChangeEvent(c *gin.Context) {
	status := c.Query("status")
	limit, err := conv.ToInt64(c.Query("limit"))
	if err != nil {
		limit = 1000
	}

	svc := service.NewModelChangeService()
	events, err := svc.ListByStatus(status, limit)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "ListModelChangeEvent", "status=`%s`", status)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", events)
}

// UpdateModelChangeEvent 更新变更事件列表
func UpdateModelChangeEvent(c *gin.Context) {
	var body updateModelChangeEventStatusSerializer
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	eventPK, err := conv.ToInt64(c.Param("event_pk"))
	if err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	svc := service.NewModelChangeService()
	err = svc.UpdateStatusByPK(eventPK, body.Status)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "UpdateModelChangeEvent", "eventPK=`%d` status=`%s`",
			eventPK, body.Status)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	util.SuccessJSONResponse(c, "ok", nil)
}
