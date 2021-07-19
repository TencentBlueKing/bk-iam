/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package impls

import (
	"errors"
	"testing"
	"time"

	"iam/pkg/cache"
	"iam/pkg/cache/memory"

	"github.com/stretchr/testify/assert"
)

func TestAppCodeAppSecretCacheKey_Key(t *testing.T) {
	k := AppCodeAppSecretCacheKey{
		AppCode:   "hello",
		AppSecret: "123",
	}
	assert.Equal(t, "hello:123", k.Key())
}

func TestVerifyAppCodeAppSecret(t *testing.T) {
	var (
		expiration = 5 * time.Minute
	)

	// valid
	retrieveFunc := func(key cache.Key) (interface{}, error) {
		return true, nil
	}
	mockCache := memory.NewCache(
		"mockCache", false, retrieveFunc, expiration)
	LocalAppCodeAppSecretCache = mockCache
	assert.True(t, VerifyAppCodeAppSecret("test", "123"))

	// error
	retrieveFunc = func(key cache.Key) (interface{}, error) {
		return false, errors.New("error here")
	}
	mockCache = memory.NewCache(
		"mockCache", false, retrieveFunc, expiration)
	LocalAppCodeAppSecretCache = mockCache
	assert.False(t, VerifyAppCodeAppSecret("test", "123"))
}
