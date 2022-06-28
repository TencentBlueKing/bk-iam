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

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/database/sdao"
	sdaoMock "iam/pkg/database/sdao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("ActionService", func() {
	var (
		ctl                               *gomock.Controller
		svc                               *actionService
		mockActionResourceTypeManager     *mock.MockActionResourceTypeManager
		mockSaaSInstanceSelectionManager  *sdaoMock.MockSaaSInstanceSelectionManager
		mockResourceTypeManager           *mock.MockResourceTypeManager
		mockSaaSActionResourceTypeManager *sdaoMock.MockSaaSActionResourceTypeManager
	)

	BeforeEach(func() {
		ctl = gomock.NewController(GinkgoT())
		mockActionResourceTypeManager = mock.NewMockActionResourceTypeManager(ctl)
		mockSaaSInstanceSelectionManager = sdaoMock.NewMockSaaSInstanceSelectionManager(ctl)
		mockResourceTypeManager = mock.NewMockResourceTypeManager(ctl)
		mockSaaSActionResourceTypeManager = sdaoMock.NewMockSaaSActionResourceTypeManager(ctl)
		svc = &actionService{
			actionResourceTypeManager:     mockActionResourceTypeManager,
			saasInstanceSelectionManager:  mockSaaSInstanceSelectionManager,
			resourceTypeManager:           mockResourceTypeManager,
			saasActionResourceTypeManager: mockSaaSActionResourceTypeManager,
		}
	})
	AfterEach(func() {
		ctl.Finish()
	})

	Describe("parseRelatedInstanceSelections", func() {
		It("rawRelatedInstanceSelections empty", func() {
			ris, err := svc.parseRelatedInstanceSelections("")
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), ris, 0)
		})
		It("json unmarshal error", func() {
			_, err := svc.parseRelatedInstanceSelections(`[{"system_id": 1, "id": "2"}]`)
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "unmarshal rawRelatedInstanceSelections=(.*) fail", err.Error())
		})
		It("ok", func() {
			ris, err := svc.parseRelatedInstanceSelections(`[{"system_id": "1", "id": "2"}]`)
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), ris, 1)
			assert.Equal(GinkgoT(), ris[0].System, "1")
			assert.Equal(GinkgoT(), ris[0].ID, "2")
		})
	})
	Describe("queryResourceTypeChain", func() {
		It("ListBySystem error", func() {
			mockSaaSInstanceSelectionManager.EXPECT().ListBySystem("system_id").Return(nil, errors.New("error"))

			_, err := svc.queryResourceTypeChain([]types.ReferenceInstanceSelection{
				{System: "system_id", ID: "id"},
			})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "saasInstanceSelectionManager.ListBySystem (.*) fail", err.Error())
		})
		It("instanceSelection not exists", func() {
			mockSaaSInstanceSelectionManager.EXPECT().ListBySystem("system_id").Return(
				[]sdao.SaaSInstanceSelection{
					{
						System:            "system_id",
						ID:                "id",
						ResourceTypeChain: `[{"system_id": "1", "id": "2"}]`,
					},
				},
				nil,
			)

			_, err := svc.queryResourceTypeChain([]types.ReferenceInstanceSelection{
				{System: "system_id", ID: "id1"},
			})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "instanceSelection not exists", err.Error())
		})
		It("unmarshal error", func() {
			mockSaaSInstanceSelectionManager.EXPECT().ListBySystem("system_id").Return(
				[]sdao.SaaSInstanceSelection{
					{
						System:            "system_id",
						ID:                "id",
						ResourceTypeChain: `[{"system_id": 1, "id": "2"}]`,
					},
				},
				nil,
			)
			_, err := svc.queryResourceTypeChain([]types.ReferenceInstanceSelection{
				{System: "system_id", ID: "id"},
			})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "unmarshal instanceSelection.ResourceTypeChain", err.Error())
		})

		It("ok", func() {
			mockSaaSInstanceSelectionManager.EXPECT().ListBySystem("system_id").Return(
				[]sdao.SaaSInstanceSelection{
					{
						System:            "system_id",
						ID:                "id",
						ResourceTypeChain: `[{"system_id": "1", "id": "2"}]`,
					},
				},
				nil,
			)

			resourceTypeChains, err := svc.queryResourceTypeChain([]types.ReferenceInstanceSelection{
				{System: "system_id", ID: "id"},
			})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), resourceTypeChains, 1)
			assert.Len(GinkgoT(), resourceTypeChains["system_id:id"], 1)
			assert.Equal(GinkgoT(), resourceTypeChains["system_id:id"][0].System, "1")
			assert.Equal(GinkgoT(), resourceTypeChains["system_id:id"][0].ID, "2")
		})
	})
	Describe("queryResourceTypePK", func() {
		It("ListByIDs error", func() {
			mockResourceTypeManager.EXPECT().ListByIDs("system_id", []string{"id"}).Return(nil, errors.New("error"))

			_, err := svc.queryResourceTypePK([]rawResourceType{{System: "system_id", ID: "id"}})
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "resourceTypeManager.ListByIDs (.*) fail", err.Error())
		})
		It("ok", func() {
			mockResourceTypeManager.EXPECT().ListByIDs("system_id", []string{"id"}).Return(
				[]dao.ResourceType{
					{PK: int64(1), System: "system_id", ID: "id"},
				},
				nil,
			)

			resourceTypePKMap, err := svc.queryResourceTypePK([]rawResourceType{{System: "system_id", ID: "id"}})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), resourceTypePKMap, 1)
			assert.Equal(GinkgoT(), resourceTypePKMap["system_id:id"], int64(1))
		})
	})

	Describe("convertToActionResourceTypes", func() {
		It("pk of action related resource type not found", func() {
			_, err := svc.convertToActionResourceTypes(
				[]sdao.SaaSActionResourceType{
					{ResourceTypeSystem: "rt_system", ResourceTypeID: "rt_id"},
				},
				[][]types.ReferenceInstanceSelection{},
				map[string][]rawResourceType{},
				map[string]int64{},
			)
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "pk of action related resource type not found", err.Error())
		})
		It("pk of resource type in chain not found", func() {
			_, err := svc.convertToActionResourceTypes(
				[]sdao.SaaSActionResourceType{
					{ResourceTypeSystem: "rt_system", ResourceTypeID: "rt_id"},
				},
				[][]types.ReferenceInstanceSelection{
					{
						{System: "system_id", ID: "id"},
					},
				},
				map[string][]rawResourceType{
					"system_id:id": {{System: "system_id", ID: "id"}},
				},
				map[string]int64{
					"rt_system:rt_id": int64(1),
				},
			)
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "pk of resource type in chain not found", err.Error())
		})
		It("ok", func() {
			actionResourceTypes, err := svc.convertToActionResourceTypes(
				[]sdao.SaaSActionResourceType{
					{ResourceTypeSystem: "rt_system", ResourceTypeID: "rt_id"},
				},
				[][]types.ReferenceInstanceSelection{
					{
						{System: "system_id", ID: "id"},
					},
				},
				map[string][]rawResourceType{
					"system_id:id": {{System: "system_id", ID: "id"}},
				},
				map[string]int64{
					"rt_system:rt_id": int64(1),
					"system_id:id":    int64(2),
				},
			)
			expected := []types.ThinActionResourceType{
				{
					PK:     int64(1),
					System: "rt_system",
					ID:     "rt_id",
					ResourceTypeOfInstanceSelections: []types.ThinResourceType{
						{
							PK:     int64(2),
							System: "system_id",
							ID:     "id",
						},
					},
				},
			}
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), actionResourceTypes, expected)
		})
	})
	Describe("ListThinActionResourceTypes", func() {
		It("ListByActionID error", func() {
			mockSaaSActionResourceTypeManager.EXPECT().ListByActionID("system_id", "action_id").Return(
				nil, errors.New("error"),
			)

			_, err := svc.ListThinActionResourceTypes("system_id", "action_id")
			assert.Error(GinkgoT(), err)
			assert.Regexp(GinkgoT(), "ListByActionID(.*) fail", err.Error())
		})
		It("ListByActionID result is empty", func() {
			mockSaaSActionResourceTypeManager.EXPECT().ListByActionID("system_id", "action_id").Return(
				nil, nil,
			)

			actionResourceTypes, err := svc.ListThinActionResourceTypes("system_id", "action_id")
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), actionResourceTypes, 0)
		})
		Context("", func() {
			var patches *gomonkey.Patches
			BeforeEach(func() {
				mockSaaSActionResourceTypeManager.EXPECT().ListByActionID("system_id", "action_id").Return(
					[]sdao.SaaSActionResourceType{
						{RelatedInstanceSelections: `[{"system_id": "system_id", "id": "id"}]`},
					}, nil,
				)
			})
			AfterEach(func() {
				patches.Reset()
			})
			It("parseRelatedInstanceSelections error", func() {
				patches = gomonkey.ApplyPrivateMethod(
					svc,
					"parseRelatedInstanceSelections",
					func(*actionService, string) ([]types.ReferenceInstanceSelection, error) {
						return nil, errors.New("error")
					},
				)

				_, err := svc.ListThinActionResourceTypes("system_id", "action_id")
				assert.Error(GinkgoT(), err)
				assert.Regexp(GinkgoT(), "parseRelatedInstanceSelections (.*) fail", err.Error())
			})
			Context("", func() {
				BeforeEach(func() {
					patches = gomonkey.ApplyPrivateMethod(
						svc,
						"parseRelatedInstanceSelections",
						func(*actionService, string) ([]types.ReferenceInstanceSelection, error) {
							return []types.ReferenceInstanceSelection{
								{System: "system_id", ID: "id"},
							}, nil
						},
					)
				})
				It("queryResourceTypeChain error", func() {
					patches.ApplyPrivateMethod(
						svc,
						"queryResourceTypeChain",
						func(*actionService, []types.ReferenceInstanceSelection) (map[string][]rawResourceType, error) {
							return nil, errors.New("error")
						},
					)

					_, err := svc.ListThinActionResourceTypes("system_id", "action_id")
					assert.Error(GinkgoT(), err)
					assert.Regexp(GinkgoT(), "queryResourceTypeChain allInstanceSelections=(.*) fail", err.Error())
				})
				Context("", func() {
					BeforeEach(func() {
						patches.ApplyPrivateMethod(
							svc,
							"queryResourceTypeChain",
							func(*actionService, []types.ReferenceInstanceSelection) (map[string][]rawResourceType, error) {
								return map[string][]rawResourceType{
									"system_id:id": {
										{System: "system_id", ID: "id"},
									},
								}, nil
							},
						)
					})
					It("queryResourceTypePK error", func() {
						patches.ApplyPrivateMethod(
							svc,
							"queryResourceTypePK",
							func(*actionService, []rawResourceType) (map[string]int64, error) {
								return nil, errors.New("error")
							},
						)

						_, err := svc.ListThinActionResourceTypes("system_id", "action_id")
						assert.Error(GinkgoT(), err)
						assert.Regexp(GinkgoT(), "queryResourceTypePK rts=(.*) fail", err.Error())
					})
					Context("", func() {
						BeforeEach(func() {
							patches.ApplyPrivateMethod(
								svc,
								"queryResourceTypePK",
								func(*actionService, []rawResourceType) (map[string]int64, error) {
									return map[string]int64{"system_id:id": int64(1)}, nil
								},
							)
						})
						It("convertToActionResourceTypes error", func() {
							patches.ApplyPrivateMethod(
								svc,
								"convertToActionResourceTypes",
								func(
									*actionService,
									[]sdao.SaaSActionResourceType,
									[][]types.ReferenceInstanceSelection,
									map[string][]rawResourceType,
									map[string]int64,
								) ([]types.ThinActionResourceType, error) {
									return nil, errors.New("error")
								},
							)

							_, err := svc.ListThinActionResourceTypes("system_id", "action_id")
							assert.Error(GinkgoT(), err)
							assert.Regexp(GinkgoT(), "convertToActionResourceTypes (.*) fail", err.Error())
						})
						It("ok", func() {
							patches.ApplyPrivateMethod(
								svc,
								"convertToActionResourceTypes",
								func(
									*actionService,
									[]sdao.SaaSActionResourceType,
									[][]types.ReferenceInstanceSelection,
									map[string][]rawResourceType,
									map[string]int64,
								) ([]types.ThinActionResourceType, error) {
									return nil, nil
								},
							)

							_, err := svc.ListThinActionResourceTypes("system_id", "action_id")
							assert.NoError(GinkgoT(), err)
						})
					})
				})
			})
		})
	})

	Describe("ListActionResourceTypeIDByActionSystem", func() {
		It("ok", func() {
			returned := []dao.ActionResourceType{
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_cmdb",
					ResourceTypeID:     "host",
				},
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_job",
					ResourceTypeID:     "job",
				},
			}

			expected := []types.ActionResourceTypeID{
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_cmdb",
					ResourceTypeID:     "host",
				},
				{
					ActionSystem:       "bk_job",
					ActionID:           "execute",
					ResourceTypeSystem: "bk_job",
					ResourceTypeID:     "job",
				},
			}

			mockActionResourceTypeManager.EXPECT().ListByActionSystem(
				gomock.Any()).Return(returned, nil)

			sgs, err := svc.ListActionResourceTypeIDByActionSystem("bk_job")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, sgs)
		})

		It("fail", func() {
			mockActionResourceTypeManager.EXPECT().ListByActionSystem(
				gomock.Any()).Return(nil, errors.New("list fail"))

			_, err := svc.ListActionResourceTypeIDByActionSystem("bk_job")
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "actionResourceTypeManager.ListByActionSystem")
		})
	})
})
