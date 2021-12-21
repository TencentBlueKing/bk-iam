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
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database"
	"iam/pkg/database/sdao"
	"iam/pkg/service/types"
)

// InstanceSelectionSVC ...
const InstanceSelectionSVC = "InstanceSelection"

// InstanceSelectionService ...
type InstanceSelectionService interface {
	ListBySystem(system string) ([]types.InstanceSelection, error)

	BulkCreate(system string, instanceSelections []types.InstanceSelection) error
	Update(system, instanceSelectionID string, instanceSelection types.InstanceSelection) error
	BulkDelete(system string, instanceSelectionIDs []string) error
}

type instanceSelectionService struct {
	saasManager sdao.SaaSInstanceSelectionManager
}

// NewInstanceSelectionService ...
func NewInstanceSelectionService() InstanceSelectionService {
	return &instanceSelectionService{
		saasManager: sdao.NewSaaSInstanceSelectionManager(),
	}
}

// ListBySystem ...
func (s *instanceSelectionService) ListBySystem(system string) (
	instanceSelections []types.InstanceSelection, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(InstanceSelectionSVC, "ListBySystem")

	saasInstanceSelections, err := s.saasManager.ListBySystem(system)
	if err != nil {
		return
	}

	for _, sis := range saasInstanceSelections {
		instanceSelection := types.InstanceSelection{
			ID:        sis.ID,
			Name:      sis.Name,
			NameEn:    sis.NameEn,
			IsDynamic: sis.IsDynamic,
		}

		err = jsoniter.UnmarshalFromString(sis.ResourceTypeChain, &instanceSelection.ResourceTypeChain)
		if err != nil {
			err = errorWrapf(err, "unmarshal sis.ResourceTypeChain=`%s` fail", sis.ResourceTypeChain)
			return
		}
		instanceSelections = append(instanceSelections, instanceSelection)
	}
	return instanceSelections, nil
}

// BulkCreate ...
func (s *instanceSelectionService) BulkCreate(system string, instanceSelections []types.InstanceSelection) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(InstanceSelectionSVC, "BulkCreate")

	// 数据转换
	dbSaaSInstanceSelections := make([]sdao.SaaSInstanceSelection, 0, len(instanceSelections))
	for _, is := range instanceSelections {
		chain, err1 := jsoniter.MarshalToString(is.ResourceTypeChain)
		if err1 != nil {
			return errorWrapf(err, "marshal is.ResourceTypeChain=`%+v` fail", is.ResourceTypeChain)
		}

		dbSaaSInstanceSelections = append(dbSaaSInstanceSelections, sdao.SaaSInstanceSelection{
			System:            system,
			ID:                is.ID,
			Name:              is.Name,
			NameEn:            is.NameEn,
			IsDynamic:         is.IsDynamic,
			ResourceTypeChain: chain,
		})
	}

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error system=`%s`", system)
	}

	// 执行插入
	err = s.saasManager.BulkCreateWithTx(tx, dbSaaSInstanceSelections)
	if err != nil {
		return errorWrapf(err, "saasManager.BulkCreateWithTx fail%s", "")
	}

	return tx.Commit()
}

// Update ...
func (s *instanceSelectionService) Update(
	system string,
	instanceSelectionID string,
	instanceSelection types.InstanceSelection,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(InstanceSelectionSVC, "Update")

	var err error

	var chain string
	chain, err = jsoniter.MarshalToString(instanceSelection.ResourceTypeChain)
	if err != nil {
		return errorWrapf(err,
			"marshal instanceSelection.ResourceTypeChain=`%+v` fail",
			instanceSelection.ResourceTypeChain)
	}

	allowBlank := database.NewAllowBlankFields()
	if instanceSelection.AllowEmptyFields.HasKey("IsDynamic") {
		allowBlank.AddKey("IsDynamic")
	}

	data := sdao.SaaSInstanceSelection{
		// PK:             0,
		// System:         "",
		// ID:             "",
		Name:              instanceSelection.Name,
		NameEn:            instanceSelection.NameEn,
		IsDynamic:         instanceSelection.IsDynamic,
		ResourceTypeChain: chain,

		AllowBlankFields: allowBlank,
	}

	return s.saasManager.Update(system, instanceSelectionID, data)
}

// BulkDelete ...
func (s *instanceSelectionService) BulkDelete(system string, instanceSelectionIDs []string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(InstanceSelectionSVC, "BulkDelete")

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error%s", "")
	}

	err = s.saasManager.BulkDeleteWithTx(tx, system, instanceSelectionIDs)
	if err != nil {
		return errorWrapf(err, "saasManager.BulkDeleteWithTx system=`%s`, actionIDs=`%+v` fail",
			system, instanceSelectionIDs)
	}

	return tx.Commit()
}
