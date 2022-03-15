/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("ActionSlz", func() {

	Describe("ActionUpdateSerializer Validate", func() {
		var slz actionUpdateSerializer
		var rrt []relatedResourceType
		BeforeEach(func() {
			slz = actionUpdateSerializer{}
			rrt = []relatedResourceType{
				{
					SystemID: "aaa",
					ID:       "bbb",
					RelatedInstanceSelections: []referenceInstanceSelection{
						{
							SystemID: "cccc",
							ID:       "cccc",
						},
					},
				},
			}
		})
		It("name empty", func() {
			a := map[string]interface{}{
				"name": "",
			}
			valid, _ := slz.validate(a)
			assert.False(GinkgoT(), valid)
		})

		It("name_en empty", func() {
			b := map[string]interface{}{
				"name_en": "",
			}
			valid, _ := slz.validate(b)
			assert.False(GinkgoT(), valid)
		})

		It("version < 1", func() {
			c := map[string]interface{}{
				"version": 0,
			}
			valid, _ := slz.validate(c)
			assert.False(GinkgoT(), valid)
		})

		It("valid relatedResourceType", func() {
			slz.RelatedResourceTypes = rrt
			d := map[string]interface{}{}
			valid, _ := slz.validate(d)
			assert.True(GinkgoT(), valid)
		})

		It("all valid", func() {
			slz2 := actionUpdateSerializer{
				Name:                 "a",
				NameEn:               "b",
				Version:              1,
				RelatedResourceTypes: rrt,
			}
			e := map[string]interface{}{
				"name":    "a",
				"name_en": "b",
				"version": 1,
			}
			valid, _ := slz2.validate(e)
			assert.True(GinkgoT(), valid)
		})

	})

	Describe("ValidateRelatedInstanceSelections", func() {
		It("invalid", func() {
			a := []referenceInstanceSelection{
				{},
			}
			valid, message := validateRelatedInstanceSelections(a, "", "")
			assert.False(GinkgoT(), valid)
			assert.Contains(GinkgoT(), message, "data of action_id")

		})
		It("valid", func() {
			d := []referenceInstanceSelection{
				{
					SystemID: "aaa",
					ID:       "aaa",
				},
			}
			valid, message := validateRelatedInstanceSelections(d, "", "")
			assert.True(GinkgoT(), valid)
			assert.Equal(GinkgoT(), "valid", message)
		})
	})

	Describe("ValidateRelatedResourceTypes", func() {
		It("empty, invalid", func() {
			a := []relatedResourceType{
				{},
			}
			valid, message := validateRelatedResourceTypes(a, "")
			assert.False(GinkgoT(), valid)
			assert.Contains(GinkgoT(), message, "data of action_id")

		})
		It("empty instanceSelections", func() {
			b := []relatedResourceType{
				{
					SystemID:                  "aaa",
					ID:                        "bbb",
					RelatedInstanceSelections: []referenceInstanceSelection{},
				},
			}
			valid, message := validateRelatedResourceTypes(b, "")
			assert.False(GinkgoT(), valid)
			assert.Contains(GinkgoT(), message, "should contain at least 1 item")

		})
		It("invalid instanceSelections", func() {
			c := []relatedResourceType{
				{
					SystemID: "aaa",
					ID:       "bbb",
					RelatedInstanceSelections: []referenceInstanceSelection{
						{
							SystemID: "cccc",
						},
					},
				},
			}
			valid, _ := validateRelatedResourceTypes(c, "")
			assert.False(GinkgoT(), valid)

		})
		It("all valid", func() {
			e := []relatedResourceType{
				{
					SystemID: "aaa",
					ID:       "bbb",
					RelatedInstanceSelections: []referenceInstanceSelection{
						{
							SystemID: "cccc",
							ID:       "cccc",
						},
					},
				},
			}
			valid, message := validateRelatedResourceTypes(e, "")
			assert.True(GinkgoT(), valid)
			assert.Equal(GinkgoT(), "valid", message)
		})
	})

	Describe("ValidateAction", func() {
		It("invalid", func() {
			a := []actionSerializer{
				{},
			}
			valid, message := validateAction(a)
			assert.False(GinkgoT(), valid)
			assert.Contains(GinkgoT(), message, "data in array")
		})

		It("with no relatedResourceTypes", func() {
			b := []actionSerializer{
				{
					ID:     "aaa",
					Name:   "aaa",
					NameEn: "aaa",
				},
			}
			valid, message := validateAction(b)
			assert.True(GinkgoT(), valid)
			assert.Equal(GinkgoT(), "valid", message)

		})
		It("with invalid relatedResourceTypes", func() {
			c := []actionSerializer{
				{
					ID:     "aaa",
					Name:   "aaa",
					NameEn: "aaa",
					RelatedResourceTypes: []relatedResourceType{
						{},
					},
				},
			}
			valid, _ := validateAction(c)
			assert.False(GinkgoT(), valid)

		})
		It("all valid", func() {
			d := []actionSerializer{
				{
					ID:     "aaa",
					Name:   "aaa",
					NameEn: "aaa",
					RelatedResourceTypes: []relatedResourceType{
						{
							SystemID: "aaa",
							ID:       "bbb",
							RelatedInstanceSelections: []referenceInstanceSelection{
								{
									SystemID: "cccc",
									ID:       "cccc",
								},
							},
						},
					},
				},
			}
			valid, _ := validateAction(d)
			assert.True(GinkgoT(), valid)

		})

	})

})
