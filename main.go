/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"iam/cmd"
)

// @title IAM API
// @version 1.0
// @description 蓝鲸权限中心后台服务 API 文档

// @securityDefinitions.apikey AppCode
// @in header
// @name X-Bk-App-Code

// @securityDefinitions.apikey AppSecret
// @in header
// @name X-Bk-App-Secret

func main() {
	cmd.Execute()
}
