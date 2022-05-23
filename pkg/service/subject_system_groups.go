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
	"database/sql"
	"errors"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// ErrNoPolicies ...
var (
	ErrNoSubjectSystemGroup = errors.New("no subject system group")
	ErrRetryFail            = errors.New("retry fail")
)

// addOrUpdateSubjectGroup 增加subject-system-group关系或更新过期时间
func (l *subjectService) addOrUpdateSubjectSystemGroup(
	tx *sqlx.Tx,
	systemID string,
	subjectPK int64,
	groupExpiredAt types.GroupExpiredAt,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "addOrUpdateSubjectSystemGroup")

	// 乐观锁, 重复提交, 最多3次
	for i := 0; i < 3; i++ {
		// 查询已有数据
		subjectSystemGroup, err := l.subjectSystemGroupManager.GetBySystemSubject(systemID, subjectPK)

		if errors.Is(err, sql.ErrNoRows) {
			// 记录不存在, 则创建
			err = l.createSubjectSystemGroup(tx, systemID, subjectPK, groupExpiredAt)

			if isMysqlDuplicateError(err) {
				continue
			}

			if err != nil {
				err = errorWrapf(
					err, "createSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupExpiredAt: %v",
					systemID, subjectPK, groupExpiredAt,
				)
			}

			return err

		} else if err != nil {
			err = errorWrapf(
				err, "subjectSystemGroupManager.GetBySystemSubject fail, systemID: %s, subjectPK: %d",
				systemID, subjectPK,
			)

			return err
		}

		// 记录存在则更新
		groupExpiredAts, err := convertToGroupExpiredAt(subjectSystemGroup.Groups)
		if err != nil {
			err = errorWrapf(err, "convertToGroupExpiredAt fail, groups=`%+v`", subjectSystemGroup.Groups)
			return err
		}

		index := findGroupIndex(groupExpiredAts, groupExpiredAt.GroupPK)
		if index == -1 {
			groupExpiredAts = append(groupExpiredAts, groupExpiredAt)
		} else {
			groupExpiredAts[index] = groupExpiredAt
		}

		rows, err := l.updateSubjectSystemGroup(tx, systemID, subjectPK, groupExpiredAts)
		if err != nil {
			err = errorWrapf(
				err, "updateSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupExpiredAts: %+v",
				systemID, subjectPK, groupExpiredAts,
			)
			return err
		}

		// 重试失败直接报错
		if rows == 0 && i == 2 {
			err = errorWrapf(
				ErrRetryFail, "retry updateSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupExpiredAts: %+v",
				systemID, subjectPK, groupExpiredAts,
			)
			return err
		} else if rows == 1 {
			return nil
		}
	}

	return nil
}

// removeSubjectSystemGroup 移除subject-system-group关系
func (l *subjectService) removeSubjectSystemGroup(
	tx *sqlx.Tx,
	systemID string,
	subjectPK, groupPK int64,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "removeSubjectSystemGroup")

	// 乐观锁, 重复提交, 最多3次
	for i := 0; i < 3; i++ {
		// 查询已有数据
		subjectSystemGroup, err := l.subjectSystemGroupManager.GetBySystemSubject(systemID, subjectPK)
		if err != nil {
			err = errorWrapf(
				err, "subjectSystemGroupManager.GetBySystemSubject fail, systemID: %s, subjectPK: %d",
				systemID, subjectPK,
			)

			return err
		}

		// 记录存在, 则更新
		groupExpiredAts, err := convertToGroupExpiredAt(subjectSystemGroup.Groups)
		if err != nil {
			err = errorWrapf(err, "convertToGroupExpiredAt fail, groups=`%+v`", subjectSystemGroup.Groups)
			return err
		}

		index := findGroupIndex(groupExpiredAts, groupPK)
		if index == -1 {
			err = errorWrapf(
				ErrNoSubjectSystemGroup, "findGroupIndex fail, systemID: %s, subjectPK: %d, groupPK: %d",
				systemID, subjectPK, groupPK,
			)
			return err
		}

		// remove
		groupExpiredAts = append(groupExpiredAts[:index], groupExpiredAts[index+1:]...)
		rows, err := l.updateSubjectSystemGroup(tx, systemID, subjectPK, groupExpiredAts)
		if err != nil {
			err = errorWrapf(
				err, "updateSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupExpiredAts: %+v",
				systemID, subjectPK, groupExpiredAts,
			)
			return err
		}

		if rows == 0 && i == 2 {
			err = errorWrapf(
				ErrRetryFail, "retry updateSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupExpiredAts: %+v",
				systemID, subjectPK, groupExpiredAts,
			)
			return err
		} else if rows == 1 {
			return nil
		}
	}

	return nil
}

func (l *subjectService) createSubjectSystemGroup(tx *sqlx.Tx, systemID string, subjectPK int64, groupExpiredAt types.GroupExpiredAt) error {
	groups, err := convertToGroupsString([]types.GroupExpiredAt{groupExpiredAt})
	if err != nil {
		return err
	}

	subjectSystemGroup := dao.SubjectSystemGroup{
		SystemID:  systemID,
		SubjectPK: subjectPK,
		Groups:    groups,
		CreateAt:  time.Now(),
	}

	return l.subjectSystemGroupManager.CreateWithTx(tx, subjectSystemGroup)
}

func (l *subjectService) updateSubjectSystemGroup(tx *sqlx.Tx, systemID string, subjectPK int64, groupExpiredAts []types.GroupExpiredAt) (int64, error) {
	groups, err := convertToGroupsString(groupExpiredAts)
	if err != nil {
		return 0, err
	}

	subjectSystemGroup := dao.SubjectSystemGroup{
		SystemID:  systemID,
		SubjectPK: subjectPK,
		Groups:    groups,
	}
	rows, err := l.subjectSystemGroupManager.UpdateWithTx(tx, subjectSystemGroup)
	if err != nil {
		return 0, err
	}
	return rows, nil
}

// convertToGroupExpiredAt 转换为结构
func convertToGroupExpiredAt(groups string) (groupExpiredAts []types.GroupExpiredAt, err error) {
	if groups == "" {
		return []types.GroupExpiredAt{}, nil
	}

	err = jsoniter.UnmarshalFromString(groups, &groupExpiredAts)
	return
}

// convertToGroupExpiredAt 转换为结构
func convertToGroupsString(groupExpiredAts []types.GroupExpiredAt) (groups string, err error) {
	return jsoniter.MarshalToString(groupExpiredAts)
}

func findGroupIndex(groupExpiredAts []types.GroupExpiredAt, groupPK int64) int {
	for i := range groupExpiredAts {
		if groupExpiredAts[i].GroupPK == groupPK {
			return i
		}
	}

	return -1
}

func isMysqlDuplicateError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	return false
}
