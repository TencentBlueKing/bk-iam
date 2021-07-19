/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateArray(t *testing.T) {
	t.Parallel()

	// not a slice
	a := "str"
	valid, message := ValidateArray(a)
	assert.False(t, valid)
	assert.NotEmpty(t, message)

	// a empty slice
	b := []string{}
	valid, message = ValidateArray(b)
	assert.False(t, valid)
	assert.Contains(t, message, "at least 1 item")

	// invalid
	type Test struct {
		Name string `json:"name" binding:"required"`
	}
	c := []Test{
		{""},
	}
	valid, message = ValidateArray(c)
	assert.False(t, valid)
	assert.Contains(t, message, "data in array")

	// valid
	d := []Test{
		{"aaaa"},
	}
	valid, message = ValidateArray(d)
	assert.True(t, valid)
	assert.Equal(t, "valid", message)
}

func TestValidIDRegex(t *testing.T) {
	t.Parallel()

	assert.True(t, ValidIDRegex.MatchString("abc"))
	assert.True(t, ValidIDRegex.MatchString("abc-def"))
	assert.True(t, ValidIDRegex.MatchString("abc_def"))
	assert.True(t, ValidIDRegex.MatchString("abc_"))
	assert.True(t, ValidIDRegex.MatchString("abc-"))
	assert.True(t, ValidIDRegex.MatchString("abc-9"))
	assert.True(t, ValidIDRegex.MatchString("abc9ed"))

	assert.False(t, ValidIDRegex.MatchString("Abc"))
	assert.False(t, ValidIDRegex.MatchString("aBc"))
	assert.False(t, ValidIDRegex.MatchString("abC"))
	assert.False(t, ValidIDRegex.MatchString("_abc"))
	assert.False(t, ValidIDRegex.MatchString("*abc"))
	assert.False(t, ValidIDRegex.MatchString("9abc"))
	assert.False(t, ValidIDRegex.MatchString("abc+"))
	assert.False(t, ValidIDRegex.MatchString("abc*"))
	assert.False(t, ValidIDRegex.MatchString("42"))
}
