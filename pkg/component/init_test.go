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
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/parnurzeal/gorequest"

	"iam/pkg/config"
	"iam/pkg/logging"
)

var _ = Describe("Init", func() {

	var start time.Time
	var result RemoteResourceResponse
	var url string

	BeforeEach(func() {
		logging.InitLogger(&config.Logger{})
		start = time.Now()
		result = RemoteResourceResponse{}
	})

	It("fill log fail detail", func() {
		url = "abc"
		data := map[string]interface{}{}
		request := gorequest.New().Timeout(1 * time.Second).Post(url).Type("json")
		// do request
		resp, respBody, errs := request.
			Send(data).
			EndStruct(&result)

		time.Sleep(20 * time.Millisecond)

		logHTTPRequest(start, request, resp, respBody, errs, &result)
	})

})
