/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pdp

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/pdp/evalctx"
	"iam/pkg/abac/pdp/evaluation"
	"iam/pkg/abac/pdp/translate"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/cacheimpls"
	"iam/pkg/logging/debug"
)

/*
PDP 模块鉴权入口结构与鉴权函数定义

		--------------
		\  entrance  \
		--------------

		--------------
		\ evaluation \
		--------------

--------------       --------------
\ condition  \       \ translate  \
--------------       --------------

*/

// PDP ...
const (
	PDP = "PDP"

	DefaultTz = "Asia/Shanghai"
)

// EmptyPolicies ...
var (
	EmptyPolicies = map[string]interface{}{}

	ErrInvalidActionResource = errors.New("validateActionResource fail")
)

// Eval 鉴权入口
func Eval(
	r *request.Request,
	entry *debug.Entry,
	withoutCache bool,
) (isPass bool, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDP, "Eval")

	// init debug entry with values
	if entry != nil {
		debug.WithValues(entry, map[string]interface{}{
			"system":       r.System,
			"subject":      r.Subject,
			"action":       r.Action,
			"resources":    r.Resources,
			"cacheEnabled": !withoutCache,
		})
	}

	// 1. PIP查询action
	debug.AddStep(entry, "Fetch action details")
	err = fillActionDetail(r)
	if err != nil {
		err = errorWrapf(err, "Fetch action detail action=`%+v` fail", r.Action)
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrInvalidAction
			return
		}

		return
	}
	debug.WithValue(entry, "action", r.Action)

	// 2. 检查请求资源与action关联的类型是否匹配 (填充后的校验)
	debug.AddStep(entry, "Validate action resource")
	if !r.ValidateActionResource() {
		err = errorWrapf(ErrInvalidActionResource,
			"ValidateActionResource systemID=`%s`, actionID=`%d`, resources=`%+v` fail, "+
				"request resources not match action",
			r.System, r.Action.ID, r.Resources)
		return false, err
	}

	// 3. PIP查询subject相关的属性
	debug.AddStep(entry, "Fetch subject details")
	err = fillSubjectDetail(r)
	if err != nil {
		// 如果用户不存在, 表现为没有权限
		// if the subject not exists
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		err = errorWrapf(err, "request fillSubjectDetail subject=`%+v`", r.Subject)

		return
	}
	debug.WithValue(entry, "subject", r.Subject)

	// 4. PRP查询subject-action相关的policies: 根据 system / subject / action 获取策略列表
	debug.AddStep(entry, "Query Policies")
	policies, err := queryPolicies(r.System, r.Subject, r.Action, withoutCache, entry)
	if err != nil {
		if errors.Is(err, ErrNoPolicies) {
			return false, nil
		}

		err = errorWrapf(err, "queryPolicies system=`%s`, subject=`%+v`, action=`%+v`, withoutCache=`%t` fail",
			r.System, r.Subject, r.Action, withoutCache)
		return false, err
	}
	debug.WithValue(entry, "policies", policies)
	debug.WithUnknownEvalPolicies(entry, policies)

	// NOTE: debug mode, do translate, for understanding easier
	if entry != nil {
		debug.WithValue(entry, "expression", "set fail")

		expr, err1 := cacheimpls.PoliciesTranslate(policies)
		if err1 == nil {
			debug.WithValue(entry, "expression", expr)
		}
	}

	debug.AddStep(entry, "Eval")
	if entry != nil {
		envs, _ := evalctx.GenTimeEnvsFromCache(DefaultTz, time.Now())
		debug.WithValue(entry, "env", envs)
	}
	var passPolicyID int64
	isPass, passPolicyID, err = evaluation.EvalPolicies(evalctx.NewEvalContext(r), policies)
	if err != nil {
		err = errorWrapf(err, "single local evaluation.EvalPolicies policies=`%+v`, request=`%+v` fail",
			policies, *r)

		return false, err
	}
	if !isPass {
		// if isPass is false, update all to `no pass`
		debug.WithNoPassEvalPolicies(entry, policies)
	} else {
		// if isPass is true, how to know which policy?
		debug.WithPassEvalPolicy(entry, passPolicyID)
	}

	return isPass, err
}

// Query 查询请求相关的Policy
func Query(
	r *request.Request,
	entry *debug.Entry,
	willCheckRemoteResource,
	withoutCache bool,
) (map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDP, "Query")

	// 1. 查询请求相关的策略
	conditions, err := queryAndPartialEvalConditions(r, entry, willCheckRemoteResource, withoutCache)
	if err != nil {
		err = errorWrapf(err, "queryAndPartialEvalConditions fail", r.Action)
		return nil, err
	}

	if len(conditions) == 0 {
		return EmptyPolicies, nil
	}

	// 2. policies表达式转换
	expr, err := translate.ConditionsTranslate(conditions)
	if err != nil {
		err = errorWrapf(err, "translate.ConditionsTranslate conditions=`%+v` fail", conditions)
		return nil, err
	}
	debug.WithValue(entry, "expression", expr)

	return expr, nil
}

