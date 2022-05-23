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
	ErrNeedRetry            = errors.New("need retry")
)

func (l *subjectService) doUpdateSubjectSystemGroup(
	tx *sqlx.Tx,
	systemID string,
	subjectPK, groupPK, expiredAt int64,
	updateGroupExpiredAtFunc func([]types.GroupExpiredAt) ([]types.GroupExpiredAt, error),
	createSubjectSystemGroupFunc func(tx *sqlx.Tx, systemID string, subjectPK, groupPK, expiredAt int64) error,
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "doUpdateSubjectSystemGroup")

	// 查询已有数据
	subjectSystemGroup, err := l.subjectSystemGroupManager.GetBySystemSubject(systemID, subjectPK)
	if errors.Is(err, sql.ErrNoRows) && createSubjectSystemGroupFunc != nil {
		// 如果需要创建, 则创建
		err = createSubjectSystemGroupFunc(tx, systemID, subjectPK, groupPK, expiredAt)
		if isMysqlDuplicateError(err) {
			return ErrNeedRetry
		}
	}

	if err != nil {
		err = errorWrapf(
			err, "subjectSystemGroupManager.GetBySystemSubject fail, systemID=`%s`, subjectPK=`%d`",
			systemID, subjectPK,
		)
		return err
	}

	// 记录存在则更新
	groupExpiredAts, err := convertToGroupExpiredAt(subjectSystemGroup.Groups)
	if err != nil {
		err = errorWrapf(err, "convertToGroupExpiredAt fail, groups=`%s`", subjectSystemGroup.Groups)
		return err
	}
	groupExpiredAts, err = updateGroupExpiredAtFunc(groupExpiredAts)
	if err != nil {
		err = errorWrapf(err, "updateGroupExpiredAtFunc fail, groupExpiredAts=`%+v`", groupExpiredAts)
		return err
	}

	// 更新记录
	rows, err := l.updateSubjectSystemGroup(tx, systemID, subjectPK, groupExpiredAts)
	if err != nil {
		err = errorWrapf(
			err, "updateSubjectSystemGroup fail, systemID=`%s`, subjectPK=`%d`, groupExpiredAts=`%+v`",
			systemID, subjectPK, groupExpiredAts,
		)
		return err
	}

	// 重试失败直接报错
	if rows == 0 {
		return ErrNeedRetry
	}

	return nil
}

// addOrUpdateSubjectGroup 增加subject-system-group关系或更新过期时间
func (l *subjectService) addOrUpdateSubjectSystemGroup(
	tx *sqlx.Tx,
	systemID string,
	subjectPK, groupPK, expiredAt int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "addOrUpdateSubjectSystemGroup")

	// 更新或创建新的关系
	updateGroupExpiredAtFunc := func(groupExpiredAts []types.GroupExpiredAt) ([]types.GroupExpiredAt, error) {
		i := findGroupIndex(groupExpiredAts, groupPK)
		if i == -1 {
			groupExpiredAts = append(groupExpiredAts, types.GroupExpiredAt{GroupPK: groupPK, ExpiredAt: expiredAt})
		} else {
			groupExpiredAts[i].ExpiredAt = expiredAt
		}
		return groupExpiredAts, nil
	}

	// 乐观锁, 重复提交, 最多3次
	for i := 0; i < 3; i++ {
		err = l.doUpdateSubjectSystemGroup(tx, systemID, subjectPK, groupPK, expiredAt, updateGroupExpiredAtFunc, l.createSubjectSystemGroup)
		if errors.Is(err, ErrNeedRetry) {
			continue
		}

		if err != nil {
			err = errorWrapf(
				err, "addOrUpdateSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupPK: %d, expiredAt: %d",
				systemID, subjectPK, groupPK, expiredAt,
			)
			return err
		}
	}

	return err
}

// removeSubjectSystemGroup 移除subject-system-group关系
func (l *subjectService) removeSubjectSystemGroup(
	tx *sqlx.Tx,
	systemID string,
	subjectPK, groupPK int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "removeSubjectSystemGroup")

	updateGroupExpiredAtFunc := func(groupExpiredAts []types.GroupExpiredAt) ([]types.GroupExpiredAt, error) {
		i := findGroupIndex(groupExpiredAts, groupPK)
		if i == -1 {
			return nil, ErrNoSubjectSystemGroup
		}
		return append(groupExpiredAts[:i], groupExpiredAts[i+1:]...), nil
	}

	// 乐观锁, 重复提交, 最多3次
	for i := 0; i < 3; i++ {
		err = l.doUpdateSubjectSystemGroup(tx, systemID, subjectPK, groupPK, 0, updateGroupExpiredAtFunc, nil)
		if errors.Is(err, ErrNeedRetry) {
			continue
		}

		if err != nil {
			err = errorWrapf(
				err, "removeSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupPK: %d",
				systemID, subjectPK, groupPK,
			)
			return err
		}
	}

	return err
}

func (l *subjectService) createSubjectSystemGroup(tx *sqlx.Tx, systemID string, subjectPK, groupPK, expiredAt int64) error {
	groups, err := convertToGroupsString([]types.GroupExpiredAt{{GroupPK: groupPK, ExpiredAt: expiredAt}})
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
