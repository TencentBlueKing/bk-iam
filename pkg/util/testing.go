/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

// TestingContent ...
var TestingContent = []byte("Hello, World!")

// NewRequestResponse ...
func NewRequestResponse() (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(TestingContent))
	w := httptest.NewRecorder()
	return r, w
}

// NewRequestResponseWithContent ...
func NewRequestResponseWithContent(content []byte) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(content))
	w := httptest.NewRecorder()
	return r, w
}

// TestingEmptyContent ...
var TestingEmptyContent = []byte("")

// NewRequestEmptyResponse ...
func NewRequestEmptyResponse() (*http.Request, *httptest.ResponseRecorder) {
	return NewRequestResponseWithContent(TestingEmptyContent)
}

// error
type errReader int

// Read ...
func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

// NewRequestErrorResponse ...
func NewRequestErrorResponse() (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodPost, "/", errReader(0))
	w := httptest.NewRecorder()
	return r, w
}

// redis
func NewTestRedisClient() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client
}

// for gin
func SetupRouter() *gin.Engine {
	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	r.Use(gin.Recovery())
	// r.GET("/ping", func(c *gin.Context) {
	//	c.String(200, "pong")
	// })
	return r
}

// NewTestRouter ...
func NewTestRouter(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
}

// CreateTestingServer ...
func CreateTestingServer(data interface{}) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respBody, _ := jsoniter.Marshal(data)
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	}))
	return ts
}

// CreateTesting500Server ...
func CreateTesting500Server() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respBody, _ := jsoniter.Marshal(map[string]interface{}{})
		w.WriteHeader(http.StatusInternalServerError)
		// w.Write([]byte("Internal Server Error"))
		w.Write(respBody)
	}))
	return ts
}

// for apitest https://github.com/steinfletcher/apitest

// JSONAssertFunc ...
type JSONAssertFunc func(map[string]interface{}) error

// type JSONAssertFunc func(Response) error

// NewJSONAssertFunc ...
func NewJSONAssertFunc(t *testing.T, assertFunc JSONAssertFunc) func(res *http.Response, req *http.Request) error {
	return func(res *http.Response, req *http.Request) error {
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err, "read body from response fail")

		defer res.Body.Close()

		var data map[string]interface{}
		// var data Response

		err = json.Unmarshal(body, &data)
		assert.NoError(t, err, "unmarshal string to json fail")

		return assertFunc(data)
	}
}

// ResponseAssertFunc ...
type ResponseAssertFunc func(Response) error

// NewResponseAssertFunc ...
func NewResponseAssertFunc(
	t *testing.T,
	responseFunc ResponseAssertFunc,
) func(res *http.Response, req *http.Request) error {
	return func(res *http.Response, req *http.Request) error {
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, err, "read body from response fail")

		defer res.Body.Close()

		var data Response

		err = json.Unmarshal(body, &data)
		assert.NoError(t, err, "unmarshal string to json fail")

		return responseFunc(data)
	}
}

// GinAPIRequest apitest with gin router
type GinAPIRequest struct {
	t       *testing.T
	request *apitest.Request
}

// CreateNewAPIRequestFunc New Gin GinAPIRequest
func CreateNewAPIRequestFunc(
	method, url string, handler gin.HandlerFunc, args ...string,
) func(t *testing.T) *GinAPIRequest {
	return func(t *testing.T) *GinAPIRequest {
		if len(args) > 1 {
			panic("args error")
		}

		handlerURL := url
		if len(args) != 0 {
			handlerURL = args[0]
		}

		r := SetupRouter()
		// reflect.ValueOf(r).MethodByName(
		// 	strings.ToUpper(method),
		// ).Call([]reflect.Value{reflect.ValueOf(handlerURL), reflect.ValueOf(handler)})
		switch strings.ToUpper(method) {
		case "GET":
			r.GET(handlerURL, handler)
		case "POST":
			r.POST(handlerURL, handler)
		case "PUT":
			r.PUT(handlerURL, handler)
		case "DELETE":
			r.DELETE(handlerURL, handler)
		case "HEAD":
			r.HEAD(handlerURL, handler)
		case "PATCH":
			r.PATCH(handlerURL, handler)
		case "OPTIONS":
			r.OPTIONS(handlerURL, handler)
		case "ANY":
			r.Any(handlerURL, handler)
		}
		// r.GET(handlerURL, handler)

		test := apitest.New().Handler(r)
		// test.Handler(r)

		reflectValues := reflect.ValueOf(test).MethodByName(
			strings.Title(method),
		).Call([]reflect.Value{reflect.ValueOf(url)})

		return &GinAPIRequest{
			t:       t,
			request: reflectValues[0].Interface().(*apitest.Request),
		}
	}
}

// JSON ...
func (g *GinAPIRequest) JSON(data interface{}) *GinAPIRequest {
	g.request.JSON(data)

	return g
}

// NoJSON ...
func (g *GinAPIRequest) NoJSON() {
	g.request.
		Expect(g.t).
		Assert(NewResponseAssertFunc(g.t, func(resp Response) error {
			assert.Equal(g.t, BadRequestError, resp.Code)
			return nil
		})).
		Status(http.StatusOK).
		End()
}

// BadRequest ...
func (g *GinAPIRequest) BadRequest(message string) {
	g.request.
		Expect(g.t).
		Assert(NewResponseAssertFunc(g.t, func(resp Response) error {
			assert.Equal(g.t, BadRequestError, resp.Code)
			assert.Equal(g.t, message, resp.Message)
			return nil
		})).
		Status(http.StatusOK).
		End()
}

// BadRequestContainsMessage assert the bad request message field should contains a specific message
func (g *GinAPIRequest) BadRequestContainsMessage(message string, args ...int) {
	var code int
	if len(args) == 1 {
		code = args[0]
	} else {
		code = BadRequestError
	}

	g.request.
		Expect(g.t).
		Assert(NewResponseAssertFunc(g.t, func(resp Response) error {
			assert.Equal(g.t, code, resp.Code)
			assert.Contains(g.t, resp.Message, message)
			return nil
		})).
		Status(http.StatusOK).
		End()
}

// SystemError ...
func (g *GinAPIRequest) SystemError() {
	g.request.
		Expect(g.t).
		Assert(NewResponseAssertFunc(g.t, func(resp Response) error {
			assert.Equal(g.t, SystemError, resp.Code)
			assert.Contains(g.t, resp.Message, "system error")
			return nil
		})).
		Status(http.StatusOK).
		End()
}

// OK ...
func (g *GinAPIRequest) OK() {
	g.request.
		Expect(g.t).
		Assert(NewResponseAssertFunc(g.t, func(resp Response) error {
			assert.Equal(g.t, NoError, resp.Code)
			return nil
		})).
		Status(http.StatusOK).
		End()
}

func ReadResponse(w *httptest.ResponseRecorder) Response {
	var got Response
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	return got
}

func CreateTestContextWithDefaultRequest(w *httptest.ResponseRecorder) *gin.Context {
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("POST", "/", new(bytes.Buffer))
	return ctx
}
