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

// AllowEmptyFields ...
type AllowEmptyFields struct {
	keys map[string]struct{}
}

// NewAllowEmptyFields ...
func NewAllowEmptyFields() AllowEmptyFields {
	return AllowEmptyFields{keys: map[string]struct{}{}}
}

// HasKey ...
func (a *AllowEmptyFields) HasKey(key string) bool {
	_, ok := a.keys[key]
	return ok
}

// AddKey ...
func (a *AllowEmptyFields) AddKey(key string) {
	a.keys[key] = struct{}{}
}

const (
	AuthTypeNone int64 = 0
	AuthTypeABAC int64 = 1
	AuthTypeRBAC int64 = 2
	AuthTypeAll  int64 = 7 // 预留一位 4, ALL 为 7
)
