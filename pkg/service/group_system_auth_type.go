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
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/database/dao"
)

// 用于授权时处理
func (l *subjectService) createOrUpdateGroupAuthType(
	tx *sqlx.Tx,
	systemID string,
	groupPK, authType int64,
) (created bool, rows int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "createOrUpdateGroupAuthType")

	groupSystemAuthType := dao.GroupSystemAuthType{
		GroupPK:  groupPK,
		SystemID: systemID,
		AuthType: authType,
		CreateAt: time.Now(),
	}

	err = l.groupSystemAuthTypeManager.CreateWithTx(tx, groupSystemAuthType)

	// 创建时如果已经存在，则更新
	if isMysqlDuplicateError(err) {
		rows, err = l.groupSystemAuthTypeManager.UpdateWithTx(tx, groupSystemAuthType)
		if err != nil {
			err = errorWrapf(
				err,
				"groupSystemAuthTypeManager.UpdateWithTx groupSystemAuthType=`%+v` fail",
				groupSystemAuthType,
			)
		}
		return false, rows, err
	} else if err != nil {
		err = errorWrapf(
			err,
			"groupSystemAuthTypeManager.CreateWithTx groupSystemAuthType=`%+v` fail",
			groupSystemAuthType,
		)
		return true, 0, err
	}

	return true, 1, err
}

// listGroupAuthSystem 查询group已授权的系统
func (l *subjectService) listGroupAuthSystem(groupPK int64) ([]string, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "listGroupAuthSystem")

	groupSystemAuthTypes, err := l.groupSystemAuthTypeManager.ListByGroup(groupPK)
	if err != nil {
		err = errorWrapf(
			err,
			"groupSystemAuthTypeManager.ListByGroup groupPK=`%d` fail",
			groupPK,
		)
		return nil, err
	}

	systems := make([]string, 0, len(groupSystemAuthTypes))
	for _, groupSystemAuthType := range groupSystemAuthTypes {
		systems = append(systems, groupSystemAuthType.SystemID)
	}

	return systems, nil
}
