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
	"errors"
	"fmt"

	"iam/pkg/api/common"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
)

// 需要db操作的校验, 统一叫 checkXXXX
type AllInstanceSelections struct {
	AllBaseInfo
	InstanceSelections []svctypes.InstanceSelection
}

// NewAllInstanceSelections ...
func NewAllInstanceSelections(instanceSelections []svctypes.InstanceSelection) *AllInstanceSelections {
	idSet := map[string]string{}
	nameSet := map[string]string{}
	nameEnSet := map[string]string{}

	for _, rt := range instanceSelections {
		idSet[rt.ID] = rt.ID
		nameSet[rt.Name] = rt.ID
		nameEnSet[rt.NameEn] = rt.ID
	}

	return &AllInstanceSelections{
		AllBaseInfo: AllBaseInfo{
			IDSet:     idSet,
			NameSet:   nameSet,
			NameEnSet: nameEnSet,
		},
		InstanceSelections: instanceSelections,
	}
}

func checkAllInstanceSelectionsQuotaAndUnique(
	systemID string,
	instanceSelections []instanceSelectionSerializer,
) error {
	svc := service.NewInstanceSelectionService()
	existingInstanceSelections, err := svc.ListBySystem(systemID)
	if err != nil {
		return errors.New("query all instance selection fail")
	}

	allInstanceSelections := NewAllInstanceSelections(existingInstanceSelections)
	for _, rt := range instanceSelections {
		if allInstanceSelections.ContainsID(rt.ID) {
			return fmt.Errorf("instance selection id[%s] already exists", rt.ID)
		}
		if allInstanceSelections.ContainsName(rt.Name) {
			return fmt.Errorf("instance selection name[%s] already exists", rt.Name)
		}
		if allInstanceSelections.ContainsNameEn(rt.NameEn) {
			return fmt.Errorf("instance selection name_en[%s] already exists", rt.NameEn)
		}
	}

	// quota
	if len(existingInstanceSelections)+len(instanceSelections) > common.GetMaxInstanceSelectionsLimit(systemID) {
		return fmt.Errorf("quota error: system %s can only have %d instance selections. [current %d, want to create %d]",
			systemID, common.GetMaxInstanceSelectionsLimit(systemID), len(existingInstanceSelections), len(existingInstanceSelections))
	}

	return nil
}

func checkInstanceSelectionUpdateUnique(systemID string, instanceSelectionID string, name string, nameEn string) error {
	svc := service.NewInstanceSelectionService()
	instanceSelections, err := svc.ListBySystem(systemID)
	if err != nil {
		return errors.New("query all instance selection fail")
	}

	allInstanceSelections := NewAllInstanceSelections(instanceSelections)
	// 先查看 instanceSelection 是否存在
	if !allInstanceSelections.ContainsID(instanceSelectionID) {
		return fmt.Errorf("instance selection id[%s] not exists", instanceSelectionID)
	}
	// check name and name_en is unique
	if name != "" && allInstanceSelections.ContainsNameExcludeSelf(name, instanceSelectionID) {
		return fmt.Errorf("instance selection name(%s) already exists", name)
	}
	if nameEn != "" && allInstanceSelections.ContainsNameEnExcludeSelf(nameEn, instanceSelectionID) {
		return fmt.Errorf("instance selection name_en(%s) already exists", nameEn)
	}
	return nil
}

func checkInstanceSelectionIDsExist(systemID string, ids []string) error {
	svc := service.NewInstanceSelectionService()
	instanceSelections, err := svc.ListBySystem(systemID)
	allInstanceSelections := NewAllInstanceSelections(instanceSelections)

	if err != nil {
		return errors.New("query all instance selection fail")
	}
	for _, id := range ids {
		if !allInstanceSelections.ContainsID(id) {
			return fmt.Errorf("instance selection id[%s] not exists", id)
		}
	}
	return nil
}
