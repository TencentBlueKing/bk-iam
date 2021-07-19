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
	"database/sql/driver"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// NewMockSqlxDB ...
func NewMockSqlxDB() (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.WithError(err).Error("An error was not expected when opening a stub database connection")
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	sqlxDB.SetMaxOpenConns(10)
	return sqlxDB, mock
}

// RunWithMock ...
func RunWithMock(t *testing.T, test func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T)) {
	runner := func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
		test(db, mock, t)
	}

	db, mock := NewMockSqlxDB()
	runner(db, mock, t)

	if err := mock.ExpectationsWereMet(); err != nil {
		log.WithError(err).Error("there were unfulfilled expectations")
	}
}

// NewMockRowsWithoutData ...
func NewMockRowsWithoutData(mock sqlmock.Sqlmock, arg interface{}) *sqlmock.Rows {
	var mockRows *sqlmock.Rows

	// 根据 Struct 的 db 标签，获取 columns
	objType := reflect.TypeOf(arg)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	var columns []string
	for i := 0; i < objType.NumField(); i++ {
		dbTagName := objType.Field(i).Tag.Get("db")
		if dbTagName != "" {
			columns = append(columns, dbTagName)
		}
	}
	// log.Infof("columns len: %d, %+v", len(columns), columns)

	mockRows = sqlmock.NewRows(columns)
	return mockRows
}

// NewMockRows ...
func NewMockRows(mock sqlmock.Sqlmock, args ...interface{}) *sqlmock.Rows {
	mockRows := NewMockRowsWithoutData(mock, args[0])

	objType := reflect.TypeOf(args[0])

	// 获取数据并写入
	for _, obj := range args {
		objValue := reflect.ValueOf(obj)
		if objType.Kind() == reflect.Ptr {
			objValue = objValue.Elem()
		}
		values := []driver.Value{}
		for i := 0; i < objType.NumField(); i++ {
			dbTagName := objType.Field(i).Tag.Get("db")
			if dbTagName != "" {
				values = append(values, objValue.Field(i).Interface())
			}
		}
		// log.Infof("values len: %d, %+v", len(values), values)

		mockRows = mockRows.AddRow(values...)
	}
	return mockRows
}
