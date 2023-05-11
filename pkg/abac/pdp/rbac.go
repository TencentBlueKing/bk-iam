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

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac"
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
	resourceNodes, err := abac.ParseResourceNode(resources[0])
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
