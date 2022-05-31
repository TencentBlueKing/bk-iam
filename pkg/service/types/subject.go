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

import "time"

// Subject ...
type Subject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SubjectMember ...
type SubjectMember struct {
	PK              int64     `json:"pk"`
	Type            string    `json:"type"`
	ID              string    `json:"id"`
	PolicyExpiredAt int64     `json:"policy_expired_at"`
	CreateAt        time.Time `json:"created_at"`
}

// SubjectGroup subject关联的组
type SubjectGroup struct {
	PK              int64     `json:"pk"`
	Type            string    `json:"type"`
	ID              string    `json:"id"`
	PolicyExpiredAt int64     `json:"policy_expired_at"`
	CreateAt        time.Time `json:"created_at"`
}

// SubjectDepartment 用户的部门PK列表
type SubjectDepartment struct {
	SubjectPK     int64
	DepartmentPKs []int64
}

// SubjectRelationPKPolicyExpiredAt ...
type SubjectRelationPKPolicyExpiredAt struct {
	PK              int64
	PolicyExpiredAt int64
}

// SubjectRelation ...
type SubjectRelation struct {
	SubjectPK       int64
	SubjectType     string
	SubjectID       string
	ParentPK        int64
	ParentType      string
	ParentID        string
	PolicyExpiredAt int64
}
