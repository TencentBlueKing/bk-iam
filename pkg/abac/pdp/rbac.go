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
	"iam/pkg/abac/types"
	"iam/pkg/logging/debug"
)

func rbacEval(
	system string,
	action types.Action,
	resources []types.Resource,
	effectGroupPKs []int64,
	withoutCache bool,
	entry *debug.Entry,
) (isPass bool, err error) {
	/*
		TODO rbac鉴权逻辑:

		1. 解析出鉴权的resources中所有的节点, 并去重, 并且需要查询到所有resource_type_pk
		2. 查询操作的 action_resource_type_pk
		3. 使用 system_id, resource_type_pk, resource_id, action_resource_type_pk
			查询 group_resource_policy, 得到 group_pk, action_pks
		4. 筛选出 action_pks 中包含 action_pk 的 group_pk
		5. 得到的group_pk与rbacGroupPKs比较, 如果包含, 则通过鉴权
	*/
	return false, nil
}
