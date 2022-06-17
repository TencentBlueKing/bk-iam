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
	"database/sql"
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/service/types"
)

type MiniResourceType struct {
	System string
	ID     string
}

func (r *MiniResourceType) UniqueKey() string {
	return fmt.Sprintf("%s:%s", r.System, r.ID)
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

	// 解析实例视图，获取到每个relation资源对应的资源实例视图涉及的资源
	miniResourceTypeOfInstancesSelections := make([][]MiniResourceType, 0, len(arts))
	allResourceTypes := []MiniResourceType{}
	for _, art := range arts {
		// 通过SaaS Action解析出实例视图里的所有资源类型
		rts, err := l.parseInvolvedResourceTypes(art.RelatedInstanceSelections)
		if err != nil {
			return nil, errorWrapf(err, "parseInvolvedResourceTypes rawString=`%+v` fail",
				art.RelatedInstanceSelections)
		}
		miniResourceTypeOfInstancesSelections = append(miniResourceTypeOfInstancesSelections, rts)
		allResourceTypes = append(allResourceTypes, rts...)
		allResourceTypes = append(allResourceTypes, MiniResourceType{
			System: art.ResourceTypeSystem,
			ID:     art.ResourceTypeID,
		})
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

		// 组装实例视图涉及的所有资源类型
		chain := miniResourceTypeOfInstancesSelections[idx]
		thinResourceTypes := make([]types.ThinResourceType, 0, len(chain))
		for _, rt := range chain {
			pk, ok := resourceTypePKMap[rt.UniqueKey()]
			if !ok {
				return nil, errorWrapf(
					fmt.Errorf("pk of resource type in chain not found, system=`%s`, id=`%s`", rt.System, rt.ID),
					"",
				)
			}
			thinResourceTypes = append(thinResourceTypes, types.ThinResourceType{PK: pk, System: rt.System, ID: rt.ID})
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
func (l *actionService) queryResourceTypePK(allResourceTypes []MiniResourceType) (map[string]int64, error) {
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

// parseInvolvedResourceTypes 用于解析出Action关联实例视图涉及的资源类型
func (l *actionService) parseInvolvedResourceTypes(rawRelatedInstanceSelections string) (
	resourceTypes []MiniResourceType, err error,
) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "parseInvolvedResourceTypes")
	// rawRelatedInstanceSelections is [{"system_id": a, "id": b}]
	if rawRelatedInstanceSelections == "" {
		return
	}

	relatedInstanceSelections := []types.ReferenceInstanceSelection{}
	err = jsoniter.UnmarshalFromString(rawRelatedInstanceSelections, &relatedInstanceSelections)
	if err != nil {
		err = errorWrapf(err, "unmarshal rawRelatedInstanceSelections=`%s` fail", rawRelatedInstanceSelections)
		return
	}

	// NOTE: here query one by one
	for _, r := range relatedInstanceSelections {
		is, err1 := l.saasInstanceSelectionManager.Get(r.System, r.ID)
		if err1 != nil {
			// NOTE: if not exists, continue
			if err1 == sql.ErrNoRows {
				continue
			}

			err = errorWrapf(err1, "saasInstanceSelectionManager.Get system=`%s`, id=`%s` fail", r.System, r.ID)
			return
		}

		chain := []map[string]string{}
		err = jsoniter.UnmarshalFromString(is.ResourceTypeChain, &chain)
		if err != nil {
			err = errorWrapf(err, "unmarshal instanceSelection.ResourceTypeChain=`%s` fail", is.ResourceTypeChain)
			return
		}

		for _, c := range chain {
			resourceTypes = append(resourceTypes, MiniResourceType{System: c["system_id"], ID: c["id"]})
		}
	}

	// 去重
	deduplicatedResourceTypes := make([]MiniResourceType, 0, len(resourceTypes))
	rtSet := set.NewFixedLengthStringSet(len(resourceTypes))
	for _, rt := range resourceTypes {
		k := rt.UniqueKey()
		if rtSet.Has(k) {
			continue
		}

		deduplicatedResourceTypes = append(deduplicatedResourceTypes, rt)
		rtSet.Add(k)
	}

	return deduplicatedResourceTypes, nil
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
