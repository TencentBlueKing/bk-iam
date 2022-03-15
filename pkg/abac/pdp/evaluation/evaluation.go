/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package evaluation

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/condition"
	"iam/pkg/abac/pdp/condition/operator"
	"iam/pkg/abac/pdp/evalctx"
	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
)

// NOTE: 目前所有的 query/eval都在这个文件中, 两个主要入口:
// - eval:  EvalPolicies
// - query: PartialEvalPolicies

// EvalPolicies 计算是否满足
func EvalPolicies(ctx *evalctx.EvalContext, policies []types.AuthPolicy) (isPass bool, policyID int64, err error) {
	currentTime := time.Now()

	for _, policy := range policies {
		isPass, err = evalPolicy(ctx, policy, currentTime)
		if err != nil {
			log.Debugf("pdp evalPolicy: ctx=`%+v`, policy=`%+v`, error=`%s`", ctx, policy, err)
		}

		if isPass {
			log.Debugf("pdp evalPolicy: ctx=`%+v`, policy=`%+v`, pass", ctx, policy)
			return isPass, policy.ID, err
		}
	}
	// TODO: 如果一条报错, 整体结果如何???
	return false, -1, nil
}

// evalPolicy 计算单个policy是否满足
func evalPolicy(ctx *evalctx.EvalContext, policy types.AuthPolicy, currentTime time.Time) (bool, error) {
	// action 不关联资源类型时, 直接返回true
	if ctx.Action.WithoutResourceType() {
		log.Debugf("pdp evalPolicy WithoutResourceType action: %s %s", ctx.System, ctx.Action.ID)
		return true, nil
	}

	// 需要传递resource却没有传, 此时直接false
	if !ctx.HasResources() {
		return false, fmt.Errorf("evalPolicy action: %s get not resource in request", ctx.Action.ID)
	}

	cond, err := cacheimpls.GetUnmarshalledResourceExpression(policy.Expression, policy.ExpressionSignature)
	if err != nil {
		log.Debugf("pdp evalPolicy policy id: %d expression: %s format error: %v",
			policy.ID, policy.Expression, err)

		return false, err
	}

	err = ctx.InitEnvironments(cond, currentTime)
	if err != nil {
		log.Errorf("pdp evalPolicy polidy id:%d expression: %s, currentTime: %s, error:%v",
			policy.ID, policy.Expression, currentTime, err)
		return false, err
	}

	isPass := cond.Eval(ctx)
	return isPass, err
}

// PartialEvalPolicies 筛选check pass的policies
func PartialEvalPolicies(
	ctx *evalctx.EvalContext,
	policies []types.AuthPolicy,
) ([]condition.Condition, []int64, error) {
	currentTime := time.Now()

	remainedConditions := make([]condition.Condition, 0, len(policies))

	passedPolicyIDs := make([]int64, 0, len(policies))
	for _, policy := range policies {
		isPass, condition, err := partialEvalPolicy(ctx, policy, currentTime)
		if err != nil {
			// TODO: 一条报错怎么处理?????
			log.Debugf("pdp PartialEvalPoliciesy policy: %+v ctx: %+v error: %s", policy, ctx, err)
		}

		if isPass {
			passedPolicyIDs = append(passedPolicyIDs, policy.ID)

			if condition != nil {
				remainedConditions = append(remainedConditions, condition)
			}
		}
	}

	return remainedConditions, passedPolicyIDs, nil
}

func partialEvalPolicy(
	ctx *evalctx.EvalContext,
	policy types.AuthPolicy,
	currentTime time.Time,
) (bool, condition.Condition, error) {
	// action 不关联资源类型时, 直接返回true
	if ctx.Action.WithoutResourceType() {
		log.Debugf("pdp evalPolicy WithoutResourceType action: %s %s", ctx.System, ctx.Action.ID)
		return true, condition.NewAnyCondition(), nil
	}

	cond, err := cacheimpls.GetUnmarshalledResourceExpression(policy.Expression, policy.ExpressionSignature)
	if err != nil {
		log.Debugf("pdp evalPolicy policy id: %d expression: %s format error: %v",
			policy.ID, policy.Expression, err)
		return false, nil, err
	}

	err = ctx.InitEnvironments(cond, currentTime)
	if err != nil {
		log.Errorf(
			"pdp evalPolicy polidy id:%d expression: %s, currentTime: %s, error:%v",
			policy.ID, policy.Expression, currentTime, err)
		return false, nil, err
	}

	// if no resource passed
	if !(ctx.HasResources() || ctx.HasEnv()) {
		return true, cond, nil
	}

	switch cond.GetName() {
	case operator.AND, operator.OR:
		ok, c := cond.(condition.LogicalCondition).PartialEval(ctx)
		return ok, c, nil
	case operator.ANY:
		return true, condition.NewAnyCondition(), nil
	default:
		key := cond.GetKeys()[0]
		dotIdx := strings.LastIndexByte(key, '.')
		if dotIdx == -1 {
			log.Errorf("policy condition key should contains dot! policy=`%+v`, condition=`%+v`, key=`%s`",
				policy, cond, key)
			// wrong policy expression, return ture with remained condition!!!!
			return true, cond, nil
		}
		_type := key[:dotIdx]
		if ctx.HasResource(_type) {
			if cond.Eval(ctx) {
				return true, condition.NewAnyCondition(), nil
			} else {
				return false, nil, nil
			}
		} else {
			// has not required resources, return ture with remained condition!!!!
			return true, cond, nil
		}
	}
}
