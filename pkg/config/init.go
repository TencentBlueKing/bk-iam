/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package config

import (
	"github.com/TencentBlueKing/gopkg/collection/set"
	log "github.com/sirupsen/logrus"
)

// SuperAppCodeSet ...
var (
	SuperAppCodeSet          *set.StringSet
	SuperUserSet             *set.StringSet
	SupportShieldFeaturesSet *set.StringSet
	SecurityAuditAppCode     *set.StringSet
)

// Worker ...
var (
	MaxSubjectActionAlterEventCheckCount               int = 3
	MaxMessageGeneratedCountPreSubjectActionAlterEvent int = 100

	MaxConsumerCountPerWorker int = 3
)

// InitSuperAppCode ...
func InitSuperAppCode(superAppCode string) {
	SuperAppCodeSet = set.SplitStringToSet(superAppCode, ",")
}

// InitSuperUser ...
func InitSuperUser(users string) {
	SuperUserSet = set.SplitStringToSet(users, ",")

	if !SuperUserSet.Has("admin") {
		SuperUserSet.Add("admin")
	}
}

// InitSupportShieldFeatures ...
func InitSupportShieldFeatures(supportShieldFeatures []string) {
	SupportShieldFeaturesSet = set.NewStringSetWithValues(supportShieldFeatures)
	// 默认支持的屏蔽的功能
	defaultSupportShieldFeatures := []string{
		// 申请功能
		"application",
		"application.custom_permission.grant", // 申请自定义权限
		"application.custom_permission.renew", // 申请自定义权限续期
		"application.group.join",              // 申请加入用户组
		"application.group.renew",             // 申请续期用户组
		"application.rating_manager.create",   // 申请新建分级管理员
		// 个人用户权限
		"user_permission",
		"user_permission.custom_permission.delete", // 自定义权限删除
		"user_permission.group.quit",               // 退出用户组
	}
	SupportShieldFeaturesSet.Append(defaultSupportShieldFeatures...)
}

// InitSecurityAuditAppCode read the value from config, parse to set
func InitSecurityAuditAppCode(securityAuditAppCode string) {
	SecurityAuditAppCode = set.SplitStringToSet(securityAuditAppCode, ",")
}

// InitWorker ...
func InitWorker(w Worker) {
	if w.MaxSubjectActionAlterEventCheckCount != 0 {
		MaxSubjectActionAlterEventCheckCount = w.MaxSubjectActionAlterEventCheckCount
	}

	if w.MaxMessageGeneratedCountPerSubjectActionAlterEvent != 0 {
		MaxMessageGeneratedCountPreSubjectActionAlterEvent = w.MaxMessageGeneratedCountPerSubjectActionAlterEvent
	}

	if w.MaxConsumerCountPerWorker != 0 {
		MaxConsumerCountPerWorker = w.MaxConsumerCountPerWorker
	}

	log.Infof(
		"init worker success, MaxMessageGeneratedCountPreSubjectActionAlterEvent=%d, "+
			"MaxSubjectActionAlterEventCheckCount=%d, "+
			"MaxConsumerCountPerWorker=%d",
		MaxMessageGeneratedCountPreSubjectActionAlterEvent,
		MaxSubjectActionAlterEventCheckCount,
		MaxConsumerCountPerWorker,
	)
}
