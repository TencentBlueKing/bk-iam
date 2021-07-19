/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package database

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func Test_bindArray(t *testing.T) {
	type Place struct {
		Country string `db:"country"`
		TelCode string `db:"telcode"`
	}
	RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		query := "INSERT INTO place (country, telcode) VALUES (:country, :telcode)"
		places := []Place{{Country: "china", TelCode: "86"}, {Country: "us", TelCode: "001"}}

		q, args, err := bindArray(sqlx.BindType(db.DriverName()), query, places, db.Mapper)
		assert.NoError(t, err)
		assert.Equal(t, q, "INSERT INTO place (country, telcode) VALUES (?, ?),(?, ?)")
		assert.Equal(t, args, []interface{}{"china", "86", "us", "001"})
	})
}
