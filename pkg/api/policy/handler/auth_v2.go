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

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/pdp"
	"iam/pkg/abac/types/request"
	"iam/pkg/api/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/logging/debug"
	"iam/pkg/util"
)

// Auth godoc
// @Summary policy auth/鉴权
// @Description eval all the policies queried by conditions: system/subject/action and resources[required]
// @ID api-v2-policy-auth
// @Tags policy
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param body body authRequest true "the policy request"
// @Success 200 {object} authResponse
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v2/policy/systems/{system_id}/auth [post]
func AuthV2(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "AuthV2")

	systemID := c.Param("system_id")

	var body authV2Request
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check blacklist
	if cacheimpls.IsSubjectInBlackList(body.Subject.Type, body.Subject.ID) {
		util.ForbiddenJSONResponse(
			c,
			fmt.Sprintf("subject(type=%s,id=%s) has been frozen", body.Subject.Type, body.Subject.ID),
		)
		return
	}

	hasSuperPerm, err := hasSystemSuperPermission(systemID, body.Subject.Type, body.Subject.ID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	if hasSuperPerm {
		util.SuccessJSONResponse(c, "ok, as super_manager or system_manager", authResponse{
			Allowed: true,
		})
		return
	}

	// 隔离结构体
	req := request.NewRequest()
	copyRequestFromAuthV2Body(req, systemID, &body)

	// 鉴权
	entry, _, isForce := common.GetDebugData(c)
	defer debug.EntryPool.Put(entry)

	allowed, err := pdp.Eval(req, entry, isForce)
	debug.WithError(entry, err)
	if err != nil {
		if errors.Is(err, pdp.ErrInvalidAction) {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}

		err = errorWrapf(err, "systemID=`%s`, body=`%+v`", systemID, body)
		util.SystemErrorJSONResponseWithDebug(c, err, entry)
		return
	}

	data := authV2Response{
		Allowed: allowed,
	}
	util.SuccessJSONResponseWithDebug(c, "ok", data, entry)
}
