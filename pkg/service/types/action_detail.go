/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

// ThinActionResourceType 用于鉴权的操作关联的资源类型.
type ThinActionResourceType struct {
	System string
	ID     string
}

// ActionDetail 用于复合缓存
type ActionDetail struct {
	// action pk
	PK int64

	// action resource types
	ResourceTypes []ThinActionResourceType
}
