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
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"iam/pkg/logging"
	"iam/pkg/util/json"
)

const (
	ArgsTruncateLength = 4096
)

// ============== tx Rollback Log ==============

// RollBackWithLog will rollback and log if error
func RollBackWithLog(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err != sql.ErrTxDone && err != nil {
		logging.GetSQLLogger().Error(err)
	}
}

// ============== slow sql logger ==============
func logSlowSQL(start time.Time, query string, args interface{}) {
	elapsed := time.Since(start)
	// to ms
	latency := float64(elapsed / time.Millisecond)

	// current, set 20ms
	if logging.GetSQLLogger().GetLevel() == log.DebugLevel || latency > 20 {
		logging.GetSQLLogger().WithFields(log.Fields{
			// replace \n\t\t: the sql in the database module
			"sql": strings.ReplaceAll(query, "\n\t\t", " "),
			// truncate the args
			"args":    truncateArgs(args, ArgsTruncateLength),
			"latency": latency,
		}).Info("-")
	}
}

func truncateArgs(args interface{}, length int) string {
	s, err := json.MarshalToString(args)
	if err != nil {
		s = fmt.Sprintf("%v", args)
	}
	return stringx.Truncate(s, length)
}

func isBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}

	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

// AllowBlankFields store the fields of the struct which allow blank
// NOTE: the key is the field name in the struct, not the db tag!
type AllowBlankFields struct {
	keys map[string]struct{}
}

// NewAllowBlankFields create a allow fields
func NewAllowBlankFields() AllowBlankFields {
	return AllowBlankFields{keys: map[string]struct{}{}}
}

// HasKey check if key exist in allowed fields
func (a *AllowBlankFields) HasKey(key string) bool {
	_, ok := a.keys[key]
	return ok
}

// AddKey add a key into allowed fields
func (a *AllowBlankFields) AddKey(key string) {
	a.keys[key] = struct{}{}
}

// ParseUpdateStruct parse a struct into updated fields
func ParseUpdateStruct(values interface{}, allowBlankFields AllowBlankFields) (string, map[string]interface{}, error) {
	var setFields []string
	updateData := map[string]interface{}{}

	// TODO: allowBlankFields maybe nil?

	v := reflect.ValueOf(values)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			dbField := v.Type().Field(i).Tag.Get("db")
			if dbField == "" {
				continue
			}

			name := v.Type().Field(i).Name

			value := v.FieldByName(name)
			// TODO: should not be the id? or some other field?
			if !isBlank(value) || allowBlankFields.HasKey(name) {
				setFields = append(setFields, fmt.Sprintf("%s=:%s", dbField, dbField))
				updateData[dbField] = v.FieldByName(name).Interface()
			}
		}
	}

	setExpr := strings.Join(setFields, ", ")

	return setExpr, updateData, nil
}

// IsMysqlDuplicateEntryError ...
func IsMysqlDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	return false
}
