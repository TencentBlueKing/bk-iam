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
	"time"
)

type updateModelChangeEventStatusSerializer struct {
	Status string `json:"status" binding:"required"`
}

type deleteModelChangeEventSerializer struct {
	Status          string `json:"status" binding:"required"`
	BeforeUpdatedAt int64  `json:"before_updated_at" binding:"omitempty,min=1,max=4102444800" example:"1592899208"`
	Limit           int64  `json:"limit" binding:"omitempty,min=1,max=100000"`
}

func (s *deleteModelChangeEventSerializer) initDefault() {
	if s.BeforeUpdatedAt == 0 {
		// 对于删除，默认可删除一个月前数据
		s.BeforeUpdatedAt = time.Now().AddDate(0, -1, 0).Unix()
	}
	if s.Limit == 0 {
		// 默认最多删除1000条数据
		s.Limit = 1000
	}
}
