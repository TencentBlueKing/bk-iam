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

	"iam/pkg/util"
)

const (
	defaultPage     = int64(1)
	defaultPageSize = int64(100)
)

type listQuerySerializer struct {
	ActionID  string `form:"action_id" binding:"required" example:"edit_host"`
	PageSize  int64  `form:"page_size" binding:"omitempty,min=10,max=500" example:"100"`
	Page      int64  `form:"page" binding:"omitempty,min=1" example:"1"`
	Timestamp int64  `form:"timestamp" binding:"omitempty,min=1" example:"1592899208"`
}

func (s *listQuerySerializer) validate() (bool, string) {
	if s.Timestamp != 0 {
		timestamp := util.TodayStartTimestamp()
		if timestamp-s.Timestamp > 24*60*60 {
			return false, fmt.Sprintf("timestamp(%d) should not less than one day before(%d)", s.Timestamp, timestamp)
		}
	}

	return true, "ok"
}

func (s *listQuerySerializer) initDefault() {
	if s.Page == 0 {
		s.Page = defaultPage
	}

	if s.PageSize == 0 {
		s.PageSize = defaultPageSize
	}

	// 起点: 当前00:00:00 之后的策略; 防止已过期的混入; 同时允许前一天的任务timestamp继续执行
	if s.Timestamp == 0 {
		// default: today 00:00:00
		s.Timestamp = util.TodayStartTimestamp()
	}
}

type thinPolicyResponse struct {
	Version    string                 `json:"version" example:"1"`
	ID         int64                  `json:"id" example:"100"`
	Subject    policyResponseSubject  `json:"subject"`
	Expression map[string]interface{} `json:"expression"`
	ExpiredAt  int64                  `json:"expired_at" example:"4102444800"`
}

type policyListResponseMetadata struct {
	System    string               `json:"system" example:"bk_test"`
	Action    policyResponseAction `json:"action"`
	Timestamp int64                `json:"timestamp" example:"1592899208"`
}

type policyListResponse struct {
	Metadata policyListResponseMetadata `json:"metadata"`
	Count    int64                      `json:"count" example:"120"`
	Results  []thinPolicyResponse       `json:"results"`
}
