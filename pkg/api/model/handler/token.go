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

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"

	"iam/pkg/cacheimpls"
	"iam/pkg/util"
)

// GetToken godoc
// @Summary token get
// @Description get the token of system
// @ID api-model-system-token-get
// @Tags model
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Success 200 {object} util.Response{data=tokenResponse}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/model/systems/{system_id}/token [get]
func GetToken(c *gin.Context) {
	// validate the body
	systemID := c.Param("system_id")

	// get info via system_id
	system, err := cacheimpls.GetSystem(systemID)
	if err != nil {
		err = errorx.Wrapf(err, "Handler", "GetToken",
			"svc.Get system_id=`%s` fail", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}
	_, ok := system.ProviderConfig["token"]
	if !ok {
		err = errors.New("the token not in the system.ProviderConfig")
		err = errorx.Wrapf(err, "Handler", "GetToken", "system=`%s`", systemID)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", tokenResponse{
		Token: system.ProviderConfig["token"].(string),
	})
}

type tokenResponse struct {
	Token string `json:"token" example:"vgclj4ddfe3ydr6eg3wjfx4vc8ogjxhi"`
}
