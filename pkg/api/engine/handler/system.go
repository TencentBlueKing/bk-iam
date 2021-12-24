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
	"iam/pkg/errorx"
	"iam/pkg/service"
	"iam/pkg/util"

	"github.com/gin-gonic/gin"
)

// GetSystem godoc
// @Summary system info
// @Description get system info
// @ID api-engine-system-info
// @Tags engine
// @Accept json
// @Produce json
// @Success 200 {object} util.Response{data=types.System}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/engine/systems/:system_id [get]
func GetSystem(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "GetSystem")

	systemID := c.Param("system_id")

	svc := service.NewSystemService()
	systemInfo, err := svc.Get(systemID)
	if err != nil {
		err = errorWrapf(err, "systemID=`%s`", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", systemInfo)
}
