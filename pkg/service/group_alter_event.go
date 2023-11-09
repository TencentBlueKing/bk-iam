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
	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// GroupAlterEventSVC ...
const GroupAlterEventSVC = "GroupAlterEventSVC"

// GroupAlterEventService ...
type GroupAlterEventService interface {
	ListBeforeCreateAt(createdAt int64, limit int64) ([]types.GroupAlterEvent, error)

	CreateByGroupAction(groupPK int64, actionPKs []int64) error
	CreateByGroupSubject(groupPK int64, subjectPKs []int64) error
	CreateBySubjectActionGroup(subjectPK, actionPK, groupPK int64) error

	BulkDeleteWithTx(tx *sqlx.Tx, uuids []string) (err error)
}

type groupAlterEventService struct {
	manager                     dao.GroupAlterEventManager
	subjectGroupManager         dao.SubjectGroupManager
	subjectTemplateGroupManager dao.SubjectTemplateGroupManager
	groupResourcePolicyManager  dao.GroupResourcePolicyManager
}

// NewGroupAlterEventService ...
func NewGroupAlterEventService() GroupAlterEventService {
	return &groupAlterEventService{
		manager:                     dao.NewGroupAlterEventManager(),
		subjectGroupManager:         dao.NewSubjectGroupManager(),
		subjectTemplateGroupManager: dao.NewSubjectTemplateGroupManager(),
		groupResourcePolicyManager:  dao.NewGroupResourcePolicyManager(),
	}
}

// ListBeforeCreateAt ...
func (s *groupAlterEventService) ListBeforeCreateAt(
	createdAt int64,
	limit int64,
) (events []types.GroupAlterEvent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupAlterEventSVC, "ListBeforeCreateAt")

	daoEvents, err := s.manager.ListBeforeCreateAt(createdAt, limit)
	if err != nil {
		err = errorWrapf(err, "manager.ListBeforeCreateAt createdAt=`%d` limit=`%d` fail", createdAt, limit)
		return
	}

	events = make([]types.GroupAlterEvent, 0, len(daoEvents))
	for _, daoEvent := range daoEvents {
		event, err := convertToSvcGroupAlterEvent(daoEvent)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return
}

func convertToSvcGroupAlterEvent(daoEvent dao.GroupAlterEvent) (types.GroupAlterEvent, error) {
	event := types.GroupAlterEvent{
		UUID:    daoEvent.UUID,
		GroupPK: daoEvent.GroupPK,
	}

	err := jsoniter.UnmarshalFromString(daoEvent.ActionPKs, &event.ActionPKs)
	if err != nil {
		return event, err
	}

	err = jsoniter.UnmarshalFromString(daoEvent.SubjectPKs, &event.SubjectPKs)
	if err != nil {
		return event, err
	}
	return event, nil
}

// CreateByGroupAction ...
func (s *groupAlterEventService) CreateByGroupAction(
	groupPK int64,
	actionPKs []int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateByGroupAction")

	if len(actionPKs) == 0 {
		return nil
	}

	subjectRelations, err := s.subjectGroupManager.ListGroupMember(groupPK)
	if err != nil {
		err = errorWrapf(err, "subjectGroupManager.ListGroupMember groupPK=`%d` fail", groupPK)
		return
	}

	// 查询 subject template group
	subjectPKs, err := s.subjectTemplateGroupManager.ListGroupDistinctSubjectPK(groupPK)
	if err != nil {
		err = errorWrapf(err, "subjectTemplateGroupManager.ListGroupDistinctSubjectPK groupPK=`%d` fail", groupPK)
		return
	}

	subjectPKset := set.NewInt64SetWithValues(subjectPKs)
	for _, r := range subjectRelations {
		subjectPKset.Add(r.SubjectPK)
	}

	subjectPKs = subjectPKset.ToSlice()
	if len(subjectPKs) == 0 {
		return nil
	}

	err = s.create(groupPK, actionPKs, subjectPKs)
	if err != nil {
		err = errorWrapf(err, "create fail groupPK=`%d` actionPKs=`%+v` subjectPKs=`%+v`", actionPKs, subjectPKs)
		return
	}

	return nil
}

// CreateByGroupSubject ...
func (s *groupAlterEventService) CreateByGroupSubject(
	groupPK int64,
	subjectPKs []int64,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateByGroupSubject")

	if len(subjectPKs) == 0 {
		return nil
	}

	actionPKsList, err := s.groupResourcePolicyManager.ListActionPKsByGroup(groupPK)
	if err != nil {
		err = errorWrapf(err, "groupResourcePolicyManager.ListActionPKsByGroup groupPK=`%d` fail", groupPK)
		return
	}

	actionPKSet := set.NewInt64Set()
	for _, actionPKsStr := range actionPKsList {
		var actionPKs []int64
		if err = jsoniter.UnmarshalFromString(actionPKsStr, &actionPKs); err != nil {
			err = errorWrapf(err, "json.Unmarshal actionPKsStr=`%s` fail", actionPKsStr)
			return
		}

		actionPKSet.Append(actionPKs...)
	}

	if actionPKSet.Size() == 0 {
		return nil
	}

	err = s.create(groupPK, actionPKSet.ToSlice(), subjectPKs)
	if err != nil {
		err = errorWrapf(
			err,
			"create fail groupPK=`%d` actionPKs=`%+v` subjectPKs=`%+v`",
			actionPKSet.ToSlice(),
			subjectPKs,
		)
		return err
	}

	return nil
}

func (s *groupAlterEventService) create(groupPK int64, actionPKs, subjectPKs []int64) error {
	actionPKStr, err := jsoniter.MarshalToString(actionPKs)
	if err != nil {
		return err
	}

	subjectPKStr, err := jsoniter.MarshalToString(subjectPKs)
	if err != nil {
		return err
	}

	event := dao.GroupAlterEvent{
		UUID:       util.GenUUID4(),
		GroupPK:    groupPK,
		ActionPKs:  actionPKStr,
		SubjectPKs: subjectPKStr,
	}

	return s.manager.Create(event)
}

func (s *groupAlterEventService) CreateBySubjectActionGroup(subjectPK, actionPK, groupPK int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateBySubjectActionGroup")
	err := s.create(groupPK, []int64{actionPK}, []int64{subjectPK})
	if err != nil {
		err = errorWrapf(err, "create fail groupPK=`%d` actionPK=`%d` subjectPK=`%d`", groupPK, actionPK, subjectPK)
		return err
	}

	return nil
}

// Delete ...
func (s *groupAlterEventService) BulkDeleteWithTx(tx *sqlx.Tx, uuids []string) (err error) {
	return s.manager.BulkDeleteWithTx(tx, uuids)
}
