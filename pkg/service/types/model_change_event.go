/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

// ModelChangeEvent is a event to store model change detail
type ModelChangeEvent struct {
	PK        int64  `json:"pk" structs:"pk"` // 自增列
	Type      string `json:"type" structs:"type"`
	Status    string `json:"status" structs:"status"`
	SystemID  string `json:"system_id" structs:"system_id"`
	ModelType string `json:"model_type" structs:"model_type"`
	ModelID   string `json:"model_id" structs:"model_id"`
	ModelPK   int64  `json:"model_pk" structs:"model_pk"`
}
