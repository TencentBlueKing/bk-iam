/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package abac

import (
	"errors"
	"fmt"
	"strings"

	"github.com/TencentBlueKing/gopkg/collection/set"

	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
)

// ParseResourceNode 解析资源节点并去重
func ParseResourceNode(resource types.Resource) ([]types.ResourceNode, error) {
	resourceNodes := make([]types.ResourceNode, 0, 2)
	nodeSet := set.NewStringSet()

	// 解析iam path
	iamPaths, ok := resource.Attribute.Get(types.IamPath)
	if ok {
		iamPathNodes, err := parseIamPath(iamPaths)
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

func parseIamPath(iamPaths interface{}) ([]types.ResourceNode, error) {
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
