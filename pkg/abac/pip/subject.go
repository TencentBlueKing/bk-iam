/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pip

import (
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/cacheimpls"
)

// SubjectPIP ...
const SubjectPIP = "SubjectPIP"

// GetSubjectPK 获取subject的PK, note this will cache in local for 1 minutes
func GetSubjectPK(_type, id string) (int64, error) {
	// pk, err := cacheimpls.GetSubjectPK(_type, id)
	pk, err := cacheimpls.GetLocalSubjectPK(_type, id)
	if err != nil {
		return pk, errorx.Wrapf(err, SubjectPIP, "GetSubjectPK",
			"cacheimpls.GetLocalSubjectPK _type=`%s`, id=`%s` fail", _type, id)
	}

	return pk, err
}

// GetSubjectDepartmentPKs ...
func GetSubjectDepartmentPKs(pk int64) (departments []int64, err error) {
	departments, err = cacheimpls.GetSubjectDepartmentPKs(pk)
	if err != nil {
		err = errorx.Wrapf(err, SubjectPIP, "GetSubjectDepartmentPKs",
			"cacheimpls.GetSubjectDepartmentPKs pk=`%d` fail", pk)
		return
	}

	return departments, err
}
