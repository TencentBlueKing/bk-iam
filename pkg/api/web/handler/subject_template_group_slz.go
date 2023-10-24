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

type subjectTemplateGroupSerializer struct {
	Type       string `json:"type" binding:"required,oneof=user department"`
	ID         string `json:"id" binding:"required"`
	TemplateID int64  `json:"template_id" binding:"required"`
	GroupID    int64  `json:"group_id" binding:"required"`
	ExpiredAt  int64  `json:"expired_at" binding:"omitempty,min=1,max=4102444800"`
}