// QueryByExtResources 使用第三方资源批量查询策略与资源的属性
func QueryByExtResources(
	r *request.Request,
	extResources []types.ExtResource,
	entry *debug.Entry,
	withoutCache bool,
) (map[string]interface{}, []types.ExtResourceWithAttribute, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDP, "QueryByExtResources")

	// 1. 查询请求相关的策略
	conditions, err := queryAndPartialEvalConditions(r, entry, false, withoutCache)
	if err != nil {
		err = errorWrapf(err, "queryAndPartialEvalConditions fail", r.Action)
		return nil, nil, err
	}

	extResourcesWithAttr := make([]types.ExtResourceWithAttribute, 0, len(extResources))
	for _, resource := range extResources {
		extResourcesWithAttr = append(extResourcesWithAttr, types.ExtResourceWithAttribute{
			System:    resource.System,
			Type:      resource.Type,
			Instances: make([]types.Instance, 0, len(resource.IDs)),
		})
	}

	// 如果策略为空, 直接返回空结果
	if len(conditions) == 0 {
		for i, resource := range extResources {
			for _, id := range resource.IDs {
				extResourcesWithAttr[i].Instances = append(extResourcesWithAttr[i].Instances, types.Instance{
					ID:        id,
					Attribute: map[string]interface{}{},
				})
			}
		}

		return EmptyPolicies, extResourcesWithAttr, nil
	}

	// 2. 批量查询 ext resource 的属性
	var remoteResources []map[string]interface{}
	for i := range extResources {
		remoteResources, err = queryExtResourceAttrs(&extResources[i], conditions)
		if err != nil {
			err = errorWrapf(err, "queryExtResourceAttrs resource=`%+v` fail", extResources[i])
			return nil, nil, err
		}

		for _, rr := range remoteResources {
			extResourcesWithAttr[i].Instances = append(extResourcesWithAttr[i].Instances, types.Instance{
				ID:        fmt.Sprint(rr["id"]),
				Attribute: rr,
			})
		}
	}

	// 3. policies表达式转换
	var expr map[string]interface{}
	expr, err = translate.ConditionsTranslate(conditions)
	if err != nil {
		err = errorWrapf(err, "translate.ConditionsTranslate conditions=`%+v` fail", conditions)
		return nil, nil, err
	}
	debug.WithValue(entry, "expression", expr)

	return expr, extResourcesWithAttr, nil
}

// QueryAuthPolicies ...
func QueryAuthPolicies(
	r *request.Request,
	entry *debug.Entry,
	withoutCache bool,
) (policies []types.AuthPolicy, err error) {
	// NOTE: the r.resources is empty here!!!!!!
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDP, "BatchResourcesEval")

	// init debug entry with values
	if entry != nil {
		debug.WithValues(entry, map[string]interface{}{
			"system":       r.System,
			"subject":      r.Subject,
			"action":       r.Action,
			"resources":    r.Resources,
			"cacheEnabled": !withoutCache,
		})
	}

	// 1. PIP查询action
	debug.AddStep(entry, "Fetch action details")
	err = fillActionDetail(r)
	if err != nil {
		err = errorWrapf(err, "Fetch action detail action=`%+v` fail", r.Action)
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrInvalidAction
			return
		}

		return
	}
	debug.WithValue(entry, "action", r.Action)

	// 3. PIP查询subject相关的属性
	debug.AddStep(entry, "Fetch subject details")
	err = fillSubjectDetail(r)
	if err != nil {
		// 如果用户不存在, 表现为没有权限
		// if the subject not exists
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrSubjectNotExists
			return
		}

		err = errorWrapf(err, "request fillSubjectDetail subject=`%+v`", r.Subject)
		return
	}
	debug.WithValue(entry, "subject", r.Subject)

	// 4. PRP查询subject-action相关的policies: 根据 system / subject / action 获取策略列表
	debug.AddStep(entry, "Query Policies")
	policies, err = queryPolicies(r.System, r.Subject, r.Action, withoutCache, entry)
	if err != nil {
		if errors.Is(err, ErrNoPolicies) {
			return
		}

		err = errorWrapf(err, "queryPolicies system=`%s`, subject=`%+v`, action=`%+v`, withoutCache=`%t` fail",
			r.System, r.Subject, r.Action, withoutCache)
		return
	}
	debug.WithValue(entry, "policies", policies)

	return policies, nil
}
