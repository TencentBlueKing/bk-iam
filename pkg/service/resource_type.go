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
	"iam/pkg/database/dao"
	"iam/pkg/database/sdao"
	"iam/pkg/service/types"
)

// NOTE: service层的error全部wrap后return, 不记录日志

// ResourceTypeSVC ...
const ResourceTypeSVC = "ResourceTypeSVC"

// ResourceTypeService ...
type ResourceTypeService interface {
	ListBySystem(system string) ([]types.ResourceType, error)

	BulkCreate(system string, resourceTypes []types.ResourceType) error
	Update(system, resourceTypeID string, resourceType types.ResourceType) error
	BulkDelete(system string, resourceTypeIDs []string) error

	Get(system string, id string) (types.ResourceType, error)
	GetPK(system string, name string) (int64, error)
	GetThinByPK(pk int64) (types.ThinResourceType, error)
}

type resourceTypeService struct {
	manager     dao.ResourceTypeManager
	saasManager sdao.SaaSResourceTypeManager
}

// NewResourceTypeService ...
func NewResourceTypeService() ResourceTypeService {
	return &resourceTypeService{
		manager:     dao.NewResourceTypeManager(),
		saasManager: sdao.NewSaaSResourceTypeManager(),
	}
}

// ListBySystem ...
func (l *resourceTypeService) ListBySystem(system string) (allResourceTypes []types.ResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ResourceTypeSVC, "ListBySystem")

	saasResourceTypes, err := l.saasManager.ListBySystem(system)
	if err != nil {
		return
	}
	resourceTypes := make([]types.ResourceType, 0, len(saasResourceTypes))
	for _, rt := range saasResourceTypes {
		resourceType := types.ResourceType{
			ID:            rt.ID,
			Name:          rt.Name,
			NameEn:        rt.NameEn,
			Description:   rt.Description,
			DescriptionEn: rt.DescriptionEn,
			Version:       rt.Version,
		}

		// NOTE: the input Parents maybe empty string!
		if rt.Parents != "" {
			err = jsoniter.UnmarshalFromString(rt.Parents, &resourceType.Parents)
			if err != nil {
				err = errorWrapf(err, "unmarshal rt.Parents=`%s` fail", rt.Parents)
				return
			}
		}

		err = jsoniter.UnmarshalFromString(rt.ProviderConfig, &resourceType.ProviderConfig)
		if err != nil {
			err = errorWrapf(err, "unmarshal rt.ProviderConfig=`%s` fail", rt.ProviderConfig)
			return
		}
		resourceTypes = append(resourceTypes, resourceType)
	}
	return resourceTypes, nil
}

// Get ...
func (l *resourceTypeService) Get(system string, resourceTypeID string) (rt types.ResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ResourceTypeSVC, "Get")

	srt, err := l.saasManager.Get(system, resourceTypeID)
	if err != nil {
		return
	}

	resourceType := types.ResourceType{
		ID:            srt.ID,
		Name:          srt.Name,
		NameEn:        srt.NameEn,
		Description:   srt.Description,
		DescriptionEn: srt.DescriptionEn,
		Version:       srt.Version,
	}

	if srt.Parents != "" {
		err = jsoniter.UnmarshalFromString(srt.Parents, &resourceType.Parents)
		if err != nil {
			err = errorWrapf(err, "unmarshal resourceType.Parents=`%s` fail", srt.Parents)
			return
		}
	}

	err = jsoniter.UnmarshalFromString(srt.ProviderConfig, &resourceType.ProviderConfig)
	if err != nil {
		err = errorWrapf(err, "unmarshal resourceType.ProviderConfig=`%s` fail", srt.ProviderConfig)
		return
	}
	return resourceType, nil
}

// GetPK ...
func (l *resourceTypeService) GetPK(system string, resourceTypeID string) (int64, error) {
	return l.manager.GetPK(system, resourceTypeID)
}

