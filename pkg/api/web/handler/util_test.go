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
	"testing"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/stretchr/testify/assert"
)

func Test_validateFields(t *testing.T) {
	type args struct {
		expected string
		actual   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty",
			args: args{
				expected: "id,name",
				actual:   "",
			},
			want: true,
		},
		{
			name: "right",
			args: args{
				expected: "id,name",
				actual:   "id,name",
			},
			want: true,
		},
		{
			name: "match",
			args: args{
				expected: "id,name,name_en",
				actual:   "id,name",
			},
			want: true,
		},
		{
			name: "wrong",
			args: args{
				expected: "id,name",
				actual:   "id,name,name_en",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateFields(tt.args.expected, tt.args.actual); got != tt.want {
				t.Errorf("validateFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterFields(t *testing.T) {
	obj := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{
		ID:   "1",
		Name: "test",
	}
	fields := set.NewStringSetWithValues([]string{"id"})

	// one
	want1 := map[string]interface{}{"id": "1"}
	got1, err := filterFields(fields, obj)
	assert.NoError(t, err)
	assert.Equal(t, want1, got1)

	// all
	want2 := map[string]interface{}{"id": "1", "name": "test"}
	fields = set.NewStringSetWithValues([]string{"id", "name"})
	got2, err := filterFields(fields, obj)
	assert.NoError(t, err)
	assert.Equal(t, want2, got2)

	// not exists
	want3 := map[string]interface{}{}
	fields = set.NewStringSetWithValues([]string{"age"})
	got3, err := filterFields(fields, obj)
	assert.NoError(t, err)
	assert.Equal(t, want3, got3)

	// empty set
	fields = set.NewStringSetWithValues([]string{})
	got4, err := filterFields(fields, obj)
	assert.NoError(t, err)
	assert.Equal(t, want3, got4)
}
