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
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"

	"iam/pkg/api/common"
)

type instanceSelectionSerializer struct {
	ID        string `json:"id" binding:"required,max=32" example:"biz_set"`
	Name      string `json:"name" binding:"required" example:"biz_set"`
	NameEn    string `json:"name_en" binding:"required" example:"biz_set"`
	IsDynamic bool   `json:"is_dynamic" binding:"omitempty" example:"false"`

	ResourceTypeChain []referenceResourceType `json:"resource_type_chain" structs:"resource_type_chain" binding:"required"`
}

type instanceSelectionUpdateSerializer struct {
	// ID     string `json:"id" binding:"required"`
	Name      string `json:"name" binding:"required" example:"biz_set"`
	NameEn    string `json:"name_en" binding:"required" example:"biz_set"`
	IsDynamic bool   `json:"is_dynamic" binding:"omitempty" example:"false"`

	ResourceTypeChain []referenceResourceType `json:"resource_type_chain" structs:"resource_type_chain" binding:"required"`
}

func (r *instanceSelectionUpdateSerializer) validate(keys map[string]interface{}) (bool, string) {
	if _, ok := keys["name"]; ok {
		if r.Name == "" {
			return false, "name should not be empty"
		}
	}

	if _, ok := keys["name_en"]; ok {
		if r.NameEn == "" {
			return false, "name_en should not be empty"
		}
	}

	if _, ok := keys["resource_type_chain"]; ok {
		if len(r.ResourceTypeChain) == 0 {
			return false, "resource_type_chain should not be empty"
		}
		if valid, message := common.ValidateArray(r.ResourceTypeChain); !valid {
			return false, fmt.Sprintf("%s: %s", "parents invald", message)
		}
	}

	return true, "valid"
}

func validateInstanceSelectionsRepeat(instanceSelections []instanceSelectionSerializer) error {
	idSet := set.NewStringSet()
	nameSet := set.NewStringSet()
	nameEnSet := set.NewStringSet()
	for _, is := range instanceSelections {
		if idSet.Has(is.ID) {
			return fmt.Errorf("instance selection id[%s] repeat", is.ID)
		}
		if nameSet.Has(is.Name) {
			return fmt.Errorf("instance selection name[%s] repeat", is.Name)
		}
		if nameEnSet.Has(is.NameEn) {
			return fmt.Errorf("instance selection name_en[%s] repeat", is.NameEn)
		}

		idSet.Add(is.ID)
		nameSet.Add(is.Name)
		nameEnSet.Add(is.NameEn)
	}
	return nil
}
