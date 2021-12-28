/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util_test

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

var _ = Describe("String", func() {

	Describe("TruncateBytes", func() {
		var s = []byte("helloworld")

		DescribeTable("TruncateBytes cases", func(expected []byte, truncatedSize int) {
			assert.Equal(GinkgoT(), expected, util.TruncateBytes(s, truncatedSize))
		},
			Entry("truncated size less than real size", []byte("he"), 2),
			Entry("truncated size equals to real size", s, 10),
			Entry("truncated size greater than real size", s, 20),
		)
	})

	Describe("TruncateBytesToString", func() {
		var s = []byte("helloworld")
		var sStr = string(s)

		DescribeTable("TruncateBytesToString cases", func(expected string, truncatedSize int) {
			assert.Equal(GinkgoT(), expected, util.TruncateBytesToString(s, truncatedSize))
		},
			Entry("truncated size less than real size", "he", 2),
			Entry("truncated size equals to real size", sStr, 10),
			Entry("truncated size greater than real size", sStr, 20),
		)
	})

})

func BenchmarkStringSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("iam:%s", "policies")
	}
}

func BenchmarkStringConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "iam:" + "policies"
	}
}

func BenchmarkIntStringSprintfD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%d", 123)
	}
}

func BenchmarkIntToStringItoa(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strconv.Itoa(123)
	}
}

func BenchmarkIntToStringFormatInt(b *testing.B) {
	x := int64(123)
	for i := 0; i < b.N; i++ {
		strconv.FormatInt(x, 10)
	}
}

func BenchmarkStringAddIntSprintf(b *testing.B) {
	x := int64(123)
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%s:%d", "abc", x)
	}
}

func BenchmarkStringAddIntFormatInt(b *testing.B) {
	x := int64(123)
	for i := 0; i < b.N; i++ {
		_ = "abc" + strconv.FormatInt(x, 10)
	}
}
