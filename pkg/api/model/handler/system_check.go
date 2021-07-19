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

import (
	"errors"
	"fmt"

	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
)

// AllSystems ...
type AllSystems struct {
	AllBaseInfo
	Systems []svctypes.System
}

// NewAllSystems ...
func NewAllSystems(systems []svctypes.System) *AllSystems {
	nameSet := map[string]string{}
	nameEnSet := map[string]string{}
	for _, sys := range systems {
		nameSet[sys.Name] = sys.ID
		nameEnSet[sys.NameEn] = sys.ID
	}

	return &AllSystems{
		AllBaseInfo: AllBaseInfo{
			IDSet:     map[string]string{},
			NameSet:   nameSet,
			NameEnSet: nameEnSet,
		},
		Systems: systems,
	}
}

func checkSystemCreateUnique(id, name, nameEn string) error {
	svc := service.NewSystemService()
	// check system is exists
	if svc.Exists(id) {
		return fmt.Errorf("system(%s) exists", id)
	}
	// check system name and name_en is unique
	systems, err := svc.ListAll()
	if err != nil {
		return errors.New("query all system fail")
	}

	allSystems := NewAllSystems(systems)
	if allSystems.ContainsName(name) {
		return fmt.Errorf("system name(%s) already exists", name)
	}
	if allSystems.ContainsNameEn(nameEn) {
		return fmt.Errorf("system name_en(%s) already exists", nameEn)
	}
	return nil
}

func checkSystemUpdateUnique(id, name, nameEn string) error {
	svc := service.NewSystemService()
	// check system name and name_en is unique
	systems, err := svc.ListAll()
	if err != nil {
		return errors.New("query all system fail")
	}

	allSystems := NewAllSystems(systems)
	if name != "" && allSystems.ContainsNameExcludeSelf(name, id) {
		return fmt.Errorf("system name(%s) already exists", name)
	}
	if nameEn != "" && allSystems.ContainsNameEnExcludeSelf(nameEn, id) {
		return fmt.Errorf("system name_en(%s) already exists", nameEn)
	}
	return nil
}
