/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dao

import (
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

var _ = Describe("GroupResourcePolicyManager", func() {
	var (
		mock    sqlmock.Sqlmock
		db      *sqlx.DB
		manager GroupResourcePolicyManager
		policy  GroupResourcePolicy
	)
	BeforeEach(func() {
		db, mock = database.NewMockSqlxDB()
		manager = &groupResourcePolicyManager{DB: db}
		policy = GroupResourcePolicy{
			GroupPK:                     int64(1),
			TemplateID:                  int64(2),
			SystemID:                    "test",
			ActionPKs:                   "[1,2,3]",
			ActionRelatedResourceTypePK: int64(3),
			ResourceTypePK:              int64(4),
			ResourceID:                  "resource_id",
		}
		policy.Signature = stringx.MD5Hash(
			fmt.Sprintf(
				"%d:%d:%s:%d:%d:%s",
				policy.GroupPK, policy.TemplateID, policy.SystemID,
				policy.ActionRelatedResourceTypePK, policy.ResourceTypePK, policy.ResourceID,
			),
		)
	})

	It("ListBySignatures", func() {
		policy.PK = int64(1)
		mockRows := database.NewMockRows(mock, []interface{}{policy}...)
		mock.ExpectQuery(
			"^SELECT pk, signature, group_pk, template_id, system_id," +
				" action_pks, action_related_resource_type_pk, resource_type_pk, resource_id" +
				" FROM rbac_group_resource_policy WHERE signature IN (.*)$",
		).WithArgs(policy.Signature).WillReturnRows(mockRows)

		policies, err := manager.ListBySignatures([]string{policy.Signature})

		assert.NoError(GinkgoT(), err)
		assert.Len(GinkgoT(), policies, 1)
		assert.Equal(GinkgoT(), int64(1), policies[0].PK)
		assert.Equal(GinkgoT(), "[1,2,3]", policies[0].ActionPKs)
	})

	It("BulkCreateWithTx", func() {
		mock.ExpectBegin()
		mock.ExpectExec(
			`INSERT INTO rbac_group_resource_policy`,
		).WithArgs(
			policy.Signature, policy.GroupPK, policy.TemplateID, policy.SystemID,
			policy.ActionPKs, policy.ActionRelatedResourceTypePK, policy.ResourceTypePK, policy.ResourceID,
		).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(GinkgoT(), err)

		err = manager.BulkCreateWithTx(tx, []GroupResourcePolicy{policy})
		tx.Commit()

		assert.NoError(GinkgoT(), err)
	})

	It("BulkUpdateActionPKsWithTx", func() {
		policy.PK = int64(1)

		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE rbac_group_resource_policy SET action_pks = (.*) WHERE pk = (.*)`)
		mock.ExpectExec(`UPDATE rbac_group_resource_policy SET action_pks = (.*) WHERE pk = (.*)`).
			WithArgs(policy.ActionPKs, policy.PK).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(GinkgoT(), err)

		err = manager.BulkUpdateActionPKsWithTx(tx, []GroupResourcePolicy{policy})
		tx.Commit()

		assert.NoError(GinkgoT(), err)
	})

	It("BulkDeleteByPKsWithTx", func() {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM rbac_group_resource_policy WHERE pk IN (.*)`).
			WithArgs(int64(1), int64(2)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		tx, err := db.Beginx()
		assert.NoError(GinkgoT(), err)

		err = manager.BulkDeleteByPKsWithTx(tx, []int64{int64(1), int64(2)})
		tx.Commit()

		assert.NoError(GinkgoT(), err)
	})

	It("ListThinByResource", func() {
		thinPolicy := ThinGroupResourcePolicy{
			GroupPK:   int64(1),
			ActionPKs: "[1,2,3]",
		}
		mockRows := database.NewMockRows(mock, []interface{}{thinPolicy}...)
		mock.ExpectQuery(
			`^SELECT 
			group_pk, action_pks
			FROM rbac_group_resource_policy
			WHERE system_id = (.*)
			AND action_related_resource_type_pk = (.*)
			AND resource_type_pk = (.*)
			AND resource_id = (.*)$`,
		).WithArgs("test", int64(1), int64(2), "resource_test").WillReturnRows(mockRows)

		policies, err := manager.ListThinByResource("test", int64(1), int64(2), "resource_test")

		assert.NoError(GinkgoT(), err)
		assert.Len(GinkgoT(), policies, 1)
		assert.Equal(GinkgoT(), int64(1), policies[0].GroupPK)
		assert.Equal(GinkgoT(), "[1,2,3]", policies[0].ActionPKs)
	})
})
