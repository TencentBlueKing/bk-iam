/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package expression

import "iam/pkg/service/types"

// Retriever define the interface of data retrieve for each layer: memory -> redis -> database
type Retriever interface {
	retrieve(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error)
	setMissing(expressions []types.AuthExpression, missingPKs []int64) error
}

type MissingRetrieveFunc func(pks []int64) (expressions []types.AuthExpression, missingPKs []int64, err error)

// expression cache的失效规则:
// 1. 只有用户自定义的(template_id=0)的, 才会更新(通过alterPolicies)
// 2. 来自于模板的(template_id!=0), 不会更新, 只会新增和删除
