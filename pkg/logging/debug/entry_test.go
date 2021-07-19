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

	"github.com/stretchr/testify/assert"
)

func TestEntry_AddStep(t *testing.T) {
	pool := newEntryPool()
	entry := pool.Get()

	entry.AddStep(NewStep("hello"))

	assert.Len(t, entry.Steps, 1)
}

func TestEntry_WithError(t *testing.T) {
	pool := newEntryPool()
	entry := pool.Get()

	msg := "hello"
	err := errors.New(msg)

	entry.WithError(err)

	assert.Equal(t, msg, entry.Error)
}

func TestEntry_WithNoPassEval(t *testing.T) {
	pool := newEntryPool()
	entry := pool.Get()

	entry.WithNoPassEval(1)

	assert.Len(t, entry.Evals, 1)
	assert.Contains(t, entry.Evals, int64(1))
	assert.Equal(t, NoPass, entry.Evals[int64(1)])
}

func TestEntry_WithPassEval(t *testing.T) {
	pool := newEntryPool()
	entry := pool.Get()

	entry.WithPassEval(1)

	assert.Len(t, entry.Evals, 1)
	assert.Contains(t, entry.Evals, int64(1))
	assert.Equal(t, Pass, entry.Evals[int64(1)])
}

func TestEntry_WithUnknownEval(t *testing.T) {
	pool := newEntryPool()
	entry := pool.Get()

	entry.WithUnknownEval(1)

	assert.Len(t, entry.Evals, 1)
	assert.Contains(t, entry.Evals, int64(1))
	assert.Equal(t, Unknown, entry.Evals[int64(1)])
}

func TestEntry_WithValue(t *testing.T) {
	pool := newEntryPool()

	entry := pool.Get()

	assert.Empty(t, entry.Context)

	entry.WithValue("hello", "world")
	assert.Len(t, entry.Context, 1)

	assert.Contains(t, entry.Context, "hello")
}

func TestEntry_WithValues(t *testing.T) {
	pool := newEntryPool()

	entry := pool.Get()

	assert.Empty(t, entry.Context)

	entry.WithValues(map[string]interface{}{
		"hello": "world",
		"a":     1,
	})
	assert.Len(t, entry.Context, 2)

	assert.Contains(t, entry.Context, "hello")
}

func TestNewStep(t *testing.T) {
	step := NewStep("hello")
	assert.Equal(t, 0, step.Index)
	assert.Equal(t, "hello", step.Name)
}

func TestEntry_AddSubDebug(t *testing.T) {
	pool := newEntryPool()
	entry := pool.Get()

	assert.Len(t, entry.SubDebugs, 0)

	entry.AddSubDebug(pool.Get())

	assert.Len(t, entry.SubDebugs, 1)

	entry.AddSubDebug(pool.Get())

	assert.Len(t, entry.SubDebugs, 2)
}
