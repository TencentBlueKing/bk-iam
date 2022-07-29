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
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/pdp/condition"
	"iam/pkg/abac/pdp/evalctx"
	"iam/pkg/abac/pdp/evaluation"
	"iam/pkg/abac/pip"
	"iam/pkg/abac/prp"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/logging/debug"
	svctypes "iam/pkg/service/types"
)

// PDPHelper ...
const PDPHelper = "PDPHelper"

// ErrNoPolicies ...
var (
	ErrNoPolicies       = errors.New("no policies")
	ErrInvalidAction    = errors.New("action.id invalid")
	ErrSubjectNotExists = errors.New("subject not exists")
)

func queryPolicies(
	system string,
	subject types.Subject,
	action types.Action,
	effectGroupPKs []int64,
	withoutCache bool,
	entry *debug.Entry,
) (policies []types.AuthPolicy, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDPHelper, "queryPolicies")

	manager := prp.NewPolicyManager()

	policies, err = manager.ListBySubjectAction(system, subject, action, effectGroupPKs, withoutCache, entry)
	if err != nil {
		err = errorWrapf(err,
			"ListBySubjectAction system=`%s`, subject=`%s`, action=`%s`, withoutCache=`%t` fail",
			system, subject, action, withoutCache)
		return
	}

	// 如果没有策略, 直接返回 false
	if len(policies) == 0 {
		err = ErrNoPolicies
		return
	}

	return
}

// queryAndPartialEvalConditions 查询请求相关的Policy
func queryAndPartialEvalConditions(
	r *request.Request,
	entry *debug.Entry,
	willCheckRemoteResource, // 是否检查请求的外部依赖资源完成性
	withoutCache bool,
) ([]condition.Condition, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDP, "queryAndPartialEvalConditions")

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

	// 1. PIP查询action的scop
	debug.AddStep(entry, "Fetch action details")
	err := fillActionDetail(r)
	if err != nil {
		err = errorWrapf(err, "Fetch action detail action=`%+v` fail", r.Action)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidAction
		}

		return nil, err
	}
	debug.WithValue(entry, "action", r.Action)

	if willCheckRemoteResource {
		// 2. 检查请求资源与action关联的外部依赖资源类型是否匹配
		debug.AddStep(entry, "Validate action remote resource")
		if !r.ValidateActionRemoteResource() {
			err = errorWrapf(ErrInvalidActionResource,
				"ValidateActionResource systemID=`%s`, actionID=`%d`, resources=`%+v` fail, "+
					"request resources not match action",
				r.System, r.Action.ID, r.Resources)

			return nil, err
		}
	}

	// 3. PIP查询subject相关的属性
	debug.AddStep(entry, "Fetch subject details")
	err = fillSubjectDepartments(r)
	if err != nil {
		// 如果用户不存在, 表现为没有权限
		// if the subject not exists
		if errors.Is(err, sql.ErrNoRows) {
			// return []types.AuthPolicy{}, nil
			return []condition.Condition{}, nil
		}

		err = errorWrapf(err, "request fillSubjectDetail subject=`%+v`", r.Subject)
		return nil, err
	}
	debug.WithValue(entry, "subject", r.Subject)

	// 3. 查询关联的group pks
	debug.AddStep(entry, "Get Effect AuthType Group PKs")
	abacGroupPKs, rbacGroupPKs, err := getEffectAuthTypeGroupPKs(r.System, r.Subject, r.Action)
	if err != nil {
		err = errorWrapf(
			err,
			"GetEffectAuthTypeGroupPKs systemID=`%s`, subject=`%+v`, action=`%+v` fail",
			r.System,
			r.Subject,
			r.Action,
		)
		return nil, err
	}
	debug.WithValue(entry, "abacGroupPks", abacGroupPKs)
	debug.WithValue(entry, "rbacGroupPks", rbacGroupPKs)

	// 4. PRP查询subject-action相关的policies
	debug.AddStep(entry, "Query Policies")
	policies, err := queryPolicies(r.System, r.Subject, r.Action, abacGroupPKs, withoutCache, entry)
	if err != nil {
		if errors.Is(err, ErrNoPolicies) {
			return nil, nil
		}

		err = errorWrapf(err, "queryPolicies system=`%s`, subject=`%+v`, action=`%+v`, withoutCache=`%t` fail",
			r.System, r.Subject, r.Action, withoutCache)

		return nil, err
	}
	debug.WithValue(entry, "policies", policies)
	debug.WithUnknownEvalPolicies(entry, policies)

	// 5. eval policies
	// 这里需要返回剩下的policies
	debug.AddStep(entry, "Filter policies by eval resources")

	// NOTE: 重要, 这个需要处理, 以降低影响?
	// if contains remote Resource
	if r.HasRemoteResources() {
		err1 := fillRemoteResourceAttrs(r, policies)
		if err1 != nil {
			return nil, errorWrapf(err1, "fillRemoteResourceAttrs fail", "")
		}
	}

	// 执行完后, 只返回 执行后的残留的 conditions
	if entry != nil {
		envs, _ := evalctx.GenTimeEnvsFromCache(DefaultTz, time.Now())
		debug.WithValue(entry, "env", envs)
	}
	conditions, passedPoliciesIDs, err := evaluation.PartialEvalPolicies(evalctx.NewEvalContext(r), policies)
	if len(conditions) == 0 {
		debug.WithNoPassEvalPolicies(entry, policies)
	}
	debug.WithPassEvalPolicyIDs(entry, passedPoliciesIDs)

	return conditions, err
}

