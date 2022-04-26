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

	"iam/pkg/config"
	"iam/pkg/util"
)

const (
	resourceCreatorActionSystemMode = "system"
	// resourceCreatorActionUserMode   = "user"
)

type actionGroupActionSerializer struct {
	ID string `json:"id" binding:"required" example:"edit"`
}

type actionGroupSerializer struct {
	Name      string                        `json:"name" binding:"required" example:"admin"`
	NameEn    string                        `json:"name_en" binding:"required" example:"admin"`
	Actions   []actionGroupActionSerializer `json:"actions,omitempty" binding:"omitempty"`
	SubGroups []actionGroupSerializer       `json:"sub_groups,omitempty" binding:"omitempty"`
}

func getAllFromActionGroupsActionIDs(actionGroups []actionGroupSerializer) []string {
	actions := []string{}

	for _, data := range actionGroups {
		if len(data.Actions) > 0 {
			for _, a := range data.Actions {
				actions = append(actions, a.ID)
			}
		}

		if len(data.SubGroups) > 0 {
			actions = append(actions, getAllFromActionGroupsActionIDs(data.SubGroups)...)
		}
	}

	return actions
}

func validateActionGroup(actionGroups []actionGroupSerializer, name string) (bool, string) {
	for index, data := range actionGroups {
		if err := binding.Validator.ValidateStruct(data); err != nil {
			var message string
			if name != "" {
				message = fmt.Sprintf("data in array[%d] name=%s, %s", index, name, util.ValidationErrorMessage(err))
			} else {
				message = fmt.Sprintf("data in array[%d], %s", index, util.ValidationErrorMessage(err))
			}
			return false, message
		}

		// actions 和 subGroups 不能同时为空, 必须有一个有数据
		if len(data.Actions) == 0 && len(data.SubGroups) == 0 {
			var message string
			if name != "" {
				message = fmt.Sprintf(
					"data in array[%d] name=%s, actions and sub_groups can't be empty at the same time",
					index, name)
			} else {
				message = fmt.Sprintf("data in array[%d], actions and sub_groups can't be empty at the same time", index)
			}
			return false, message
		}

		if len(data.SubGroups) > 0 {
			valid, message := validateActionGroup(data.SubGroups, data.Name)
			if !valid {
				return false, message
			}

			// currently only support 2-levels
			for _, sg := range data.SubGroups {
				if len(sg.SubGroups) > 0 {
					return false, "more than 2-levels action_groups, currently only support 2-levels"
				}
			}
		}
	}

	return true, "valid"
}

type actionRelationActionSerializer struct {
	ID string `json:"id" binding:"required" example:"edit"`
}

type actionRelationSerializer struct {
	ID      string                           `json:"id" binding:"required" example:"edit"`
	Parents []actionRelationActionSerializer `json:"parents" binding:"required"`
}

func getAllActionIDsFromActionRelations(actionRelations []actionRelationSerializer) []string {
	actions := []string{}

	for _, data := range actionRelations {
		actions = append(actions, data.ID)
		if len(data.Parents) > 0 {
			for _, a := range data.Parents {
				actions = append(actions, a.ID)
			}
		}
	}

	return actions
}

func validateActionRelations(systemID string, actionRelations []actionRelationSerializer) (bool, string) {
	// 1. all action_id should exists
	actionIDs := getAllActionIDsFromActionRelations(actionRelations)
	if err := checkActionIDsExist(systemID, actionIDs); err != nil {
		return false, util.ValidationErrorMessage(err)
	}

	// 2. no duplicated
	for _, ar := range actionRelations {
		uniqIDs := set.NewStringSet()
		if len(ar.Parents) == 0 {
			return false, fmt.Sprintf("the parents of action %s should not be empty", ar.ID)
		}

		for _, a := range ar.Parents {
			if a.ID == ar.ID {
				return false, fmt.Sprintf("the parent of action %s should not be himself", ar.ID)
			}

			uniqIDs.Add(a.ID)
		}

		if len(ar.Parents) != uniqIDs.Size() {
			return false, fmt.Sprintf("action %s's parents has duplicate action", ar.ID)
		}
	}

	return true, "valid"
}

type resourceCreatorSingleActionSerializer struct {
	ID       string `json:"id" binding:"required" example:"edit"`
	Required bool   `json:"required" binding:"required" example:"true"`
}

