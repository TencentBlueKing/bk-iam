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
type AllResourceTypes struct {
	AllBaseInfo
	ResourceTypes []svctypes.ResourceType
}

// NewAllResourceTypes ...
func NewAllResourceTypes(resourceTypes []svctypes.ResourceType) *AllResourceTypes {
	idSet := map[string]string{}
	nameSet := map[string]string{}
	nameEnSet := map[string]string{}

	for _, rt := range resourceTypes {
		idSet[rt.ID] = rt.ID
		nameSet[rt.Name] = rt.ID
		nameEnSet[rt.NameEn] = rt.ID
	}

	return &AllResourceTypes{
		AllBaseInfo: AllBaseInfo{
			IDSet:     idSet,
			NameSet:   nameSet,
			NameEnSet: nameEnSet,
		},
		ResourceTypes: resourceTypes,
	}
}

func checkAllResourceTypesQuotaAndUnique(systemID string, inResourceTypes []resourceTypeSerializer) error {
	svc := service.NewResourceTypeService()
	resourceTypes, err := svc.ListBySystem(systemID)
	if err != nil {
		return errors.New("query all resource type fail")
	}

	allResourceTypes := NewAllResourceTypes(resourceTypes)
	for _, rt := range inResourceTypes {
		if allResourceTypes.ContainsID(rt.ID) {
			return fmt.Errorf("resource type id[%s] already exists", rt.ID)
		}
		if allResourceTypes.ContainsName(rt.Name) {
			return fmt.Errorf("resource type name[%s] already exists", rt.Name)
		}
		if allResourceTypes.ContainsNameEn(rt.NameEn) {
			return fmt.Errorf("resource type name_en[%s] already exists", rt.NameEn)
		}
	}

	// quota
	if len(resourceTypes)+len(inResourceTypes) > common.GetMaxResourceTypesLimit(systemID) {
		return fmt.Errorf("quota error: system %s can only have %d resource types.[current %d, want to create %d]",
			systemID, common.GetMaxResourceTypesLimit(systemID), len(resourceTypes), len(inResourceTypes))
	}

	return nil
}

func checkResourceTypeUpdateUnique(systemID string, resourceTypeID string, name string, nameEn string) error {
	svc := service.NewResourceTypeService()
	resourceTypes, err := svc.ListBySystem(systemID)
	if err != nil {
		return errors.New("query all resource type fail")
	}

	allResourceTypes := NewAllResourceTypes(resourceTypes)
	// 先查看resourceType是否存在
	if !allResourceTypes.ContainsID(resourceTypeID) {
		return fmt.Errorf("resource type id[%s] not exists", resourceTypeID)
	}
	// check name and name_en is unique
	if name != "" && allResourceTypes.ContainsNameExcludeSelf(name, resourceTypeID) {
		return fmt.Errorf("resource type name(%s) already exists", name)
	}
	if nameEn != "" && allResourceTypes.ContainsNameEnExcludeSelf(nameEn, resourceTypeID) {
		return fmt.Errorf("resource type name_en(%s) already exists", nameEn)
	}
	return nil
}

func checkResourceTypeIDsExist(systemID string, ids []string) error {
	svc := service.NewResourceTypeService()
	resourceTypes, err := svc.ListBySystem(systemID)
	allResourceTypes := NewAllResourceTypes(resourceTypes)

	if err != nil {
		return errors.New("query all resource type fail")
	}
	for _, id := range ids {
		if !allResourceTypes.ContainsID(id) {
			return fmt.Errorf("resource type id[%s] not exists", id)
		}
	}
	return nil
}
