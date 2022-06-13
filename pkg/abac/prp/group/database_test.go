/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package group

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/service/mock"
	"iam/pkg/service/types"
)

var _ = Describe("database", func() {
	Describe("serviceGroupAuthTypeRetriever", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("service.ListGroupAuthBySystemGroupPKs fail", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListGroupAuthBySystemGroupPKs("test", []int64{1, 2}).Return(
				nil, errors.New("error"),
			)

			retriever := &serviceGroupAuthTypeRetriever{
				systemID: "test",
				service:  mockGroupService,
			}

			_, err := retriever.Retrieve([]int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "error")
		})

		It("ok", func() {
			mockGroupService := mock.NewMockGroupService(ctl)
			mockGroupService.EXPECT().ListGroupAuthBySystemGroupPKs("test", []int64{1, 2}).Return(
				[]types.GroupAuthType{
					{GroupPK: 1, AuthType: types.AuthTypeABAC},
				}, nil,
			)

			retriever := &serviceGroupAuthTypeRetriever{
				systemID: "test",
				service:  mockGroupService,
			}

			groupAuthTypes, err := retriever.Retrieve([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupAuthType{
				{GroupPK: 1, AuthType: types.AuthTypeABAC},
				{GroupPK: 2, AuthType: types.AuthTypeNone},
			}, groupAuthTypes)
		})
	})
})
