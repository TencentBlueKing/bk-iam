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
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

func Test_engineRbacPolicyManager_GetMaxPKBeforeUpdatedAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		now := int64(1617457847)

		mockRows := sqlmock.NewRows([]string{"MAX(pk)"}).AddRow(int64(1))
		mock.ExpectQuery(
			`SELECT .* FROM rbac_group_resource_policy WHERE updated_at <= .*`,
		).WithArgs(now).WillReturnRows(mockRows)

		manager := &engineRbacPolicyManager{DB: db}
		pk, err := manager.GetMaxPKBeforeUpdatedAt(now)

		assert.Equal(t, int64(1), pk)
		assert.NoError(t, err)
	})
}

func Test_engineRbacPolicyManager_ListPKBetweenUpdatedAt(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		begin := time.Now().Unix()
		end := time.Now().Unix()

		mockRows := sqlmock.NewRows([]string{"pk"}).AddRow(int64(1)).AddRow(int64(2))
		mock.ExpectQuery(
			`SELECT pk FROM rbac_group_resource_policy WHERE updated_at BETWEEN .* AND .*`,
		).WithArgs(begin, end).WillReturnRows(mockRows)

		manager := &engineRbacPolicyManager{DB: db}
		pks, err := manager.ListPKBetweenUpdatedAt(begin, end)

		assert.Equal(t, []int64{1, 2}, pks)
		assert.NoError(t, err)
	})
}

func Test_engineRbacPolicyManager_ListBetweenPK(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		now := time.Unix(1617457847, 0)

		mockRows := sqlmock.NewRows([]string{
			"pk", "group_pk", "system_id", "template_id", "action_pks", "action_related_resource_type_pk", "resource_type_pk", "resource_id", "updated_at",
		}).AddRow(int64(1), int64(1), "", int64(1), "", int64(1), int64(1), "", now)
		mock.ExpectQuery(
			`SELECT
			pk,
			group_pk,
			system_id,
			template_id,
			action_pks,
			action_related_resource_type_pk,
			resource_type_pk,
			resource_id,
			updated_at
			FROM rbac_group_resource_policy
			WHERE pk BETWEEN .* AND .*`,
		).WithArgs(int64(1), int64(100)).WillReturnRows(mockRows)

		manager := &engineRbacPolicyManager{DB: db}
		policies, err := manager.ListBetweenPK(
			int64(1), int64(100),
		)

		expected := EngineRbacPolicy{
			PK:                          int64(1),
			GroupPK:                     int64(1),
			TemplateID:                  int64(1),
			SystemID:                    "",
			ActionPKs:                   "",
			ActionRelatedResourceTypePK: int64(1),
			ResourceTypePK:              int64(1),
			ResourceID:                  "",
			UpdatedAt:                   now,
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, policies[0])
	})
}

func Test_engineRbacPolicyManager_ListByPKs(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		now := time.Unix(1617457847, 0)

		mockRows := sqlmock.NewRows([]string{
			"pk", "group_pk", "system_id", "template_id", "action_pks", "action_related_resource_type_pk", "resource_type_pk", "resource_id", "updated_at",
		}).AddRow(int64(1), int64(1), "", int64(1), "", int64(1), int64(1), "", now)
		mock.ExpectQuery(
			`SELECT
			pk,
			group_pk,
			system_id,
			template_id,
			action_pks,
			action_related_resource_type_pk,
			resource_type_pk,
			resource_id,
			updated_at
			FROM rbac_group_resource_policy
			WHERE pk IN`,
		).WithArgs(int64(1), int64(2)).WillReturnRows(mockRows)

		manager := &engineRbacPolicyManager{DB: db}
		policies, err := manager.ListByPKs([]int64{1, 2})

		expected := EngineRbacPolicy{
			PK:                          int64(1),
			GroupPK:                     int64(1),
			TemplateID:                  int64(1),
			SystemID:                    "",
			ActionPKs:                   "",
			ActionRelatedResourceTypePK: int64(1),
			ResourceTypePK:              int64(1),
			ResourceID:                  "",
			UpdatedAt:                   now,
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, policies[0])
	})
}
