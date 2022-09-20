/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pap

import "time"

// Subject ...
type Subject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SubjectDepartment ...
type SubjectDepartment struct {
	SubjectID     string   `json:"id"`
	DepartmentIDs []string `json:"departments"`
}

// GroupMember ...
type GroupMember struct {
	PK        int64     `json:"pk"`
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	ExpiredAt int64     `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SubjectGroup subject关联的组
type SubjectGroup struct {
	PK        int64     `json:"pk"`
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	ExpiredAt int64     `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupSubject ...
type GroupSubject struct {
	PK        int64 `json:"pk"`
	Subject   Subject
	Group     Subject
	ExpiredAt int64     `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}
