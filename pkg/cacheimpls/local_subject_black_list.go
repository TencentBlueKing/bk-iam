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
	"errors"

	"github.com/TencentBlueKing/gopkg/cache"
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"

	log "github.com/sirupsen/logrus"

	"iam/pkg/service"
)

/*
 * > 冻结解冻功能, 全局的黑名单
 */

var (
	globalSubjectBlackListKey = cache.NewStringKey("black_list")
	emptySubjectBlackList     = set.NewInt64Set()
)

func retrieveSubjectBlackList(key cache.Key) (interface{}, error) {
	// The key is not used in this function
	svc := service.NewSubjectBlackListService()
	subjectPKs, err := svc.ListSubjectPK()
	if err != nil {
		// NOTE: 获取失败, 打日志, 黑名单失效, 但是不影响正常逻辑
		log.WithError(err).Error("retrieveSubjectBlackList do svc.ListSubjectPK fail, the blacklist will not working!")
		return emptySubjectBlackList, nil
	}
	if len(subjectPKs) == 0 {
		return emptySubjectBlackList, nil
	}

	return set.NewInt64SetWithValues(subjectPKs), nil
}

// ListSubjectRoleSystemID ...
func IsSubjectInBlackList(subjectType, subjectID string) bool {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(CacheLayer, "IsSubjectInBlackList")

	subjectPK, err := GetLocalSubjectPK(subjectType, subjectID)
	if err != nil {
		// NOTE: 获取失败, 打日志, 黑名单失效, 但是不影响正常逻辑
		err = errorWrapf(err, "GetLocalSubjectPK subjectType=`%s`, subjectID=`%s` fail", subjectType, subjectID)
		// TODO: 这里无效的用户都会打?
		log.WithError(err).Warnf("the blacklist check will not working!")
		return false
	}

	value, err := LocalSubjectBlackListCache.Get(globalSubjectBlackListKey)
	if err != nil {
		// NOTE: 获取失败, 打日志, 黑名单失效, 但是不影响正常逻辑
		log.WithError(err).Warn("LocalSubjectBlackListCache get black list fail, the blacklist check will not working!")
		return false
	}

	var ok bool
	subjectBlackList, ok := value.(set.Int64Set)
	if !ok {
		// NOTE: 获取失败, 打日志, 黑名单失效, 但是不影响正常逻辑
		err = errors.New("not []set.Int64Set in cache")
		log.WithError(err).
			Warn("LocalSubjectBlackListCache convert black list fail, the blacklist check will not working!")
		return false
	}
	return subjectBlackList.Has(subjectPK)
}

// DeleteSubjectBlackListCache ...
func DeleteSubjectBlackListCache() error {
	return LocalSubjectBlackListCache.Delete(globalSubjectBlackListKey)
}
