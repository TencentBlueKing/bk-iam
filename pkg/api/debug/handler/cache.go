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
	"strconv"

	"github.com/gin-gonic/gin"

	"iam/pkg/abac/prp/expression"
	pl "iam/pkg/abac/prp/policy"
	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

type queryPolicyCacheSerializer struct {
	System      string `form:"system"       binding:"required"`
	SubjectType string `form:"subject_type" binding:"required"`
	SubjectID   string `form:"subject_id"   binding:"required"`

	Action string `form:"action"`
}

// QueryPolicyCache ...
func QueryPolicyCache(c *gin.Context) {
	var body queryPolicyCacheSerializer
	if err := c.ShouldBindQuery(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	subjectPK, err := cacheimpls.GetLocalSubjectPK(body.SubjectType, body.SubjectID)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// action不传, 则得到快照列表  => actionPKs => 需要转成 actions
	if body.Action == "" {
		// hash key=iam:policies:{system}:{subject_pk}
		// will return all the fields
		errs := []error{}

		hashKey := body.System + ":" + strconv.FormatInt(subjectPK, 10)
		keys, err0 := cacheimpls.PolicyCache.HKeys(hashKey)
		errs = append(errs, err0)

		actions := make([]types.ThinAction, 0, len(keys))
		if err0 == nil && len(keys) > 0 {
			svc := service.NewActionService()

			for _, key := range keys {
				actionPK, err1 := strconv.ParseInt(key, 10, 64)
				if err1 != nil {
					errs = append(errs, err1)

					continue
				}

				action, err2 := svc.GetThinActionByPK(actionPK)
				if err2 != nil {
					errs = append(errs, err2)

					continue
				}
				actions = append(actions, action)
			}
		}

		util.SuccessJSONResponse(c, "ok", gin.H{
			"subject_pk": subjectPK,
			"keys":       keys,
			"actions":    actions,
			"errs":       errs,
		})
		return
	}

	actionPK, err := cacheimpls.GetActionPK(body.System, body.Action)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	// action传, 则得到policy详情 => 注意expressionPK需要解析, 获取得到详情?(这是debug接口, 不考虑过大)
	errs := []error{}
	policies := []gin.H{}
	ps, noCacheSubjectPKs, err := pl.DebugRawGetPolicyFromCache(body.System, actionPK, []int64{subjectPK})
	errs = append(errs, err)

	expressionPKs := make([]int64, 0, len(ps))
	for _, p := range ps {
		policies = append(policies, gin.H{
			"PK":           p.PK,
			"SubjectPK":    p.SubjectPK,
			"ExpressionPK": p.ExpressionPK,
			"ExpiredAt":    p.ExpiredAt,
		})
		expressionPKs = append(expressionPKs, p.ExpressionPK)
	}
	svc := service.NewPolicyService()
	expressions, err1 := svc.ListExpressionByPKs(expressionPKs)
	errs = append(errs, err1)

	util.SuccessJSONResponse(c, "ok", gin.H{
		"subject_pk": subjectPK,
		"action_pk":  actionPK,
		// if in cache but exipred, the notInCache will be false
		"notInCache": len(noCacheSubjectPKs) == 1,

		"policies":    policies,
		"expressions": expressions,
		"errs":        errs,
	})
}

// QueryExpressionCache ...
func QueryExpressionCache(c *gin.Context) {
	pks, ok := c.GetQuery("pks")
	if !ok {
		util.BadRequestErrorJSONResponse(c, "pks required")
		return
	}
	expressionPKs, err := util.StringToInt64Slice(pks, ",")
	if err != nil {
		util.BadRequestErrorJSONResponse(c, err.Error())
		return
	}

	// expressionPK
	exprs, noCachePKs, err := expression.DebugRawGetExpressionFromCache(expressionPKs)

	util.SuccessJSONResponse(c, "ok", gin.H{
		"pks":         expressionPKs,
		"expressions": exprs,
		"noCachePKs":  noCachePKs,
		"err":         err,
	})
}
