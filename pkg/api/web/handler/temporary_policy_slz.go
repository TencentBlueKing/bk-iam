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

import "iam/pkg/api/common"

// 临时权限 request body
type temporaryPoliciesSerializer struct {
	Subject  subject  `json:"subject"  binding:"required"`
	Policies []policy `json:"policies" binding:"required"`
}

func (slz *temporaryPoliciesSerializer) validate() (bool, string) {
	if len(slz.Policies) > 0 {
		if valid, message := common.ValidateArray(slz.Policies); !valid {
			return false, message
		}
	}
	return true, ""
}

type temporaryPoliciesDeleteSerializer struct {
	SubjectType string  `json:"subject_type" binding:"required"`
	SubjectID   string  `json:"subject_id"   binding:"required"`
	SystemID    string  `json:"system_id"    binding:"required"`
	IDs         []int64 `json:"ids"          binding:"required,gt=0"`
}
