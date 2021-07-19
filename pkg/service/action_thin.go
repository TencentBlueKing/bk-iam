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

import (
	"iam/pkg/database/dao"
	"iam/pkg/errorx"
	"iam/pkg/service/types"
)

func convertToThinActions(actions []dao.Action) []types.ThinAction {
	thinActions := make([]types.ThinAction, 0, len(actions))
	for _, a := range actions {
		thinActions = append(thinActions, types.ThinAction{
			PK:     a.PK,
			ID:     a.ID,
			System: a.System,
		})
	}
	return thinActions
}

// ListThinActionByPKs 批量查询ActionID
func (l *actionService) ListThinActionByPKs(pks []int64) ([]types.ThinAction, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListThinActionByPKs")

	actions, err := l.manager.ListByPKs(pks)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListBySPKs pks=`%v` fail", pks)
	}

	return convertToThinActions(actions), nil
}

// ListThinActionBySystem ...
func (l *actionService) ListThinActionBySystem(system string) ([]types.ThinAction, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListThinActionBySystem")

	actions, err := l.manager.ListBySystem(system)
	if err != nil {
		return nil, errorWrapf(err, "manager.ListBySystem system=`%s` fail", system)
	}

	return convertToThinActions(actions), nil
}