// fillSubjectDepartments ...
func fillSubjectDepartments(r *request.Request) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Request", "fillSubjectDetail")

	_type := r.Subject.Type
	id := r.Subject.ID

	pk, err := pip.GetSubjectPK(_type, id)
	if err != nil {
		err = errorWrapf(err, "GetSubjectPK _type=`%s`, id=`%s` fail", _type, id)
		return err
	}

	departments, err := pip.GetSubjectDepartmentPKs(pk)
	if err != nil {
		err = errorWrapf(err, "GetSubjectDetail pk=`%d` fail", pk)
		return err
	}

	r.Subject.FillAttributes(pk, departments)
	return nil
}

// fillActionDetail ...
func fillActionDetail(r *request.Request) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("Request", "fillActionDetail")
	system := r.System
	id := r.Action.ID

	// TODO: to local cache? but, how to notify the changes in /api/query? do nothing currently!
	pk, authType, actionResourceTypes, err := pip.GetActionDetail(system, id)
	if err != nil {
		err = errorWrapf(err, "GetActionDetail system=`%s`, id=`%s` fail", system, id)
		return err
	}

	r.Action.FillAttributes(pk, authType, actionResourceTypes)
	return nil
}

// getEffectAuthTypeGroupPKs 获取authType分组的group pks
func getEffectAuthTypeGroupPKs(
	systemID string,
	subject types.Subject,
	action types.Action,
) (abacGroupPKs []int64, rbacGroupPKs []int64, err error) {
	groupPKs, err := prp.GetEffectGroupPKs(systemID, subject)
	if err != nil {
		return nil, nil, err
	}

	actionAuthType, err := action.Attribute.GetAuthType()
	if err != nil {
		return nil, nil, err
	}

	// NOTE: 当前只有action的auth type是RBAC时才需要分离出 abac/rbac group pks
	if actionAuthType == svctypes.AuthTypeRBAC {
		abacGroupPKs, rbacGroupPKs, err = prp.SplitGroupPKsByAuthType(systemID, groupPKs)
		if err != nil {
			return nil, nil, err
		}

		return abacGroupPKs, rbacGroupPKs, nil
	}

	return groupPKs, nil, nil
}
