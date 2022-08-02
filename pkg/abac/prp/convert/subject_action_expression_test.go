/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package convert

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("Handler", func() {
	Describe("convertToSubjectActionExpression", func() {
		It("ok", func() {
			patches := gomonkey.ApplyFunc(time.Now, func() time.Time {
				return time.Time{}
			})
			defer patches.Reset()
			patches.ApplyFunc(cacheimpls.GetAction, func(_ int64) (types.ThinAction, error) {
				return types.ThinAction{
					System: "system",
					ID:     "action",
					PK:     1,
				}, nil
			})
			patches.ApplyFunc(cacheimpls.GetLocalActionDetail, func(_, _ string) (types.ActionDetail, error) {
				return types.ActionDetail{
					ResourceTypes: []types.ThinActionResourceType{
						{System: "system", ID: "resource_type"},
					},
				}, nil
			})
			patches.ApplyFunc(cacheimpls.GetLocalResourceTypePK, func(_, _ string) (int64, error) {
				return 1, nil
			})
			patches.ApplyFunc(cacheimpls.GetThinResourceType, func(_ int64) (types.ThinResourceType, error) {
				return types.ThinResourceType{
					System: "system",
					ID:     "resource_type2",
					PK:     2,
				}, nil
			})

			obj := types.SubjectActionGroupResource{
				SubjectPK: 1,
				ActionPK:  1,
				GroupResource: map[int64]types.ResourceExpiredAt{
					1: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							1: {"resource1", "resource2"},
						},
					},
					2: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							2: {"resource3", "resource4"},
						},
					},
				},
			}

			expression, err := SubjectActionGroupResourceToExpression(obj)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), expression.Expression, 209)
		})
	})
})
