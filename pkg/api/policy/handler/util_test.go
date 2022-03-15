/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"errors"
	"testing"
	"time"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/cache/memory"
	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/cacheimpls"
	"iam/pkg/config"
)

var baseReq = baseRequest{
	System: "iam",
	Subject: subject{
		Type: "user",
		ID:   "admin",
	},
}

func Test_copyRequestFromAuthBody(t *testing.T) {
	t.Parallel()

	type args struct {
		req  *request.Request
		body *authRequest
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "right",
			args: args{
				req: request.NewRequest(),
				body: &authRequest{
					baseRequest: baseReq,
					Resources: []resource{{
						System: "iam",
						Type:   "host",
						ID:     "1",
						Attribute: map[string]interface{}{
							"key": "value",
						},
					}},
					Action: action{
						ID: "test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyRequestFromAuthBody(tt.args.req, tt.args.body)
			assert.Equal(t, &request.Request{
				System: "iam",
				Subject: types.Subject{
					Type:      "user",
					ID:        "admin",
					Attribute: types.NewSubjectAttribute(),
				},
				Action: types.Action{
					ID:        "test",
					Attribute: types.NewActionAttribute(),
				},
				Resources: []types.Resource{{
					System: "iam",
					Type:   "host",
					ID:     "1",
					Attribute: map[string]interface{}{
						"key": "value",
					},
				}},
			}, tt.args.req)
		})
	}
}

func Test_copyRequestFromQueryBody(t *testing.T) {
	t.Parallel()

	type args struct {
		req  *request.Request
		body *queryRequest
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "right",
			args: args{
				req: request.NewRequest(),
				body: &queryRequest{
					baseRequest: baseReq,
					Resources: []resource{{
						System: "iam",
						Type:   "host",
						ID:     "1",
						Attribute: map[string]interface{}{
							"key": "value",
						},
					}},
					Action: action{
						ID: "test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyRequestFromQueryBody(tt.args.req, tt.args.body)
			assert.Equal(t, &request.Request{
				System: "iam",
				Subject: types.Subject{
					Type:      "user",
					ID:        "admin",
					Attribute: types.NewSubjectAttribute(),
				},
				Action: types.Action{
					ID:        "test",
					Attribute: types.NewActionAttribute(),
				},
				Resources: []types.Resource{{
					System: "iam",
					Type:   "host",
					ID:     "1",
					Attribute: map[string]interface{}{
						"key": "value",
					},
				}},
			}, tt.args.req)
		})
	}
}

func Test_copyRequestFromQueryByActionsBody(t *testing.T) {
	t.Parallel()

	type args struct {
		req  *request.Request
		body *queryByActionsRequest
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "right",
			args: args{
				req: request.NewRequest(),
				body: &queryByActionsRequest{
					baseRequest: baseReq,
					Resources: []resource{{
						System: "iam",
						Type:   "host",
						ID:     "1",
						Attribute: map[string]interface{}{
							"key": "value",
						},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyRequestFromQueryByActionsBody(tt.args.req, tt.args.body)
			assert.Equal(t, &request.Request{
				System: "iam",
				Subject: types.Subject{
					Type:      "user",
					ID:        "admin",
					Attribute: types.NewSubjectAttribute(),
				},
				Action: types.Action{
					Attribute: types.NewActionAttribute(),
				},
				Resources: []types.Resource{{
					System: "iam",
					Type:   "host",
					ID:     "1",
					Attribute: map[string]interface{}{
						"key": "value",
					},
				}},
			}, tt.args.req)
		})
	}
}

func TestAnyExpression(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "any", AnyExpression["op"])
	assert.Equal(t, "", AnyExpression["field"])
	assert.Equal(t, []interface{}{}, AnyExpression["value"])
}

var _ = Describe("util", func() {

	Describe("validateSystemSuperUser", func() {
		var patches *gomonkey.Patches
		BeforeEach(func() {
			config.InitSuperUser("")
		})
		AfterEach(func() {
			patches.Reset()
		})

		It("validateSystemSuperUser error", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.ListSubjectRoleSystemID,
				func(subjectType, subjectID string) ([]string, error) {
					return nil, errors.New("test")
				})

			ok, err := hasSystemSuperPermission("bk_cmdb", "user", "admin1")
			assert.False(GinkgoT(), ok)
			assert.Error(GinkgoT(), err)
		})

		It("false", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.ListSubjectRoleSystemID,
				func(subjectType, subjectID string) ([]string, error) {
					return []string{}, nil
				})

			ok, err := hasSystemSuperPermission("bk_cmdb", "user", "admin1")
			assert.False(GinkgoT(), ok)
			assert.NoError(GinkgoT(), err)
		})

		It("ok, super_user", func() {
			ok, err := hasSystemSuperPermission("bk_cmdb", "user", "admin")
			assert.True(GinkgoT(), ok)
			assert.NoError(GinkgoT(), err)
		})

		It("ok, system_manager", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.ListSubjectRoleSystemID,
				func(subjectType, subjectID string) ([]string, error) {
					return []string{"bk_cmdb"}, nil
				})

			ok, err := hasSystemSuperPermission("bk_cmdb", "user", "admin1")
			assert.True(GinkgoT(), ok)
			assert.NoError(GinkgoT(), err)
		})

		It("ok, super_manager", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.ListSubjectRoleSystemID,
				func(subjectType, subjectID string) ([]string, error) {
					return []string{"SUPER"}, nil
				})

			ok, err := hasSystemSuperPermission("bk_cmdb", "user", "admin1")
			assert.True(GinkgoT(), ok)
			assert.NoError(GinkgoT(), err)
		})
	})

	Describe("buildResourceID", func() {

		It("empty", func() {
			uid := buildResourceID([]resource{})
			assert.Equal(GinkgoT(), uid, "")
		})

		It("single", func() {
			uid := buildResourceID([]resource{
				{
					System: "test",
					Type:   "foo",
					ID:     "123",
				},
			})
			assert.Equal(GinkgoT(), uid, "test,foo,123")
		})

		It("multi", func() {
			uid := buildResourceID([]resource{
				{
					System: "test",
					Type:   "foo",
					ID:     "123",
				},
				{
					System: "test2",
					Type:   "bar",
					ID:     "456",
				},
			})
			assert.Equal(GinkgoT(), uid, "test,foo,123/test2,bar,456")
		})
	})
})

func Test_validateSystemMatchClient(t *testing.T) {
	var (
		expiration = 5 * time.Minute
	)
	retrieveFunc := func(key cache.Key) (interface{}, error) {
		return []string{"test"}, nil
	}
	mockCache := memory.NewCache(
		"mockCache", false, retrieveFunc, expiration, nil)
	cacheimpls.LocalSystemClientsCache = mockCache

	type args struct {
		systemID string
		clientID string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "right",
			args: args{
				systemID: "test",
				clientID: "test",
			},
			want: true,
		},
		{
			name: "empty_system",
			args: args{
				systemID: "",
				clientID: "test",
			},
			want: false,
		},
		{
			name: "empty_client",
			args: args{
				systemID: "",
				clientID: "test",
			},
			want: false,
		},
		{
			name: "wrong",
			args: args{
				systemID: "test",
				clientID: "test1",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateSystemMatchClient(tt.args.systemID, tt.args.clientID) == nil; got != tt.want {
				t.Errorf("validateSystemMatchClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
