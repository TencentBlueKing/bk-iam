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

type resourceProviderConfig struct {
	// TODO: valid path?
	Path string `json:"path" structs:"path" binding:"required,uri" example:"/api/v1/resources/biz_set/query"`
}

type resourceTypeSerializer struct {
	ID     string `json:"id"      binding:"required,max=32" example:"biz_set"`
	Name   string `json:"name"    binding:"required"        example:"biz_set"`
	NameEn string `json:"name_en" binding:"required"        example:"biz_set"`

	Description   string `json:"description"    binding:"omitempty"             example:"biz_set is a"`
	DescriptionEn string `json:"description_en" binding:"omitempty"             example:"biz_set is a"`
	Sensitivity   int64  `json:"sensitivity"    binding:"omitempty,gte=0,lte=9" example:"0"`

	// can be empty
	Parents []referenceResourceType `json:"parents"`

	ProviderConfig resourceProviderConfig `json:"provider_config" binding:"required"`

	Version int64 `json:"version" binding:"omitempty,gte=1" example:"1"`
}

type resourceTypeUpdateSerializer struct {
	// ID     string `json:"id" binding:"required"`
	Name          string `json:"name"           binding:"omitempty"             example:"biz_set"`
	NameEn        string `json:"name_en"        binding:"omitempty"             example:"biz_set"`
	Description   string `json:"description"    binding:"omitempty"             example:"biz_set is a"`
	DescriptionEn string `json:"description_en" binding:"omitempty"             example:"biz_set is a"`
	Sensitivity   int64  `json:"sensitivity"    binding:"omitempty,gte=0,lte=9" example:"0"`

	// can be empty
	Parents []referenceResourceType `json:"parents"`

	ProviderConfig *resourceProviderConfig `json:"provider_config" binding:"omitempty"`

	Version int64 `json:"version" binding:"omitempty,gte=1" example:"1"`
}

func (r *resourceTypeUpdateSerializer) validate(keys map[string]interface{}) (bool, string) {
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

	if _, ok := keys["version"]; ok {
		if r.Version < 1 {
			return false, "version should be an integer, greater equals to 1"
		}
	}
	if _, ok := keys["parents"]; ok {
		if len(r.Parents) > 0 {
			if valid, message := common.ValidateArray(r.Parents); !valid {
				return false, fmt.Sprintf("%s: %s", "parents invald", message)
			}
		}
	}
	if _, ok := keys["provider_config"]; ok {
		if r.ProviderConfig.Path == "" {
			return false, "provider_config should contains key: path, and path should not be empty"
		}
	}

	return true, "valid"
}

func validateResourceTypesRepeat(resourceTypes []resourceTypeSerializer) error {
	idSet := set.NewStringSet()
	nameSet := set.NewStringSet()
	nameEnSet := set.NewStringSet()
	for _, rt := range resourceTypes {
		if idSet.Has(rt.ID) {
			return fmt.Errorf("resource type id[%s] repeat", rt.ID)
		}
		if nameSet.Has(rt.Name) {
			return fmt.Errorf("resource type name[%s] repeat", rt.Name)
		}
		if nameEnSet.Has(rt.NameEn) {
			return fmt.Errorf("resource type name_en[%s] repeat", rt.NameEn)
		}

		idSet.Add(rt.ID)
		nameSet.Add(rt.Name)
		nameEnSet.Add(rt.NameEn)
	}
	return nil
}
