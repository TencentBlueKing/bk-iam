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

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// ErrConcurrencyConflict ...
var ErrConcurrencyConflict = errors.New("concurrency conflict")

// AlterGroupAuthType 变更group的auth type
func (s *groupService) AlterGroupAuthType(
	tx *sqlx.Tx,
	systemID string,
	groupPK int64,
	authType int64,
) (changed bool, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectSVC, "AlterGroupAuthType")

	if authType == types.AuthTypeNone {
		count, err := s.authTypeManger.DeleteBySystemGroupWithTx(tx, systemID, groupPK)
		if err != nil {
			return false, errorWrapf(
				err, "authTypeManger.DeleteBySystemGroupWithTx systemID=`%s` groupPK=`%d` fail",
				systemID, groupPK,
			)
		}

		// 用户组auth type变更为none, 则删除用户组成员与系统的关联
		if count == 1 {
			// 查询用户组所有的成员并删除subject system group
			members, err := s.manager.ListGroupMember(groupPK)
			if err != nil {
				return false, errorWrapf(err, "manager.ListGroupMember groupPK=`%d` fail", groupPK)
			}

			subjectSet := set.NewInt64Set()
			for _, relation := range members {
				subjectSet.Add(relation.SubjectPK)
			}

			// 查询用户组模版成员
			relations, err := s.subjectTemplateGroupManager.ListThinRelationWithMaxExpiredAtByGroupPK(groupPK)
			if err != nil {
				return false, errorWrapf(
					err,
					"subjectTemplateGroupManager.ListThinRelationWithMaxExpiredAtByGroupPK groupPK=`%d` fail",
					groupPK,
				)
			}
			for _, relation := range relations {
				subjectSet.Add(relation.SubjectPK)
			}

			for _, subjectPK := range subjectSet.ToSlice() {
				err := s.removeSubjectSystemGroup(tx, subjectPK, systemID, map[int64]int64{groupPK: 0})
				if err != nil {
					return false, errorWrapf(
						err, "removeSubjectSystemGroup member=`%d` systemID=`%s` groupPK=`%d` fail",
						subjectPK, systemID, groupPK,
					)
				}
			}

			return true, nil
		}
	} else {
		created, count, err := s.createOrUpdateGroupAuthType(tx, systemID, groupPK, authType)
		if err != nil {
			return false, errorWrapf(
				err, "createOrUpdateGroupAuthType systemID=`%s` groupPK=`%d` authType=`%d` fail",
				systemID, groupPK, authType,
			)
		}

		if created {
			// 查询用户组所有的成员并添加subject system group
			members, err := s.manager.ListGroupMember(groupPK)
			if err != nil {
				return false, errorWrapf(err, "manager.ListGroupMember groupPK=`%d` fail", groupPK)
			}

			// 查询用户组模版成员
			relations, err := s.subjectTemplateGroupManager.ListThinRelationWithMaxExpiredAtByGroupPK(groupPK)
			if err != nil {
				return false, errorWrapf(
					err,
					"subjectTemplateGroupManager.ListThinRelationWithMaxExpiredAtByGroupPK groupPK=`%d` fail",
					groupPK,
				)
			}

			nowTS := time.Now().Unix()
			subjectExpiredAtMap := make(map[int64]int64, len(relations)+len(members))
			for _, relation := range members {
				if relation.ExpiredAt < nowTS {
					continue
				}

				subjectExpiredAtMap[relation.SubjectPK] = relation.ExpiredAt
			}

			for _, relation := range relations {
				if relation.ExpiredAt < nowTS {
					continue
				}

				// 取过期时间大的
				if relation.ExpiredAt > subjectExpiredAtMap[relation.SubjectPK] {
					subjectExpiredAtMap[relation.SubjectPK] = relation.ExpiredAt
				}
			}

			for subjectPK, expiredAt := range subjectExpiredAtMap {
				err := s.addOrUpdateSubjectSystemGroup(
					tx, subjectPK, systemID, map[int64]int64{groupPK: expiredAt},
				)
				if err != nil {
					return false, errorWrapf(
						err, "addOrUpdateSubjectSystemGroup member=`%d` systemID=`%s` groupPK=`%d` fail",
						subjectPK, systemID, groupPK,
					)
				}
			}
		}

		// 返回有变更
		if count == 1 {
			return true, nil
		}
	}

	return false, nil
}

