/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util_test

import (
	"errors"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

var _ = Describe("Validation", func() {

	// Describe("ValidationFieldError", func() {
	//	DescribeTable("ValidationFieldError cases", func(expected string, err util.ValidationFieldError) {
	//		assert.True(GinkgoT(), strings.Contains(err.String(), expected))
	//	},
	//		Entry("empty slice", "", []int64{}, ","),
	//		Entry("slice with 1 value", "1", []int64{1}, ","),
	//		Entry("slice with 3 values", "1,2,3", []int64{1, 2, 3}, ","),
	//	)
	// })

	Describe("ValidationErrorMessage", func() {

		It("io.EOF", func() {
			s := util.ValidationErrorMessage(io.EOF)
			assert.Equal(GinkgoT(), "EOF, json decode fail", s)
		})

		It("not a field error", func() {
			s := util.ValidationErrorMessage(errors.New("anError"))

			contains := strings.Contains(s, "json decode or validate fail")
			assert.True(GinkgoT(), contains)
		})

		It("empty", func() {
			s := util.ValidationErrorMessage(validator.ValidationErrors{})
			assert.Equal(GinkgoT(), "validationErrs with no error message", s)
		})

		Context("fieldError: string", func() {
			type A struct {
				X string `validate:"required,min=2,max=6,email,len=5"`
			}

			It("required", func() {
				a := &A{}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X is required", msg)
			})

			It("min", func() {
				a := &A{"1"}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X must be longer than 2", msg)
			})

			It("max", func() {
				a := &A{"1234567"}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X cannot be longer than 6", msg)
			})

			It("email", func() {
				a := &A{"123456"}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "Invalid email format", msg)
			})

			It("len", func() {
				a := &A{"1@3.cn"}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X must be 5 characters long", msg)
			})

		})

		Context("fieldError: int", func() {
			type A struct {
				X int `validate:"gt=10,gte=12,lt=20,lte=18"`
			}

			It("gt", func() {
				a := &A{1}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X must greater than 10", msg)
			})

			It("gte", func() {
				a := &A{11}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X must greater or equals to 12", msg)
			})

			It("lt", func() {
				a := &A{50}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X must less than 20", msg)
			})

			It("lte", func() {
				a := &A{19}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X must less or equals to 18", msg)
			})

		})
		//
		Context("fieldError: oneof", func() {
			type A struct {
				X string `validate:"oneof=a b c"`
			}

			It("oneof invalid", func() {
				a := &A{"d"}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X must be one of 'a b c'", msg)

			})

		})

		Context("fieldError not include", func() {
			type A struct {
				X string `validate:"ipv4"`
			}

			It("ipv4 not include", func() {
				a := &A{"d"}
				err := validator.New().Struct(a)
				assert.Error(GinkgoT(), err)

				e, ok := err.(validator.ValidationErrors)
				assert.True(GinkgoT(), ok)

				msg := util.ValidationErrorMessage(e)
				assert.Equal(GinkgoT(), "X is not valid, condition: ipv4", msg)

			})

		})

	})

})