// BulkCreate ...
func (l *resourceTypeService) BulkCreate(system string, resourceTypes []types.ResourceType) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ResourceTypeSVC, "BulkCreate")

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error system=`%s`", system)
	}

	// 数据转换
	dbResourceTypes := make([]dao.ResourceType, 0, len(resourceTypes))
	dbSaaSResourceTypes := make([]sdao.SaaSResourceType, 0, len(resourceTypes))
	for _, rt := range resourceTypes {
		parents, err1 := jsoniter.MarshalToString(rt.Parents)
		if err1 != nil {
			return errorWrapf(err, "marshal rt.Parents=`%+v` fail", rt.Parents)
		}
		providerConfig, err1 := jsoniter.MarshalToString(rt.ProviderConfig)
		if err1 != nil {
			return errorWrapf(err, "marshal rt.ProviderConfig=`%+v` fail", rt.ProviderConfig)
		}
		dbResourceTypes = append(dbResourceTypes, dao.ResourceType{
			System: system,
			ID:     rt.ID,
		})
		dbSaaSResourceTypes = append(dbSaaSResourceTypes, sdao.SaaSResourceType{
			System:         system,
			ID:             rt.ID,
			Name:           rt.Name,
			NameEn:         rt.NameEn,
			Description:    rt.Description,
			DescriptionEn:  rt.DescriptionEn,
			Sensitivity:    rt.Sensitivity,
			Parents:        parents,
			ProviderConfig: providerConfig,
		})
	}

	// 执行插入
	err = l.manager.BulkCreateWithTx(tx, dbResourceTypes)
	if err != nil {
		return errorWrapf(err, "manager.BulkCreateWithTx fail%s", "")
	}

	err = l.saasManager.BulkCreateWithTx(tx, dbSaaSResourceTypes)
	if err != nil {
		return errorWrapf(err, "saasManager.BulkCreateWithTx fail%s", "")
	}

	return tx.Commit()
}

// Update ...
func (l *resourceTypeService) Update(
	system, resourceTypeID string,
	resourceType types.ResourceType,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ResourceTypeSVC, "Update")

	var err error

	// BUG: if resourceType.Parents is empty and allowBlank => will be set to "" instead of [] => we need []
	var parents string
	// if len(resourceType.Parents) > 0 {
	parents, err = jsoniter.MarshalToString(resourceType.Parents)
	if err != nil {
		return errorWrapf(err, "marshal resourceType.Parent=`%+v` fail", resourceType.Parents)
	}
	// }

	var providerConfig string
	if len(resourceType.ProviderConfig) > 0 {
		providerConfig, err = jsoniter.MarshalToString(resourceType.ProviderConfig)
		if err != nil {
			return errorWrapf(err, "marshal resourceType.ProviderConfig=`%+v` fail", resourceType.ProviderConfig)
		}
	}

	allowBlank := database.NewAllowBlankFields()
	// resourceType.Parents is set => set saasResourceType.Parents
	if resourceType.AllowEmptyFields.HasKey("Parents") {
		allowBlank.AddKey("Parents")
	}
	if resourceType.AllowEmptyFields.HasKey("Description") {
		allowBlank.AddKey("Description")
	}
	if resourceType.AllowEmptyFields.HasKey("DescriptionEn") {
		allowBlank.AddKey("DescriptionEn")
	}
	if resourceType.AllowEmptyFields.HasKey("Sensitivity") {
		allowBlank.AddKey("Sensitivity")
	}

	data := sdao.SaaSResourceType{
		// PK:             0,
		// System:         "",
		// ID:             "",
		Name:           resourceType.Name,
		NameEn:         resourceType.NameEn,
		Description:    resourceType.Description,
		DescriptionEn:  resourceType.DescriptionEn,
		Sensitivity:    resourceType.Sensitivity,
		Version:        resourceType.Version,
		Parents:        parents,
		ProviderConfig: providerConfig,

		AllowBlankFields: allowBlank,
	}

	return l.saasManager.Update(system, resourceTypeID, data)
}

// BulkDelete ...
func (l *resourceTypeService) BulkDelete(system string, resourceTypeIDs []string) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ResourceTypeSVC, "BulkDelete")

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)

	if err != nil {
		return errorWrapf(err, "define tx error system=`%s`", system)
	}

	err = l.manager.BulkDeleteWithTx(tx, system, resourceTypeIDs)
	if err != nil {
		return errorWrapf(err, "manager.BulkDeleteWithTx fail%s", "")
	}

	err = l.saasManager.BulkDeleteWithTx(tx, system, resourceTypeIDs)
	if err != nil {
		return errorWrapf(err, "saasManager.BulkDeleteWithTx fail%s", "")
	}

	return tx.Commit()
}

// GetByPK ...
func (l *resourceTypeService) GetThinByPK(pk int64) (resourceType types.ThinResourceType, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ResourceTypeSVC, "GetByPK")

	dbResourceType, err := l.manager.GetByPK(pk)
	if err != nil {
		return resourceType, errorWrapf(err, "manager.GetByPK fail, pk=`%d`", pk)
	}

	resourceType = types.ThinResourceType{
		PK:     dbResourceType.PK,
		System: dbResourceType.System,
		ID:     dbResourceType.ID,
	}

	return
}
