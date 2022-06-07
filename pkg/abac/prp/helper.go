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
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
)

/*
NOTE:
 - 当前部门不会直接配置权限, 只能通过加入用户组的方式配置; 所以 dept PKs 不加入最终生效的pks

TODO:
 - 当前  cacheimpls.ListSubjectEffectGroups pipeline获取的性能有问题, 需要考虑走cache?

*/

func getEffectSubjectPKs(systemID string, subject types.Subject) ([]int64, error) {
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
		if sg.PolicyExpiredAt > now {
			groupPKSet.Add(sg.PK)
		}
	}

	// 2. collect all pks
	effectSubjectPKs := make([]int64, 0, 1+groupPKSet.Size())
	// 将用户自身添加进去
	effectSubjectPKs = append(effectSubjectPKs, subjectPK)
	// 用户加入的用户组 + 用户继承组织加入的用户组
	effectSubjectPKs = append(effectSubjectPKs, groupPKSet.ToSlice()...)

	return effectSubjectPKs, nil
}
