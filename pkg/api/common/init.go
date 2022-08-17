/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	log "github.com/sirupsen/logrus"

	"iam/pkg/config"
)

const (
	// model quota
	maxActionsLimitKey            = "max_actions_limit"
	maxResourceTypesLimitKey      = "max_resource_types_limit"
	maxInstanceSelectionsLimitKey = "max_instance_selections_limit"

	DefaultMaxActionsLimit            = 100
	DefaultMaxResourceTypesLimit      = 50
	DefaultMaxInstanceSelectionsLimit = 50

	// web quota
	maxSubjectGroupsLimitKey = "max_subject_groups_limit"

	DefaultMaxSubjectGroupsLimit = 100

	// triggers
	triggerDisableCreateSystemClientValidationKey = "disable_create_system_client_validation"

	DisableCreateSystemClientValidation = false
)

var (
	quota        = config.Quota{}
	customQuotas = make(map[string]config.Quota)

	switches = make(map[string]bool)
)

// InitQuota ...
func InitQuota(q config.Quota, cq map[string]config.Quota) {
	quota = q
	customQuotas = cq

	log.Infof("init quota: %+v, customQuotas: %+v", quota, customQuotas)
}

// InitSwitch ...
func InitSwitch(ss map[string]bool) {
	switches = ss
	log.Infof("init switch: %+v", switches)
}

func makeGetModelLimitFunc(key string, defaultLimit int) func(string) int {
	return func(systemID string) int {
		// custom
		if cq, ok := customQuotas[systemID]; ok {
			if limit, ok := cq.Model[key]; ok && limit > 0 {
				return limit
			}
		}
		// config file default
		if limit, ok := quota.Model[key]; ok && limit > 0 {
			return limit
		}
		// default
		return defaultLimit
	}
}

func makeGetWebLimitFunc(key string, defaultLimit int) func(string) int {
	return func(systemID string) int {
		// custom
		if cq, ok := customQuotas[systemID]; ok {
			if limit, ok := cq.Web[key]; ok && limit > 0 {
				return limit
			}
		}
		// config file default
		if limit, ok := quota.Web[key]; ok && limit > 0 {
			return limit
		}
		// default
		return defaultLimit
	}
}

var (
	// Model
	GetMaxActionsLimit            = makeGetModelLimitFunc(maxActionsLimitKey, DefaultMaxActionsLimit)
	GetMaxResourceTypesLimit      = makeGetModelLimitFunc(maxResourceTypesLimitKey, DefaultMaxResourceTypesLimit)
	GetMaxInstanceSelectionsLimit = makeGetModelLimitFunc(
		maxInstanceSelectionsLimitKey,
		DefaultMaxInstanceSelectionsLimit,
	)

	// Web
	GetMaxSubjectGroupsLimit = makeGetWebLimitFunc(maxSubjectGroupsLimitKey, DefaultMaxSubjectGroupsLimit)
)

func makeGetSwitchFunc(key string, defaultValue bool) func() bool {
	return func() bool {
		if b, ok := switches[key]; ok {
			return b
		}
		return defaultValue
	}
}

var GetSwitchDisableCreateSystemClientValidation = makeGetSwitchFunc(
	triggerDisableCreateSystemClientValidationKey,
	DisableCreateSystemClientValidation,
)
