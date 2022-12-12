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
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/TencentBlueKing/gopkg/stringx"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/database/sdao"
	"iam/pkg/service/types"
	"iam/pkg/util/json"
)

// SystemSVC ...
const SystemSVC = "SystemSVC"

// SystemService ...
type SystemService interface {
	Get(id string) (types.System, error)
	Exists(id string) bool
	ListAll() ([]types.System, error)

	Create(system types.System) error
	Update(id string, system types.System) error
}

type systemService struct {
	manager     dao.SystemManager
	saasManager sdao.SaaSSystemManager
}

// NewSystemService ...
func NewSystemService() SystemService {
	return &systemService{
		manager:     dao.NewSystemManager(),
		saasManager: sdao.NewSaaSSystemManager(),
	}
}

// Get ...
func (l *systemService) Get(id string) (system types.System, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemSVC, "Get")

	saasSystem, err := l.saasManager.Get(id)
	if err != nil {
		err = errorWrapf(err, "saasManager.Get id=`%s` fail", id)
		return
	}

	system.ID = saasSystem.ID
	system.Name = saasSystem.Name
	system.NameEn = saasSystem.NameEn
	system.Description = saasSystem.Description
	system.DescriptionEn = saasSystem.DescriptionEn
	system.Clients = saasSystem.Clients

	err = json.UnmarshalFromString(saasSystem.ProviderConfig, &system.ProviderConfig)
	if err != nil {
		err = errorWrapf(err, "unmarshal system.ProviderConfig=`%s` fail", saasSystem.ProviderConfig)
		return
	}

	return
}

// Exists ...
func (l *systemService) Exists(id string) bool {
	_, err := l.manager.Get(id)
	return err == nil
}

// ListAll ...
func (l *systemService) ListAll() (allSystems []types.System, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemSVC, "ListAll")

	saasSystems, err := l.saasManager.ListAll()
	if err != nil {
		err = errorWrapf(err, "saasManager.ListAll fail%s", "")
		return
	}

	for _, sys := range saasSystems {
		system := types.System{
			ID:            sys.ID,
			Name:          sys.Name,
			NameEn:        sys.NameEn,
			Description:   sys.Description,
			DescriptionEn: sys.DescriptionEn,
			Clients:       sys.Clients,
		}
		err = json.UnmarshalFromString(sys.ProviderConfig, &system.ProviderConfig)
		if err != nil {
			err = errorWrapf(err, "unmarshal system.ProviderConfig=`%s` fail", sys.ProviderConfig)
			return
		}

		allSystems = append(allSystems, system)
	}

	return allSystems, nil
}

// Create ...
func (l *systemService) Create(system types.System) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemSVC, "Create")

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error%s", "")
	}

	// 数据转换
	dbSystem := dao.System{
		ID: system.ID,
	}

	// NOTE: generate token here
	system.ProviderConfig["token"] = stringx.Random(32)

	providerConfig, err := json.MarshalToString(system.ProviderConfig)
	if err != nil {
		return errorWrapf(err, "marshal system.ProviderConfig=`%+v` fail", system.ProviderConfig)
	}

	dbSaaSSystem := sdao.SaaSSystem{
		ID:             system.ID,
		Name:           system.Name,
		NameEn:         system.NameEn,
		Description:    system.Description,
		DescriptionEn:  system.DescriptionEn,
		Clients:        system.Clients,
		ProviderConfig: providerConfig,
	}

	// 执行插入
	err = l.manager.CreateWithTx(tx, dbSystem)
	if err != nil {
		return errorWrapf(err, "manager.CreateWithTx dbSystem=`%+v` fail", dbSystem)
	}
	err = l.saasManager.CreateWithTx(tx, dbSaaSSystem)
	if err != nil {
		return errorWrapf(err, "saasManager.CreateWithTx dbSaaSSystem=`%+v` fail", dbSaaSSystem)
	}

	return tx.Commit()
}

// Update ...
func (l *systemService) Update(id string, system types.System) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SystemSVC, "Update")

	var err error
	var providerConfigStr string
	if len(system.ProviderConfig) > 0 {
		providerConfig := map[string]interface{}{}

		// NOTE: should only update the fields: host/auth, not token
		s, _ := l.saasManager.Get(id)
		if s.ProviderConfig != "" {
			err = json.UnmarshalFromString(s.ProviderConfig, &providerConfig)
			if err != nil {
				return errorWrapf(err, "unmarshal system.Provider=`%s` fail", s.ProviderConfig)
			}

			for key, value := range system.ProviderConfig {
				providerConfig[key] = value
			}
		} else {
			providerConfig = system.ProviderConfig
		}

		providerConfigStr, err = json.MarshalToString(providerConfig)
		if err != nil {
			return errorWrapf(err, "marshal system.Provider=`%+v` fail", providerConfig)
		}
	}

	allowBlank := database.NewAllowBlankFields()
	// action.Type is set => set saasAction.Type
	if system.AllowEmptyFields.HasKey("Description") {
		allowBlank.AddKey("Description")
	}
	if system.AllowEmptyFields.HasKey("DescriptionEn") {
		allowBlank.AddKey("DescriptionEn")
	}

	dbSaaSSystem := sdao.SaaSSystem{
		Name:           system.Name,
		NameEn:         system.NameEn,
		Description:    system.Description,
		DescriptionEn:  system.DescriptionEn,
		Clients:        system.Clients,
		ProviderConfig: providerConfigStr,

		AllowBlankFields: allowBlank,
	}
	return l.saasManager.Update(id, dbSaaSSystem)
}
