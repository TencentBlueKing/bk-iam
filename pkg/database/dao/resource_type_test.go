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
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

var _ = Describe("", func() {
	var (
		mock    sqlmock.Sqlmock
		db      *sqlx.DB
		manager ResourceTypeManager
	)
	BeforeEach(func() {
		db, mock = database.NewMockSqlxDB()
		manager = &resourceTypeManager{DB: db}
	})
	It("ListByIDs", func() {
		mockRows := database.NewMockRows(mock, []interface{}{
			ResourceType{
				PK:     int64(1),
				System: "system_id",
				ID:     "id",
			},
		}...)
		mock.ExpectQuery(
			"^SELECT pk, system_id, id FROM resource_type WHERE system_id = (.*) AND id IN (.*)$",
		).WithArgs("system_id", "id").WillReturnRows(mockRows)

		resourceTypes, err := manager.ListByIDs("system_id", []string{"id"})
		assert.NoError(GinkgoT(), err)
		assert.Len(GinkgoT(), resourceTypes, 1)
		assert.Equal(GinkgoT(), int64(1), resourceTypes[0].PK)
	})
})
