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

import "iam/pkg/service/types"

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// EnginePolicySVC is the layer-object name
const EnginePolicySVC = "EnginePolicySVC"

// EnginePolicyService provide the func for iam-engine
type EnginePolicyService interface {
	GetMaxPKBeforeUpdatedAt(updatedAt int64) (int64, error)
	ListPKBetweenUpdatedAt(beginUpdatedAt, endUpdatedAt int64) ([]int64, error)
	ListBetweenPK(expiredAt, minPK, maxPK int64) (policies []types.EnginePolicy, err error)
	ListByPKs(pks []int64) (policies []types.EnginePolicy, err error)
}
