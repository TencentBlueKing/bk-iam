/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-gopkg available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package errorx

import (
	"errors"
)

// Errorx is a struct for wrap raw err with message
type Errorx struct {
	message string
	err     error
}

// Error return the error message
func (e Errorx) Error() string {
	return e.message
}

// Is reports whether any error in err's chain matches target.
func (e Errorx) Is(target error) bool {
	if target == nil || e.err == nil {
		return e.err == target
	}

	return errors.Is(e.err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func (e *Errorx) Unwrap() error {
	u, ok := e.err.(interface {
		Unwrap() error
	})
	if !ok {
		return e.err
	}

	return u.Unwrap()
}
