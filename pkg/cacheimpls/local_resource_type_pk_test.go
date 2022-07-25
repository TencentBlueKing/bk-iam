/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"errors"
	"testing"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/cache/memory"
	"github.com/stretchr/testify/assert"
)

func TestGetLocalResourceTypePK(t *testing.T) {
	expiration := 5 * time.Minute

	// valid
	retrieveFunc := func(key cache.Key) (interface{}, error) {
		return int64(64), nil
	}
	mockCache := memory.NewCache(
		"mockCache", false, retrieveFunc, expiration, nil)
	LocalResourceTypePKCache = mockCache

	pk, err := GetLocalResourceTypePK("test", "id")
	assert.NoError(t, err)
	assert.Equal(t, int64(64), pk)

	// error
	retrieveFunc = func(key cache.Key) (interface{}, error) {
		return int64(0), errors.New("error here")
	}
	mockCache = memory.NewCache(
		"mockCache", false, retrieveFunc, expiration, nil)
	LocalResourceTypePKCache = mockCache

	_, err = GetLocalResourceTypePK("test", "id")
	assert.Error(t, err)
}
