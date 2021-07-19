/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package errorx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type NoIsWrapError struct {
	message string
	err     error
}

func (e NoIsWrapError) Error() string {
	return e.message
}

func TestIAMError_Is(t *testing.T) {
	// err vs iamerror
	e1 := errors.New("a")

	e2 := IAMError{
		message: "iam_e2",
		err:     e1,
	}

	assert.False(t, errors.Is(e1, e2))
	assert.True(t, errors.Is(e2, e1))

	// iamerror vs iamerror
	e3 := IAMError{
		message: "iam_e3",
		err:     e2,
	}
	assert.True(t, errors.Is(e3, e1))
	assert.True(t, errors.Is(e3, e2))

	assert.False(t, errors.Is(e2, e3))
	assert.False(t, errors.Is(e2, e3))

	// noIsWrapError vs iamerror
	e4 := NoIsWrapError{
		message: "no_is_wrap",
		err:     e1,
	}
	e5 := IAMError{
		message: "iam_e5",
		err:     e4,
	}

	assert.True(t, errors.Is(e5, e4))
	assert.False(t, errors.Is(e4, e5))
}
