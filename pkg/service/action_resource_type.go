/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/service/types"
)

type miniResourceType struct {
	System string
	ID     string
}

func (r *miniResourceType) UniqueKey() string {
	return r.System + ":" + r.ID
}

// ListThinActionResourceTypes 获取操作关联的资源类型
func (l *actionService) ListThinActionResourceTypes(
	system, actionID string,
) (actionResourceTypes []types.ThinActionResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListThinActionResourceTypes")

	// 实例视图相关信息只有SaaS表有，所以直接查询SaaS的即可
	arts, err := l.saasActionResourceTypeManager.ListByActionID(system, actionID)
	if err != nil {
		return nil, errorWrapf(err, "ListByActionID(system=%s, actionID=%s) fail", system, actionID)
	}

	if len(arts) == 0 {
		return
	}

	// 解析实例视图，用于后面查询实例视图的ResourceTypeChain
	relatedInstanceSelectionsList := make([][]types.ReferenceInstanceSelection, 0, len(arts))
	allInstanceSelections := []types.ReferenceInstanceSelection{}
	allResourceTypes := []miniResourceType{}
	for _, art := range arts {
		ris, err := l.parseRelatedInstanceSelections(art.RelatedInstanceSelections)
		if err != nil {
			return nil, errorWrapf(err, "parseRelatedInstanceSelections rawString=`%+v` fail",
				art.RelatedInstanceSelections)
		}

		// relatedInstanceSelectionsList是为了记录顺序，后续便于填充
		relatedInstanceSelectionsList = append(relatedInstanceSelectionsList, ris)

		// allInstanceSelections是为了批量查询实例视图的ResourceTypeChain
		allInstanceSelections = append(allInstanceSelections, ris...)

		// allResourceTypes 是为了后续查询resource type pk
		allResourceTypes = append(allResourceTypes, miniResourceType{
			System: art.ResourceTypeSystem,
			ID:     art.ResourceTypeID,
		})
	}

	// 查询实例视图的ResourceTypeChain
	allResourceTypeChains, err := l.queryResourceTypeChain(allInstanceSelections)
	if err != nil {
		return nil, errorWrapf(err, "queryResourceTypeChain allInstanceSelections=`%+v` fail",
			allInstanceSelections)
	}

	// 获取到每个relation资源对应的资源实例视图涉及的资源
	for _, resourceTypeChain := range allResourceTypeChains {
		allResourceTypes = append(allResourceTypes, resourceTypeChain...)
	}

	// 查询所有ResourceType的PK
	resourceTypePKMap, err := l.queryResourceTypePK(allResourceTypes)
	if err != nil {
		return nil, errorWrapf(err, "queryResourceTypePK rts=`%+v` fail", allResourceTypes)
	}

	// 组装数据
	actionResourceTypes = make([]types.ThinActionResourceType, 0, len(arts))
	for idx, art := range arts {
		// Action关联的资源类型的PK
		actionResourceTypePK, ok := resourceTypePKMap[fmt.Sprintf("%s:%s", art.ResourceTypeSystem, art.ResourceTypeID)]
		if !ok {
			return nil, errorWrapf(
				fmt.Errorf(
					"pk of action related resource type not found, system=`%s`, id=`%s`",
					art.ResourceTypeSystem, art.ResourceTypeID,
				),
				"",
			)
		}

		// 关联的实例视图
		relatedInstanceSelections := relatedInstanceSelectionsList[idx]
		// 遍历每个实例视图，获取其资源Chain，并获取Chain里的资源类型
		thinResourceTypes := []types.ThinResourceType{}
		resourceTypePKSet := set.NewInt64Set()
		for _, is := range relatedInstanceSelections {
			resourceTypeChain := allResourceTypeChains[is.System+":"+is.ID]
			// 遍历每个资源类型
			for _, rt := range resourceTypeChain {
				pk, ok := resourceTypePKMap[rt.UniqueKey()]
				if !ok {
					return nil, errorWrapf(
						fmt.Errorf("pk of resource type in chain not found, system=`%s`, id=`%s`", rt.System, rt.ID),
						"",
					)
				}
				// 这里只是为了获取涉及的资源类型，所以不需要重复
				if resourceTypePKSet.Has(pk) {
					continue
				}
				resourceTypePKSet.Add(pk)
				thinResourceTypes = append(
					thinResourceTypes,
					types.ThinResourceType{PK: pk, System: rt.System, ID: rt.ID},
				)
			}
		}

		actionResourceTypes = append(actionResourceTypes, types.ThinActionResourceType{
			PK:                               actionResourceTypePK,
			System:                           art.ResourceTypeSystem,
			ID:                               art.ResourceTypeID,
			ResourceTypeOfInstanceSelections: thinResourceTypes,
		})
	}

	return actionResourceTypes, nil
}

// queryResourceTypePK : 后续用于填充[]types.ThinActionResourceType数据结构里，所有涉及资源类型的PK
func (l *actionService) queryResourceTypePK(allResourceTypes []miniResourceType) (map[string]int64, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "queryResourceTypePK")

	// 需要按照系统分组资源类型
	systemToResourceTypeIDs := map[string]*set.StringSet{}
	for _, rt := range allResourceTypes {
		if _, ok := systemToResourceTypeIDs[rt.System]; !ok {
			systemToResourceTypeIDs[rt.System] = set.NewStringSet()
		}
		systemToResourceTypeIDs[rt.System].Add(rt.ID)
	}

	resourceTypePKMap := make(map[string]int64, len(allResourceTypes))
	// 按照系统查询所有资源类型PK
	for systemID, resourceTypeIDSet := range systemToResourceTypeIDs {
		rts, err := l.resourceTypeManager.ListByIDs(systemID, resourceTypeIDSet.ToSlice())
		if err != nil {
			return resourceTypePKMap, errorWrapf(
				err,
				"resourceTypeManager.ListByIDs system=`%s`, ids=`%v` fail", systemID, resourceTypeIDSet,
			)
		}

		for _, rt := range rts {
			k := fmt.Sprintf("%s:%s", rt.System, rt.ID)
			resourceTypePKMap[k] = rt.PK
		}
	}

	return resourceTypePKMap, nil
}

