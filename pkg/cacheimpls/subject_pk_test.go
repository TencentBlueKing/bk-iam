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
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/redis"
	"iam/pkg/service"
	"iam/pkg/service/mock"
)

func TestSubjectIDCacheKey_Key(t *testing.T) {
	key := SubjectIDCacheKey{
		Type: "user",
		ID:   "admin",
	}
	assert.Equal(t, "user:admin", key.Key())
}

func TestGetSubjectPK(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	var (
		expiration = 5 * time.Minute
	)

	mockService := mock.NewMockSubjectService(ctl)
	mockService.EXPECT().GetPK("user", "admin").Return(int64(64), nil).AnyTimes()

	patches := gomonkey.ApplyFunc(service.NewSubjectService,
		func() service.SubjectService {
			return mockService
		})
	defer patches.Reset()

	mockCache := redis.NewMockCache("mockCache", expiration)

	SubjectPKCache = mockCache

	pk, err := GetSubjectPK("user", "admin")
	assert.NoError(t, err)
	assert.Equal(t, int64(64), pk)
}