// createOrUpdateGroupAuthType ...
func (s *groupService) createOrUpdateGroupAuthType(
	tx *sqlx.Tx,
	systemID string,
	groupPK, authType int64,
) (created bool, count int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "createOrUpdateGroupAuthType")

	groupSystemAuthType, err := s.authTypeManger.GetBySystemGroup(systemID, groupPK)
	if errors.Is(err, sql.ErrNoRows) {
		groupSystemAuthType = dao.GroupSystemAuthType{
			SystemID: systemID,
			GroupPK:  groupPK,
			AuthType: authType,
		}
		err = s.authTypeManger.CreateWithTx(tx, groupSystemAuthType)
		if err == nil {
			return true, 1, nil
		}

		if database.IsMysqlDuplicateEntryError(err) {
			return false, 0, ErrConcurrencyConflict
		}
	}

	if err != nil {
		err = errorWrapf(
			err,
			"groupSystemAuthTypeManager.GetBySystemGroup systemID=`%s` groupPK=`%d` fail",
			systemID,
			groupPK,
		)
		return false, 0, err
	}

	// 类型相同, 不需要更新
	if groupSystemAuthType.AuthType == authType {
		return false, 0, nil
	}

	groupSystemAuthType.AuthType = authType
	count, err = s.authTypeManger.UpdateWithTx(tx, groupSystemAuthType)
	if err != nil {
		err = errorWrapf(
			err, "groupSystemAuthTypeManager.UpdateWithTx groupSystemAuthType=`%+v` fail",
			groupSystemAuthType,
		)
		return false, 0, err
	}

	// 并发更新冲突
	if count == 0 {
		return false, 0, ErrConcurrencyConflict
	}

	return false, count, nil
}

// ListGroupAuthSystemIDs 查询group已授权的系统
func (s *groupService) ListGroupAuthSystemIDs(groupPK int64) ([]string, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListGroupAuthSystemIDs")

	groupSystemAuthTypes, err := s.authTypeManger.ListByGroup(groupPK)
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

func (s *groupService) GetGroupOneAuthSystem(groupPK int64) (string, error) {
	return s.authTypeManger.GetOneAuthSystemByGroup(groupPK)
}

// ListGroupAuthBySystemGroupPKs 查询groups的授权类型
func (s *groupService) ListGroupAuthBySystemGroupPKs(systemID string, groupPKs []int64) ([]types.GroupAuthType, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupSVC, "ListGroupAuthBySystemGroupPKs")

	groupAuthTypes := make([]types.GroupAuthType, 0, len(groupPKs))
	// 分段查询, 避免SQL参数过多
	for _, index := range util.Chunks(len(groupPKs), 1000) {
		daoGroupAuthTypes, err := s.authTypeManger.ListAuthTypeBySystemGroups(systemID, groupPKs[index.Begin:index.End])
		if err != nil {
			err = errorWrapf(
				err, "authTypeManger.ListAuthTypeBySystemGroups systemID=`%s` groupPKs=`%+v` fail",
				systemID, groupPKs,
			)
			return nil, err
		}

		for _, daoGroupAuthType := range daoGroupAuthTypes {
			groupAuthTypes = append(groupAuthTypes, types.GroupAuthType{
				GroupPK:  daoGroupAuthType.GroupPK,
				AuthType: daoGroupAuthType.AuthType,
			})
		}
	}

	return groupAuthTypes, nil
}
