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

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/parnurzeal/gorequest"
)

// RemoteResourceTimeout ...
const (
	RemoteResourceTimeout = 30 * time.Second

	ipRegexString = "\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}"
	replaceToIP   = "*.*.*.*"
)

var (
	ipRegex = regexp.MustCompile(ipRegexString)
)

// RemoteResourceRequest ...
type RemoteResourceRequest struct {
	URL     string
	Headers map[string]string
}

// RemoteResourceResponse ...
type RemoteResourceResponse struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Data    []map[string]interface{} `json:"data"`
}

// Error ...
func (r *RemoteResourceResponse) Error() error {
	if r.Code == 0 {
		return nil
	}

	return fmt.Errorf("response error[code=`%d`,  message=`%s`]", r.Code, r.Message)
}

// RemoteResourceClient ...
type RemoteResourceClient interface {
	QueryResources(req RemoteResourceRequest, system string, _type string, ids []string,
		fields []string) ([]map[string]interface{}, error)
	GetResources(req RemoteResourceRequest, system string, _type string, ids []string,
		fields []string) ([]map[string]interface{}, error)
}

type remoteResourceClient struct {
}

// NewRemoteResourceClient ...
func NewRemoteResourceClient() RemoteResourceClient {
	return &remoteResourceClient{}
}

// QueryResources ...
func (c *remoteResourceClient) QueryResources(
	req RemoteResourceRequest,
	system, _type string,
	ids []string,
	fields []string,
) ([]map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("RemoteResourceClient", "QueryResources")

	var err error
	url := req.URL

	data := map[string]interface{}{
		"type":   _type,
		"method": "fetch_instance_info",
		"filter": map[string]interface{}{
			"ids":   ids,
			"attrs": fields,
		},
	}

	result := RemoteResourceResponse{}
	start := time.Now()
	callbackFunc := NewMetricCallback(system, start)

	request := gorequest.New().Timeout(RemoteResourceTimeout).Post(url).Type("json")
	// set headers
	if len(req.Headers) > 0 {
		for key, value := range req.Headers {
			request.Header.Set(key, value)
		}
	}
	// do request
	resp, respBody, errs := request.
		Send(data).
		EndStruct(&result, callbackFunc)

	logHTTPRequest(start, request, resp, respBody, errs, &result)

	if len(errs) != 0 {
		// 敏感信息泄漏 ip+端口号, 替换为 *.*.*.*
		errsMessage := fmt.Sprintf("gorequest errorx=`%s`", errs)
		errsMessage = ipRegex.ReplaceAllString(errsMessage, replaceToIP)
		err = errors.New(errsMessage)

		err = errorWrapf(err, "errsCount=`%d`", len(errs))
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("query resources from %s not 200", system)
		return nil, errorWrapf(err, "status=%d", resp.StatusCode)
	}
	if result.Code != 0 {
		err = errors.New(result.Message)
		err = errorWrapf(err, "result.Code=%d", result.Code)
		return nil, err
	}
	return result.Data, nil
}

// GetResources ...
func (c *remoteResourceClient) GetResources(
	req RemoteResourceRequest,
	system string,
	_type string,
	ids []string,
	fields []string,
) ([]map[string]interface{}, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("RemoteResourceClient", "QueryResources")

	data, err := c.QueryResources(req, system, _type, ids, fields)
	if err != nil {
		return nil, errorWrapf(err, "queryResources system=`%s`, type=`%s`, ids=`%v`, fields=`%v` fail",
			system, _type, ids, fields)
	}

	if len(data) < 1 {
		err = errors.New("get resource got empty data")
		return nil, errorWrapf(err, "queryResources system=`%s`, type=`%s`, ids=`%v`, fields=`%v` fail",
			system, _type, ids, fields)
	}
	return data, nil
}
