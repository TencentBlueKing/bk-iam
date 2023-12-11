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
	"reflect"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/database/dao/mock"
	"iam/pkg/service/types"
)

var _ = Describe("GroupService", func() {
	Describe("ListGroupMember", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListGroupMember fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListGroupMember(int64(1)).Return(
				nil, errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListGroupMember(1)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListGroupMember")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListGroupMember(int64(1)).Return(
				[]dao.SubjectRelation{}, nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			groupMembers, err := manager.ListGroupMember(1)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.GroupMember{}, groupMembers)
		})
	})

	Describe("BulkDeleteBySubjectPKsWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.BulkDeleteBySubjectPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)

			mockSubjectService.EXPECT().BulkDeleteBySubjectPKs(gomock.Any(), []int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteBySubjectPKs")
		})

		It("subjectSystemGroupManager.DeleteBySubjectPKsWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)

			mockSubjectService.EXPECT().BulkDeleteBySubjectPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectSystemGroupService := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupService.EXPECT().DeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager:                   mockSubjectService,
				subjectSystemGroupManager: mockSubjectSystemGroupService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "DeleteBySubjectPKsWithTx")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)

			mockSubjectService.EXPECT().BulkDeleteBySubjectPKs(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			mockSubjectSystemGroupService := mock.NewMockSubjectSystemGroupManager(ctl)
			mockSubjectSystemGroupService.EXPECT().DeleteBySubjectPKsWithTx(gomock.Any(), []int64{1, 2}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				manager:                   mockSubjectService,
				subjectSystemGroupManager: mockSubjectSystemGroupService,
			}

			err := manager.BulkDeleteBySubjectPKsWithTx(nil, []int64{1, 2})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("UpdateGroupMembersExpiredAtWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkUpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.UpdateGroupMembersExpiredAtWithTx(nil, int64(1), []types.SubjectTemplateGroup{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "UpdateExpiredAtWithTx")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkUpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			patches := gomonkey.ApplyMethod(reflect.TypeOf(manager), "ListGroupAuthSystemIDs",
				func(s *groupService, groupPK int64) ([]string, error) {
					return []string{}, nil
				})
			defer patches.Reset()

			err := manager.UpdateGroupMembersExpiredAtWithTx(nil, int64(1), []types.SubjectTemplateGroup{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkCreateGroupMembersWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   2,
					ExpiredAt: 3,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.BulkCreateGroupMembersWithTx(nil, int64(1), []types.SubjectTemplateGroup{
				{
					SubjectPK: 1,
					GroupPK:   2,
					ExpiredAt: 3,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreateWithTx")
		})

		It("manager.UpdateExpiredAtWithTx ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   2,
					ExpiredAt: 3,
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			patches := gomonkey.ApplyMethod(reflect.TypeOf(manager), "ListGroupAuthSystemIDs",
				func(s *groupService, groupPK int64) ([]string, error) {
					return []string{}, nil
				})
			defer patches.Reset()

			err := manager.BulkCreateGroupMembersWithTx(nil, int64(1), []types.SubjectTemplateGroup{
				{
					SubjectPK: 1,
					GroupPK:   2,
					ExpiredAt: 3,
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkDeleteGroupMembers", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByGroupMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
				int64(0), errors.New("error"),
			)

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})
			defer patches.Reset()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.BulkDeleteGroupMembers(int64(1), []int64{2}, []int64{3})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteByGroupMembersWithTx")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkDeleteByGroupMembersWithTx(gomock.Any(), int64(1), []int64{2}).Return(
				int64(1), nil,
			)

			mockSubjectService.EXPECT().BulkDeleteByGroupMembersWithTx(gomock.Any(), int64(1), []int64{3}).Return(
				int64(1), nil,
			)

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().
				GetMaxExpiredAtBySubjectGroup(gomock.Any(), gomock.Any(), int64(0)).
				Return(
					time.Now().Unix()+10, nil,
				).
				AnyTimes()

			db, mock := database.NewMockSqlxDB()
			mock.ExpectBegin()
			mock.ExpectCommit()
			tx, _ := db.Beginx()

			patches := gomonkey.ApplyFunc(database.GenerateDefaultDBTx, func() (*sqlx.Tx, error) {
				return tx, nil
			})

			manager := &groupService{
				manager:                     mockSubjectService,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			patches.ApplyMethod(reflect.TypeOf(manager), "ListGroupAuthSystemIDs",
				func(s *groupService, groupPK int64) ([]string, error) {
					return []string{}, nil
				})
			defer patches.Reset()

			_, err := manager.BulkDeleteGroupMembers(int64(1), []int64{2}, []int64{3})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("ListThinSubjectGroupsBySubjectPKs", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListThinRelationAfterExpiredAtBySubjectPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListThinRelationAfterExpiredAtBySubjectPKs([]int64{1, 2}, gomock.Any()).Return(
				nil, errors.New("error"),
			)

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListEffectThinSubjectGroupsBySubjectPKs([]int64{1, 2})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListThinRelationAfterExpiredAtBySubjectPKs")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().ListThinRelationAfterExpiredAtBySubjectPKs([]int64{1, 2}, gomock.Any()).Return(
				[]dao.ThinSubjectRelation{
					{
						GroupPK:   1,
						ExpiredAt: 2,
					},
				}, nil,
			)

			manager := &groupService{
				manager: mockSubjectService,
			}

			subjectGroups, err := manager.ListEffectThinSubjectGroupsBySubjectPKs([]int64{1, 2})
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), []types.ThinSubjectGroup{{GroupPK: 1, ExpiredAt: 2}}, subjectGroups)
		})
	})

	Describe("ListEffectThinSubjectGroupsBySubjectPKGroupPKs", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.ListEffectThinSubjectGroupsBySubjectPKGroupPKs fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				ListRelationBySubjectPKGroupPKs(int64(123), []int64{1}).
				Return(
					nil, errors.New("error"),
				).
				AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			_, err := manager.ListEffectSubjectGroupsBySubjectPKGroupPKs(int64(123), []int64{1})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListRelationBySubjectPKGroupPKs")
		})

		It("success", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				ListRelationBySubjectPKGroupPKs(int64(123), []int64{1}).
				Return(
					[]dao.SubjectRelation{{
						SubjectPK: 123,
						GroupPK:   1,
						ExpiredAt: 1,
					}}, nil,
				).
				AnyTimes()

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().ListRelationBySubjectPKGroupPKs(int64(123), []int64{1}).
				Return(
					[]dao.SubjectTemplateGroup{{
						SubjectPK: 123,
						GroupPK:   1,
						ExpiredAt: 1,
					}, {
						SubjectPK: 123,
						GroupPK:   2,
						ExpiredAt: 1,
					}}, nil,
				).
				AnyTimes()

			manager := &groupService{
				manager:                     mockSubjectService,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			subjectGroups, err := manager.ListEffectSubjectGroupsBySubjectPKGroupPKs(int64(123), []int64{1})
			assert.NoError(GinkgoT(), err)
			assert.ElementsMatch(GinkgoT(), []types.SubjectGroup{{
				GroupPK:   1,
				ExpiredAt: 1,
			}, {
				GroupPK:   2,
				ExpiredAt: 1,
			}}, subjectGroups)
		})
	})

	Describe("GetMaxExpiredAtBySubjectGroup", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.GetExpiredAtBySubjectGroup fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(
					int64(0), errors.New("error"),
				)

			manager := &groupService{
				manager: mockSubjectService,
			}

			expiredAt, err := manager.GetMaxExpiredAtBySubjectGroup(int64(1), int64(2), int64(0))
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "GetExpiredAtBySubjectGroup")
			assert.Equal(GinkgoT(), int64(0), expiredAt)
		})

		It("not found", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(
					int64(0), sql.ErrNoRows,
				)

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2), int64(0)).
				Return(
					int64(0), sql.ErrNoRows,
				)

			manager := &groupService{
				manager:                     mockSubjectService,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			expiredAt, err := manager.GetMaxExpiredAtBySubjectGroup(int64(1), int64(2), int64(0))
			assert.Error(GinkgoT(), err)
			assert.True(GinkgoT(), errors.Is(err, ErrGroupMemberNotFound))
			assert.Equal(GinkgoT(), int64(0), expiredAt)
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(
					int64(10), nil,
				)

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2), int64(0)).
				Return(
					int64(0), sql.ErrNoRows,
				)

			manager := &groupService{
				manager:                     mockSubjectService,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			expiredAt, err := manager.GetMaxExpiredAtBySubjectGroup(int64(1), int64(2), int64(0))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(10), expiredAt)
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().
				GetExpiredAtBySubjectGroup(int64(1), int64(2)).
				Return(
					int64(1), nil,
				)

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().
				GetMaxExpiredAtBySubjectGroup(int64(1), int64(2), int64(0)).
				Return(
					int64(10), nil,
				)

			manager := &groupService{
				manager:                     mockSubjectService,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			expiredAt, err := manager.GetMaxExpiredAtBySubjectGroup(int64(1), int64(2), int64(0))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), int64(10), expiredAt)
		})
	})

	Describe("BulkCreateSubjectTemplateGroupWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.BulkCreateWithTx fail", func() {
			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			err := manager.BulkCreateSubjectTemplateGroupWithTx(nil, []types.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkCreateWithTx")
		})

		It("manager.BulkCreateWithTx ok", func() {
			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().BulkCreateWithTx(gomock.Any(), []dao.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			patches := gomonkey.ApplyMethod(reflect.TypeOf(manager), "ListGroupAuthSystemIDs",
				func(s *groupService, groupPK int64) ([]string, error) {
					return []string{}, nil
				})
			defer patches.Reset()

			err := manager.BulkCreateSubjectTemplateGroupWithTx(nil, []types.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("UpdateSubjectGroupExpiredAtWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.UpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkUpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				manager: mockSubjectService,
			}

			err := manager.UpdateSubjectGroupExpiredAtWithTx(nil, []types.SubjectTemplateGroup{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "manager.BulkUpdateExpiredAtWithTx")
		})

		It("subjectTemplateGroupManager.BulkUpdateExpiredAtWithTx fail", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkUpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}).Return(
				nil,
			).AnyTimes()

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().
				BulkUpdateExpiredAtByRelationWithTx(gomock.Any(), []dao.SubjectRelation{
					{
						SubjectPK: 1,
						GroupPK:   1,
						ExpiredAt: 2,
					},
				}).
				Return(
					errors.New("error"),
				).
				AnyTimes()

			manager := &groupService{
				manager:                     mockSubjectService,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			err := manager.UpdateSubjectGroupExpiredAtWithTx(nil, []types.SubjectTemplateGroup{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}, true)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subjectTemplateGroupManager.BulkUpdateExpiredAtByRelationWithTx")
		})

		It("ok", func() {
			mockSubjectService := mock.NewMockSubjectGroupManager(ctl)
			mockSubjectService.EXPECT().BulkUpdateExpiredAtWithTx(gomock.Any(), []dao.SubjectRelation{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}).Return(
				nil,
			).AnyTimes()

			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().
				BulkUpdateExpiredAtByRelationWithTx(gomock.Any(), []dao.SubjectRelation{
					{
						SubjectPK: 1,
						GroupPK:   1,
						ExpiredAt: 2,
					},
				}).
				Return(
					nil,
				).
				AnyTimes()

			manager := &groupService{
				manager:                     mockSubjectService,
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			err := manager.UpdateSubjectGroupExpiredAtWithTx(nil, []types.SubjectTemplateGroup{
				{
					SubjectPK: 1,
					GroupPK:   1,
					ExpiredAt: 2,
				},
			}, true)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("BulkDeleteSubjectTemplateGroupWithTx", func() {
		var ctl *gomock.Controller
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())
		})
		AfterEach(func() {
			ctl.Finish()
		})

		It("manager.BulkDeleteWithTx fail", func() {
			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().BulkDeleteWithTx(gomock.Any(), []dao.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			}).Return(
				errors.New("error"),
			).AnyTimes()

			manager := &groupService{
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			err := manager.BulkDeleteSubjectTemplateGroupWithTx(nil, []types.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			})
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "BulkDeleteWithTx")
		})

		It("manager.BulkDeleteWithTx ok", func() {
			mockSubjectTemplateGroupManager := mock.NewMockSubjectTemplateGroupManager(ctl)
			mockSubjectTemplateGroupManager.EXPECT().BulkDeleteWithTx(gomock.Any(), []dao.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			}).Return(
				nil,
			).AnyTimes()

			manager := &groupService{
				subjectTemplateGroupManager: mockSubjectTemplateGroupManager,
			}

			patches := gomonkey.ApplyMethod(reflect.TypeOf(manager), "ListGroupAuthSystemIDs",
				func(s *groupService, groupPK int64) ([]string, error) {
					return []string{}, nil
				})
			defer patches.Reset()

			err := manager.BulkDeleteSubjectTemplateGroupWithTx(nil, []types.SubjectTemplateGroup{
				{
					SubjectPK:  1,
					TemplateID: 2,
					GroupPK:    2,
					ExpiredAt:  3,
				},
			})
			assert.NoError(GinkgoT(), err)
		})
	})
})

