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

type subjectTemplateSerializer struct {
	SubjectType string `json:"subject_type" binding:"required"`
	SubjectID   string `json:"subject_id" binding:"required"`
	SystemID    string `json:"system_id" binding:"required"`
	TemplateID  int64  `json:"template_id" binding:"required,min=1"`
}

type createAndDeleteTemplatePolicySerializer struct {
	Subject         subject  `json:"subject" binding:"required"`
	SystemID        string   `json:"system_id" binding:"required"`
	TemplateID      int64    `json:"template_id" binding:"required,min=1"`
	CreatePolicies  []policy `json:"create_policies" binding:"required"`
	DeletePolicyIDs []int64  `json:"delete_policy_ids" binding:"required"`
}

func (slz *createAndDeleteTemplatePolicySerializer) validate() (bool, string) {
	if len(slz.CreatePolicies) > 0 {
		if valid, message := common.ValidateArray(slz.CreatePolicies); !valid {
			return false, message
		}
	}
	return true, ""
}

type updateTemplatePolicySerializer struct {
	Subject        subject        `json:"subject" binding:"required"`
	SystemID       string         `json:"system_id" binding:"required"`
	TemplateID     int64          `json:"template_id" binding:"required,min=1"`
	UpdatePolicies []updatePolicy `json:"update_policies" binding:"required"`
}

func (slz *updateTemplatePolicySerializer) validate() (bool, string) {
	if valid, message := common.ValidateArray(slz.UpdatePolicies); !valid {
		return false, message
	}

	return true, ""
}
