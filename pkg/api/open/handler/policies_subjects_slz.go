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

import "iam/pkg/abac/prp"

type subjectsSerializer struct {
	IDs string `form:"ids" binding:"required" example:"1,2,3"`

	Type string `form:"type" json:"type" binding:"omitempty,oneof=abac rbac" example:"abac"`
}

func (s *subjectsSerializer) initDefault() {
	if s.Type == "" {
		s.Type = prp.EngineListPolicyTypeAbac
	}
}

type policyIDSubject struct {
	PolicyID int64                 `json:"id" example:"100"`
	Subject  policyResponseSubject `json:"subject"`
}

type policySubjectsResponse []policyIDSubject
