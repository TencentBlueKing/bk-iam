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
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"iam/pkg/service"
	"iam/pkg/service/mock"
	svctypes "iam/pkg/service/types"
)

func TestGetSubjectByPK(t *testing.T) {
	expiration := 5 * time.Minute

	// valid
	retrieveFunc := func(key cache.Key) (interface{}, error) {
		return svctypes.Subject{}, nil
	}
	mockCache := memory.NewCache(
		"mockCache", false, retrieveFunc, expiration, nil)
	LocalSubjectCache = mockCache

	_, err := GetSubjectByPK(1)
	assert.NoError(t, err)

	// error
	retrieveFunc = func(key cache.Key) (interface{}, error) {
		return false, errors.New("error here")
	}
	mockCache = memory.NewCache(
		"mockCache", false, retrieveFunc, expiration, nil)
	LocalSubjectCache = mockCache

	_, err = GetSubjectByPK(1)
	assert.Error(t, err)
}

func TestBatchGetSubjectByPKsFail(t *testing.T) {
	ctl := gomock.NewController(t)
	mockSvc := mock.NewMockSubjectService(ctl)
	mockSvc.EXPECT().ListByPKs([]int64{1}).Return(nil, errors.New("error here"))

	patches := gomonkey.ApplyFunc(service.NewSubjectService,
		func() service.SubjectService {
			return mockSvc
		})
	defer patches.Reset()

	expiration := 5 * time.Minute

	// valid
	retrieveFunc := func(key cache.Key) (interface{}, error) {
		return svctypes.Subject{}, nil
	}
	mockCache := memory.NewCache(
		"mockCache", false, retrieveFunc, expiration, nil)
	LocalSubjectCache = mockCache

	_, err := BatchGet([]int64{1})
	assert.Error(t, err)
}

func TestBatchGetSubjectByPKsOK(t *testing.T) {
	ctl := gomock.NewController(t)
	mockSvc := mock.NewMockSubjectService(ctl)
	mockSvc.EXPECT().ListByPKs([]int64{2, 3}).Return([]svctypes.Subject{
		{
			PK:   3,
			Type: "department",
			ID:   "department",
			Name: "department",
		},
	}, nil)

	patches := gomonkey.ApplyFunc(service.NewSubjectService,
		func() service.SubjectService {
			return mockSvc
		})
	defer patches.Reset()

	expiration := 5 * time.Minute

	// valid
	retrieveFunc := func(key cache.Key) (interface{}, error) {
		return svctypes.Subject{}, nil
	}
	mockCache := memory.NewCache(
		"mockCache", false, retrieveFunc, expiration, nil)
	LocalSubjectCache = mockCache

	LocalSubjectCache.Set(SubjectPKCacheKey{
		PK: 1,
	}, svctypes.Subject{
		PK:   1,
		Type: "user",
		ID:   "admin",
		Name: "admin",
	})

	subjects, err := BatchGet([]int64{1, 2, 3})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(subjects))
}
