/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

/*
 * copy from logrus https://github.com/sirupsen/logrus/blob/master/json_formatter_test.go,
 * modify the assert to testify
 * under MIT License
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 Simon Eskildsen
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package logging

import (
	"errors"
	"fmt"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util/json"
)

// TODO: logrus

// ! copy from logrus, change to testify

func TestErrorNotLost(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("error", errors.New("wild walrus")))
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	assert.Equal(t, entry["error"], "wild walrus")
}

func TestErrorNotLostOnFieldNotNamedError(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("omg", errors.New("wild walrus")))
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	assert.Equal(t, entry["omg"], "wild walrus")
}

func TestFieldClashWithTime(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("time", "right now!"))
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	assert.Equal(t, entry["fields.time"], "right now!")

	assert.Equal(t, entry["time"], "0001-01-01T00:00:00Z")
}

func TestFieldClashWithMsg(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("msg", "something"))
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	assert.Equal(t, entry["fields.msg"], "something")
}

func TestFieldClashWithLevel(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("level", "something"))
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	assert.Equal(t, entry["fields.level"], "something")
}

func TestFieldClashWithRemappedFields(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{
		FieldMap: FieldMap{
			FieldKeyTime:  "@timestamp",
			FieldKeyLevel: "@level",
			FieldKeyMsg:   "@message",
		},
	}

	b, err := formatter.Format(logrus.WithFields(logrus.Fields{
		"@timestamp": "@timestamp",
		"@level":     "@level",
		"@message":   "@message",
		"timestamp":  "timestamp",
		"level":      "level",
		"msg":        "msg",
	}))
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	for _, field := range []string{"timestamp", "level", "msg"} {
		assert.Equal(t, field, entry[field])

		remappedKey := fmt.Sprintf("fields.%s", field)
		assert.NotContains(t, entry, remappedKey)
	}

	for _, field := range []string{"@timestamp", "@level", "@message"} {
		assert.NotEqual(t, field, entry[field])

		remappedKey := fmt.Sprintf("fields.%s", field)

		assert.Contains(t, entry, remappedKey)
		assert.Equal(t, field, entry[remappedKey])
	}
}

func TestFieldsInNestedDictionary(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{
		DataKey: "args",
	}

	logEntry := logrus.WithFields(logrus.Fields{
		"level": "level",
		"test":  "test",
	})
	logEntry.Level = logrus.InfoLevel

	b, err := formatter.Format(logEntry)
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	args := entry["args"].(map[string]interface{})

	for _, field := range []string{"test", "level"} {
		assert.Contains(t, args, field)
		assert.Equal(t, args[field], field)
	}

	for _, field := range []string{"test", "fields.level"} {
		assert.NotContains(t, entry, field)
	}

	// with nested object, "level" shouldn't clash
	assert.Equal(t, "info", entry["level"])
}

func TestJSONEntryEndsWithNewline(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("level", "something"))
	assert.NoError(t, err)

	assert.Equal(t, "\n", string(b[len(b)-1]))
}

func TestJSONMessageKey(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{
		FieldMap: FieldMap{
			FieldKeyMsg: "message",
		},
	}

	b, err := formatter.Format(&logrus.Entry{Message: "oh hai"})
	assert.NoError(t, err)

	s := string(b)
	assert.Contains(t, s, "message")
	assert.Contains(t, s, "oh hai")
}

func TestJSONLevelKey(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{
		FieldMap: FieldMap{
			FieldKeyLevel: "somelevel",
		},
	}

	b, err := formatter.Format(logrus.WithField("level", "something"))
	assert.NoError(t, err)

	s := string(b)
	assert.Contains(t, s, "somelevel")
}

func TestJSONTimeKey(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{
		FieldMap: FieldMap{
			FieldKeyTime: "timeywimey",
		},
	}

	b, err := formatter.Format(logrus.WithField("level", "something"))
	assert.NoError(t, err)

	s := string(b)
	assert.Contains(t, s, "timeywimey")
}

func TestFieldDoesNotClashWithCaller(t *testing.T) {
	t.Parallel()

	logrus.SetReportCaller(false)
	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("func", "howdy pardner"))
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	assert.Equal(t, entry["func"], "howdy pardner")
}

func TestFieldClashWithCaller(t *testing.T) {
	// t.Parallel()

	logrus.SetReportCaller(true)
	formatter := &JSONFormatter{}
	e := logrus.WithField("func", "howdy pardner")
	e.Caller = &runtime.Frame{Function: "somefunc"}
	b, err := formatter.Format(e)
	assert.NoError(t, err)

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	assert.NoError(t, err)

	assert.Equal(t, entry["fields.func"], "howdy pardner")

	assert.Equal(t, entry["func"], "somefunc")

	logrus.SetReportCaller(false) // return to default value
}

func TestJSONDisableTimestamp(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{
		DisableTimestamp: true,
	}

	b, err := formatter.Format(logrus.WithField("level", "something"))
	assert.NoError(t, err)

	s := string(b)
	assert.NotContains(t, s, FieldKeyTime)
}

func TestJSONEnableTimestamp(t *testing.T) {
	t.Parallel()

	formatter := &JSONFormatter{}

	b, err := formatter.Format(logrus.WithField("level", "something"))
	assert.NoError(t, err)

	s := string(b)
	assert.Contains(t, s, FieldKeyTime)
}
