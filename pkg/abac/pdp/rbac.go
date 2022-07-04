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
	"strconv"
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

	// 检验资源类型是否匹配
	err = validResourceType(resources, actionResourceTypes)
	if err != nil {
		err = errorWrapf(
			err,
			"validResourceType fail, resources=`%+v`, actionResourceTypes=`%+v`",
			resources,
			actionResourceTypes,
		)
		return
	}

	// 2. 解析资源实例的以及属性, 返回出可能被授权的资源实例节点
	debug.AddStep(entry, "Parse Resource Nodes")
	// 从resources中解析出用于rbac鉴权的资源实例节点 NOTE: 支持rbac鉴权的资源类型只能有一个
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

	debug.AddStep(entry, "Get Action Resource Type PK")
	actionResourceTypePK, err := cacheimpls.GetLocalResourceTypePK(
		actionResourceTypes[0].System,
		actionResourceTypes[0].Type,
	)
	if err != nil {
		err = errorWrapf(
			err,
			"cacheimpls.GetLocalResourceTypePK fail, actionResourceType=`%+v`",
			actionResourceTypes[0],
		)
		return
	}
	debug.WithValue(entry, "actionResourceTypePK", actionResourceTypePK)

	// 4. 遍历查询资源实例节点授权的groupPKs
	debug.AddStep(entry, "Get Resource ActionAuthorized Group PKs")
	effectGroupPKSet := set.NewInt64SetWithValues(effectGroupPKs)
	for i, resourceNode := range resourceNodes {
		var groupPKs []int64
		if withoutCache {
			svc := service.NewGroupResourcePolicyService()
			var actionGroupPKs map[int64][]int64
			actionGroupPKs, err = svc.GetAuthorizedActionGroupMap(
				system, actionResourceTypePK, resourceNode.TypePK, resourceNode.ID,
			)
			if err != nil {
				err = errorWrapf(
					err,
					"svc.GetAuthorizedActionGroupMap fail, system=`%s` action=`%+v` resource=`%+v`",
					system,
					action,
					resourceNode,
				)
				return
			}

			groupPKs = actionGroupPKs[actionPK]
		} else {
			groupPKs, err = cacheimpls.GetResourceActionAuthorizedGroupPKs(
				system,
				actionPK,
				actionResourceTypePK,
				resourceNode.TypePK,
				resourceNode.ID,
			)
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
		}

		debug.WithValue(entry, "loop"+strconv.Itoa(i), map[string]interface{}{
			"resourceNode": resourceNode,
			"groupPKs":     groupPKs,
		})

		// 5. 判断资源实例授权的用户组是否在effectGroupPKs中
		for _, groupPK := range groupPKs {
			if effectGroupPKSet.Has(groupPK) {
				debug.WithValue(entry, "passGroupPK", groupPK)
				return true, nil
			}
		}
	}
	return false, nil
}

func validResourceType(resources []types.Resource, actionResourceTypes []types.ActionResourceType) error {
	if len(resources) != 1 || len(actionResourceTypes) != 1 {
		// NOTE: 能做RBAC鉴权的操作的资源类型只能有一个
		return errors.New("rbacEval only support action with one resource type")
	}

	resource := resources[0]
	actionResourceType := actionResourceTypes[0]
	if actionResourceType.System != resource.System || actionResourceType.Type != resource.Type {
		return fmt.Errorf(
			"resource type not match, actionResourceType=`%+v`, resource=`%+v`",
			actionResourceType,
			resource,
		)
	}

	return nil
}

// parseResourceNode 解析资源节点并去重
func parseResourceNode(
	resource types.Resource,
	actionResourceType types.ActionResourceType,
) ([]types.ResourceNode, error) {
	resourceNodes := make([]types.ResourceNode, 0, 2)
	nodeSet := set.NewStringSet()

	// 解析iam path
	iamPaths, ok := resource.Attribute.Get(types.IamPath)
	if ok {
		iamPathNodes, err := parseIamPath(iamPaths, actionResourceType)
		if err != nil {
			return nil, err
		}

		for _, node := range iamPathNodes {
			uniqueID := node.UniqueID()
			if !nodeSet.Has(uniqueID) {
				resourceNodes = append(resourceNodes, node)
				nodeSet.Add(uniqueID)
			}
		}
	}

	resourceTypePK, err := cacheimpls.GetLocalResourceTypePK(resource.System, resource.Type)
	if err != nil {
		return nil, err
	}

	node := types.ResourceNode{
		System: resource.System,
		Type:   resource.Type,
		ID:     resource.ID,
		TypePK: resourceTypePK,
	}

	uniqueID := node.UniqueID()
	if !nodeSet.Has(uniqueID) {
		resourceNodes = append(resourceNodes, node)
		nodeSet.Add(uniqueID)
	}

	return resourceNodes, nil
}

func parseIamPath(
	iamPaths interface{},
	actionResourceType types.ActionResourceType,
) ([]types.ResourceNode, error) {
	resourceNodes := make([]types.ResourceNode, 0, 2)
	paths := make([]string, 0, 2)
	switch vs := iamPaths.(type) {
	case []interface{}:
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
			if len(parts) != 3 {
				return nil, fmt.Errorf(
					"iamPath=`%s` is not valid, example: `/system_id,resource_type_id,resource_id/`",
					path,
				)
			}

			systemID := parts[0]
			resourceTypeID := parts[1]
			resourceID := parts[2]

			resourceTypePK, err := cacheimpls.GetLocalResourceTypePK(systemID, resourceTypeID)
			if err != nil {
				return nil, err
			}

			node := types.ResourceNode{
				System: systemID,
				Type:   resourceTypeID,
				ID:     resourceID,
				TypePK: resourceTypePK,
			}

			resourceNodes = append(resourceNodes, node)
		}
	}
	return resourceNodes, nil
}
