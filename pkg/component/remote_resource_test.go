/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package component

import (
	"testing"

	"iam/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestRemoteResourceClient_QueryResources(t *testing.T) {
	// TODO: change to table-driven unit test

	// 1. 500
	ts := util.CreateTesting500Server()
	defer ts.Close()
	client := NewRemoteResourceClient()
	req := RemoteResourceRequest{
		URL:     ts.URL,
		Headers: nil,
	}

	resources, err := client.QueryResources(req, "paas", "app", []string{"1", "2"}, []string{"name"})
	assert.Error(t, err)
	assert.Nil(t, resources)

	// 2. code != 0
	ts1 := util.CreateTestingServer(map[string]interface{}{
		"code":    140000,
		"message": "fail",
	})
	defer ts1.Close()
	client1 := NewRemoteResourceClient()
	req1 := RemoteResourceRequest{
		URL:     ts1.URL,
		Headers: nil,
	}
	resources, err = client1.QueryResources(req1, "paas", "app", []string{"1", "2"}, []string{"name"})
	assert.Error(t, err)
	assert.Equal(t, "[RemoteResourceClient:QueryResources] result.Code=140000 => [Raw:Error] fail", err.Error())
	assert.Empty(t, resources)

	// 3. code == 0
	ts2 := util.CreateTestingServer(map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data": []map[string]string{
			{"name": "tom"},
		},
	})
	defer ts2.Close()
	client2 := NewRemoteResourceClient()
	req2 := RemoteResourceRequest{
		URL:     ts2.URL,
		Headers: nil,
	}
	resources, err = client2.QueryResources(req2, "paas", "app", []string{"1", "2"}, []string{"name"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resources))
	assert.Equal(t, "tom", resources[0]["name"])
}
