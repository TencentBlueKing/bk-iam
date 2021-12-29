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
	"strings"

	"github.com/TencentBlueKing/gopkg/collection/set"
)

const (
	resourceTypeSupportFields = "id,name,name_en,description,description_en,parents,provider_config,version"
	resourceTypeDefaultFields = "id,name,name_en"
)

type resourceTypeSerializer struct {
	Systems string `form:"systems" binding:"required"`
	queryViaFields
}

func (s *resourceTypeSerializer) validate() (bool, string) {
	valid, message := s.validateFields(resourceTypeSupportFields, resourceTypeDefaultFields)
	if !valid {
		return valid, message
	}

	if s.Systems == "" {
		return false, "systems must not empty"
	}

	return true, ""
}

func (s *resourceTypeSerializer) systems() []string {
	return strings.Split(s.Systems, ",")
}

func (s *resourceTypeSerializer) fieldsSet() *set.StringSet {
	return set.SplitStringToSet(s.Fields, ",")
}
