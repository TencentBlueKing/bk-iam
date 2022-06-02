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
	"github.com/gin-gonic/gin/binding"

	"iam/pkg/api/common"
	"iam/pkg/util"
)

type relatedResourceType struct {
	SystemID string `json:"system_id" binding:"required" example:"bk_cmdb"`
	ID       string `json:"id" binding:"required,max=32" example:"host"`

	NameAlias   string `json:"name_alias" example:""`
	NameAliasEn string `json:"name_alias_en" example:""`

	// 实例选择方式/范围: ["all", "instance", "attribute"]
	SelectionMode string `json:"selection_mode" binding:"omitempty,oneof=all instance attribute" example:"instance"`

	RelatedInstanceSelections []referenceInstanceSelection `json:"related_instance_selections" binding:"omitempty"`
}

// relatedEnvironment, currently only support `current_timestamp`.
// if we support more types, should add a `validate` method, each type has different operators.
type relatedEnvironment struct {
	// NOTE: currently only support period_daily, will support current_timestamp later
	//       and no operators now!
	//       only one field, but should be a struct! keep extensible in the future
	Type string `json:"type" binding:"oneof=period_daily" example:"period_daily"`
	// Operators []string `json:"operators" binding:"omitempty,unique"`
}

type actionSerializer struct {
	ID     string `json:"id" binding:"required,max=32" example:"biz_create"`
	Name   string `json:"name" binding:"required" example:"biz_create"`
	NameEn string `json:"name_en" binding:"required" example:"biz_create"`

	Description   string `json:"description" binding:"omitempty" example:"biz_create is"`
	DescriptionEn string `json:"description_en" binding:"omitempty" example:"biz_create is"`
	Sensitivity   int64  `json:"sensitivity" binding:"omitempty,gte=0,lte=9" example:"0"`

	Type string `json:"type" binding:"omitempty,oneof=create edit view delete list manage execute debug use"`

	RelatedResourceTypes []relatedResourceType `json:"related_resource_types"`
	RelatedActions       []string              `json:"related_actions"`
	RelatedEnvironments  []relatedEnvironment  `json:"related_environments" binding:"omitempty"`

	Version int64 `json:"version" binding:"omitempty,gte=1" example:"1"`
}

type actionUpdateSerializer struct {
	Name          string `json:"name" example:"biz_create"`
	NameEn        string `json:"name_en" example:"biz_create"`
	Description   string `json:"description" binding:"omitempty" example:"biz_create is"`
	DescriptionEn string `json:"description_en" binding:"omitempty" example:"biz_create is"`
	Sensitivity   int64  `json:"sensitivity" binding:"omitempty,gte=0,lte=9" example:"0"`

	Type string `json:"type" binding:"omitempty,oneof=create edit view delete list manage execute debug use"`

	RelatedResourceTypes []relatedResourceType `json:"related_resource_types"`
	RelatedActions       []string              `json:"related_actions"`
	RelatedEnvironments  []relatedEnvironment  `json:"related_environments" binding:"omitempty"`

	Version int64 `json:"version" binding:"omitempty,gte=1" example:"1"`
}

func (a *actionUpdateSerializer) validate(keys map[string]interface{}) (bool, string) {
	if _, ok := keys["name"]; ok {
		if a.Name == "" {
			return false, "name should not be empty"
		}
	}

	if _, ok := keys["name_en"]; ok {
		if a.NameEn == "" {
			return false, "name_en should not be empty"
		}
	}

	// NOTE: type can be set to ""
	// if _, ok := keys["type"]; ok {
	// 	if a.Type == "" {
	// 		return false, "type should not be empty"
	// 	}
	// }

	if _, ok := keys["version"]; ok {
		if a.Version < 1 {
			return false, "version should be an integer, greater equals to 1"
		}
	}

	if len(a.RelatedResourceTypes) > 0 {
		// NOTE: here, has no actionID in validate()
		valid, message := validateRelatedResourceTypes(a.RelatedResourceTypes, "")
		if !valid {
			return false, message
		}
	}

	if len(a.RelatedEnvironments) > 0 {
		valid, message := validateRelatedEnvironments(a.RelatedEnvironments, "")
		if !valid {
			return false, message
		}
	}

	return true, "valid"
}

