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

	"iam/pkg/abac/pdp"
	"iam/pkg/abac/types/request"
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
// @Router /api/v2/policy/systems/{system_id}/auth/ [post]
func AuthV2(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "AuthV2")
	entry, _, isForce := getDebugData(c)
	defer debug.EntryPool.Put(entry)

	systemID := c.Param("system_id")

	var body authV2Request
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check blacklist
	if shouldReturnIfSubjectInBlackList(c, body.Subject.Type, body.Subject.ID) {
		return
	}

	// check super permission
	if shouldReturnIfSubjectHasSystemSuperPermission(
		c,
		systemID,
		body.Subject.Type,
		body.Subject.ID,
		func() interface{} {
			return authV2Response{
				Allowed: true,
			}
		},
	) {
		return
	}

	// 隔离结构体
	req := request.NewRequest()
	copyRequestFromAuthV2Body(req, systemID, &body)

	// 鉴权
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

// BatchAuthV2ByActions godoc
// @Summary batch auth by actions/批量鉴权接口
// @Description batch auth by actions
// @ID api-v2-policy-batch-auth-by-actions
// @Tags policy
// @Accept json
// @Produce json
// @Param body body authV2ByActionsRequest true "the batch auth by actions request"
// @Success 200 {object} authByActionsResponse
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v2/policy/systems/{system_id}/auth_by_actions/ [post]
func BatchAuthV2ByActions(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchAuthV2ByActions")
	entry, isDebug, isForce := getDebugData(c)
	defer debug.EntryPool.Put(entry)

	var body authV2ByActionsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check system
	systemID := c.Param("system_id")
	clientID := util.GetClientID(c)
	if err := ValidateSystemMatchClient(systemID, clientID); err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	// check blacklist
	if shouldReturnIfSubjectInBlackList(c, body.Subject.Type, body.Subject.ID) {
		return
	}
	// check super permission
	if shouldReturnIfSubjectHasSystemSuperPermission(
		c,
		systemID,
		body.Subject.Type,
		body.Subject.ID,
		func() interface{} {
			data := make(authByActionsResponse, len(body.Actions))
			for _, action := range body.Actions {
				data[action.ID] = true
			}
			return data
		},
	) {
		return
	}

	// 查询  subject-system-action的policies, 然后执行鉴权!
	result := make(authByActionsResponse, len(body.Actions))
	for _, action := range body.Actions {
		req := request.NewRequest()
		copyRequestFromAuthV2ByActionsBody(req, systemID, &body)
		req.Action.ID = action.ID

		var subEntry *debug.Entry
		if isDebug {
			// NOTE: no need to call EntryPool.Put here, the global entry will do the put
			subEntry = debug.EntryPool.Get()
		}

		allowed, err := pdp.Eval(req, subEntry, isForce)
		debug.WithError(subEntry, err)
		if err != nil {
			if errors.Is(err, pdp.ErrInvalidAction) {
				util.BadRequestErrorJSONResponse(c, err.Error())
				return
			}

			err = errorWrapf(err, "systemID=`%s`, body=`%+v`", systemID, body)
			util.SystemErrorJSONResponseWithDebug(c, err, entry)
			return
		}

		result[action.ID] = allowed

		debug.AddSubDebug(entry, subEntry)
	}

	util.SuccessJSONResponseWithDebug(c, "ok", result, entry)
}
