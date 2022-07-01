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
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pip"
	"iam/pkg/cacheimpls"
)

var _ = Describe("Subject", func() {
	Describe("GetSubjectPK", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("GetSubjectPK fail", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalSubjectPK, func(_type, id string) (pk int64, err error) {
				return -1, errors.New("get subject_pk fail")
			})

			_, err := pip.GetSubjectPK("user", "tome")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "get subject_pk fail")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalSubjectPK, func(_type, id string) (pk int64, err error) {
				return 123, nil
			})

			pk, err := pip.GetSubjectPK("user", "tom")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(123), pk)
		})
	})

	Describe("GetSubjectDepartmentPKs", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
			patches.Reset()
		})

		It("GetSubjectDepartmentPKs fail", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalSubjectDepartmentPKs, func(pk int64) ([]int64, error) {
				return nil, errors.New("get GetLocalSubjectDepartmentPKs fail")
			})

			_, err := pip.GetSubjectDepartmentPKs(123)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "get GetLocalSubjectDepartmentPKs fail")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.GetLocalSubjectDepartmentPKs, func(pk int64) ([]int64, error) {
				return []int64{1, 2, 3}, nil
			})

			depts, err := pip.GetSubjectDepartmentPKs(123)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []int64{1, 2, 3}, depts)
		})
	})
})
