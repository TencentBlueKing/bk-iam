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
	"iam/pkg/errorx"
	"iam/pkg/service/types"
	"iam/pkg/util"
	"strings"
)

// PrepareRequest ...
func PrepareRequest(
	system types.System,
	resourceType types.ResourceType,
) (req RemoteResourceRequest, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf("RemoteResourceClient", "PrepareRequest")

	// 1. parse the providerConfig
	systemProviderConfig, err := util.MapValueInterfaceToString(system.ProviderConfig)
	if err != nil {
		err = errorWrapf(err, "MapValueInterfaceToString system.ProviderConfig=`%s` fail", system.ProviderConfig)
		return
	}
	resourceTypeProviderConfig, err := util.MapValueInterfaceToString(resourceType.ProviderConfig)
	if err != nil {
		err = errorWrapf(err, "MapValueInterfaceToString resource.ProviderConfig=`%s` fail",
			resourceType.ProviderConfig)
		return
	}

	// NOTE: `token` in System.ProviderConfig is sensitive

	// 2. get data from providerConfig
	host, ok := systemProviderConfig["host"]
	if !ok {
		err = errorWrapf(err, "key `host` not in system.ProviderConfig systemID=`%s`", system.ID)
		return
	}
	auth, ok := systemProviderConfig["auth"]
	if !ok {
		err = errorWrapf(err, "key `auth` not in system.ProviderConfig systemID=`%s`", system.ID)
		return
	}
	token := ""
	if auth != "none" {
		token, ok = systemProviderConfig["token"]
		if !ok {
			err = errorWrapf(err, "key `token` not in system.ProviderConfig systemID=`%s`", system.ID)
			return
		}
	}

	path, ok := resourceTypeProviderConfig["path"]
	if !ok {
		err = errorWrapf(err, "key `path` not in resourceType.ProviderConfig=`%s`", resourceTypeProviderConfig)
		return
	}

	// make the request
	req.URL = strings.TrimRight(host, "/") + "/" + strings.TrimLeft(path, "/")
	// NOTE: currently only support none and basic auth
	if token != "" {
		req.Headers = map[string]string{
			"Authorization": util.BasicAuthAuthorizationHeader("bk_iam", token),
		}
	}

	return req, nil
}
