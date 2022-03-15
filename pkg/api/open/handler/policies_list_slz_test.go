/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

func Test_listQuerySerializer_validate(t *testing.T) {
	t.Parallel()

	s := listQuerySerializer{}

	ok, _ := s.validate()
	assert.True(t, ok)

	s.Timestamp = 1
	ok, _ = s.validate()
	assert.False(t, ok)
}

func Test_listQuerySerializer_initDefault(t *testing.T) {
	t.Parallel()

	s := listQuerySerializer{}
	s.initDefault()

	assert.Equal(t, defaultPage, s.Page)
	assert.Equal(t, defaultPageSize, s.PageSize)
	assert.Equal(t, util.TodayStartTimestamp(), s.Timestamp)

	s1 := listQuerySerializer{
		ActionID:  "abc",
		PageSize:  50,
		Page:      2,
		Timestamp: 1,
	}
	s1.initDefault()

	assert.Equal(t, int64(2), s1.Page)
	assert.Equal(t, int64(50), s1.PageSize)
	assert.Equal(t, int64(1), s1.Timestamp)
}
