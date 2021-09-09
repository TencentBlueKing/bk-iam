/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package backend

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTTLCache(t *testing.T) {
	c := newTTLCache(5*time.Second, 10*time.Second)
	assert.NotNil(t, c)

	c = newTTLCache(5*time.Second, 0*time.Second)
	assert.NotNil(t, c)
}

func TestMemoryBackend(t *testing.T) {
	be := NewMemoryBackend("test", 5*time.Second, nil)
	assert.NotNil(t, be)

	_, found := be.Get("not_exists")
	assert.False(t, found)

	be.Set("hello", "world", time.Duration(0))
	value, found := be.Get("hello")
	assert.True(t, found)
	assert.Equal(t, "world", value)

	be.Delete("hello")
	_, found = be.Get("hello")
	assert.False(t, found)
}
