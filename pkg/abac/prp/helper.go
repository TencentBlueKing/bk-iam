/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

import (
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/pdp/translate"
	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	svctypes "iam/pkg/service/types"
)

/*
NOTE:
 - 当前部门不会直接配置权限, 只能通过加入用户组的方式配置; 所以 dept PKs 不加入最终生效的pks

TODO:
 - 当前  cacheimpls.ListSubjectEffectGroups pipeline获取的性能有问题, 需要考虑走cache?

*/

func GetEffectGroupPKs(systemID string, subject types.Subject) ([]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PRP, "getEffectSubjectPKs")

	subjectPK, err := subject.Attribute.GetPK()
	if err != nil {
		err = errorWrapf(err, "subject.Attribute.GetPK subject=`%+v` fail", subject)
		return nil, err
	}

	// 通过subject对象获取dept pks
	deptPKs, err := subject.GetDepartmentPKs()
	if err != nil {
		err = errorWrapf(err, "subject.GetDepartmentPKs subject=`%+v` fail", subject)
		return nil, err
	}

	subjectPKs := make([]int64, 0, len(deptPKs)+1)
	subjectPKs = append(subjectPKs, subjectPK)
	subjectPKs = append(subjectPKs, deptPKs...)

	// 用户继承组织加入的用户组 => 多个部门属于同一个组, 所以需要去重
	now := time.Now().Unix()
	groupPKSet := set.NewInt64Set()
	subjectGroups, err := cacheimpls.ListSystemSubjectEffectGroups(systemID, subjectPKs)
	if err != nil {
		err = errorWrapf(err, "ListSubjectEffectGroups deptPKs=`%+v` fail", deptPKs)
		return nil, err
	}
	for _, sg := range subjectGroups {
		if sg.ExpiredAt > now {
			groupPKSet.Add(sg.GroupPK)
		}
	}

	return groupPKSet.ToSlice(), nil
}

// translateExpressions translate expression to json format
func translateExpressions(
	expressionPKs []int64,
) (expressionMap map[int64]map[string]interface{}, err error) {
	// when the pk is -1, will translate to any
	pkExpressionStrMap := map[int64]string{
		// NOTE: -1 for the `any`
		-1: "",
	}
	if len(expressionPKs) > 0 {
		manager := NewPolicyManager()

		var exprs []svctypes.AuthExpression
		exprs, err = manager.GetExpressionsFromCache(-1, expressionPKs)
		if err != nil {
			err = fmt.Errorf("policyManager.GetExpressionsFromCache pks=`%+v` fail. err=%w", expressionPKs, err)
			return
		}

		for _, e := range exprs {
			pkExpressionStrMap[e.PK] = e.Expression
		}
	}

	// translate one by one
	expressionMap = make(map[int64]map[string]interface{}, len(pkExpressionStrMap))
	for pk, expr := range pkExpressionStrMap {
		// TODO: 如何优化这里的性能?
		// TODO: 理论上, signature一样的只需要转一次
		// e.Signature
		translatedExpr, err1 := translate.PolicyExpressionTranslate(expr)
		if err1 != nil {
			err = fmt.Errorf("translate.PolicyExpressionTranslate expr=`%s` fail. err=%w", expr, err1)
			return
		}
		expressionMap[pk] = translatedExpr
	}
	return expressionMap, nil
}
