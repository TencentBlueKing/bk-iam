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
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// ThinSubjectGroup keep the minimum fields of a group, with the group subject_pk and expired_at
type ThinSubjectGroup struct {
	// PK is the subject_pk of group
	PK              int64 `json:"pk" msgpack:"p"`
	PolicyExpiredAt int64 `json:"policy_expired_at" msgpack:"pe"`
}

// SubjectDetail 用户鉴权时的复合缓存, 如果要加新成员, 必须变更cache名字, 防止从已有缓存数据拿不到对应的字段产生bug
type SubjectDetail struct {
	DepartmentPKs []int64            `json:"department_pks" msgpack:"dps"`
	SubjectGroups []ThinSubjectGroup `json:"subject_groups" msgpack:"sg"`
}

//custom the msgpack marshal/unmarshal, for better performance
var _ msgpack.Marshaler = (*SubjectDetail)(nil)

// MarshalMsgpack to   1,2,3,4|pk1:12,pk2:34,pk3:45
func (s *SubjectDetail) MarshalMsgpack() ([]byte, error) {
	var b1 strings.Builder
	maxIdx1 := len(s.DepartmentPKs) - 1
	for idx, deptPK := range s.DepartmentPKs {
		b1.WriteString(strconv.FormatInt(deptPK, 10))
		if idx != maxIdx1 {
			b1.WriteRune(',')
		}
	}

	var b2 strings.Builder

	maxIdx2 := len(s.SubjectGroups) - 1
	for idx, sg := range s.SubjectGroups {
		b2.WriteString(strconv.FormatInt(sg.PK, 10))
		b2.WriteByte(':')
		b2.WriteString(strconv.FormatInt(sg.PolicyExpiredAt, 10))
		if idx != maxIdx2 {
			b2.WriteByte(',')
		}
	}

	b3 := b1.String() + "|" + b2.String()
	return msgpack.Marshal(b3)
}

var _ msgpack.Unmarshaler = (*SubjectDetail)(nil)

// UnmarshalMsgpack from 1,2,3,4|pk1:12,pk2:34,pk3:45  to subjectDetail
func (s *SubjectDetail) UnmarshalMsgpack(b []byte) error {
	var b3 string
	err := msgpack.Unmarshal(b, &b3)
	if err != nil {
		return err
	}

	sepIndex := strings.IndexByte(b3, '|')

	// unpack the department pks
	deptPKParts := strings.Split(b3[:sepIndex], ",")
	s.DepartmentPKs = make([]int64, 0, len(deptPKParts))
	for _, deptPKStr := range deptPKParts {
		if deptPKStr == "" {
			continue
		}
		deptPK, err := strconv.ParseInt(deptPKStr, 10, 64)
		if err != nil {
			return err
		}
		s.DepartmentPKs = append(s.DepartmentPKs, deptPK)
	}

	// unpack the groups
	subjectGroupParts := strings.Split(b3[sepIndex+1:], ",")
	s.SubjectGroups = make([]ThinSubjectGroup, 0, len(subjectGroupParts))
	for _, sgStr := range subjectGroupParts {
		if sgStr == "" {
			continue
		}

		sgSepIndex := strings.IndexByte(sgStr, ':')

		gPK, err := strconv.ParseInt(sgStr[:sgSepIndex], 10, 64)
		if err != nil {
			return err
		}
		expire, err := strconv.ParseInt(sgStr[sgSepIndex+1:], 10, 64)
		if err != nil {
			return err
		}

		s.SubjectGroups = append(s.SubjectGroups, ThinSubjectGroup{
			PK:              gPK,
			PolicyExpiredAt: expire,
		})
	}
	return nil
}
