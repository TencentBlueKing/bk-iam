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

	log "github.com/sirupsen/logrus"

	"iam/pkg/abac/pdp/condition"
	pdptypes "iam/pkg/abac/pdp/types"
	"iam/pkg/abac/types"
)

/*
求值逻辑, 包括:

对Policy的condition求值
*/

// EvalPolicies 计算是否满足
func EvalPolicies(ctx *pdptypes.ExprContext, policies []types.AuthPolicy) (isPass bool, policyID int64, err error) {
	for _, policy := range policies {
		isPass, err = EvalPolicy(ctx, policy)
		if err != nil {
			log.Debugf("pdp evalPolicies EvalPolicy policy: %+v ctx: %+v error: %s", policy, ctx, err)
		}

		if isPass {
			log.Debugf("pdp evalPolicies EvalPolicy policy: %+v ctx: %+v pass", policy, ctx)
			return isPass, policy.ID, err
		}
	}
	return isPass, -1, err
}

// FilterPolicies 筛选check pass的policies
func FilterPolicies(ctx *pdptypes.ExprContext, policies []types.AuthPolicy) ([]types.AuthPolicy, error) {
	passPolicies := make([]types.AuthPolicy, 0, len(policies))
	var (
		isPass bool
		err    error
	)
	for _, policy := range policies {
		isPass, err = EvalPolicy(ctx, policy)
		if err != nil {
			log.Debugf("pdp filterPolicies EvalPolicy policy: %+v ctx: %+v error: %s", policy, ctx, err)
		}

		if isPass {
			log.Debugf("pdp filterPolicies EvalPolicy policy: %+v ctx: %+v pass", policy, ctx)
			passPolicies = append(passPolicies, policy)
		}
	}
	return passPolicies, err
}

// EvalPolicy 计算单个policy是否满足
func EvalPolicy(ctx *pdptypes.ExprContext, policy types.AuthPolicy) (bool, error) {
	// action 不关联资源类型时, 直接返回true
	if ctx.Action.WithoutResourceType() {
		log.Debugf("pdp EvalPolicy WithoutResourceType action: %s %s", ctx.System, ctx.Action.ID)
		return true, nil
	}

	// 如果请求中没有相关的资源信息
	if ctx.Resource == nil {
		return false, fmt.Errorf("evalPolicy action: %s get resource nil", ctx.Action.ID)
	}

	// TODO: newExpression, 两阶段计算

	cond, err := condition.ParseResourceConditionFromExpression(ctx.Resource,
		policy.Expression,
		policy.ExpressionSignature)
	if err != nil {
		log.Debugf("pdp EvalPolicy policy id: %d expression: %s format error: %v",
			policy.ID, policy.Expression, err)
		return false, err
	}

	isPass := cond.Eval(ctx)
	return isPass, err
}
