/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/database/sdao"
	"iam/pkg/util/json"
)

// SystemConfigSVC ...
const SystemConfigSVC = "SystemConfigSVC"

// ConfigKeyActionGroups ...
const (
	// NOTE:  这里用复数!
	// 操作组

	ConfigKeyActionGroups           = "action_groups"
	ConfigKeyResourceCreatorActions = "resource_creator_actions"
	ConfigKeyCommonActions          = "common_actions"
	ConfigKeyFeatureShieldRules     = "feature_shield_rules"
	ConfigKeySystemManagers         = "system_managers"
	ConfigKeyCustomFrontendSettings = "custom_frontend_settings"

	ConfigTypeJSON = "json"
)

// SystemConfigService ...
type SystemConfigService interface {
	get(system string, key string) (interface{}, error)
	// actionGroup

	GetActionGroups(system string) ([]interface{}, error)
	CreateOrUpdateActionGroups(system string, actionGroup []interface{}) error

	// resourceCreatorAction

	GetResourceCreatorActions(system string) (map[string]interface{}, error)
	CreateOrUpdateResourceCreatorActions(system string, resourceCreatorAction map[string]interface{}) error

	// ConfigKeyCommonActions

	GetCommonActions(system string) ([]interface{}, error)
	CreateOrUpdateCommonActions(system string, actions []interface{}) error

	// featureShieldRules

	GetFeatureShieldRules(system string) ([]interface{}, error)
	CreateOrUpdateFeatureShieldRules(system string, featureShieldRules []interface{}) error

	// systemMangers

	GetSystemManagers(system string) (sm []interface{}, err error)
	CreateOrUpdateSystemManagers(system string, systemManagers []interface{}) (err error)

	// custom frontend settings

	GetCustomFrontendSettings(system string) (settings map[string]interface{}, err error)
	CreateOrUpdateCustomFrontendSettings(system string, settings map[string]interface{}) (err error)
}

type systemConfigService struct {
	manager sdao.SaaSSystemConfigManager
}

// NewSystemConfigService ...
func NewSystemConfigService() SystemConfigService {
	return &systemConfigService{
		manager: sdao.NewSaaSSystemConfigManager(),
	}
}

func (s *systemConfigService) get(system string, key string) (data interface{}, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemConfigSVC, "get")

	sc, err := s.manager.Get(system, key)
	if err != nil {
		err = errorWrapf(err, "s.manager.Get system=`%s`, key=`%s` fail", system, key)
		return
	}

	if sc.Type != ConfigTypeJSON {
		err = fmt.Errorf("systemConfigSVC not support type != json yet")
		return
	}

	err = json.UnmarshalFromString(sc.Value, &data)
	if err != nil {
		err = errorWrapf(err, "unmarshal system=`%s`, key=`%s`, value=`%s` fail", system, key, sc.Value)
	}
	return
}

func (s *systemConfigService) getSliceConfig(system string, configKey string) (dataI []interface{}, err error) {
	var data interface{}
	data, err = s.get(system, configKey)
	if err != nil {
		return
	}

	var ok bool
	dataI, ok = data.([]interface{})
	if !ok {
		err = errors.New("data in db not type []interface{}")
		return
	}
	return
}

func (s *systemConfigService) getMapConfig(system string, configKey string) (dataI map[string]interface{},
	err error,
) {
	var data interface{}
	data, err = s.get(system, configKey)
	if err != nil {
		return
	}

	var ok bool
	dataI, ok = data.(map[string]interface{})
	if !ok {
		err = errors.New("data in db not type map[string]interface{}")
		return
	}
	return
}

func (s *systemConfigService) create(system, key, _type string, data interface{}) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemConfigSVC, "create")
	var (
		value string
		err   error
	)
	if _type == ConfigTypeJSON {
		value, err = json.MarshalToString(data)
		if err != nil {
			return errorWrapf(err, "unmarshal data=`%+v` fail", data)
		}
	}

	sc := sdao.SaaSSystemConfig{
		System: system,
		Name:   key,
		Type:   _type,
		Value:  value,
	}
	return s.manager.Create(sc)
}

