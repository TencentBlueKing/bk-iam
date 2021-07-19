/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package impls

import (
	"iam/pkg/cache"
	"iam/pkg/database/edao"

	log "github.com/sirupsen/logrus"
)

// AppCodeAppSecretCacheKey ...
type AppCodeAppSecretCacheKey struct {
	AppCode   string
	AppSecret string
}

// Key ...
func (k AppCodeAppSecretCacheKey) Key() string {
	return k.AppCode + ":" + k.AppSecret
}

func retrieveAppCodeAppSecret(key cache.Key) (interface{}, error) {
	k := key.(AppCodeAppSecretCacheKey)

	manager := edao.NewAppSecretManager()
	return manager.Exists(k.AppCode, k.AppSecret)
}

// VerifyAppCodeAppSecret ...
func VerifyAppCodeAppSecret(appCode, appSecret string) bool {
	key := AppCodeAppSecretCacheKey{
		AppCode:   appCode,
		AppSecret: appSecret,
	}
	exists, err := LocalAppCodeAppSecretCache.GetBool(key)
	if err != nil {
		log.Errorf("get app_code_app_secret from memory cache fail, key=%s, err=%s", key.Key(), err)
		return false
	}
	return exists
}
