/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package sdao

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"iam/pkg/database"
)

func Test_actionManager_GetPK(t *testing.T) {
	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		mockQuery := `^SELECT auth_type FROM saas_action WHERE system_id = (.*) AND id = (.*) LIMIT 1`
		mockRows := sqlmock.NewRows([]string{"auth_type"}).
			AddRow("abac")
		mock.ExpectQuery(mockQuery).WithArgs("iam", "edit").WillReturnRows(mockRows)

		manager := &saasActionManager{DB: db}
		authType, err := manager.GetAuthType("iam", "edit")

		assert.NoError(t, err, "query from db fail.")
		assert.Equal(t, authType, "abac")
	})
}
