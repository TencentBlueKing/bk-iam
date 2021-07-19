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

	"github.com/agiledragon/gomonkey"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/impls"

	"iam/pkg/abac/pip"
)

var _ = Describe("Resource", func() {
	Describe("QueryRemoteResourceAttribute", func() {
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

		It("keys empty", func() {
			d, err := pip.QueryRemoteResourceAttribute("bk_test", "app", "demo123", []string{})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), d, 1)
			id, ok := d["id"]
			assert.True(GinkgoT(), ok)
			assert.Equal(GinkgoT(), "demo123", id)
		})

		It("keys only have id", func() {
			d, err := pip.QueryRemoteResourceAttribute(
				"bk_test", "app", "demo123", []string{"id"})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), d, 1)
			id, ok := d["id"]
			assert.True(GinkgoT(), ok)
			assert.Equal(GinkgoT(), "demo123", id)
		})

		It("GetRemoteResource fail", func() {
			patches = gomonkey.ApplyFunc(impls.GetRemoteResource,
				func(system, _type, id string, keys []string) (map[string]interface{}, error) {
					return nil, errors.New("get remote resource fail")
				})

			_, err := pip.QueryRemoteResourceAttribute(
				"bk_test", "app", "demo123", []string{"id", "name"})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "get remote resource fail")
		})

		It("ok", func() {
			want := map[string]interface{}{
				"hello": 1,
			}
			patches = gomonkey.ApplyFunc(impls.GetRemoteResource,
				func(system, _type, id string, keys []string) (map[string]interface{}, error) {
					return want, nil
				})

			r, err := pip.QueryRemoteResourceAttribute(
				"bk_test", "app", "demo123", []string{"id", "name"})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, r)

		})

	})

	Describe("BatchQueryRemoteResourcesAttribute", func() {
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
		wantIDAttrs := []map[string]interface{}{
			{
				"id": "demo123",
			},
			{
				"id": "demo456",
			},
		}

		It("keys empty", func() {
			d, err := pip.BatchQueryRemoteResourcesAttribute(
				"bk_test", "app", []string{"demo123", "demo456"}, []string{})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), d, 2)
			assert.Equal(GinkgoT(), wantIDAttrs, d)
		})

		It("keys only have id", func() {
			d, err := pip.BatchQueryRemoteResourcesAttribute(
				"bk_test", "app", []string{"demo123", "demo456"}, []string{"id"})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), d, 2)
			assert.Equal(GinkgoT(), wantIDAttrs, d)
		})

		It("ListRemoteResources fail", func() {
			patches = gomonkey.ApplyFunc(impls.ListRemoteResources,
				func(system, _type string, ids []string, keys []string) ([]map[string]interface{}, error) {
					return nil, errors.New("list remote resource fail")
				})

			_, err := pip.BatchQueryRemoteResourcesAttribute(
				"bk_test", "app", []string{"demo123", "demo456"}, []string{"id", "name"})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "list remote resource fail")
		})

		It("ok", func() {
			want := []map[string]interface{}{
				{
					"hello": 1,
				},
				{
					"hello": 2,
				},
			}
			patches = gomonkey.ApplyFunc(impls.ListRemoteResources,
				func(system, _type string, ids []string, keys []string) ([]map[string]interface{}, error) {
					return want, nil
				})

			r, err := pip.BatchQueryRemoteResourcesAttribute(
				"bk_test", "app", []string{"demo123", "demo456"}, []string{"id", "name"})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, r)

		})

	})

})
