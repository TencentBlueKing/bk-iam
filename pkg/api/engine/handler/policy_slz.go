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
	"fmt"

	"iam/pkg/abac/prp"
	"iam/pkg/util"
)

// -- listPolicy

type listPolicySerializer struct {
	Timestamp int64 `form:"timestamp" json:"timestamp" binding:"omitempty,min=1" example:"1592899208"`

	MinID int64 `form:"min_id" json:"min_id" binding:"omitempty,min=1" example:"1"`
	MaxID int64 `form:"max_id" json:"max_id" binding:"omitempty,min=1" example:"10001"`

	IDs string `form:"ids" json:"ids" binding:"omitempty" example:"1,2,3"`

	Type string `form:"type" json:"type" binding:"omitempty,oneof=abac rbac" example:"abac"`
}

// getIDs parse string ids to slice
func (s *listPolicySerializer) getIDs() ([]int64, error) {
	return util.StringToInt64Slice(s.IDs, ",")
}

func (s *listPolicySerializer) hasIDs() bool {
	return s.IDs != ""
}

func (s *listPolicySerializer) validate() (bool, string) {
	// 如果有ids, ids的优先级最高
	if s.hasIDs() {
		ids, err := s.getIDs()
		if err != nil {
			return false, fmt.Sprintf("ids(%s) should be int64 slice", s.IDs)
		}

		if len(ids) > 200 {
			return false, fmt.Sprintf("ids length=%d, should less than 200", len(ids))
		}

		return true, "ok"
	}

	// validate min_id and max_id
	if s.MinID == 0 {
		return false, "min_id should greater than 0"
	}
	if s.MaxID == 0 {
		return false, "max_id should greater than 0"
	}
	// NOTE: should be >, not >=, some case MinID will be equal to MaxID
	if s.MinID > s.MaxID {
		return false, "min_id should less than max_id"
	}

	if s.Timestamp != 0 {
		timestamp := util.TodayStartTimestamp()
		if timestamp-s.Timestamp > 24*60*60 {
			return false, fmt.Sprintf("timestamp(%d) should not less than one day before(%d)", s.Timestamp, timestamp)
		}
	}

	return true, "ok"
}

func (s *listPolicySerializer) initDefault() {
	// 起点: 当前00:00:00 之后的策略; 防止已过期的混入; 同时允许前一天的任务timestamp继续执行
	if s.Timestamp == 0 {
		// default: today 00:00:00
		s.Timestamp = util.TodayStartTimestamp()
	}

	if s.Type == "" {
		s.Type = prp.PolicyTypeAbac
	}
}

type policyResponseSubject struct {
	Type string `json:"type" example:"user"`
	ID   string `json:"id"   example:"admin"`
	Name string `json:"name" example:"Administer"`
}

type policyResponseAction struct {
	ID string `json:"id" example:"edit"`
}

type enginePolicyResponse struct {
	Version    string                 `json:"version"     example:"1"`
	ID         int64                  `json:"id"          example:"100"`
	System     string                 `json:"system"      example:"bk_cmdb"`
	Actions    []policyResponseAction `json:"actions"`
	Subject    policyResponseSubject  `json:"subject"`
	Expression map[string]interface{} `json:"expression"`
	TemplateID int64                  `json:"template_id"`
	ExpiredAt  int64                  `json:"expired_at"  example:"4102444800"`
	UpdatedAt  int64                  `json:"updated_at"  example:"4102444800"`
}

type policyListResponse struct {
	Metadata listPolicySerializer   `json:"metadata"`
	Results  []enginePolicyResponse `json:"results"`
}

// -- listPolicyPKs

type listPolicyIDsSerializer struct {
	BeginUpdatedAt int64  `form:"begin_updated_at" json:"begin_updated_at" binding:"min=1"`
	EndUpdatedAt   int64  `form:"end_updated_at"   json:"end_updated_at"   binding:"min=1"`
	Type           string `form:"type"             json:"type"             binding:"omitempty,oneof=abac rbac"`
}

func (s *listPolicyIDsSerializer) validate() (bool, string) {
	if s.BeginUpdatedAt >= s.EndUpdatedAt {
		return false, "begin_updated_at should less than end_updated_at"
	}

	if s.EndUpdatedAt-s.BeginUpdatedAt > 3600 {
		return false, "the time gap between begin_update_at and end_update_at should be less than 1 hour(3600 seconds)"
	}

	return true, "ok"
}

func (s *listPolicyIDsSerializer) initDefault() {
	if s.Type == "" {
		s.Type = prp.PolicyTypeAbac
	}
}

type listPolicyIDsResponse struct {
	IDs []int64 `json:"ids"`
}

// --a getMaxPolicyPK

type getMaxPolicyIDSerializer struct {
	UpdatedAt int64  `form:"updated_at" json:"updated_at" binding:"min=1"                     example:"1592899208"`
	Type      string `form:"type"       json:"type"       binding:"omitempty,oneof=abac rbac" example:"abac"`
}

func (s *getMaxPolicyIDSerializer) initDefault() {
	if s.Type == "" {
		s.Type = prp.PolicyTypeAbac
	}
}

type getMaxPolicyIDResponse struct {
	ID int64 `json:"id"`
}
