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
	"database/sql"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("SubjectActionGroupResourceService", func() {
	Describe("inter func", func() {
		It("convertToSvcSubjectActionGroupResource", func() {
			daoObj := dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     2,
				ActionPK:      3,
				GroupResource: `{"4":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}

			obj, err := convertToSvcSubjectActionGroupResource(daoObj)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), types.SubjectActionGroupResource{
				SubjectPK: 2,
				ActionPK:  3,
				GroupResource: map[int64]types.ExpiredAtResource{
					4: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							5: {"6"},
						},
					},
				},
			}, obj)
		})

		It("convertToDaoSubjectActionGroupResource", func() {
			obj := types.SubjectActionGroupResource{
				SubjectPK: 2,
				ActionPK:  3,
				GroupResource: map[int64]types.ExpiredAtResource{
					4: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							5: {"6"},
						},
					},
				},
			}

			daoObj, err := convertToDaoSubjectActionGroupResource(0, obj)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), dao.SubjectActionGroupResource{
				PK:            0,
				SubjectPK:     2,
				ActionPK:      3,
				GroupResource: `{"4":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}, daoObj)
		})
	})

	Describe("CreateOrUpdateWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})
		It("manager.GetBySubjectAction fail", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionGroupResource{}, errors.New("error"))

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			_, err := svc.CreateOrUpdateWithTx(nil, int64(1), int64(2), int64(3), int64(10), nil)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySubjectAction")
		})

		It("createWithTx fail", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionGroupResource{}, sql.ErrNoRows)
			mockManager.EXPECT().CreateWithTx(gomock.Any(), dao.SubjectActionGroupResource{
				PK:            0,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{"3":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}).Return(errors.New("error"))

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			_, err := svc.CreateOrUpdateWithTx(nil, int64(1), int64(2), int64(3), int64(10), map[int64][]string{
				5: {"6"},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "createWithTx")
		})

		It("create ok", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionGroupResource{}, sql.ErrNoRows)
			mockManager.EXPECT().CreateWithTx(gomock.Any(), dao.SubjectActionGroupResource{
				PK:            0,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{"3":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}).Return(nil)

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			obj, err := svc.CreateOrUpdateWithTx(nil, int64(1), int64(2), int64(3), int64(10), map[int64][]string{
				5: {"6"},
			})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), types.SubjectActionGroupResource{
				SubjectPK: 1,
				ActionPK:  2,
				GroupResource: map[int64]types.ExpiredAtResource{
					3: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							5: {"6"},
						},
					},
				},
			}, obj)
		})

		It("updateWithTx fail", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().GetBySubjectAction(int64(1), int64(2)).Return(dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{}`,
			}, nil)
			mockManager.EXPECT().UpdateGroupResourceWithTx(gomock.Any(), dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{"3":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}).Return(errors.New("error"))

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			_, err := svc.CreateOrUpdateWithTx(nil, int64(1), int64(2), int64(3), int64(10), map[int64][]string{
				5: {"6"},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "updateWithTx")
		})

		It("update ok", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().GetBySubjectAction(int64(1), int64(2)).Return(dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{}`,
			}, nil)
			mockManager.EXPECT().UpdateGroupResourceWithTx(gomock.Any(), dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{"3":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}).Return(nil)

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			obj, err := svc.CreateOrUpdateWithTx(nil, int64(1), int64(2), int64(3), int64(10), map[int64][]string{
				5: {"6"},
			})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), types.SubjectActionGroupResource{
				SubjectPK: 1,
				ActionPK:  2,
				GroupResource: map[int64]types.ExpiredAtResource{
					3: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							5: {"6"},
						},
					},
				},
			}, obj)
		})
	})

	Describe("DeleteGroupWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})
		It("manager.GetBySubjectAction fail", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionGroupResource{}, errors.New("error"))

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			_, err := svc.DeleteGroupWithTx(nil, int64(1), int64(2), int64(3))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySubjectAction")
		})

		It("update fail", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().GetBySubjectAction(int64(1), int64(2)).Return(dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{"3":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}, nil)
			mockManager.EXPECT().UpdateGroupResourceWithTx(gomock.Any(), dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{}`,
			}).Return(errors.New("error"))

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			_, err := svc.DeleteGroupWithTx(nil, int64(1), int64(2), int64(3))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateWithTx")
		})

		It("ok", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().GetBySubjectAction(int64(1), int64(2)).Return(dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{"3":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}, nil)
			mockManager.EXPECT().UpdateGroupResourceWithTx(gomock.Any(), dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{}`,
			}).Return(nil)

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			obj, err := svc.DeleteGroupWithTx(nil, int64(1), int64(2), int64(3))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), types.SubjectActionGroupResource{
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: map[int64]types.ExpiredAtResource{},
			}, obj)
		})
	})

	Describe("Get", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})
		It("manager.GetBySubjectAction fail", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionGroupResource{}, errors.New("error"))

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			_, err := svc.Get(int64(1), int64(2))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySubjectAction")
		})

		It("ok", func() {
			mockManager := mock.NewMockSubjectActionGroupResourceManager(ctl)
			mockManager.EXPECT().GetBySubjectAction(int64(1), int64(2)).Return(dao.SubjectActionGroupResource{
				PK:            1,
				SubjectPK:     1,
				ActionPK:      2,
				GroupResource: `{"3":{"expired_at":10,"resources":{"5":["6"]}}}`,
			}, nil)

			svc := &subjectActionGroupResourceService{
				manager: mockManager,
			}

			obj, err := svc.Get(int64(1), int64(2))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), types.SubjectActionGroupResource{
				SubjectPK: 1,
				ActionPK:  2,
				GroupResource: map[int64]types.ExpiredAtResource{
					3: {
						ExpiredAt: 10,
						Resources: map[int64][]string{
							5: {"6"},
						},
					},
				},
			}, obj)
		})
	})
})
