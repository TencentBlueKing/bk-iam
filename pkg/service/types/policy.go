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

// ======== FOR AUTH ========

// AuthPolicy for auth
type AuthPolicy struct {
	PK           int64 `msgpack:"p"`
	SubjectPK    int64 `msgpack:"s"`
	ExpressionPK int64 `msgpack:"e1"`
	ExpiredAt    int64 `msgpack:"e2"`
}

// AuthExpression for auth
type AuthExpression struct {
	PK         int64  `msgpack:"p"`
	Expression string `msgpack:"ea"`
	Signature  string `msgpack:"s"`
}

// IsEmpty ...
func (a AuthExpression) IsEmpty() bool {
	return a.PK == 0
}

// ======== FOR OTHERS ========

// QueryPolicy ...
type QueryPolicy struct {
	PK           int64
	SubjectPK    int64
	ActionPK     int64
	ExpressionPK int64
	ExpiredAt    int64
}

// EngineQueryPolicy query policy for iam engine
// type EngineQueryPolicy struct {
// 	QueryPolicy
// 	TemplateID int64
// 	UpdatedAt  int64
// }

type EnginePolicy struct {
	Version string
	ID      int64
	System  string
	// abac, policy with single action
	// rbac, policy with multiple actions
	ActionPKs []int64

	SubjectPK int64

	ExpressionStr string
	ExpressionPK  int64
	TemplateID    int64
	ExpiredAt     int64
	UpdatedAt     int64
}

// Policy ...
type Policy struct {
	Version string
	ID      int64

	SubjectPK  int64
	ActionPK   int64
	Expression string
	Signature  string

	ExpiredAt  int64
	TemplateID int64
}

// ThinPolicy ...
type ThinPolicy struct {
	Version string
	ID      int64

	ActionPK  int64
	ExpiredAt int64
}
