/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package middleware

import (
	"errors"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"

	"iam/pkg/cache/impls"
)

var _ = Describe("client_jwt", func() {

	Describe("getClientIDFromJWTToken", func() {
		var patches *gomonkey.Patches
		BeforeEach(func() {
		})
		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})
		It("hit cache", func() {
			patches = gomonkey.ApplyFunc(impls.GetJWTTokenClientID, func(string) (string, error) {
				return "abc", nil
			})

			clientID, err := getClientIDFromJWTToken("aaa", []byte(""))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "abc", clientID)
		})

		It("miss cache, error in verify", func() {
			patches = gomonkey.ApplyFunc(impls.GetJWTTokenClientID, func(string) (string, error) {
				return "", errors.New("an error")
			})

			_, err := getClientIDFromJWTToken("aaa", []byte(""))
			assert.Error(GinkgoT(), err)
		})

		It("miss cache, verify ok", func() {
			patches = gomonkey.ApplyFunc(impls.GetJWTTokenClientID, func(string) (string, error) {
				return "", errors.New("an error")
			})
			patches.ApplyFunc(impls.SetJWTTokenClientID, func(string, string) {
			})
			patches.ApplyFunc(verifyClientID, func(string, []byte) (string, error) {
				return "abc", nil
			})

			clientID, err := getClientIDFromJWTToken("aaa", []byte(""))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "abc", clientID)

			//clientID, err = impls.GetJWTTokenClientID("aaa")
			//assert.NoError(GinkgoT(), err)
			//assert.Equal(GinkgoT(), "abc", clientID)
		})

	})

	Describe("verifyClientID", func() {
		var patches *gomonkey.Patches
		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		It("parse fail", func() {
			patches = gomonkey.ApplyFunc(parseBKJWTToken, func(string, []byte) (jwt.MapClaims, error) {
				return nil, errors.New("an error")
			})

			_, err := verifyClientID("aaa", []byte(""))
			assert.Error(GinkgoT(), err)
		})

		It("parse ok", func() {
			patches = gomonkey.ApplyFunc(parseBKJWTToken, func(string, []byte) (jwt.MapClaims, error) {
				return jwt.MapClaims{
					"app": map[string]interface{}{
						"app_code": "bk_test",
						"verified": true,
					},
				}, nil
			})

			clientID, err := verifyClientID("aaa", []byte(""))
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), "bk_test", clientID)
		})

	})

})

const testPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnzyis1ZjfNB0bBgKFMSv
vkTtwlvBsaJq7S5wA+kzeVOVpVWwkWdVha4s38XM/pa/yr47av7+z3VTmvDRyAHc
aT92whREFpLv9cj5lTeJSibyr/Mrm/YtjCZVWgaOYIhwrXwKLqPr/11inWsAkfIy
tvHWTxZYEcXLgAXFuUuaS3uF9gEiNQwzGTU1v0FqkqTBr4B8nW3HCN47XUu0t8Y0
e+lf4s4OxQawWD79J9/5d3Ry0vbV3Am1FtGJiJvOwRsIfVChDpYStTcHTCMqtvWb
V6L11BWkpzGXSW4Hv43qa+GSYOD2QU68Mb59oSk2OB+BtOLpJofmbGEGgvmwyCI9
MwIDAQAB
-----END PUBLIC KEY-----`

// NOTE: copied from apigateway jwt parse and generate
func TestParseBKJWTToken(t *testing.T) {
	var data1 = []struct {
		jwt               string
		publicKey         string
		expectedJWTHeader map[string]interface{}
		expectedClaims    jwt.MapClaims
		willError         bool
	}{
		{
			// jwt header without kid
			jwt: `eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCIsImlhdCI6MTU2MDQ4MzA5NH0.eyJpc3MiOiJBUElHVyIsIm` +
				`FwcCI6eyJhcHBfY29kZSI6ImFwaWd3LXRlc3QifSwidXNlciI6eyJ1c2VybmFtZSI6ImFkbWluIn0sImV4c` +
				`CI6MTg3Mjc2ODI3OSwibmJmIjoxNTYwNDgzMDk0fQ.Sy5CyTO5mBoINnMkhQ0ZqM-Zcsp1kv-wnEmEmhZOY` +
				`W-KDl_qipIHekNWqkuMkfZWB9I5O1kEPWA3ApY9SUwfosaTE2ZEahH9fM1WNgHlB_sd_cOxYXJ0CATPI_aY` +
				`D96cdbRXiRIEr57J_OmnQhI4Xk4nmIuP7NZb1lOmy2Qm711fhFAcIpp_U1gu98f5IBvoDxl9XfgJCa_-ZPl` +
				`5zOPdwfnKN29fUiXDkmJmTwcf6verC53OhJN3liRx-myjHgZ8JIKRkkUwbp8L1gkapIvY_WMjwBiMJ7feGa` +
				`61BNvnOYsPzG-CnnAFpZT8H5Cgy-Raq4I8afhbsT-yPq71N0QD9A`,
			expectedJWTHeader: nil,
			expectedClaims:    nil,
			willError:         true,
		},
		{
			// invalid public key
			jwt: `eyJhbGciOiJSUzUxMiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QiLCJpYXQiOjE1NjA0ODMxMTZ9.eyJpc3MiO` +
				`iJBUElHVyIsImFwcCI6eyJhcHBfY29kZSI6ImFwaWd3LXRlc3QifSwidXNlciI6eyJ1c2VybmFtZSI6ImFkbWlu` +
				`In0sImV4cCI6MTg3Mjc2ODI3OSwibmJmIjoxNTYwNDgzMTE2fQ.ask_dn8tdpjbEi_WELTVwHYWGPr8TFaAtGnZ` +
				`cxPwLlieMvCkSMMjrT1bmCIoOBqpL0XiMxkd_7XwQqOtI4fNol-PyGUV60Bfe8sABt59KpNNFGvpe1L-uSmPw2r` +
				`6r-GD4gnO_NJnSHn-BKtz2CKc451HWIBa1iHBEbj3wJX8XXjGj8SR4TVAuljUQASnH4RC2EUz7SqguWDYqNyPHR` +
				`kwERwq1aO6e9tCdIxrQaRh3CKNG5_OBkeH5IGC9avBNKy-j-Tp_l6HwFvXHWxq86bu84lujvaMnHBLvBtbFID2d` +
				`xIcdF6VTltfV69vpd3Zyc91A8Co0vMQgW4Sdjym7s3nUw`,
			publicKey:         "invalid",
			expectedJWTHeader: nil,
			expectedClaims:    nil,
			willError:         true,
		},
		{
			// invalid public key, format correct
			jwt: `eyJhbGciOiJSUzUxMiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QiLCJpYXQiOjE1NjA0ODMxMTZ9.eyJpc3MiOiJ` +
				`BUElHVyIsImFwcCI6eyJhcHBfY29kZSI6ImFwaWd3LXRlc3QifSwidXNlciI6eyJ1c2VybmFtZSI6ImFkbWluIn0s` +
				`ImV4cCI6MTg3Mjc2ODI3OSwibmJmIjoxNTYwNDgzMTE2fQ.ask_dn8tdpjbEi_WELTVwHYWGPr8TFaAtGnZcxPwLl` +
				`ieMvCkSMMjrT1bmCIoOBqpL0XiMxkd_7XwQqOtI4fNol-PyGUV60Bfe8sABt59KpNNFGvpe1L-uSmPw2r6r-GD4gn` +
				`O_NJnSHn-BKtz2CKc451HWIBa1iHBEbj3wJX8XXjGj8SR4TVAuljUQASnH4RC2EUz7SqguWDYqNyPHRkwERwq1aO6` +
				`e9tCdIxrQaRh3CKNG5_OBkeH5IGC9avBNKy-j-Tp_l6HwFvXHWxq86bu84lujvaMnHBLvBtbFID2dxIcdF6VTltfV` +
				`69vpd3Zyc91A8Co0vMQgW4Sdjym7s3nUw`,
			publicKey:         strings.Replace(testPublicKey, "9", "8", 1),
			expectedJWTHeader: nil,
			expectedClaims:    nil,
			willError:         true,
		},
		{
			// exp expired
			jwt: `eyJhbGciOiJSUzUxMiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QiLCJpYXQiOjE1NjA0ODMyMDd9.eyJpc3MiOiJBUE` +
				`lHVyIsImFwcCI6eyJhcHBfY29kZSI6ImFwaWd3LXRlc3QifSwidXNlciI6eyJ1c2VybmFtZSI6ImFkbWluIn0sImV4cC` +
				`I6MTU2MDQ4MzIwNywibmJmIjoxNTYwNDgzMjA3fQ.XpfknmZUPGiG6qUloMRdZ_aBmazbMxJtMtMYXNiWYaUG7X4E09e` +
				`HKYaJilTLp0tVmiet1YxUnEYPW0gzbeZKP46sEt5qcbBtrov0VjCkH14PgjNGo2Rx6-S443Jz_LBRUw8R8XcmOMnA4_X` +
				`lqIQVFqEB5M36suSeFucmLLgBoYdVNT_7TyYNe-A0RtFIFXkDDiMWWesHqxqP1aSfOa_TRnd7Fq_i3krPXO5IBuYtMdn` +
				`xh7Dbz29bNCEeithKyx8KrzHQ3NnSMouG6eiw1XCVyETkausEEa5s9siDWDFkHPByXfSuMiYzvUnaFNnCB4H7fwpv6tg` +
				`WPoeOE8BQGaKwMQ`,
			publicKey:         testPublicKey,
			expectedJWTHeader: nil,
			expectedClaims:    nil,
			willError:         true,
		},
		{
			// nbf invalid
			jwt: `eyJhbGciOiJSUzUxMiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QiLCJpYXQiOjE1NjA0ODMyNDZ9.eyJpc3MiOiJBUE` +
				`lHVyIsImFwcCI6eyJhcHBfY29kZSI6ImFwaWd3LXRlc3QifSwidXNlciI6eyJ1c2VybmFtZSI6ImFkbWluIn0sImV4cC` +
				`I6MTg3Mjc2ODI3OSwibmJmIjoxODcyNzY4Mjc5fQ.JbL1-0K4wU26yzho-WotHLNnx6bkqR27Yi_up6L5VP_PvRklPZQ` +
				`648fmphpPK5OeBNKpZ1pYVaO9KgTZaVDToZ1f0YRO_Pali6Mt7Q8SIaRmbeM4N9pVnyeSBVdS4I8c3baZYQSgBGgrgt8` +
				`pwLPDe_FJ8Baz_Ftwb5uJQkQYPikjw60-GAuvAhtDyS5FbIwNXFY1KQDLOeVIbWlIwxgSQZwx6CiJXiawhqMJ-7ssAK4` +
				`RXnlBoNI_A9KlIpNu0motRezmn1r0XwE-sTt1eBxgU9HveZYqj9uuU2H1nMlAMtrDzswuQJPZJrmfp6DlWWd0tlwr-4d` +
				`tR3o4qQlDW5Hzrw`,
			publicKey:         testPublicKey,
			expectedJWTHeader: nil,
			expectedClaims:    nil,
			willError:         true,
		},
		{
			// token invalid
			jwt:               `invalid`,
			publicKey:         testPublicKey,
			expectedJWTHeader: nil,
			expectedClaims:    nil,
			willError:         true,
		},
		{
			// ok
			jwt: `eyJhbGciOiJSUzUxMiIsImtpZCI6InRlc3QiLCJ0eXAiOiJKV1QiLCJpYXQiOjE1NjA0ODMxMTZ9.eyJpc3MiOiJBU` +
				`ElHVyIsImFwcCI6eyJhcHBfY29kZSI6ImFwaWd3LXRlc3QifSwidXNlciI6eyJ1c2VybmFtZSI6ImFkbWluIn0sImV4cCI6` +
				`MTg3Mjc2ODI3OSwibmJmIjoxNTYwNDgzMTE2fQ.ask_dn8tdpjbEi_WELTVwHYWGPr8TFaAtGnZcxPwLlieMvCkSMMjrT1b` +
				`mCIoOBqpL0XiMxkd_7XwQqOtI4fNol-PyGUV60Bfe8sABt59KpNNFGvpe1L-uSmPw2r6r-GD4gnO_NJnSHn-BKtz2CKc451` +
				`HWIBa1iHBEbj3wJX8XXjGj8SR4TVAuljUQASnH4RC2EUz7SqguWDYqNyPHRkwERwq1aO6e9tCdIxrQaRh3CKNG5_OBkeH5I` +
				`GC9avBNKy-j-Tp_l6HwFvXHWxq86bu84lujvaMnHBLvBtbFID2dxIcdF6VTltfV69vpd3Zyc91A8Co0vMQgW4Sdjym7s3nUw`,
			publicKey: testPublicKey,
			expectedJWTHeader: map[string]interface{}{
				"alg": "RS512",
				"kid": "test",
				"typ": "JWT",
				// "iat": 1560481523,
			},
			expectedClaims: map[string]interface{}{
				"iss": "APIGW",
				"app": map[string]interface{}{
					"app_code": "apigw-test",
				},
				"user": map[string]interface{}{
					"username": "admin",
				},
				// "exp": 1872768279,
				// "nbf": 1560481523,
			},
			willError: false,
		},
	}
	for _, test := range data1 {
		claims, err := parseBKJWTToken(test.jwt, []byte(test.publicKey))

		if err == nil {
			// iat/exp/nbf is float in parse result
			//	delete(tokenHeader, "iat")
			delete(claims, "exp")
			delete(claims, "nbf")
		}

		if test.willError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, test.expectedClaims, claims)
	}
}
