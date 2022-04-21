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
	"github.com/TencentBlueKing/gopkg/collection/set"
)

const (
	actionSupportFields = "id,name,name_en,related_resource_types,version,type," +
		"description,description_en,related_actions,related_environments"
	actionDefaultFields = "id,name,name_en"
)

var (
	actionSupportFieldSet = set.SplitStringToSet(actionSupportFields, ",")
)

type actionQuerySerializer struct {
	queryViaFields
}

func (s *actionQuerySerializer) validate() (bool, string) {
	return s.validateFields(actionSupportFields, actionDefaultFields)
}
