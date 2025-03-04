/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service"
)

func retrieveSubjectDepartment(key cache.Key) (interface{}, error) {
	k := key.(SubjectPKCacheKey)

	departmentSvc := service.NewDepartmentService()
	departments, err := departmentSvc.GetSubjectDepartmentPKs(k.PK)
	if err != nil {
		return nil, err
	}

	return departments, nil
}

// GetSubjectDepartmentPKs ...
func GetSubjectDepartmentPKs(pk int64) (departmentPKs []int64, err error) {
	key := SubjectPKCacheKey{
		PK: pk,
	}

	err = SubjectDepartmentCache.GetInto(key, &departmentPKs, retrieveSubjectDepartment)
	err = errorx.Wrapf(err, CacheLayer, "GetSubjectDepartmentPKs",
		"SubjectDepartmentCache.Get key=`%s` fail", key.Key())
	return
}
