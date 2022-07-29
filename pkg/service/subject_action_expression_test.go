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

var _ = Describe("SubjectActionExpressionService", func() {
	Describe("CreateOrUpdateWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})
		It("manager.GetBySubjectAction fail", func() {
			mockManager := mock.NewMockSubjectActionExpressionManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionExpression{}, errors.New("error"))

			svc := &subjectActionExpressionService{
				manager: mockManager,
			}

			err := svc.CreateOrUpdateWithTx(nil, types.SubjectActionExpression{
				SubjectPK: 1,
				ActionPK:  2,
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetBySubjectAction")
		})

		It("createWithTx fail", func() {
			mockManager := mock.NewMockSubjectActionExpressionManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionExpression{}, sql.ErrNoRows)
			mockManager.EXPECT().CreateWithTx(gomock.Any(), dao.SubjectActionExpression{
				PK:         0,
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{}`,
				ExpiredAt:  10,
			}).Return(errors.New("error"))

			svc := &subjectActionExpressionService{
				manager: mockManager,
			}

			err := svc.CreateOrUpdateWithTx(nil, types.SubjectActionExpression{
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{}`,
				ExpiredAt:  10,
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "CreateWithTx")
		})

		It("create ok", func() {
			mockManager := mock.NewMockSubjectActionExpressionManager(ctl)
			mockManager.EXPECT().
				GetBySubjectAction(int64(1), int64(2)).
				Return(dao.SubjectActionExpression{}, sql.ErrNoRows)
			mockManager.EXPECT().CreateWithTx(gomock.Any(), dao.SubjectActionExpression{
				PK:         0,
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{}`,
				ExpiredAt:  10,
			}).Return(nil)

			svc := &subjectActionExpressionService{
				manager: mockManager,
			}

			err := svc.CreateOrUpdateWithTx(nil, types.SubjectActionExpression{
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{}`,
				ExpiredAt:  10,
			})
			assert.NoError(GinkgoT(), err)
		})

		It("updateWithTx fail", func() {
			mockManager := mock.NewMockSubjectActionExpressionManager(ctl)
			mockManager.EXPECT().GetBySubjectAction(int64(1), int64(2)).Return(dao.SubjectActionExpression{
				PK:         1,
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{}`,
				ExpiredAt:  10,
			}, nil)
			mockManager.EXPECT().
				UpdateExpressionExpiredAtWithTx(gomock.Any(), int64(1), `{"OR":[{"content":[]}`, int64(10)).
				Return(errors.New("error"))

			svc := &subjectActionExpressionService{
				manager: mockManager,
			}

			err := svc.CreateOrUpdateWithTx(nil, types.SubjectActionExpression{
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{"OR":[{"content":[]}`,
				ExpiredAt:  10,
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateWithTx")
		})

		It("update ok", func() {
			mockManager := mock.NewMockSubjectActionExpressionManager(ctl)
			mockManager.EXPECT().GetBySubjectAction(int64(1), int64(2)).Return(dao.SubjectActionExpression{
				PK:         1,
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{}`,
				ExpiredAt:  10,
			}, nil)
			mockManager.EXPECT().
				UpdateExpressionExpiredAtWithTx(gomock.Any(), int64(1), `{"OR":[{"content":[]}`, int64(10)).
				Return(nil)

			svc := &subjectActionExpressionService{
				manager: mockManager,
			}

			err := svc.CreateOrUpdateWithTx(nil, types.SubjectActionExpression{
				SubjectPK:  1,
				ActionPK:   2,
				Expression: `{"OR":[{"content":[]}`,
				ExpiredAt:  10,
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("ListBySubjectAction", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})
		It("manager.ListBySubjectAction fail", func() {
			mockManager := mock.NewMockSubjectActionExpressionManager(ctl)
			mockManager.EXPECT().
				ListBySubjectAction([]int64{1}, int64(2)).
				Return([]dao.SubjectActionExpression{}, errors.New("error"))

			svc := &subjectActionExpressionService{
				manager: mockManager,
			}

			_, err := svc.ListBySubjectAction([]int64{1}, 2)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListBySubjectAction")
		})

		It("ok", func() {
			mockManager := mock.NewMockSubjectActionExpressionManager(ctl)
			mockManager.EXPECT().
				ListBySubjectAction([]int64{1}, int64(2)).
				Return([]dao.SubjectActionExpression{
					{
						PK:        1,
						ExpiredAt: 0,
					},
					{
						PK:        2,
						ExpiredAt: 10,
					},
				}, nil)

			svc := &subjectActionExpressionService{
				manager: mockManager,
			}

			expressions, err := svc.ListBySubjectAction([]int64{1}, 2)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.SubjectActionExpression{
				{
					PK:        2,
					ExpiredAt: 10,
				},
			}, expressions)
		})
	})
})
