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
	"iam/pkg/api/common"
)

// Query for
type policySerializer struct {
	SubjectType string `form:"subject_type" json:"subject_type" binding:"required"`
	SubjectID   string `form:"subject_id" json:"subject_id" binding:"required"`
	TemplateID  int64  `form:"template_id" json:"template_id" binding:"omitempty"`
}

// 变更 request body
type policiesAlterSerializer struct {
	Subject         subject        `json:"subject" binding:"required"`
	CreatePolicies  []policy       `json:"create_policies" binding:"required"`
	UpdatePolicies  []updatePolicy `json:"update_policies" binding:"required"`
	DeletePolicyIDs []int64        `json:"delete_policy_ids" binding:"required"`
}

type subject struct {
	Type string `json:"type" binding:"required"`
	ID   string `json:"id" binding:"required"`
}

type policy struct {
	ActionID           string `json:"action_id" binding:"required"`
	ResourceExpression string `json:"resource_expression" binding:"required"`
	ExpiredAt          int64  `json:"expired_at" binding:"required,min=0,max=4102444800"`

	// NOTE: this field not used!
	Environment string `json:"environment" binding:"omitempty"`
}

type updatePolicy struct {
	ID int64 `json:"id" binding:"required"`
	policy
}

func (slz *policiesAlterSerializer) validate() (bool, string) {
	if len(slz.CreatePolicies) > 0 {
		if valid, message := common.ValidateArray(slz.CreatePolicies); !valid {
			return false, message
		}
	}

	if len(slz.UpdatePolicies) > 0 {
		if valid, message := common.ValidateArray(slz.UpdatePolicies); !valid {
			return false, message
		}
	}

	return true, ""
}

type policiesDeleteSerializer struct {
	policySerializer
	SystemID string  `json:"system_id" binding:"required"`
	IDs      []int64 `json:"ids" binding:"required,gt=0"`
}

type queryListPolicySerializer struct {
	SubjectType     string `form:"subject_type" json:"subject_type" binding:"required"`
	SubjectID       string `form:"subject_id" json:"subject_id" binding:"required"`
	BeforeExpiredAt int64  `form:"before_expired_at" json:"before_expired_at" binding:"required,min=0"`
}
