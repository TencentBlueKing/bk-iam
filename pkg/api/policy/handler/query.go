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
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/api/common"
	"iam/pkg/cacheimpls"
	"iam/pkg/logging/debug"
	"iam/pkg/util"
)

// 策略相关的api

// Query godoc
// @Summary policy query/策略查询
// @Description query the policy by conditions: system/subject/action and resources[optional]
// @ID api-policy-query
// @Tags policy
// @Accept json
// @Produce json
// @Param body body queryRequest true "the policy request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/policy/query [post]
func Query(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "Query")

	var body queryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check system
	systemID := body.System
	clientID := util.GetClientID(c)
	if err := ValidateSystemMatchClient(systemID, clientID); err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

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
		util.SuccessJSONResponse(c, "ok, as super_manager or system_manager", AnyExpression)
		return
	}

	// 隔离结构体
	req := request.NewRequest()
	copyRequestFromQueryBody(req, &body)

	// debug
	entry, _, isForce := common.GetDebugData(c)
	defer debug.EntryPool.Put(entry)

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

// BatchQueryByActions godoc
// @Summary batch query by actions/批量查询策略
// @Description batch query policies by actions
// @ID api-policy-batch-query-by-actions
// @Tags policy
// @Accept json
// @Produce json
// @Param body body queryByActionsRequest true "the batch query by action request"
// @Success 200 {array} actionPoliciesResponse
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/policy/query_by_actions [post]
func BatchQueryByActions(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchQueryByActions")

	var body queryByActionsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check system
	systemID := body.System
	clientID := util.GetClientID(c)
	if err := ValidateSystemMatchClient(systemID, clientID); err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	policies := make([]actionPoliciesResponse, 0, len(body.Actions))

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
		for _, action := range body.Actions {
			policies = append(policies, actionPoliciesResponse{
				Action:    actionInResponse(action),
				Condition: AnyExpression,
			})
		}
		util.SuccessJSONResponse(c, "ok, as super_manager or system_manager", policies)
		return
	}

	// enable debug
	entry, isDebug, isForce := common.GetDebugData(c)
	defer debug.EntryPool.Put(entry)

	// TODO: 这里, subject/resource都是一致的, 只是action是多个, 所以其中pdp.Query会存在重复查询/重复计算?
	for _, action := range body.Actions {
		req := request.NewRequest()
		copyRequestFromQueryByActionsBody(req, &body)
		req.Action.ID = action.ID

		var subEntry *debug.Entry
		if isDebug {
			// NOTE: no need to call EntryPool.Put here, the global entry will do the put
			subEntry = debug.EntryPool.Get()
		}

		expr, err := pdp.Query(req, subEntry, true, isForce)
		debug.WithError(subEntry, err)
		if err != nil {
			err = errorWrapf(err, "systemID=`%s`, request.Action.ID=`%s`, body=`%+v`", systemID, action.ID, body)
			util.SystemErrorJSONResponseWithDebug(c, err, subEntry)
			return
		}

		policies = append(policies, actionPoliciesResponse{
			Action:    actionInResponse(action),
			Condition: expr,
		})

		debug.AddSubDebug(entry, subEntry)
	}

	util.SuccessJSONResponseWithDebug(c, "ok", policies, entry)
}

// QueryByExtResources godoc
// @Summary policy query by ext resources/批量第三方依赖策略查询
// @Description query the policy by conditions: system/subject/action and resources[optional]
// @ID api-policy-query-by-ext-resources
// @Tags policy
// @Accept json
// @Produce json
// @Param body body queryByExtResourcesRequest true "the policy request"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/policy/query_by_ext_resources [post]
func QueryByExtResources(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "Query")

	var body queryByExtResourcesRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	if valid, message := body.Validate(); !valid {
		util.BadRequestErrorJSONResponse(c, message)
		return
	}

	// check system
	systemID := body.System
	clientID := util.GetClientID(c)
	if err := ValidateSystemMatchClient(systemID, clientID); err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

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
		extResourcesWithAttr := make([]types.ExtResourceWithAttribute, 0, len(body.ExtResources))
		for _, extResource := range body.ExtResources {
			extResourceWithAttr := types.ExtResourceWithAttribute{
				System:    extResource.System,
				Type:      extResource.Type,
				Instances: make([]types.Instance, 0, len(extResource.IDs)),
			}

			for _, id := range extResource.IDs {
				extResourceWithAttr.Instances = append(extResourceWithAttr.Instances, types.Instance{
					ID:        id,
					Attribute: map[string]interface{}{},
				})
			}
			extResourcesWithAttr = append(extResourcesWithAttr, extResourceWithAttr)
		}

		util.SuccessJSONResponse(c, "ok, as super_manager or system_manager", map[string]interface{}{
			"expression":    AnyExpression,
			"ext_resources": extResourcesWithAttr,
		})
		return
	}

	// 隔离结构体
	req := request.NewRequest()
	copyRequestFromQueryBody(req, &body.queryRequest)

	// Debug
	entry, _, isForce := common.GetDebugData(c)
	defer debug.EntryPool.Put(entry)

	// 结构体隔离转换
	extResources := make([]types.ExtResource, 0, len(body.ExtResources))
	for _, r := range body.ExtResources {
		extResources = append(extResources, types.ExtResource{
			System: r.System,
			Type:   r.Type,
			IDs:    r.IDs,
		})
	}

	expr, extResourcesWithAttr, err := pdp.QueryByExtResources(req, extResources, entry, isForce)
	debug.WithError(entry, err)
	if err != nil {
		err = errorWrapf(err, "systemID=`%s`, body=`%+v`", systemID, body)
		util.SystemErrorJSONResponseWithDebug(c, err, entry)
		return
	}

	util.SuccessJSONResponseWithDebug(c, "ok", gin.H{
		"expression":    expr,
		"ext_resources": extResourcesWithAttr,
	}, entry)
}
