/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
)

var _ = Describe("Context", func() {

	Describe("NewExprContext", func() {
		It("ok", func() {
			ctx := &request.Request{}
			resource := &types.Resource{}
			ec := NewExprContext(ctx, resource)
			assert.NotNil(GinkgoT(), ec)
		})

	})

	var c *ExprContext
	BeforeEach(func() {
		request := &request.Request{
			System: "iam",
			Subject: types.Subject{
				Type: "user",
				ID:   "admin",
			},
			Action: types.Action{
				ID: "execute_job",
			},
		}
		resource := &types.Resource{
			System:    "iam",
			Type:      "job",
			ID:        "job1",
			Attribute: map[string]interface{}{"key": "value1"},
		}
		c = NewExprContext(request, resource)
	})

	Describe("GetAttr", func() {
		It("ok", func() {
			a, err := c.GetAttr("id")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "job1", a)
		})

	})

	Describe("getResourceAttr", func() {
		It("ok id", func() {
			a, err := c.getResourceAttr("id")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "job1", a)
		})
		It("ok type", func() {
			a, err := c.getResourceAttr("type")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), nil, a)
		})
		It("ok others", func() {
			a, err := c.getResourceAttr("key")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "value1", a)
		})

		It("fail not exists", func() {
			a, err := c.getResourceAttr("notExists")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), nil, a)
		})

	})

	//Describe("GetFullNameAttr", func() {
	//
	//	It("invalid name", func() {
	//		_, err := c.GetFullNameAttr("a")
	//		assert.Error(GinkgoT(), err)
	//	})
	//
	//	It("name not support", func() {
	//		_, err := c.GetFullNameAttr("environment.a")
	//		assert.Error(GinkgoT(), err)
	//		assert.Contains(GinkgoT(), err.Error(), "name not support")
	//	})
	//
	//	It("ok resource attr", func() {
	//		a, err := c.GetFullNameAttr("resource.id")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), "job1", a)
	//	})
	//
	//	It("ok action attr", func() {
	//		a, err := c.GetFullNameAttr("action.id")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), "execute_job", a)
	//	})
	//
	//	It("ok subject attr", func() {
	//		a, err := c.GetFullNameAttr("subject.id")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), "admin", a)
	//	})
	//
	//})

	//Describe("getActionAttr", func() {
	//	It("ok id", func() {
	//		a, err := c.getActionAttr("id")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), "execute_job", a)
	//	})
	//	It("fail not support", func() {
	//		a, err := c.getActionAttr("notExists")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), nil, a)
	//	})
	//
	//})
	//Describe("getSubjectAttr", func() {
	//	It("ok id", func() {
	//		a, err := c.getSubjectAttr("id")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), "admin", a)
	//	})
	//	It("ok type", func() {
	//		a, err := c.getSubjectAttr("type")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), "user", a)
	//	})
	//
	//	It("fail not exists", func() {
	//		a, err := c.getSubjectAttr("notExists")
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), nil, a)
	//	})
	//
	//})
})