func validateRelatedInstanceSelections(
	data []referenceInstanceSelection, actionID string,
	relatedResourceTypeID string,
) (bool, string) {
	for index, data := range data {
		if err := binding.Validator.ValidateStruct(data); err != nil {
			message := fmt.Sprintf("data of action_id=%s releated_resource_type[%s] instance_selections[%d], %s",
				actionID, relatedResourceTypeID, index, util.ValidationErrorMessage(err))
			return false, message
		}
	}
	return true, "valid"
}

func validateRelatedEnvironments(data []relatedEnvironment, actionID string) (bool, string) {
	typeID := set.NewStringSet()
	for index, d := range data {
		if err := binding.Validator.ValidateStruct(d); err != nil {
			message := fmt.Sprintf("data of action_id=%s related_environments[%d], %s",
				actionID, index, util.ValidationErrorMessage(err))
			return false, message
		}

		// 校验 data.ID 没有重复
		if typeID.Has(d.Type) {
			message := fmt.Sprintf("data of action_id=%s related_environments[%d] id should not repeat",
				actionID, index)
			return false, message
		}

		typeID.Add(d.Type)
	}
	return true, "valid"
}

func validateRelatedResourceTypes(data []relatedResourceType, actionID string) (bool, string) {
	resourceTypeID := set.NewStringSet()
	for index, d := range data {
		if err := binding.Validator.ValidateStruct(d); err != nil {
			message := fmt.Sprintf("data of action_id=%s related_resource_types[%d], %s",
				actionID, index, util.ValidationErrorMessage(err))
			return false, message
		}

		// 校验 data.ID 没有重复
		if resourceTypeID.Has(d.ID) {
			message := fmt.Sprintf("data of action_id=%s related_resource_types[%d] id"+
				" should not repeat",
				actionID, index)
			return false, message
		}

		resourceTypeID.Add(d.ID)

		relatedResourceTypeID := fmt.Sprintf("system_id=%s,id=%s", d.SystemID, d.ID)

		// selection_mode = attribute的时候, related_instance_selections 可以为空,
		// 其他情况: instance OR all, 不能为空
		if d.SelectionMode != SelectionModeAttribute {
			if len(d.RelatedInstanceSelections) > 0 {
				// validate if not empty
				valid, message := validateRelatedInstanceSelections(
					d.RelatedInstanceSelections,
					actionID,
					relatedResourceTypeID,
				)
				if !valid {
					return false, message
				}
			} else {
				// instance OR all: should contain at least 1 item
				message := fmt.Sprintf("data of action_id=%s related_resource_types[%d] instance_selections"+
					" should contain at least 1 item",
					actionID, index)
				return false, message
			}
		}
	}
	return true, "valid"
}

func validateAction(body []actionSerializer) (bool, string) {
	for index, data := range body {
		if err := binding.Validator.ValidateStruct(data); err != nil {
			message := fmt.Sprintf("data in array[%d], %s", index, util.ValidationErrorMessage(err))
			return false, message
		}

		if !common.ValidIDRegex.MatchString(data.ID) {
			message := fmt.Sprintf("data in array[%d] id=%s, %s", index, data.ID, common.ErrInvalidID)
			return false, message
		}

		// related_resource_types, validate if not empty
		if len(data.RelatedResourceTypes) > 0 {
			valid, message := validateRelatedResourceTypes(data.RelatedResourceTypes, data.ID)
			if !valid {
				return false, message
			}
		}

		if len(data.RelatedEnvironments) > 0 {
			valid, message := validateRelatedEnvironments(data.RelatedEnvironments, data.ID)
			if !valid {
				return false, message
			}
		}
	}
	return true, "valid"
}

func validateActionsRepeat(actions []actionSerializer) error {
	idSet := set.NewStringSet()
	nameSet := set.NewStringSet()
	nameEnSet := set.NewStringSet()
	for _, ac := range actions {
		if idSet.Has(ac.ID) {
			return fmt.Errorf("action id[%s] repeat", ac.ID)
		}
		if nameSet.Has(ac.Name) {
			return fmt.Errorf("action name[%s] repeat", ac.Name)
		}
		if nameEnSet.Has(ac.NameEn) {
			return fmt.Errorf("action name_en[%s] repeat", ac.NameEn)
		}

		idSet.Add(ac.ID)
		nameSet.Add(ac.Name)
		nameEnSet.Add(ac.NameEn)
	}
	return nil
}
