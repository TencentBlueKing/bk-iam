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

type policyGetSerializer struct {
	PolicyID int64 `uri:"policy_id"`
}

type policyGetResponse struct {
	Version    string                 `json:"version" example:"1"`
	ID         int64                  `json:"id" example:"100"`
	System     string                 `json:"system" example:"bk_test"`
	Subject    policyResponseSubject  `json:"subject"`
	Action     policyResponseAction   `json:"action"`
	Expression map[string]interface{} `json:"expression"`
	ExpiredAt  int64                  `json:"expired_at" example:"4102444800"`
}

type policyResponseSubject struct {
	Type string `json:"type" example:"user"`
	ID   string `json:"id" example:"admin"`
	Name string `json:"name" example:"Administer"`
}

type policyResponseAction struct {
	ID string `json:"id" example:"edit"`
}
