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

import (
	"errors"
	"testing"

	"iam/pkg/abac/types"

	"github.com/stretchr/testify/assert"
)

func TestNil(t *testing.T) {
	var e *Entry

	WithValue(e, "hello", "world")
	WithValues(e, map[string]interface{}{
		"hello": "world",
	})

	WithUnknownEvalPolicies(e, nil)
	WithPassEvalPolicies(e, nil)
	WithNoPassEvalPolicies(e, nil)

	WithPassEvalPolicy(e, 1)
	WithNoPassEvalPolicy(e, 1)
	WithError(e, nil)
	AddStep(e, "hello")
	AddSubDebug(e, nil)

	e1 := NewSubDebug(e)
	assert.Nil(t, e1)
}

func TestNotNil(t *testing.T) {
	pool := newEntryPool()
	e := pool.Get()

	WithValue(e, "hello", "world")
	assert.Contains(t, e.Context, "hello")

	WithValues(e, map[string]interface{}{
		"hello1": "world",
	})
	assert.Contains(t, e.Context, "hello1")

	WithUnknownEvalPolicies(e, []types.AuthPolicy{
		{ID: 1},
	})
	WithPassEvalPolicies(e, []types.AuthPolicy{
		{ID: 2},
	})
	WithNoPassEvalPolicies(e, []types.AuthPolicy{
		{ID: 3},
	})
	WithPassEvalPolicy(e, 4)
	WithNoPassEvalPolicy(e, 5)

	assert.Len(t, e.Evals, 5)
	assert.Equal(t, e.Evals[int64(1)], Unknown)
	assert.Equal(t, e.Evals[int64(2)], Pass)
	assert.Equal(t, e.Evals[int64(3)], NoPass)
	assert.Equal(t, e.Evals[int64(4)], Pass)
	assert.Equal(t, e.Evals[int64(5)], NoPass)

	msg := "this is a error"
	WithError(e, errors.New(msg))
	assert.Equal(t, e.Error, msg)

	AddStep(e, "hello")
	assert.Len(t, e.Steps, 1)

	e1 := NewSubDebug(e)
	assert.Len(t, e.SubDebugs, 1)
	assert.NotNil(t, e1)

	AddSubDebug(e, pool.Get())
	assert.Len(t, e.SubDebugs, 2)
}
