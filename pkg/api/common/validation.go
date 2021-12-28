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
	"errors"
	"fmt"
	"regexp"

	"github.com/TencentBlueKing/gopkg/conv"
	"github.com/gin-gonic/gin/binding"

	"iam/pkg/util"
)

// Some Tag Validators: https://godoc.org/gopkg.in/go-playground/validator.v9#hdr-Baked_In_Validators_and_Tags

// ValidateArray ...
func ValidateArray(data interface{}) (bool, string) {
	array, err := conv.ToSlice(data)
	if err != nil {
		return false, err.Error()
	}

	if len(array) == 0 {
		return false, "the array should contain at least 1 item"
	}
	for index, item := range array {
		if err := binding.Validator.ValidateStruct(item); err != nil {
			message := fmt.Sprintf("data in array[%d], %s", index, util.ValidationErrorMessage(err))
			return false, message
		}
	}
	return true, "valid"
}

const (
	// 小写字母开头, 可以包含小写字母/数字/下划线/连字符
	validIDString = "^[a-z]+[a-z0-9_-]*$"
)

// ValidIDRegex ...
var (
	ValidIDRegex = regexp.MustCompile(validIDString)

	ErrInvalidID = errors.New("invalid id: id should begin with a lowercase letter, " +
		"contains lowercase letters(a-z), numbers(0-9), underline(_) or hyphen(-)")
)
