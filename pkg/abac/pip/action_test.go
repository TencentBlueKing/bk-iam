/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pip_test

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pip"
	"iam/pkg/cacheimpls"
	"iam/pkg/service/types"
)

var _ = Describe("Action", func() {

	Describe("GetActionDetail", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("GetActionPK fail", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetActionDetail, func(system, id string) (types.ActionDetail, error) {
				return types.ActionDetail{}, errors.New("get GetActionDetail fail")
			})

			_, _, err := pip.GetActionDetail("bk_test", "edit")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "get GetActionDetail fail")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetActionDetail, func(system, id string) (types.ActionDetail, error) {
				return types.ActionDetail{PK: 123, ResourceTypes: []types.ThinActionResourceType{
					{
						System: "test",
						ID:     "abc",
					},
				}}, nil
			})

			pk, rts, err := pip.GetActionDetail("bk_test", "edit")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(123), pk)
			assert.Len(GinkgoT(), rts, 1)
		})

	})
})
