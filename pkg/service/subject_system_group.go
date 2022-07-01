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
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// ErrNoPolicies ...
var (
	ErrNoSubjectSystemGroup = errors.New("no subject system group")
	ErrNeedRetry            = errors.New("need retry")
)

// RetryCount ...
const RetryCount = 3

func (l *groupService) doUpdateSubjectSystemGroup(
	tx *sqlx.Tx,
	systemID string,
	subjectPK, groupPK, expiredAt int64,
	createIfNotExists bool,
	updateGroupExpiredAtFunc func(groupExpiredAtMap map[int64]int64) (map[int64]int64, error),
) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "doUpdateSubjectSystemGroup")

	// 查询已有数据
	subjectSystemGroup, err := l.subjectSystemGroupManager.GetBySystemSubject(systemID, subjectPK)
	if createIfNotExists && errors.Is(err, sql.ErrNoRows) {
		// 查不到数据时, 如果需要创建, 则创建
		err = l.createSubjectSystemGroup(tx, systemID, subjectPK, groupPK, expiredAt)
		if database.IsMysqlDuplicateEntryError(err) {
			return ErrNeedRetry
		}

		if err == nil {
			return nil
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
	groups, err := updateGroupsString(subjectSystemGroup.Groups, updateGroupExpiredAtFunc)
	if err != nil {
		err = errorWrapf(err, "updateGroupsString fail, groups=`%s`", subjectSystemGroup.Groups)
		return err
	}

	subjectSystemGroup.Groups = groups
	count, err := l.subjectSystemGroupManager.UpdateWithTx(tx, subjectSystemGroup)
	if err != nil {
		err = errorWrapf(
			err,
			"subjectSystemGroupManager.UpdateWithTx fail, subjectSystemGroup=`%+v`",
			subjectSystemGroup,
		)
		return err
	}

	// 数据未更新时需要重试
	if count == 0 {
		return ErrNeedRetry
	}

	return nil
}

// addOrUpdateSubjectSystemGroup 增加subject-system-group关系或更新过期时间
func (l *groupService) addOrUpdateSubjectSystemGroup(
	tx *sqlx.Tx,
	subjectPK int64,
	systemID string,
	groupPK, expiredAt int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "addOrUpdateSubjectSystemGroup")

	// 更新或创建新的关系
	addOrUpdateFunc := func(groupExpiredAtMap map[int64]int64) (map[int64]int64, error) {
		groupExpiredAtMap[groupPK] = expiredAt
		return groupExpiredAtMap, nil
	}

	// 乐观锁, 重复提交, 最多3次
	for i := 0; i < RetryCount; i++ {
		err = l.doUpdateSubjectSystemGroup(tx, systemID, subjectPK, groupPK, expiredAt, true, addOrUpdateFunc)
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrNeedRetry) {
			continue
		}

		if err != nil {
			err = errorWrapf(
				err, "addOrUpdateSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupPK: %d, expiredAt: %d",
				systemID, subjectPK, groupPK, expiredAt,
			)
			return
		}
	}

	return err
}

// removeSubjectSystemGroup 移除subject-system-group关系
func (l *groupService) removeSubjectSystemGroup(
	tx *sqlx.Tx,
	subjectPK int64,
	systemID string,
	groupPK int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "removeSubjectSystemGroup")

	removeFunc := func(groupExpiredAtMap map[int64]int64) (map[int64]int64, error) {
		_, ok := groupExpiredAtMap[groupPK]
		if !ok {
			return nil, ErrNoSubjectSystemGroup
		}
		delete(groupExpiredAtMap, groupPK)
		return groupExpiredAtMap, nil
	}

	// 乐观锁, 重复提交, 最多3次
	for i := 0; i < RetryCount; i++ {
		err = l.doUpdateSubjectSystemGroup(tx, systemID, subjectPK, groupPK, 0, false, removeFunc)
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrNeedRetry) {
			continue
		}

		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrNoSubjectSystemGroup) {
			// 数据不存在时记录日志
			log.Warningf("removeSubjectSystemGroup not exists systemID=`%s`, subjectPK=`%d`, parentPK=`%d`",
				systemID, subjectPK, groupPK)
			return nil
		}

		if err != nil {
			err = errorWrapf(
				err, "removeSubjectSystemGroup fail, systemID: %s, subjectPK: %d, groupPK: %d",
				systemID, subjectPK, groupPK,
			)
			return
		}
	}

	return err
}

func (l *groupService) createSubjectSystemGroup(
	tx *sqlx.Tx,
	systemID string,
	subjectPK, groupPK, expiredAt int64,
) error {
	groups, err := jsoniter.MarshalToString(map[int64]int64{groupPK: expiredAt})
	if err != nil {
		return err
	}

	subjectSystemGroup := dao.SubjectSystemGroup{
		SystemID:  systemID,
		SubjectPK: subjectPK,
		Groups:    groups,
	}

	return l.subjectSystemGroupManager.CreateWithTx(tx, subjectSystemGroup)
}

// updateGroupsString 更新groups字符串
func updateGroupsString(
	groups string,
	updateGroupExpiredAtFunc func(map[int64]int64) (map[int64]int64, error),
) (string, error) {
	var groupExpiredAtMap map[int64]int64 = make(map[int64]int64)
	if groups != "" {
		err := jsoniter.UnmarshalFromString(groups, &groupExpiredAtMap)
		if err != nil {
			return "", err
		}
	}

	groupExpiredAtMap, err := updateGroupExpiredAtFunc(groupExpiredAtMap)
	if err != nil {
		return "", err
	}
	groups, err = jsoniter.MarshalToString(groupExpiredAtMap)
	if err != nil {
		return "", err
	}
	return groups, nil
}

// ListEffectThinSubjectGroups 批量获取 subject 有效的 groups(未过期的)
func (l *groupService) ListEffectThinSubjectGroups(
	systemID string,
	pks []int64,
) (subjectGroups map[int64][]types.ThinSubjectGroup, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListEffectThinSubjectGroups")

	subjectGroups = make(map[int64][]types.ThinSubjectGroup, len(pks))

	relations, err := l.subjectSystemGroupManager.ListSubjectGroups(systemID, pks)
	if err != nil {
		return subjectGroups, errorWrapf(err, "manager.ListRelationByPKs pks=`%+v` fail", pks)
	}

	// 筛选未过期的用户组
	now := time.Now().Unix()
	for _, r := range relations {
		subjectPK := r.SubjectPK

		thinSubjectGroup, err := convertSystemSubjectGroupsToThinSubjectGroup(r.Groups)
		if err != nil {
			err = errorWrapf(
				err, "convertSystemSubjectGroupsToThinSubjectGroup fail, groups=`%s`", r.Groups,
			)
			return nil, err
		}

		for _, group := range thinSubjectGroup {
			if group.PolicyExpiredAt > now {
				subjectGroups[subjectPK] = append(subjectGroups[subjectPK], group)
			}
		}
	}

	return subjectGroups, nil
}

func convertSystemSubjectGroupsToThinSubjectGroup(
	groups string,
) (thinSubjectGroup []types.ThinSubjectGroup, err error) {
	var groupExpiredAtMap map[int64]int64 = make(map[int64]int64)
	if groups != "" {
		err := jsoniter.UnmarshalFromString(groups, &groupExpiredAtMap)
		if err != nil {
			return nil, err
		}
	}

	for groupPK, expiredAt := range groupExpiredAtMap {
		thinSubjectGroup = append(thinSubjectGroup, types.ThinSubjectGroup{
			GroupPK:         groupPK,
			PolicyExpiredAt: expiredAt,
		})
	}

	return thinSubjectGroup, nil
}
