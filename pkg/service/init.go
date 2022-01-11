/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"context"
	"time"
)

func InitExpressionCleanTask(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(24) * time.Hour)
	svc := NewPolicyService()
	go func() {
		for {
			select {
			case <-ticker.C:
				// TODO 记录保存日志，需要有一个新的logger
				svc.DeleteUnQuotedExpression()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
