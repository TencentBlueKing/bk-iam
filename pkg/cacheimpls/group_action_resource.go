/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"fmt"
	"strconv"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/service"
)

// GroupActionCacheKey ...
type GroupActionCacheKey struct {
	GroupPK  int64
	ActionPK int64
}

func (k GroupActionCacheKey) Key() string {
	return strconv.FormatInt(k.GroupPK, 10) + ":" + strconv.FormatInt(k.ActionPK, 10)
}

// GetGroupActionAuthorizedResource 查询group 授权 action 的所有资源实例
// Key: group_pk:action_pk
// value: {"resource_type_pk": ["resource_id"]}
func GetGroupActionAuthorizedResource(groupPK, actionPK int64) (authorizedResources map[int64][]string, err error) {
	key := GroupActionCacheKey{
		GroupPK:  groupPK,
		ActionPK: actionPK,
	}

	err = GroupActionResourceCache.GetInto(key, &authorizedResources, retrieveGroupActionAuthorizedResource)
	return
}

func retrieveGroupActionAuthorizedResource(key cache.Key) (interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "retrieveAuthorizedResourceByGroupAction")

	k := key.(GroupActionCacheKey)

	// 查询 action 的 system_id, action_related_resource_type_pk
	// 查询操作信息
	action, err := GetAction(k.ActionPK)
	if err != nil {
		err = errorWrapf(err, "GetAction fail actionPK=`%d`", k.ActionPK)
		return nil, err
	}

	actionDetail, err := GetLocalActionDetail(action.System, action.ID)
	if err != nil {
		err = errorWrapf(err, "GetActionDetail fail system=`%s`, action=`%s`", action.System, action.ID)
		return nil, err
	}

	if len(actionDetail.ResourceTypes) != 1 {
		err = errorWrapf(
			fmt.Errorf(
				"rbac action must related one resource type, but got %d, actionPK=`%d`",
				len(actionDetail.ResourceTypes),
				k.ActionPK,
			),
			"",
		)
		return nil, err
	}

	resourceType := actionDetail.ResourceTypes[0]
	// NOTE: RBAC 操作只能关联1个资源类型
	// 获取资源类型的PK
	resourceTypePK, err := GetLocalResourceTypePK(resourceType.System, resourceType.ID)
	if err != nil {
		err = errorWrapf(
			err,
			"GetLocalResourceTypePK fail system=`%s`, resourceTypeID=`%s`",
			resourceType.System,
			resourceType.ID,
		)
		return nil, err
	}

	// 查询 group action 授权的实例
	groupResourcePolicySvc := service.NewGroupResourcePolicyService()
	resources, err := groupResourcePolicySvc.ListResourceByGroupAction(
		k.GroupPK,
		action.System,
		k.ActionPK,
		resourceTypePK,
	)
	if err != nil {
		err = errorWrapf(
			err,
			"groupResourcePolicySvc.ListResourceByGroupAction fail groupPK=%d, actionPK=%d",
			k.GroupPK,
			k.ActionPK,
		)
		return nil, err
	}

	authorizedResources := make(map[int64][]string, 2)
	for _, resource := range resources {
		authorizedResources[resource.ResourceTypePK] = append(
			authorizedResources[resource.ResourceTypePK],
			resource.ResourceID,
		)
	}

	return authorizedResources, nil
}

// DeleteGroupActionAuthorizedResourceCache ...
func DeleteGroupActionAuthorizedResourceCache(groupPK, actionPK int64) error {
	key := GroupActionCacheKey{
		GroupPK:  groupPK,
		ActionPK: actionPK,
	}

	return GroupActionResourceCache.Delete(key)
}
