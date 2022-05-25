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
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/go-sql-driver/mysql"
	jsoniter "github.com/json-iterator/go"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

var _ = Describe("Utils", func() {

	Describe("truncateArgs", func() {

		Context("a string", func() {
			var a string
			BeforeEach(func() {
				a = `abc`
			})
			It("less than", func() {
				b := truncateArgs(a, 10)
				assert.Equal(GinkgoT(), `"abc"`, b)
			})
			It("just equals", func() {
				b := truncateArgs(a, 5)
				assert.Equal(GinkgoT(), `"abc"`, b)
			})
			It("greater than", func() {
				b := truncateArgs(a, 2)
				assert.Equal(GinkgoT(), `"a`, b)
			})
		})

		Context("a interface", func() {
			var a []int64
			BeforeEach(func() {
				a = []int64{1, 2, 3, 4, 5, 6}
			})
			It("less than", func() {
				b := truncateArgs(a, 20)
				assert.Equal(GinkgoT(), `[1,2,3,4,5,6]`, b)
			})
			It("just equals", func() {
				b := truncateArgs(a, 22)
				assert.Equal(GinkgoT(), `[1,2,3,4,5,6]`, b)
			})
			It("greater than", func() {
				b := truncateArgs(a, 2)
				assert.Equal(GinkgoT(), `[1`, b)
			})
		})

	})

	Describe("IsMysqlDuplicateEntryError", func() {
		It("true", func() {
			assert.True(GinkgoT(), IsMysqlDuplicateEntryError(&mysql.MySQLError{
				Number: 1062,
			}))
		})

		It("false", func() {
			assert.False(GinkgoT(), IsMysqlDuplicateEntryError(errors.New("error")))
		})

		It("nil false", func() {
			assert.False(GinkgoT(), IsMysqlDuplicateEntryError(nil))
		})
	})

})

func TestLogSlowSQL(t *testing.T) {
	// TODO

}

func TestAllowBlankFields(t *testing.T) {
	a := NewAllowBlankFields()

	assert.False(t, a.HasKey("hello"))

	a.AddKey("hello")
	assert.True(t, a.HasKey("hello"))
}

func TestIsBlank(t *testing.T) {
	assert.True(t, isBlank(reflect.ValueOf("")))
	assert.True(t, isBlank(reflect.ValueOf(false)))
	assert.True(t, isBlank(reflect.ValueOf(0)))
	assert.True(t, isBlank(reflect.ValueOf(0.0)))
}

func TestParseUpdateStruct(t *testing.T) {
	type TestAction struct {
		AllowBlankFields

		System  string `db:"system_id"`
		Name    string `db:"name"`
		NameEn  string `db:"name_en"`
		Version int64  `db:"version"`
	}

	allowBlankFields := NewAllowBlankFields()
	allowBlankFields.AddKey("Name")
	a := TestAction{
		AllowBlankFields: allowBlankFields,
		System:           "test",
		Name:             "",
		NameEn:           "",
		Version:          1,
	}

	expr, data, err := ParseUpdateStruct(a, a.AllowBlankFields)
	assert.Equal(t, "system_id=:system_id, name=:name, version=:version", expr)
	assert.NoError(t, err)

	systemID, ok := data["system_id"]
	assert.True(t, ok)
	assert.Equal(t, "test", systemID)

	name, ok := data["name"]
	assert.True(t, ok)
	assert.Equal(t, "", name)

	_, ok = data["name_en"]
	assert.False(t, ok)

	version, ok := data["version"]
	assert.True(t, ok)
	assert.Equal(t, int64(1), version)
}

func truncateInterface(v interface{}) string {
	s := fmt.Sprintf("%v", v)
	return stringx.Truncate(s, 10)
}

func truncateInterfaceViaJSON(v interface{}) string {
	s, err := jsoniter.MarshalToString(v)
	if err != nil {
		s = fmt.Sprintf("%v", v)
	}
	return stringx.Truncate(s, 10)
}
func truncateInterfaceViaJSONToBytes(v interface{}) string {
	s, err := jsoniter.Marshal(v)
	if err != nil {
		s = []byte(fmt.Sprintf("%v", v))
	}
	return util.TruncateBytesToString(s, 10)
}

func BenchmarkTruncateInterface(b *testing.B) {
	x := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateInterface(x)
	}
}

func BenchmarkTruncateInterfaceViaJson(b *testing.B) {
	x := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateInterfaceViaJSON(x)
	}
}

func BenchmarkTruncateInterfaceViaJsonToBytes(b *testing.B) {
	x := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateInterfaceViaJSONToBytes(x)
	}
}
