/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package debug

import "iam/pkg/abac/types"

// WithValue ...
func WithValue(e *Entry, key string, value interface{}) {
	if e == nil {
		return
	}

	e.WithValue(key, value)
}

// WithValues ...
func WithValues(e *Entry, data map[string]interface{}) {
	if e == nil {
		return
	}

	e.WithValues(data)
}

// WithUnknownEvalPolicies ...
func WithUnknownEvalPolicies(e *Entry, policies []types.AuthPolicy) {
	if e == nil {
		return
	}

	for _, policy := range policies {
		e.WithUnknownEval(policy.ID)
	}
}

// WithPassEvalPolicies ...
func WithPassEvalPolicies(e *Entry, policies []types.AuthPolicy) {
	if e == nil {
		return
	}

	for _, policy := range policies {
		e.WithPassEval(policy.ID)
	}
}

// WithNoPassEvalPolicies ...
func WithNoPassEvalPolicies(e *Entry, policies []types.AuthPolicy) {
	if e == nil {
		return
	}

	for _, policy := range policies {
		e.WithNoPassEval(policy.ID)
	}
}

// WithPassEvalPolicy ...
func WithPassEvalPolicy(e *Entry, policyID int64) {
	if e == nil {
		return
	}

	e.WithPassEval(policyID)
}

func WithPassEvalPolicyIDs(e *Entry, policyIDs []int64) {
	if e == nil {
		return
	}

	if len(policyIDs) == 0 {
		return
	}

	for _, pid := range policyIDs {
		e.WithPassEval(pid)
	}
}

// WithNoPassEvalPolicy ...
func WithNoPassEvalPolicy(e *Entry, policyID int64) {
	if e == nil {
		return
	}

	e.WithNoPassEval(policyID)
}

// WithError ...
func WithError(e *Entry, err error) {
	if e == nil {
		return
	}

	e.WithError(err)
}

// AddSubDebug ...
func AddSubDebug(e *Entry, subEntry *Entry) {
	if e == nil || subEntry == nil {
		return
	}

	e.AddSubDebug(subEntry)
}

// AddStep ...
func AddStep(e *Entry, name string) {
	if e == nil {
		return
	}

	e.AddStep(NewStep(name))
}

// NewSubDebug ...
func NewSubDebug(e *Entry) *Entry {
	if e == nil {
		return nil
	}

	entry := EntryPool.Get()
	e.AddSubDebug(entry)
	return entry
}