type resourceCreatorActionConfig struct {
	ID               string                                  `json:"id" binding:"required" example:"host"`
	Actions          []resourceCreatorSingleActionSerializer `json:"actions" binding:"required,gt=0"`
	SubResourceTypes []resourceCreatorActionConfig           `json:"sub_resource_types,omitempty" binding:"omitempty"`
}

func (r *resourceCreatorActionConfig) getAllActionIDResourceTypeID() []ActionIDResourceTypeID {
	actionIDResourceTypeIDs := []ActionIDResourceTypeID{}

	// 当前层级
	for _, action := range r.Actions {
		actionIDResourceTypeIDs = append(actionIDResourceTypeIDs, ActionIDResourceTypeID{
			ActionID:       action.ID,
			ResourceTypeID: r.ID,
		})
	}

	// 子资源类型的
	for _, srt := range r.SubResourceTypes {
		artus := srt.getAllActionIDResourceTypeID()
		actionIDResourceTypeIDs = append(actionIDResourceTypeIDs, artus...)
	}

	return actionIDResourceTypeIDs
}

func (r *resourceCreatorActionConfig) validate() error {
	// actions 不能为空数据等
	if err := binding.Validator.ValidateStruct(r); err != nil {
		return err
	}
	// 校验递归数据
	for _, rcac := range r.SubResourceTypes {
		if err := rcac.validate(); err != nil {
			return err
		}
	}
	return nil
}

type resourceCreatorActionSerializer struct {
	// 选择支持的方式：接入系统 和 用户，对于用户，则需要在授权接口上额外传入祖先creator信息
	Mode   string                        `json:"mode,omitempty" binding:"omitempty,oneof=system user" example:"system"`
	Config []resourceCreatorActionConfig `json:"config" binding:"required"`
}

func (r *resourceCreatorActionSerializer) getAllActionIDResourceTypeIDFromConfig() []ActionIDResourceTypeID {
	actionIDResourceTypeIDs := []ActionIDResourceTypeID{}

	for _, rca := range r.Config {
		artus := rca.getAllActionIDResourceTypeID()
		actionIDResourceTypeIDs = append(actionIDResourceTypeIDs, artus...)
	}
	return actionIDResourceTypeIDs
}

func (r *resourceCreatorActionSerializer) validate() error {
	// 校验配置是否OK
	for _, rcac := range r.Config {
		if err := rcac.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r *resourceCreatorActionSerializer) setDefaultValue() {
	// 模式目前仅仅支持接入系统默认模式
	if r.Mode == "" {
		r.Mode = resourceCreatorActionSystemMode
	}
}

func (r *resourceCreatorActionSerializer) toMapInterface() map[string]interface{} {
	// 转换为map
	return map[string]interface{}{
		"mode":   r.Mode,
		"config": r.Config,
	}
}

type actionIDSerializer struct {
	ID string `json:"id" binding:"required"`
}

type commonActionSerializer struct {
	Name    string               `json:"name" binding:"required" example:"admin"`
	NameEn  string               `json:"name_en" binding:"required" example:"admin"`
	Actions []actionIDSerializer `json:"actions" binding:"required,gt=1"`
}

func getAllFromCommonActions(commonActions []commonActionSerializer) []string {
	actions := []string{}

	for _, data := range commonActions {
		for _, a := range data.Actions {
			actions = append(actions, a.ID)
		}
	}

	return actions
}

type featureShieldRuleSerializer struct {
	Effect  string             `json:"effect" binding:"required,oneof=deny allow" example:"deny"`
	Feature string             `json:"feature" binding:"required" example:"application.custom_permission"`
	Action  actionIDSerializer `json:"action" binding:"required"`
}

func (f *featureShieldRuleSerializer) validate() error {
	if err := binding.Validator.ValidateStruct(f); err != nil {
		return err
	}
	// 检查是否是支持屏蔽的功能
	if !config.SupportShieldFeaturesSet.Has(f.Feature) {
		return fmt.Errorf("feature[%s] not support shield", f.Feature)
	}
	return nil
}

func validateFeatureShieldRules(featureShieldRules []featureShieldRuleSerializer) (bool, string) {
	if len(featureShieldRules) == 0 {
		return false, "the array should contain at least 1 item"
	}
	for _, fsr := range featureShieldRules {
		if err := fsr.validate(); err != nil {
			return false, err.Error()
		}
	}
	return true, "valid"
}
