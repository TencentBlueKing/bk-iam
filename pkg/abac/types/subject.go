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

import (
	"time"
)

// Subject 被授权对象
type Subject struct {
	Type      string
	ID        string
	Attribute *SubjectAttribute
}

// SubjectGroup 用户-组/部门-组, PK is the group subject pk, PolicyExpiredAt is the expired
// for abac(prp), only need to know two fields to do query/auth
type SubjectGroup struct {
	PK              int64 `json:"pk"`
	PolicyExpiredAt int64 `json:"policy_expired_at"`
}

// NewSubject ...
func NewSubject() Subject {
	return Subject{
		Attribute: NewSubjectAttribute(),
	}
}

// FillAttributes 填充subject的属性
func (s *Subject) FillAttributes(pk int64, groups []SubjectGroup, departments []int64) {
	s.Attribute.SetPK(pk)
	s.Attribute.SetGroups(groups)
	s.Attribute.SetDepartments(departments)
}

// GetEffectGroupPKs 获取有效的用户组PK
func (s *Subject) GetEffectGroupPKs() ([]int64, error) {
	groups, err := s.Attribute.GetGroups()
	if err != nil {
		return nil, err
	}

	nowUnix := time.Now().Unix()
	pks := make([]int64, 0, len(groups))

	for _, group := range groups {
		// 仅仅在有效期内才需要
		if group.PolicyExpiredAt > nowUnix {
			pks = append(pks, group.PK)
		}
	}
	return pks, nil
}

// GetDepartmentPKs 获取部门PK
func (s *Subject) GetDepartmentPKs() ([]int64, error) {
	return s.Attribute.GetDepartments()
}

// Reset 重置数据
func (s *Subject) Reset() {
	s.ID = ""
	s.Type = ""
	s.Attribute.Reset()
}
