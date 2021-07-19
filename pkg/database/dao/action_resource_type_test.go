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

	"iam/pkg/database"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func Test_actionResourceTypeManager_ListResourceTypeByAction(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			ActionResourceType{
				ResourceTypeSystem: "bk_cmdb",
				ResourceTypeID:     "host",
			},
			ActionResourceType{
				ResourceTypeSystem: "bk_job",
				ResourceTypeID:     "job",
			},
		}
		mockQuery := `^SELECT resource_type_system_id, resource_type_id
			FROM action_resource_type WHERE action_system_id = (.*) AND action_id = (.*)$`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs("bk_job", "job_execute").WillReturnRows(mockRows)

		manager := &actionResourceTypeManager{DB: db}
		actionResourceTypes, err := manager.ListResourceTypeByAction("bk_job", "job_execute")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(actionResourceTypes), 2)
		assert.Equal(t, actionResourceTypes[0], mockData[0].(ActionResourceType))
		assert.Equal(t, actionResourceTypes[1], mockData[1].(ActionResourceType))
	})
}

func Test_actionResourceTypeManager_ListByActionSystem(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockData := []interface{}{
			ActionResourceType{
				ResourceTypeSystem: "bk_cmdb",
				ResourceTypeID:     "host",
			},
			ActionResourceType{
				ResourceTypeSystem: "bk_job",
				ResourceTypeID:     "job",
			},
		}
		mockQuery := `^SELECT (.*) FROM action_resource_type WHERE action_system_id = (.*)
			ORDER BY pk$`
		mockRows := database.NewMockRows(mock, mockData...)
		mock.ExpectQuery(mockQuery).WithArgs("bk_job").WillReturnRows(mockRows)

		manager := &actionResourceTypeManager{DB: db}
		actionResourceTypes, err := manager.ListByActionSystem("bk_job")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, len(actionResourceTypes), 2)
		assert.Equal(t, actionResourceTypes[0], mockData[0].(ActionResourceType))
		assert.Equal(t, actionResourceTypes[1], mockData[1].(ActionResourceType))
	})
}
