/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pdp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	"iam/pkg/logging/debug"
	"iam/pkg/service"
)

func rbacEval(
	system string,
	action types.Action,
	resources []types.Resource,
	effectGroupPKs []int64,
	withoutCache bool,
	parentEntry *debug.Entry,
) (isPass bool, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDP, "rbacEval")

	entry := debug.NewSubDebug(parentEntry)
	if entry != nil {
		debug.WithValue(entry, "cacheEnabled", !withoutCache)
	}

	// 1. 拿action关联的类型
	debug.AddStep(entry, "Get Action ResourceTypes")
	actionResourceTypes, err := action.Attribute.GetResourceTypes()
	if err != nil {
		err = errorWrapf(err, "action.Attribute.GetResourceTypes fail, action=`%+v`", action)
		return
	}
	debug.WithValue(entry, "actionResourceTypes", actionResourceTypes)

	if len(resources) != 1 || len(actionResourceTypes) != 1 {
		// NOTE: 能做RBAC鉴权的操作的资源类型只能有一个
		err = errorWrapf(errors.New("rbacEval only support action with one resource type"), "")
		return
	}

	// 2. 解析资源实例的以及属性, 返回出可能被授权的资源实例节点
	debug.AddStep(entry, "Parse Resource Nodes")
	// 从resources中解析出用于rbac鉴权的资源实例节点
	resourceNodes, err := parseResourceNode(resources[0], actionResourceTypes[0])
	if err != nil {
		err = errorWrapf(
			err,
			"parseResourceNode fail, resource=`%+v` actionResourceType=`%+v`",
			resources[0],
			actionResourceTypes[0],
		)
		return
	}
	debug.WithValue(entry, "resourceNodes", resourceNodes)

	// 3. 拿action的PK
	debug.AddStep(entry, "Get Action PK")
	actionPK, err := action.Attribute.GetPK()
	if err != nil {
		err = errorWrapf(err, "action.Attribute.GetPK fail, action=`%+v`", action)
		return
	}
	debug.WithValue(entry, "actionPK", actionPK)
	actionResourceTypePK := actionResourceTypes[0].PK

	// 4. 遍历查询资源实例节点授权的groupPKs
	effectGroupPKSet := set.NewInt64SetWithValues(effectGroupPKs)
	for _, resourceNode := range resourceNodes {
		debug.AddStep(entry, "Get Resource ActionAuthorized Group PKs")
		var groupPKs []int64
		if withoutCache {
			groupPKs, err = cacheimpls.GetResourceActionAuthorizedGroupPKs(
				system,
				actionPK,
				actionResourceTypePK,
				resourceNode.TypePK,
				resourceNode.ID,
			)
		} else {
			svc := service.NewGroupResourcePolicyService()
			var actionGroupPKs map[int64][]int64
			actionGroupPKs, err = svc.GetAuthorizedActionGroupMap(
				system, actionResourceTypePK, resourceNode.TypePK, resourceNode.ID,
			)
			if err == nil {
				groupPKs = actionGroupPKs[actionPK]
			}
		}
		if err != nil {
			err = errorWrapf(
				err,
				"GetResourceActionAuthorizedGroupPKs fail, system=`%s` action=`%+v` resource=`%+v`",
				system,
				action,
				resourceNode,
			)
			return
		}
		debug.WithValue(entry, "resourceNode", resourceNode)
		debug.WithValue(entry, "groupPKs", groupPKs)

		// 5. 判断资源实例授权的用户组是否在effectGroupPKs中
		for _, groupPK := range groupPKs {
			if effectGroupPKSet.Has(groupPK) {
				return true, nil
			}
		}
	}
	return false, nil
}

// parseResourceNode 解析资源节点并去重
func parseResourceNode(
	resource types.Resource,
	actionResourceType types.ActionResourceType,
) ([]types.ResourceNode, error) {
	resourceNodes := make([]types.ResourceNode, 0, 2)
	nodeSet := set.NewStringSet()

	// 生成resource type -> system/pk 映射
	resourceTypeMap := make(map[string]types.ThinResourceType, 4)
	for _, rt := range actionResourceType.ResourceTypeOfInstanceSelections {
		resourceTypeMap[rt.ID] = rt
	}

	// 解析资源属性
	iamPaths, ok := resource.Attribute.Get(types.IamPath)
	if ok {
		paths := make([]string, 0, 2)
		switch vs := iamPaths.(type) {
		case []interface{}: // 处理属性为array的情况
			for _, v := range vs {
				if s, ok := v.(string); ok {
					paths = append(paths, s)
				} else {
					return nil, errors.New("iamPath is not string")
				}
			}
		case string:
			paths = append(paths, vs)
		default:
			return nil, errors.New("iamPath is not string or array")
		}

		for _, path := range paths {
			if path == "" {
				continue
			}

			nodes := strings.Split(strings.Trim(path, "/"), "/")
			for _, node := range nodes {
				parts := strings.Split(node, ",")
				if len(parts) != 2 {
					return nil, errors.New("iamPath is not valid")
				}

				resourceTypeID := parts[0]
				rt, ok := resourceTypeMap[resourceTypeID]
				if !ok {
					return nil, fmt.Errorf("iamPath resource type not found, resourceTypeID=%s", resourceTypeID)
				}

				node := types.ResourceNode{
					System: rt.System,
					Type:   resourceTypeID,
					ID:     parts[1],
					TypePK: rt.PK,
				}

				if !nodeSet.Has(node.String()) {
					resourceNodes = append(resourceNodes, node)
					nodeSet.Add(node.String())
				}
			}
		}
	}

	if actionResourceType.System != resource.System || actionResourceType.Type != resource.Type {
		return nil, fmt.Errorf(
			"resource type not match, actionResourceType=`%+v`, resource=`%+v`",
			actionResourceType,
			resource,
		)
	}

	node := types.ResourceNode{
		System: resource.System,
		Type:   resource.Type,
		ID:     resource.ID,
		TypePK: actionResourceType.PK,
	}

	if !nodeSet.Has(node.String()) {
		resourceNodes = append(resourceNodes, node)
		nodeSet.Add(node.String())
	}

	return resourceNodes, nil
}
