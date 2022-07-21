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

// 策略相关的api

// Query godoc
// @Summary policy query/策略查询
// @Description query the policy by conditions: system/subject/action and resources[optional]
// @ID api-v2-policy-query
// @Tags policy
// @Accept json
// @Produce json
// @Param system_id path string true "System ID"
// @Param body body queryRequest true "the policy request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v2/policy/systems/{system_id}/query [post]
func QueryV2(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "Query")
	entry, _, isForce := getDebugData(c)
	defer debug.EntryPool.Put(entry)

	systemID := c.Param("system_id")

	var body queryV2Request
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check blacklist
	if checkIfSubjectInBlackList(c, body.Subject.Type, body.Subject.ID) {
		return
	}

	// check super permission
	if checkSystemSuperPermission(c, systemID, body.Subject.Type, body.Subject.ID, func() interface{} {
		return AnyExpression
	}) {
		return
	}

	// 隔离结构体
	req := request.NewRequest()
	copyRequestFromQueryV2Body(req, systemID, &body)

	// 如果传的筛选的资源实例为空, 则不判断外部依赖资源是否满足
	willCheckRemoteResource := true
	if len(req.Resources) == 0 {
		willCheckRemoteResource = false
	}

	expr, err := pdp.Query(req, entry, willCheckRemoteResource, isForce)
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

	util.SuccessJSONResponseWithDebug(c, "ok", expr, entry)
}
