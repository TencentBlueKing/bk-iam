/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package rbac

import (
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service"
	"iam/pkg/service/types"
	"iam/pkg/task/handler"
)

const rbacDatabaseLayer = "rbacDatabaseLayer"

type RbacPolicyDatabaseRetriever struct {
	subjectActionExpressionService    service.SubjectActionExpressionService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
}

func NewRbacPolicyDatabaseRetriever() RbacPolicyRetriever {
	return &RbacPolicyDatabaseRetriever{
		subjectActionExpressionService:    service.NewSubjectActionExpressionService(),
		subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
	}
}

func (r *RbacPolicyDatabaseRetriever) ListBySubjectAction(
	subjectPKs []int64,
	actionPK int64,
) ([]types.SubjectActionExpression, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(rbacDatabaseLayer, "ListBySubjectAction")
	expressions, err := r.subjectActionExpressionService.ListBySubjectAction(subjectPKs, actionPK)
	if err != nil {
		err = errorWrapf(
			err,
			"subjectActionExpressionService.ListBySubjectAction fail, subjectPKs=`%+v`, actionPK=`%d`",
			subjectPKs,
			actionPK,
		)
		return nil, err
	}

	nowUnix := time.Now().Unix()
	validExpressions := make([]types.SubjectActionExpression, 0, len(expressions))
	for _, expression := range expressions {
		// NOTE: expriredAt为0表示所有的用户组都已过期, 忽略该无效数据
		if expression.ExpiredAt == 0 {
			continue
		}

		if expression.ExpiredAt > nowUnix {
			validExpressions = append(validExpressions, expression)
			continue
		}

		// 已过期的数据，从原始数据中转换获取
		exp, err := r.refreshSubjectActionExpression(expression.SubjectPK, actionPK)
		if err != nil {
			err = errorWrapf(
				err,
				"refreshSubjectActionExpression fail, subjectPK=`%d`, actionPK=`%d`",
				expression.SubjectPK,
				actionPK,
			)
			return nil, err
		}

		if exp.ExpiredAt != 0 {
			validExpressions = append(validExpressions, exp)
		}
	}

	return validExpressions, nil
}

func (r *RbacPolicyDatabaseRetriever) refreshSubjectActionExpression(
	subjectPK, actionPK int64,
) (expression types.SubjectActionExpression, err error) {
	// query subject action group resource from db
	obj, err := r.subjectActionGroupResourceService.Get(subjectPK, actionPK)
	if err != nil {
		return
	}

	// to subject action expression
	return handler.ConvertSubjectActionGroupResourceToExpression(obj)
}