// func (s *systemConfigService) update(systemConfig sdao.SaaSSystemConfig, data map[string]interface{}) error {
func (s *systemConfigService) update(systemConfig sdao.SaaSSystemConfig, data interface{}) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemConfigSVC, "update")
	var (
		value string
		err   error
	)
	if systemConfig.Type == ConfigTypeJSON {
		value, err = json.MarshalToString(data)
		if err != nil {
			return errorWrapf(err, "unmarshal data=`%+v` fail", data)
		}
	}

	systemConfig.Value = value
	return s.manager.Update(systemConfig)
}

// nolint:unparam
func (s *systemConfigService) createOrUpdate(system, key, _type string, data interface{}) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemConfigSVC, "createOrUpdate")

	sc, err := s.manager.Get(system, key)
	// not exist do add
	if errors.Is(err, sql.ErrNoRows) {
		err = s.create(system, key, _type, data)
		if err != nil {
			err = errorWrapf(err, "s.create system=`%s`, key=`%s`, type=`%s`, data=`%+v` fail",
				system, key, _type, data)
		}
		return
	}
	if err != nil {
		err = errorWrapf(err, "s.manager.Get system=`%s`, key=`%s` fail", system, key)
		return
	}

	// exist, do update
	err = s.update(sc, data)
	if err != nil {
		err = errorWrapf(err, "s.update systemConfig=`%+v`, data=`%+v` fail", sc, data)
		return
	}
	return
}

// GetActionGroups ...
func (s *systemConfigService) GetActionGroups(system string) (ag []interface{}, err error) {
	return s.getSliceConfig(system, ConfigKeyActionGroups)
}

// CreateOrUpdateActionGroups ...
func (s *systemConfigService) CreateOrUpdateActionGroups(system string, actionGroup []interface{}) (err error) {
	return s.createOrUpdate(system, ConfigKeyActionGroups, ConfigTypeJSON, actionGroup)
}

// GetResourceCreatorActions ...
func (s *systemConfigService) GetResourceCreatorActions(system string) (rca map[string]interface{}, err error) {
	return s.getMapConfig(system, ConfigKeyResourceCreatorActions)
}

// CreateOrUpdateResourceCreatorActions ...
func (s *systemConfigService) CreateOrUpdateResourceCreatorActions(
	system string,
	resourceCreatorAction map[string]interface{},
) (err error) {
	return s.createOrUpdate(system, ConfigKeyResourceCreatorActions, ConfigTypeJSON, resourceCreatorAction)
}

// GetCommonActions ...
func (s *systemConfigService) GetCommonActions(system string) (ca []interface{}, err error) {
	return s.getSliceConfig(system, ConfigKeyCommonActions)
}

// CreateOrUpdateCommonActions ...
func (s *systemConfigService) CreateOrUpdateCommonActions(system string, commonActions []interface{}) (err error) {
	return s.createOrUpdate(system, ConfigKeyCommonActions, ConfigTypeJSON, commonActions)
}

// GetFeatureShieldRules ...
func (s *systemConfigService) GetFeatureShieldRules(system string) ([]interface{}, error) {
	return s.getSliceConfig(system, ConfigKeyFeatureShieldRules)
}

// CreateOrUpdateFeatureShieldRules ...
func (s *systemConfigService) CreateOrUpdateFeatureShieldRules(
	system string,
	featureShieldRules []interface{},
) (err error) {
	return s.createOrUpdate(system, ConfigKeyFeatureShieldRules, ConfigTypeJSON, featureShieldRules)
}

// GetSystemManagers ...
func (s *systemConfigService) GetSystemManagers(system string) (sm []interface{}, err error) {
	return s.getSliceConfig(system, ConfigKeySystemManagers)
}

// CreateOrUpdateSystemManagers ...
func (s *systemConfigService) CreateOrUpdateSystemManagers(system string, systemManagers []interface{}) (err error) {
	return s.createOrUpdate(system, ConfigKeySystemManagers, ConfigTypeJSON, systemManagers)
}

// GetCustomFrontendSettings ...
func (s *systemConfigService) GetCustomFrontendSettings(system string) (settings map[string]interface{}, err error) {
	return s.getMapConfig(system, ConfigKeyCustomFrontendSettings)
}

// CreateOrUpdateCustomFrontendSettings ...
func (s *systemConfigService) CreateOrUpdateCustomFrontendSettings(
	system string, settings map[string]interface{},
) (err error) {
	return s.createOrUpdate(system, ConfigKeyCustomFrontendSettings, ConfigTypeJSON, settings)
}
