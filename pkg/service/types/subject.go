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

// SubjectDepartment 用户的部门PK列表
type SubjectDepartment struct {
	SubjectPK     int64   `json:"subject_pk"`
	DepartmentPKs []int64 `json:"department_pks"`
}

// GroupMember ...
type GroupMember struct {
	PK        int64     `json:"pk"`
	SubjectPK int64     `json:"subject_pk"`
	ExpiredAt int64     `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SubjectGroup subject关联的组
type SubjectGroup struct {
	PK        int64     `json:"pk"`
	GroupPK   int64     `json:"group_pk"`
	ExpiredAt int64     `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ThinSubjectGroup keep the minimum fields of a group, with the group subject_pk and expired_at
type ThinSubjectGroup struct {
	// GroupPK is the subject_pk of group
	GroupPK   int64 `json:"group_pk" msgpack:"p"`
	ExpiredAt int64 `json:"expired_at" msgpack:"pe"`
}

// GroupAuthType 用于鉴权查询
type GroupAuthType struct {
	GroupPK  int64 `json:"group_pk"`
	AuthType int64 `json:"auth_type"`
}

// SubjectRelationForUpdate 用于更新 subject-relation
type SubjectRelationForUpdate struct {
	PK        int64 `json:"pk"`
	SubjectPK int64 `json:"subject_pk"`
	ExpiredAt int64 `json:"expired_at"`
}

// SubjectRelationForCreate 用于创建 subject-relation
type SubjectRelationForCreate struct {
	SubjectPK int64 `json:"subject_pk"`
	GroupPK   int64 `json:"group_pk"`
	ExpiredAt int64 `json:"expired_at"`
}