// parseRelatedInstanceSelections, 解析出实例视图（并不包含resource_type_chain）
func (l *actionService) parseRelatedInstanceSelections(rawRelatedInstanceSelections string) (
	relatedInstanceSelections []types.ReferenceInstanceSelection, err error,
) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "parseRelatedInstanceSelections")
	// rawRelatedInstanceSelections is [{"system_id": a, "id": b}]
	if rawRelatedInstanceSelections == "" {
		return
	}

	err = jsoniter.UnmarshalFromString(rawRelatedInstanceSelections, &relatedInstanceSelections)
	if err != nil {
		err = errorWrapf(err, "unmarshal rawRelatedInstanceSelections=`%s` fail", rawRelatedInstanceSelections)
		return
	}

	return
}

func (l *actionService) queryResourceTypeChain(ris []types.ReferenceInstanceSelection) (
	resourceTypeChains map[string][]miniResourceType, err error,
) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "queryResourceTypeChain")
	// 按系统分组查询
	systemIDSet := set.NewStringSet()
	for _, is := range ris {
		systemIDSet.Add(is.System)
	}
	systemIDs := systemIDSet.ToSlice()

	// 按系统查询，并记录每个实例视图对应的resourceTypeChain
	instanceSelectionToResourceTypeChainMap := map[string]string{}
	for _, systemID := range systemIDs {
		instanceSelections, err := l.saasInstanceSelectionManager.ListBySystem(systemID)
		if err != nil {
			return resourceTypeChains, errorWrapf(
				err, "saasInstanceSelectionManager.ListBySystem systemID=`%s` fail", systemID,
			)
		}

		// Note: 这里并不会将resourceTypeChain的json string解析为 struct，而是待确定是查询所需的再进行解析
		for _, is := range instanceSelections {
			key := is.System + ":" + is.ID
			instanceSelectionToResourceTypeChainMap[key] = is.ResourceTypeChain
		}
	}

	// 遍历请求的每个实例视图，组装对应的ResourceTypeChain数据
	resourceTypeChains = make(map[string][]miniResourceType, len(ris))
	for _, is := range ris {
		key := is.System + ":" + is.ID
		rawResourceTypeChain, ok := instanceSelectionToResourceTypeChainMap[key]
		if !ok {
			return resourceTypeChains, errorWrapf(
				err, "instanceSelection not exists, systemID=`%s` id=`%s`", is.System, is.ID,
			)
		}

		chain := []map[string]string{}
		err = jsoniter.UnmarshalFromString(rawResourceTypeChain, &chain)
		if err != nil {
			err = errorWrapf(err, "unmarshal instanceSelection.ResourceTypeChain=`%s` fail", rawResourceTypeChain)
			return
		}

		resourceTypeChain := make([]miniResourceType, 0, len(chain))
		for _, c := range chain {
			resourceTypeChain = append(resourceTypeChain, miniResourceType{System: c["system_id"], ID: c["id"]})
		}

		resourceTypeChains[key] = resourceTypeChain
	}

	return resourceTypeChains, nil
}

// ListActionResourceTypeIDByResourceTypeSystem ...
func (l *actionService) ListActionResourceTypeIDByResourceTypeSystem(resourceTypeSystem string) (
	actionResourceTypeIDs []types.ActionResourceTypeID, err error,
) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListActionResourceTypeIDByResourceTypeSystem")

	actionResourceTypes, err := l.actionResourceTypeManager.ListByResourceTypeSystem(resourceTypeSystem)
	if err != nil {
		err = errorWrapf(err, "actionResourceTypeManager.ListByResourceTypeSystem resourceTypeSystem=`%s` fail",
			resourceTypeSystem)
		return actionResourceTypeIDs, err
	}
	for _, art := range actionResourceTypes {
		actionResourceTypeIDs = append(actionResourceTypeIDs, types.ActionResourceTypeID{
			ActionSystem:       art.ActionSystem,
			ActionID:           art.ActionID,
			ResourceTypeSystem: art.ResourceTypeSystem,
			ResourceTypeID:     art.ResourceTypeID,
		})
	}
	return
}

// ListActionResourceTypeIDByActionSystem ...
func (l *actionService) ListActionResourceTypeIDByActionSystem(actionSystem string) (
	actionResourceTypeIDs []types.ActionResourceTypeID, err error,
) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListActionResourceTypeIDByActionSystem")

	actionResourceTypes, err := l.actionResourceTypeManager.ListByActionSystem(actionSystem)
	if err != nil {
		err = errorWrapf(err, "actionResourceTypeManager.ListByActionSystem actionSystem=`%s` fail",
			actionSystem)
		return actionResourceTypeIDs, err
	}
	for _, art := range actionResourceTypes {
		actionResourceTypeIDs = append(actionResourceTypeIDs, types.ActionResourceTypeID{
			ActionSystem:       art.ActionSystem,
			ActionID:           art.ActionID,
			ResourceTypeSystem: art.ResourceTypeSystem,
			ResourceTypeID:     art.ResourceTypeID,
		})
	}
	return
}
