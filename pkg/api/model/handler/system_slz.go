/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

type systemProviderConfig struct {
	// TODO: valid host?
	Host string `json:"host" structs:"host" binding:"required,url" example:"http://bkpaas.service.consul"`
	Auth string `json:"auth" structs:"auth" binding:"required,oneof=none basic" example:"basic"`

	Healthz string `json:"healthz" structs:"healthz" binding:"omitempty" example:"/healthz"`
}

type systemSerializer struct {
	ID            string `json:"id" binding:"required,max=32" example:"bk_paas"`
	Name          string `json:"name" binding:"required" example:"bk_paas"`
	NameEn        string `json:"name_en" binding:"required" example:"bk_paas"`
	Description   string `json:"description" binding:"omitempty" example:"Platform as A Service"`
	DescriptionEn string `json:"description_en" binding:"omitempty" example:"Platform as A Service"`
	Clients       string `json:"clients" binding:"required" example:"bk_paas,bk_esb"`

	ProviderConfig systemProviderConfig `json:"provider_config" binding:"required"`
}

type systemUpdateSerializer struct {
	Name          string `json:"name" binding:"omitempty" example:"bk_paas"`
	NameEn        string `json:"name_en" binding:"omitempty" example:"bk_paas"`
	Description   string `json:"description" binding:"omitempty" example:"Platform as A Service"`
	DescriptionEn string `json:"description_en" binding:"omitempty" example:"Platform as A Service"`
	Clients       string `json:"clients" binding:"omitempty" example:"bk_paas,bk_esb"`

	ProviderConfig *systemProviderConfig `json:"provider_config" binding:"omitempty"`
}

func (s *systemUpdateSerializer) validate(keys map[string]interface{}) (bool, string) {
	if _, ok := keys["name"]; ok {
		if s.Name == "" {
			return false, "name should not be empty"
		}
	}

	if _, ok := keys["name_en"]; ok {
		if s.NameEn == "" {
			return false, "name_en should not be empty"
		}
	}

	if _, ok := keys["clients"]; ok {
		if s.Clients == "" {
			return false, "clients should not be empty"
		}
	}

	if _, ok := keys["provider_config"]; ok {
		if s.ProviderConfig.Host == "" {
			return false, "provider_config should contains key: host, and host should not be empty"
		}
	}
	return true, "ok"
}

type systemResponse struct {
	ID             string                 `json:"id" example:"bk_paas"`
	Name           string                 `json:"name" example:"bk_paas"`
	NameEn         string                 `json:"name_en" example:"bk_paas"`
	Description    string                 `json:"description" example:"Platform as A Service"`
	DescriptionEn  string                 `json:"description_en" example:"Platform as A Service"`
	Clients        string                 `json:"clients" example:"bk_paas,bk_esb"`
	ProviderConfig map[string]interface{} `json:"provider_config"`
}

type systemCreateResponse struct {
	ID string `json:"id" example:"bk_paas"`
}

type systemClientsResponse struct {
	Clients string `json:"clients" example:"bk_paas,bk_esb"`
}
