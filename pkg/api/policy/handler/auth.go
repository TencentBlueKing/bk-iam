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

	"github.com/gin-gonic/gin"

	"iam/pkg/abac/pdp"
	"iam/pkg/abac/pdp/evalctx"
	"iam/pkg/abac/pdp/evaluation"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/cache/impls"
	"iam/pkg/errorx"
	"iam/pkg/logging/debug"
	"iam/pkg/util"
)

// Auth godoc
// @Summary policy auth/鉴权
// @Description eval all the policies queried by conditions: system/subject/action and resources[required]
// @ID api-policy-auth
// @Tags policy
// @Accept json
// @Produce json
// @Param body body authRequest true "the policy request"
// @Success 200 {object} authResponse
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/policy/auth [post]
func Auth(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "Auth")

	var body authRequest
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

	hasSuperPerm, err := hasSystemSuperPermission(systemID, body.Subject.Type, body.Subject.ID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	if hasSuperPerm {
		util.SuccessJSONResponse(c, "ok", authResponse{
			Allowed: true,
		})
		return
	}

	// 隔离结构体
	var req = request.NewRequest()
	copyRequestFromAuthBody(req, &body)

	// 鉴权
	var entry *debug.Entry

	if _, isDebug := c.GetQuery("debug"); isDebug {
		entry = debug.EntryPool.Get()
		defer debug.EntryPool.Put(entry)
	}
	_, isForce := c.GetQuery("force")

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

	data := authResponse{
		Allowed: allowed,
	}
	util.SuccessJSONResponseWithDebug(c, "ok", data, entry)
}

// BatchAuthByActions godoc
// @Summary batch auth by actions/批量鉴权接口
// @Description batch auth by actions
// @ID api-policy-batch-auth-by-actions
// @Tags policy
// @Accept json
// @Produce json
// @Param body body authByActionsRequest true "the batch auth by actions request"
// @Success 200 {object} authByActionsResponse
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/policy/auth_by_actions [post]
func BatchAuthByActions(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchAuthByActions")

	var body authByActionsRequest
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

	result := make(authByActionsResponse, len(body.Actions))

	// super admin and system admin
	hasSuperPerm, err := hasSystemSuperPermission(systemID, body.Subject.Type, body.Subject.ID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	if hasSuperPerm {
		for _, action := range body.Actions {
			result[action.ID] = true
		}
		util.SuccessJSONResponse(c, "ok", result)
		return
	}

	// enable debug
	var entry *debug.Entry
	_, isDebug := c.GetQuery("debug")
	if isDebug {
		entry = debug.EntryPool.Get()
		defer debug.EntryPool.Put(entry)
	}

	_, isForce := c.GetQuery("force")

	// 查询  subject-system-action的policies, 然后执行鉴权!
	for _, action := range body.Actions {
		req := request.NewRequest()
		copyRequestFromAuthByActionsBody(req, &body)
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

// BatchAuthByResources godoc
// @Summary batch auth by resources/批量鉴权接口
// @Description batch auth by resources
// @ID api-policy-batch-auth-by-resources
// @Tags policy
// @Accept json
// @Produce json
// @Param body body authByResourcesRequest true "the batch auth by resources request"
// @Success 200 {object} authByResourcesResponse
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/policy/auth_by_resources [post]
func BatchAuthByResources(c *gin.Context) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Handler", "BatchAuthByResources")

	var body authByResourcesRequest
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

	data := make(authByResourcesResponse, len(body.ResourcesList))

	// super admin and system admin
	hasSuperPerm, err := hasSystemSuperPermission(systemID, body.Subject.Type, body.Subject.ID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	if hasSuperPerm {
		for _, r := range body.ResourcesList {
			data[buildResourceID(r)] = true
		}

		util.SuccessJSONResponse(c, "ok", data)
		return
	}

	// 隔离结构体
	var req = request.NewRequest()
	copyRequestFromAuthByResourcesBody(req, &body)

	// 鉴权
	var entry *debug.Entry

	if _, isDebug := c.GetQuery("debug"); isDebug {
		entry = debug.EntryPool.Get()
		defer debug.EntryPool.Put(entry)
	}
	_, isForce := c.GetQuery("force")

	// TODO: 这里下沉到下一层, 不应该直接依赖evaluation, 只应该依赖pdp
	// query policies
	policies, err := pdp.QueryAuthPolicies(req, entry, isForce)
	if err != nil {
		debug.WithError(entry, err)
		if errors.Is(err, pdp.ErrInvalidAction) {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}
		// no permission =>
		if errors.Is(err, pdp.ErrSubjectNotExists) || errors.Is(err, pdp.ErrNoPolicies) {
			// 没有权限
			for _, r := range body.ResourcesList {
				data[buildResourceID(r)] = false
			}

			util.SuccessJSONResponseWithDebug(c, "ok", data, entry)
			return
		}

		// else, system error
		err = errorWrapf(err, "systemID=`%s`, body=`%+v`", systemID, body)
		util.SystemErrorJSONResponseWithDebug(c, err, entry)
		return
	}

	// do eval for each resource
	for _, resources := range body.ResourcesList {
		// TODO: 这里下沉到下一层, 不应该直接依赖evaluation, 只应该依赖pdp
		// copy the req, reset and assign the resources
		r := req
		r.Resources = make([]types.Resource, 0, len(resources))
		for _, resource := range resources {
			r.Resources = append(r.Resources, types.Resource{
				System:    resource.System,
				Type:      resource.Type,
				ID:        resource.ID,
				Attribute: resource.Attribute,
			})
		}

		// do eval
		isAllowed, _, err := evaluation.EvalPolicies(evalctx.NewEvalContext(r), policies)
		if err != nil {
			err = errorWrapf(err, " pdp.EvalPolicies req=`%+v`, policies=`%+v` fail", r, policies)
			util.SystemErrorJSONResponseWithDebug(c, err, entry)
			return
		}

		data[buildResourceID(resources)] = isAllowed
	}

	// NOTE: debug mode, do translate, for understanding easier
	if entry != nil && len(body.ResourcesList) > 0 {
		debug.WithValue(entry, "expression", "set fail")
		expr, err1 := impls.PoliciesTranslate(policies)
		if err1 == nil {
			debug.WithValue(entry, "expression", expr)
		}
	}

	util.SuccessJSONResponseWithDebug(c, "ok", data, entry)
}
