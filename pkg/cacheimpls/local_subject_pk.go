/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cacheimpls

import (
	"iam/pkg/cache"
	"iam/pkg/errorx"
)

func retrieveSubjectPKFromRedis(key cache.Key) (interface{}, error) {
	k := key.(SubjectIDCacheKey)
	return GetSubjectPK(k.Type, k.ID)
}

// GetLocalSubjectPK ...
func GetLocalSubjectPK(_type, id string) (pk int64, err error) {
	key := SubjectIDCacheKey{
		Type: _type,
		ID:   id,
	}
	pk, err = LocalSubjectPKCache.GetInt64(key)
	if err != nil {
		err = errorx.Wrapf(err, CacheLayer, "GetLocalSubjectPK",
			"LocalSubjectPKCache.Get _type=`%s`, id=`%s` fail", _type, id)
	}
	return
}

// DeleteLocalSubjectPK ...
func DeleteLocalSubjectPK(_type, id string) error {
	key := SubjectIDCacheKey{
		Type: _type,
		ID:   id,
	}
	return LocalSubjectPKCache.Delete(key)
}
