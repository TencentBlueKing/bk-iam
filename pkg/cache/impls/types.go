/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package impls

import (
	"strconv"
)

// 公共的 CacheKey; struct中内容相同的, 不应该重复key

// ActionIDCacheKey ...
type ActionIDCacheKey struct {
	SystemID string
	ActionID string
}

// Key ...
func (k ActionIDCacheKey) Key() string {
	return k.SystemID + ":" + k.ActionID
}

// SubjectPKCacheKey ...
type SubjectPKCacheKey struct {
	PK int64
}

// Key ...
func (k SubjectPKCacheKey) Key() string {
	return strconv.FormatInt(k.PK, 10)
}
