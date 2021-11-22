/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package types

// ActionResourceType ...
type ActionResourceType struct {
	System      string `json:"system_id" structs:"system_id"`
	ID          string `json:"id" structs:"id"`
	NameAlias   string `json:"name_alias" structs:"name_alias"`
	NameAliasEn string `json:"name_alias_en" structs:"name_alias_en"`

	SelectionMode string `json:"selection_mode" structs:"selection_mode"`
	// for input, to db storage
	RelatedInstanceSelections []map[string]interface{} `json:"-" structs:"related_instance_selections"`

	// for output, to api display
	InstanceSelections []map[string]interface{} `json:"instance_selections" structs:"instance_selections"`
}

type ActionEnvironment struct {
	Type      string   `json:"type" structs:"type"`
	Operators []string `json:"operators" structs:"operators"`
}

// ReferenceInstanceSelection ...
type ReferenceInstanceSelection struct {
	System        string `json:"system_id" structs:"system_id"`
	ID            string `json:"id" structs:"id"`
	IgnoreIAMPath bool   `json:"ignore_iam_path" structs:"ignore_iam_path"`
}

// Action ...
type Action struct {
	AllowEmptyFields

	ID                   string               `json:"id" structs:"id"`
	Name                 string               `json:"name" structs:"name"`
	NameEn               string               `json:"name_en" structs:"name_en"`
	Description          string               `json:"description" structs:"description"`
	DescriptionEn        string               `json:"description_en" structs:"description_en"`
	Type                 string               `json:"type" structs:"type"`
	Version              int64                `json:"version" structs:"version"`
	RelatedResourceTypes []ActionResourceType `json:"related_resource_types" structs:"related_resource_types"`
	RelatedActions       []string             `json:"related_actions" structs:"related_actions"`
	RelatedEnvironments  []ActionEnvironment  `json:"related_environments" structs:"related_environments"`
}

// ThinAction ...
type ThinAction struct {
	PK     int64
	System string
	ID     string
}

// ActionResourceTypeID ...
type ActionResourceTypeID struct {
	ActionSystem       string
	ActionID           string
	ResourceTypeSystem string
	ResourceTypeID     string
}

// ActionInstanceSelectionID ...
type ActionInstanceSelectionID struct {
	ActionSystem            string
	ActionID                string
	InstanceSelectionSystem string
	InstanceSelectionID     string
}
