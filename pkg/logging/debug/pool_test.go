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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newEntryPool(t *testing.T) {
	pool := newEntryPool()

	e := pool.Get()
	assert.NotNil(t, e)
	assert.Empty(t, e.Context)
	assert.Empty(t, e.Steps)
	assert.Empty(t, e.Evals)

	e.WithValue("hello", "world")
	assert.NotEmpty(t, e.Context)

	pool.Put(e)

	e1 := pool.Get()
	assert.NotNil(t, e1)
	assert.Empty(t, e1.Context)
	assert.Empty(t, e1.Steps)
	assert.Empty(t, e1.Evals)

	e1.AddSubDebug(pool.Get())
	assert.Len(t, e1.SubDebugs, 1)
	e1.AddSubDebug(nil)
	assert.Len(t, e1.SubDebugs, 1)

	pool.Put(e1)

	// nil, do nothing
	pool.Put(nil)
}
