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

	"iam/pkg/cache"
	"iam/pkg/cache/redis"
	"iam/pkg/component"
	"iam/pkg/component/mock"
	"iam/pkg/service/types"

	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRemoteResourceCacheKey_Key(t *testing.T) {
	k := RemoteResourceCacheKey{
		System: "test",
		Type:   "host",
		ID:     "1",
		Fields: "id;name",
	}

	assert.Equal(t, stringx.MD5Hash("test:host:1:id;name"), k.Key())
}

func TestGetCMDBResource(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	var (
		expiration = 5 * time.Minute
	)

	system := types.System{
		ID:            "",
		Name:          "",
		NameEn:        "",
		Description:   "",
		DescriptionEn: "",
		Clients:       "",
		ProviderConfig: map[string]interface{}{
			"host": "",
			"auth": "none",
		},
	}

	SystemCache = redis.NewMockCache("mockCache", expiration)
	SystemCache.Set(cache.NewStringKey("test"), system, 0)

	resourceType := types.ResourceType{
		ID:            "",
		Name:          "",
		NameEn:        "",
		Description:   "",
		DescriptionEn: "",
		Parents:       nil,
		ProviderConfig: map[string]interface{}{
			"path": "/api/v1/resources",
		},
		Version: 0,
	}

	ResourceTypeCache = redis.NewMockCache("mockCache", expiration)
	ResourceTypeCache.Set(ResourceTypeCacheKey{"test", "app"}, resourceType, 0)

	req, _ := component.PrepareRequest(system, resourceType)

	mockService := mock.NewMockRemoteResourceClient(ctl)
	mockService.EXPECT().GetResources(req, "test", "app", []string{"checklist"}, []string{"name"}).Return(
		[]map[string]interface{}{{
			"id": "checklist",
		}}, nil).AnyTimes()

	component.BKRemoteResource = mockService

	mockCache := redis.NewMockCache("mockCache", expiration)

	RemoteResourceCache = mockCache

	resource, err := GetRemoteResource("test", "app", "checklist", []string{"name"})
	assert.NoError(t, err)
	assert.Equal(t, "checklist", resource["id"])
}
