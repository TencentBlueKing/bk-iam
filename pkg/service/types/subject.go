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

// GroupSubject 关系数据
type GroupSubject struct {
	SubjectPK int64 `json:"subject_pk"`
	GroupPK   int64 `json:"group_pk"`
	ExpiredAt int64 `json:"expired_at"`
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

type GroupAlterEvent struct {
	UUID       string  `json:"uuid"`
	GroupPK    int64   `json:"group_pk"`
	ActionPKs  []int64 `json:"action_pks"`
	SubjectPKs []int64 `json:"subject_pks"`
}

// ResourceExpiredAt ...
type ResourceExpiredAt struct {
	Resources map[int64][]string `json:"resources"` // resource_type_pk -> resource_ids
	ExpiredAt int64              `json:"expired_at"`
}

// SubjectActionGroupResource ...
type SubjectActionGroupResource struct {
	PK int64 `db:"pk"`

	SubjectPK     int64                       `json:"subject_pk"`
	ActionPK      int64                       `json:"action_pk"`
	GroupResource map[int64]ResourceExpiredAt `json:"group_resource"` // group_pk -> ExpiredAtResource
}

// DeleteGroupResource ...
func (s *SubjectActionGroupResource) DeleteGroupResource(groupPK int64) {
	delete(s.GroupResource, groupPK)
}

// UpdateGroupResource ...
func (s *SubjectActionGroupResource) UpdateGroupResource(groupPK int64, resources map[int64][]string, expiredAt int64) {
	s.GroupResource[groupPK] = ResourceExpiredAt{
		Resources: resources,
		ExpiredAt: expiredAt,
	}
}

// SubjectActionExpression ...
type SubjectActionExpression struct {
	PK         int64  `json:"pk" msgpack:"p"`
	SubjectPK  int64  `json:"subject_pk" msgpack:"s1"`
	ActionPK   int64  `json:"action_pk" msgpack:"a"`
	Expression string `json:"expression" msgpack:"e1"`
	Signature  string `json:"signature" msgpack:"s2"`
	ExpiredAt  int64  `json:"expired_at" msgpack:"e2"`
}

// SubjectActionGroupMessage ...
type SubjectActionGroupMessage struct {
	SubjectPK int64   `json:"subject_pk"`
	ActionPK  int64   `json:"action_pk"`
	GroupPKs  []int64 `json:"group_pks"`
}

// SubjectActionAlterEvent ...
type SubjectActionAlterEvent struct {
	UUID       string                      `json:"uuid"`
	Messages   []SubjectActionGroupMessage `json:"messages"`
	Status     int64                       `json:"status"`
	CheckCount int64                       `json:"check_count"`
}

const (
	SubjectActionAlterEventStatusCreated int64 = iota
	SubjectActionAlterEventStatusPushed
	SubjectActionAlterEventStatusProcessing
)
