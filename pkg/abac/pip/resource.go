/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pip

import (
	"iam/pkg/cacheimpls"
	"iam/pkg/errorx"
)

// ResourcePIP ...
const ResourcePIP = "ResourcePIP"

// QueryRemoteResourceAttribute 查询被依赖资源的属性
func QueryRemoteResourceAttribute(system, _type, id string, keys []string) (map[string]interface{}, error) {
	// if no keys, return without query
	// 如果不需要属性, iam不查询第三方, 不负责校验id的存在与否
	if len(keys) == 0 || (len(keys) == 1 && keys[0] == "id") {
		return map[string]interface{}{
			"id": id,
		}, nil
	}

	resource, err := cacheimpls.GetRemoteResource(system, _type, id, keys)
	if err != nil {
		err = errorx.Wrapf(err, ResourcePIP, "QueryRemoteResourceAttribute",
			"cacheimpls.GetRemoteResource system=`%s`, _type=`%s`, id=`%s`, keys=`%+v` fail",
			system, _type, id, keys)
		return nil, err
	}

	return resource, nil
}

// BatchQueryRemoteResourcesAttribute 批量查询资源的属性 without cache
func BatchQueryRemoteResourcesAttribute(
	system, _type string, ids []string, keys []string,
) ([]map[string]interface{}, error) {
	if len(keys) == 0 || (len(keys) == 1 && keys[0] == "id") {
		resources := make([]map[string]interface{}, 0, len(ids))
		for _, id := range ids {
			resources = append(resources, map[string]interface{}{
				"id": id,
			})
		}
		return resources, nil
	}

	resources, err := cacheimpls.ListRemoteResources(system, _type, ids, keys)
	if err != nil {
		err = errorx.Wrapf(err, ResourcePIP, "BatchQueryRemoteResourcesAttribute",
			"listRemoteResources system=`%s`, _type=`%s`, ids=`%+v`, keys=`%+v` fail",
			system, _type, ids, keys)
		return nil, err
	}

	return resources, nil
}
