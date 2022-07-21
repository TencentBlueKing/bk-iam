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
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/pdp"
	"iam/pkg/abac/pdp/evalctx"
	"iam/pkg/abac/pdp/evaluation"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/cacheimpls"
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
	entry, _, isForce := getDebugData(c)
	defer debug.EntryPool.Put(entry)

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
			return authResponse{
				Allowed: true,
			}
		},
	) {
		return
	}

	// 隔离结构体
	req := request.NewRequest()
	copyRequestFromAuthBody(req, &body)

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
	entry, isDebug, isForce := getDebugData(c)
	defer debug.EntryPool.Put(entry)

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
	entry, _, isForce := getDebugData(c)
	defer debug.EntryPool.Put(entry)

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
			data := make(authByResourcesResponse, len(body.ResourcesList))
			for _, r := range body.ResourcesList {
				data[buildResourceID(r)] = true
			}
			return data
		},
	) {
		return
	}

	// 隔离结构体
	req := request.NewRequest()
	copyRequestFromAuthByResourcesBody(req, &body)

	/*
		TODO 2种变更方式
		1. 走RBAC逻辑
			- 遍历resource走rbac鉴权, 如果有资源pass, 则记录结果, no pass则回落到abac鉴权
		2. 走query的逻辑, 分别查询abac与rbac的策略并合并, 遍历走abac鉴权
	*/

	// TODO: 这里下沉到下一层, 不应该直接依赖evaluation, 只应该依赖pdp
	// query policies
	data := make(authByResourcesResponse, len(body.ResourcesList))
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

	// TODO: move to pdp/entrance.go
	if entry != nil {
		envs, _ := evalctx.GenTimeEnvsFromCache(pdp.DefaultTz, time.Now())
		debug.WithValue(entry, "env", envs)
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
		expr, err1 := cacheimpls.PoliciesTranslate(policies)
		if err1 == nil {
			debug.WithValue(entry, "expression", expr)
		}
	}

	util.SuccessJSONResponseWithDebug(c, "ok", data, entry)
}
