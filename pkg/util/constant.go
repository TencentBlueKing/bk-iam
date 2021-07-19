/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util

// RequestIDKey ...
const (
	RequestIDKey       = "request_id"
	RequestIDHeaderKey = "X-Request-Id"

	ClientIDKey = "client_id"

	ErrorIDKey = "err"

	// NeverExpiresUnixTime 永久有效期，使用2100.01.01 00:00:00 的unix time作为永久有效期的表示，单位秒
	// time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
	NeverExpiresUnixTime = 4102444800
)
