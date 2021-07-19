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

// 参考来源：https://docs.aws.amazon.com/en_pv/IAM/latest/UserGuide/reference_policies_elements_condition_operators.html
/*
const (
	// 权限中心已支持的操作
	StringEquals  = "StringEquals"
	StringPrefix  = "StringPrefix"
	NumericEquals = "NumericEquals"
	Bool          = "Bool"
	Any           = "Any"
	// 暂未支持的操作
	// 字符串
	StringNotEquals           = "StringNotEquals"
	StringEqualsIgnoreCase    = "StringEqualsIgnoreCase"
	StringNotEqualsIgnoreCase = "StringNotEqualsIgnoreCase"
	StringLike                = "StringLike"
	StringNotLike             = "StringNotLike"
	// 数字
	NumericNotEquals         = "NumericNotEquals"
	NumericLessThan          = "NumericLessThan"
	NumericLessThanEquals    = "NumericLessThanEquals"
	NumericGreaterThan       = "NumericGreaterThan"
	NumericGreaterThanEquals = "NumericGreaterThanEquals"
	// 时间
	DateEquals = "DateEquals"
	DateNotEquals = "DateNotEquals"
	DateLessThan = "DateLessThan"
	DateLessThanEquals = "DateLessThanEquals"
	DateGreaterThan = "DateGreaterThan"
	DateGreaterThanEquals = "DateGreaterThanEquals"
	// IP
	IpAddress = "IpAddress"
	NotIpAddress = "NotIpAddress"
	// 可用于组合
	ForAnyValue = "ForAnyValue"
	ForAllValues = "ForAllValues"
	IfExists = "IfExists"
)
*/

// Policy ...
type Policy struct {
	Version string
	ID      int64
	System  string

	Subject Subject
	Action  Action
	// PRP 暂时不解析ResourceExpression里的
	Expression string
	ExpiredAt  int64
	TemplateID int64
}

// SaaSPolicy ...
type SaaSPolicy struct {
	Version string `json:"version"`
	ID      int64  `json:"id"`

	System    string `json:"system"`
	ActionID  string `json:"action_id"`
	ExpiredAt int64  `json:"expired_at"`
}

// AuthPolicy ...
type AuthPolicy struct {
	Version string
	ID      int64

	Expression          string
	ExpressionSignature string
	ExpiredAt           int64
}

// PolicyPKExpiredAt ...
type PolicyPKExpiredAt struct {
	PK        int64
	ExpiredAt int64
}
