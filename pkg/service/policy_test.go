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
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

var _ = Describe("PolicyService", func() {
	Describe("ListAuthBySubjectAction cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.AuthPolicy{
				{
					PK: 1,
				},
				{
					PK: 2,
				},
			}
			expected := []types.AuthPolicy{
				{
					PK: 1,
				},
				{
					PK: 2,
				},
			}

			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListAuthBySubjectAction([]int64{1, 2}, int64(1), gomock.Any()).Return(
				returned, nil,
			)
			svc := policyService{
				manager: mockPolicyManager,
			}

			policies, err := svc.ListAuthBySubjectAction([]int64{1, 2}, int64(1))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, policies)
		})

		It("error", func() {
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListAuthBySubjectAction([]int64{1, 2}, int64(1), gomock.Any()).Return(
				nil, errors.New("error"),
			)
			svc := policyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListAuthBySubjectAction([]int64{1, 2}, int64(1))
			assert.Error(GinkgoT(), err)

		})
	})

	Describe("ListExpressionByPKs cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.AuthExpression{
				{
					PK: 1,
				},
				{
					PK: 2,
				},
			}
			expected := []types.AuthExpression{
				{
					PK: 1,
				},
				{
					PK: 2,
				},
			}

			mockExpressionManager := mock.NewMockExpressionManager(ctl)
			mockExpressionManager.EXPECT().ListAuthByPKs([]int64{1, 2}).Return(returned, nil)

			svc := policyService{
				expressionManger: mockExpressionManager,
			}
			expressions, err := svc.ListExpressionByPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, expressions)
		})

		It("error", func() {
			mockExpressionManager := mock.NewMockExpressionManager(ctl)
			mockExpressionManager.EXPECT().ListAuthByPKs([]int64{1, 2}).Return(nil, errors.New("error"))

			svc := policyService{
				expressionManger: mockExpressionManager,
			}
			_, err := svc.ListExpressionByPKs([]int64{1, 2})
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("ListThinBySubjectSystemTemplate cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.Policy{
				{
					PK: 1,
				},
				{
					PK: 2,
				},
			}

			expected := []types.ThinPolicy{
				{
					Version: "1",
					ID:      1,
				},
				{
					Version: "1",
					ID:      2,
				},
			}

			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBySubjectActionTemplate(int64(1), []int64{1}, int64(1)).Return(returned, nil)

			svc := policyService{
				manager: mockPolicyManager,
			}

			policies, err := svc.ListThinBySubjectActionTemplate(int64(1), []int64{1}, int64(1))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, policies)
		})

		It("error", func() {
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBySubjectActionTemplate(int64(1), []int64{1}, int64(1)).Return(nil,
				errors.New("error"))

			svc := policyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListThinBySubjectActionTemplate(int64(1), []int64{1}, int64(1))
			assert.Error(GinkgoT(), err)
		})
	})

	Describe("ListAuthBySubjectAction cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.Policy{
				{
					PK: 1,
				},
				{
					PK: 2,
				},
			}

			expected := []types.ThinPolicy{
				{
					Version: "1",
					ID:      1,
				},
				{
					Version: "1",
					ID:      2,
				},
			}

			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBySubjectTemplateBeforeExpiredAt(int64(1), int64(0), int64(10)).Return(
				returned, nil,
			)
			svc := policyService{
				manager: mockPolicyManager,
			}

			policies, err := svc.ListThinBySubjectTemplateBeforeExpiredAt(int64(1), int64(0), int64(10))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, policies)
		})

		It("error", func() {
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBySubjectTemplateBeforeExpiredAt(int64(1), int64(0), int64(10)).Return(
				nil, errors.New("error"),
			)
			svc := policyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListThinBySubjectTemplateBeforeExpiredAt(int64(1), int64(0), int64(10))
			assert.Error(GinkgoT(), err)

		})
	})

	Describe("DeleteByPKs cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.Policy{
				{
					PK:           1,
					ExpressionPK: 1,
					TemplateID:   0,
				},
				{
					PK:           2,
					ExpressionPK: 2,
					TemplateID:   1,
				},
			}
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBySubjectPKAndPKs(int64(1), []int64{1, 2}).Return(returned, nil)
			mockPolicyManager.EXPECT().BulkDeleteByTemplatePKsWithTx(
				gomock.Any(), int64(1), int64(0), []int64{1}).Return(int64(1), nil)
			mockExpressionManager := mock.NewMockExpressionManager(ctl)
			mockExpressionManager.EXPECT().BulkDeleteByPKsWithTx(gomock.Any(), []int64{1}).Return(int64(1), nil)

			svc := policyService{
				manager:          mockPolicyManager,
				expressionManger: mockExpressionManager,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			err := svc.DeleteByPKs(int64(1), []int64{1, 2})
			assert.NoError(GinkgoT(), err)

			err = dbMock.ExpectationsWereMet()
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("AlterCustomPolicies cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBySubjectPKAndPKs(int64(1), []int64{1, 2}).Return(
				[]dao.Policy{
					{
						PK:           1,
						SubjectPK:    1,
						ActionPK:     3,
						ExpressionPK: 1,
						ExpiredAt:    1,
					},
					{
						PK:           2,
						SubjectPK:    1,
						ActionPK:     4,
						ExpressionPK: 2,
						ExpiredAt:    1,
						TemplateID:   1,
					},
				}, nil,
			)
			mockPolicyManager.EXPECT().ListBySubjectPKAndPKs(int64(1), []int64{}).Return([]dao.Policy{}, nil)
			mockPolicyManager.EXPECT().BulkDeleteByTemplatePKsWithTx(
				gomock.Any(), int64(1), int64(0), []int64{}).Return(int64(0), nil)
			mockExpressionManager := mock.NewMockExpressionManager(ctl)
			mockExpressionManager.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.Expression{
				{
					Type:       0,
					Expression: "test",
					Signature:  "098f6bcd4621d373cade4e832627b4f6",
				},
				{
					Type:       0,
					Expression: "test",
					Signature:  "098f6bcd4621d373cade4e832627b4f6",
				},
			}).Return([]int64{1, 2}, nil)

			mockPolicyManager.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.Policy{
				{
					SubjectPK:    1,
					ActionPK:     1,
					ExpressionPK: 1,
					ExpiredAt:    1,
				},
				{
					SubjectPK:    1,
					ActionPK:     2,
					ExpressionPK: 2,
					ExpiredAt:    1,
				},
			}).Return(nil)

			mockExpressionManager.EXPECT().BulkUpdateWithTx(gomock.Any(), []dao.Expression{
				{
					PK:         1,
					Type:       0,
					Expression: "test",
					Signature:  "098f6bcd4621d373cade4e832627b4f6",
				},
			}).Return(nil)

			mockExpressionManager.EXPECT().BulkDeleteByPKsWithTx(gomock.Any(), []int64{}).Return(int64(0), nil)

			svc := policyService{
				manager:          mockPolicyManager,
				expressionManger: mockExpressionManager,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			createPolicies := []types.Policy{
				{
					Version:    "1",
					SubjectPK:  1,
					ActionPK:   1,
					Expression: "test",
					Signature:  "",
					ExpiredAt:  1,
				},
				{
					Version:    "1",
					SubjectPK:  1,
					ActionPK:   2,
					Expression: "test",
					Signature:  "",
					ExpiredAt:  1,
				},
			}

			updatePolicies := []types.Policy{
				{
					Version:    "1",
					ID:         1,
					SubjectPK:  1,
					ActionPK:   3,
					Expression: "test",
					Signature:  "",
					ExpiredAt:  1,
				},
				{
					Version:    "1",
					ID:         2,
					SubjectPK:  1,
					ActionPK:   4,
					Expression: "test",
					Signature:  "",
					ExpiredAt:  1,
				},
			}

			set := util.NewInt64Set()
			set.Add(1)
			set.Add(2)

			_, err := svc.AlterCustomPolicies(1, createPolicies, updatePolicies, []int64{}, set)
			assert.NoError(GinkgoT(), err)

			// _, err = dbMock.ExpectationsWereMet()
			err = dbMock.ExpectationsWereMet()
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("ListQueryByPKs cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.Policy{
				{
					PK:           1,
					ExpressionPK: 1,
					TemplateID:   0,
				},
				{
					PK:           2,
					ExpressionPK: 2,
					TemplateID:   1,
				},
			}
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(returned, nil)

			svc := policyService{
				manager: mockPolicyManager,
			}

			policies, err := svc.ListQueryByPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Len(GinkgoT(), policies, 2)
		})

		It("fail", func() {
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(nil, errors.New("list fail"))

			svc := policyService{
				manager: mockPolicyManager,
			}

			_, err := svc.ListQueryByPKs([]int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "manager.ListByPKs")
		})
	})

	Describe("UpdateExpiredAt cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			returned := []dao.Policy{
				{
					PK:           1,
					ExpressionPK: 1,
					TemplateID:   0,
					ExpiredAt:    0,
				},
				{
					PK:           2,
					ExpressionPK: 2,
					TemplateID:   1,
					ExpiredAt:    0,
				},
			}
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(returned, nil)
			mockPolicyManager.EXPECT().BulkUpdateExpiredAtWithTx(gomock.Any(), gomock.Any()).Return(nil)

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			svc := policyService{
				manager: mockPolicyManager,
			}

			err := svc.UpdateExpiredAt([]types.QueryPolicy{{
				PK:           1,
				ExpressionPK: 1,
			}, {
				PK:           2,
				ExpressionPK: 1,
			}})
			assert.NoError(GinkgoT(), err)
		})

		It("ListByPKs fail", func() {
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(nil, errors.New("list fail"))

			svc := policyService{
				manager: mockPolicyManager,
			}

			err := svc.UpdateExpiredAt([]types.QueryPolicy{{
				PK:           1,
				ExpressionPK: 1,
			}, {
				PK:           2,
				ExpressionPK: 1,
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListByPKs")
		})

		It("UpdateExpiredAt", func() {
			returned := []dao.Policy{
				{
					PK:           1,
					ExpressionPK: 1,
					TemplateID:   0,
					ExpiredAt:    0,
				},
				{
					PK:           2,
					ExpressionPK: 2,
					TemplateID:   1,
					ExpiredAt:    0,
				},
			}
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListByPKs([]int64{1, 2}).Return(returned, nil)
			mockPolicyManager.EXPECT().BulkUpdateExpiredAtWithTx(gomock.Any(), gomock.Any()).Return(errors.New("update fail"))

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			svc := policyService{
				manager: mockPolicyManager,
			}

			err := svc.UpdateExpiredAt([]types.QueryPolicy{{
				PK:           1,
				ExpressionPK: 1,
			}, {
				PK:           2,
				ExpressionPK: 1,
			}})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateExpiredAt")
		})
	})

	Describe("CreateAndDeleteTemplatePolicies cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockExpressionManager := mock.NewMockExpressionManager(ctl)
			mockExpressionManager.EXPECT().ListDistinctBySignaturesType(gomock.Any(), int64(1)).Return([]dao.Expression{
				{
					PK:         1,
					Expression: "test",
					Signature:  "098f6bcd4621d373cade4e832627b4f6",
				},
			}, nil)
			mockExpressionManager.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.Expression{
				{
					Type:       1,
					Expression: "expression",
					Signature:  "63973cd3ad7ccf2c8d5dce94b215f683",
				},
			}).Return([]int64{2}, nil)

			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.Policy{
				{
					SubjectPK:    1,
					ActionPK:     1,
					ExpressionPK: 1,
					ExpiredAt:    1,
					TemplateID:   1,
				},
				{
					SubjectPK:    1,
					ActionPK:     2,
					ExpressionPK: 2,
					ExpiredAt:    1,
					TemplateID:   1,
				},
			}).Return(nil)
			mockPolicyManager.EXPECT().BulkDeleteByTemplatePKsWithTx(
				gomock.Any(), int64(1), int64(1), []int64{}).Return(int64(0), nil)

			svc := policyService{
				manager:          mockPolicyManager,
				expressionManger: mockExpressionManager,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			createPolicies := []types.Policy{
				{
					Version:    "1",
					SubjectPK:  1,
					ActionPK:   1,
					Expression: "test",
					Signature:  "",
					ExpiredAt:  1,
					TemplateID: 1,
				},
				{
					Version:    "1",
					SubjectPK:  1,
					ActionPK:   2,
					Expression: "expression",
					Signature:  "",
					ExpiredAt:  1,
					TemplateID: 1,
				},
			}

			set := util.NewInt64Set()
			set.Add(1)
			set.Add(2)

			err := svc.CreateAndDeleteTemplatePolicies(1, 1, createPolicies, []int64{}, set)
			assert.NoError(GinkgoT(), err)

			// _, err = dbMock.ExpectationsWereMet()
			err = dbMock.ExpectationsWereMet()
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("UpdateTemplatePolicies cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockExpressionManager := mock.NewMockExpressionManager(ctl)
			mockExpressionManager.EXPECT().ListDistinctBySignaturesType(gomock.Any(), int64(1)).Return([]dao.Expression{
				{
					PK:         1,
					Expression: "test",
					Signature:  "098f6bcd4621d373cade4e832627b4f6",
				},
			}, nil)
			mockExpressionManager.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.Expression{
				{
					Type:       1,
					Expression: "expression",
					Signature:  "63973cd3ad7ccf2c8d5dce94b215f683",
				},
			}).Return([]int64{2}, nil)

			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().ListBySubjectPKAndPKs(int64(1), []int64{1, 2}).Return(
				[]dao.Policy{
					{
						PK:           1,
						SubjectPK:    1,
						ActionPK:     1,
						ExpressionPK: 3,
						ExpiredAt:    1,
						TemplateID:   1,
					},
					{
						PK:           2,
						SubjectPK:    1,
						ActionPK:     2,
						ExpressionPK: 4,
						ExpiredAt:    1,
						TemplateID:   1,
					},
				}, nil,
			)
			mockPolicyManager.EXPECT().BulkUpdateExpressionPKWithTx(gomock.Any(), []dao.Policy{
				{
					PK:           1,
					SubjectPK:    1,
					ActionPK:     1,
					ExpressionPK: 1,
					ExpiredAt:    1,
					TemplateID:   1,
				},
				{
					PK:           2,
					SubjectPK:    1,
					ActionPK:     2,
					ExpressionPK: 2,
					ExpiredAt:    1,
					TemplateID:   1,
				},
			}).Return(nil)

			svc := policyService{
				manager:          mockPolicyManager,
				expressionManger: mockExpressionManager,
			}

			db, dbMock := database.NewMockSqlxDB()
			dbMock.ExpectBegin()
			dbMock.ExpectCommit()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, db.Beginx)
			defer patches.Reset()

			updatePolicies := []types.Policy{
				{
					Version:    "1",
					ID:         1,
					SubjectPK:  1,
					ActionPK:   1,
					Expression: "test",
					Signature:  "",
					ExpiredAt:  1,
					TemplateID: 1,
				},
				{
					Version:    "1",
					ID:         2,
					SubjectPK:  1,
					ActionPK:   2,
					Expression: "expression",
					Signature:  "",
					ExpiredAt:  1,
					TemplateID: 1,
				},
			}

			set := util.NewInt64Set()
			set.Add(1)
			set.Add(2)

			err := svc.UpdateTemplatePolicies(1, updatePolicies, set)
			assert.NoError(GinkgoT(), err)

			// _, err = dbMock.ExpectationsWereMet()
			err = dbMock.ExpectationsWereMet()
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("DeleteTemplatePolicies cases", func() {
		var ctl *gomock.Controller

		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctl.Finish()
		})

		It("ok", func() {
			mockPolicyManager := mock.NewMockPolicyManager(ctl)
			mockPolicyManager.EXPECT().BulkDeleteBySubjectTemplate(int64(1), int64(1)).Return(nil)

			svc := policyService{
				manager: mockPolicyManager,
			}

			err := svc.DeleteTemplatePolicies(int64(1), int64(1))
			assert.NoError(GinkgoT(), err)
		})
	})

})
