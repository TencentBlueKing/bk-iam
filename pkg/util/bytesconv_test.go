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
	"math/rand"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/util"
)

var testString = "Albert Einstein: Logic will get you from A to B. Imagination will take you everywhere."
var testBytes = []byte(testString)

func rawBytesToStr(b []byte) string {
	return string(b)
}

func rawStrToBytes(s string) []byte {
	return []byte(s)
}

const letterBytesForTest = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytesForTest) {
			sb.WriteByte(letterBytesForTest[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

var _ = Describe("Bytesconv", func() {

	Describe("StringToBytes", func() {
		It("a normal string", func() {
			b := util.StringToBytes("abc")
			assert.Equal(GinkgoT(), []byte("abc"), b)
		})

		It("random generate", func() {
			for i := 0; i < 100; i++ {
				s := RandStringBytesMaskImprSrcSB(64)

				assert.Equal(GinkgoT(), rawStrToBytes(s), util.StringToBytes(s))
			}
		})
	})

	Describe("BytesToString", func() {
		It("a normal bytes", func() {
			s := util.BytesToString([]byte("abc"))
			assert.Equal(GinkgoT(), "abc", s)
		})

		It("random generate", func() {
			data := make([]byte, 1024)
			for i := 0; i < 100; i++ {
				rand.Read(data)
				assert.Equal(GinkgoT(), rawBytesToStr(data), util.BytesToString(data))
			}
		})
	})
})

// go test -v -run=none -bench=^BenchmarkBytesConv -benchmem=true

func BenchmarkBytesConvBytesToStrRaw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rawBytesToStr(testBytes)
	}
}

func BenchmarkBytesConvBytesToStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		util.BytesToString(testBytes)
	}
}

func BenchmarkBytesConvStrToBytesRaw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rawStrToBytes(testString)
	}
}

func BenchmarkBytesConvStrToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		util.StringToBytes(testString)
	}
}
