/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

// PolicyID Rule:
// abac, table `policy` auto increment ID, 0 - 500000000
// rbac, table `rbac_group_resource_policy` auto increment ID, but scope = 500000000 - 1000000000

const rbacIDBegin = 500000000

const (
	PolicyTypeAbac = "abac"
	PolicyTypeRbac = "rbac"
)
