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

	"github.com/TencentBlueKing/gopkg/stringx"

	"iam/pkg/api/common"
)

type subject struct {
	Type string `json:"type" binding:"required" example:"user"`
	ID   string `json:"id" binding:"required" example:"admin"`
}

type action struct {
	ID string `json:"id" binding:"required" example:"edit"`
}

type resource struct {
	System    string                 `json:"system" binding:"required" example:"bk_paas"`
	Type      string                 `json:"type" binding:"required" example:"app"`
	ID        string                 `json:"id" binding:"required" example:"framework"`
	Attribute map[string]interface{} `json:"attribute" binding:"required"`
}

// UID ...
func (r *resource) UID() string {
	s := fmt.Sprintf("%s:%s:%s:%v", r.System, r.Type, r.ID, r.Attribute)
	return stringx.MD5Hash(s)
}

// query for ext resources 附加查询的资源实例
type extResource struct {
	System string   `json:"system" binding:"required" example:"bk_paas"`
	Type   string   `json:"type" binding:"required" example:"app"`
	IDs    []string `json:"ids" binding:"required,gt=0"`
}

type baseRequest struct {
	System  string  `json:"system" binding:"required" example:"bk_paas"`
	Subject subject `json:"subject" binding:"required"`
}

// ====== auth
type authRequest struct {
	baseRequest
	// required
	Resources []resource `json:"resources" binding:"required"`
	Action    action     `json:"action" binding:"required"`
}

type authResponse struct {
	Allowed bool `json:"allowed" example:"false"`
}

// ====== auth v2
type authV2Request struct {
	Subject   subject    `json:"subject" binding:"required"`
	Action    action     `json:"action" binding:"required"`
	Resources []resource `json:"resources" binding:"required"`
}

type authV2Response struct {
	Allowed bool `json:"allowed" example:"false"`
}

// ======= auth by actions

type authByActionsRequest struct {
	baseRequest
	// can't be empty
	Resources []resource `json:"resources" binding:"required"`
	Actions   []action   `json:"actions" binding:"required,max=10"`
}

type authByActionsResponse map[string]bool

// ======= auth by resources

type authByResourcesRequest struct {
	baseRequest
	Action        action       `json:"action" binding:"required"`
	ResourcesList [][]resource `json:"resources_list" binding:"required,max=100"`
}

type authByResourcesResponse map[string]bool

// ====== query
type queryRequest struct {
	baseRequest
	// can be empty
	Resources []resource `json:"resources" binding:"omitempty"`
	Action    action     `json:"action" binding:"required"`
}

// ====== query v2

type queryV2Request struct {
	Subject subject `json:"subject" binding:"required"`
	Action  action  `json:"action" binding:"required"`
	// can be empty
	Resources []resource `json:"resources" binding:"omitempty"`
}

// ======= query by actions

type queryByActionsRequest struct {
	baseRequest
	// can be empty
	Resources []resource `json:"resources" binding:"omitempty"`
	Actions   []action   `json:"actions" binding:"required"`
}

type actionInResponse struct {
	ID string `json:"id" example:"edit"`
}

type actionPoliciesResponse struct {
	Action    actionInResponse       `json:"action"`
	Condition map[string]interface{} `json:"condition"`
}

// ======= query by ext resources
type queryByExtResourcesRequest struct {
	queryRequest
	ExtResources []extResource `json:"ext_resources" binding:"required,lte=1000"`
}

// Validate ...
func (q *queryByExtResourcesRequest) Validate() (bool, string) {
	if len(q.ExtResources) == 0 {
		return true, ""
	}
	if valid, message := common.ValidateArray(q.ExtResources); !valid {
		return false, message
	}
	return true, ""
}

type GenSuperPermissionResponseData func() interface{}