var _ = Describe("SubjectSystemGroupHelper", func() {
	var helper *subjectSystemGroupHelper

	BeforeEach(func() {
		helper = &subjectSystemGroupHelper{
			subjectSystemGroup: make(map[string]map[int64]int64),
		}
	})

	Describe("Adding groups", func() {
		Context("When adding a new group", func() {
			It("should add the group correctly", func() {
				helper.Add(1, "system1", 1, 1000)
				expected := map[string]map[int64]int64{
					"1:system1": {1: 1000},
				}
				assert.Equal(GinkgoT(), helper.subjectSystemGroup, expected)
			})
		})

		Context("When adding multiple groups", func() {
			It("should add the groups correctly", func() {
				helper.Add(1, "system1", 1, 1000)
				helper.Add(1, "system1", 2, 2000)
				expected := map[string]map[int64]int64{
					"1:system1": {
						1: 1000,
						2: 2000,
					},
				}
				assert.Equal(GinkgoT(), helper.subjectSystemGroup, expected)
			})
		})
	})

	Describe("Generating key", func() {
		Context("With valid subjectPK and systemID", func() {
			It("should generate the correct key", func() {
				key := helper.generateKey(1, "system1")
				assert.Equal(GinkgoT(), key, "1:system1")
			})
		})
	})

	Describe("Parsing key", func() {
		Context("With a valid key", func() {
			It("should parse the key correctly", func() {
				subjectPK, systemID, err := helper.ParseKey("1:system1")
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), subjectPK, int64(1))
				assert.Equal(GinkgoT(), systemID, "system1")
			})
		})

		Context("With an invalid key", func() {
			It("should return an error", func() {
				_, _, err := helper.ParseKey("invalid")
				assert.Error(GinkgoT(), err)
			})
		})
	})
})
