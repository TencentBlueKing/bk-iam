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
	"errors"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// GroupAlterEventSVC ...
const GroupAlterEventSVC = "GroupAlterEventSVC"

// GroupAlterEventService ...
type GroupAlterEventService interface {
	ListByGroup(groupPK int64) ([]types.GroupAlterEvent, error)

	CreateByGroupAction(groupPK int64, actionPKs []int64) (types.GroupAlterEvent, error)
	CreateByGroupSubject(groupPK int64, subjectPKs []int64) (types.GroupAlterEvent, error)
}

type groupAlterEventService struct {
	manager                    dao.GroupAlterEventManager
	subjectGroupManager        dao.SubjectGroupManager
	groupResourcePolicyManager dao.GroupResourcePolicyManager
}

// NewGroupAlterEventService ...
func NewGroupAlterEventService() GroupAlterEventService {
	return &groupAlterEventService{
		manager:                    dao.NewGroupAlterEventManager(),
		subjectGroupManager:        dao.NewSubjectGroupManager(),
		groupResourcePolicyManager: dao.NewGroupResourcePolicyManager(),
	}
}

// CreateByGroupAction ...
func (s *groupAlterEventService) CreateByGroupAction(
	groupPK int64,
	actionPKs []int64,
) (event types.GroupAlterEvent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateByGroupAction")

	if len(actionPKs) == 0 {
		return event, errorWrapf(errors.New("actionPKs is empty"), "")
	}

	subjectRelations, err := s.subjectGroupManager.ListGroupMember(groupPK)
	if err != nil {
		err = errorWrapf(err, "subjectGroupManager.ListGroupMember groupPK=`%d` fail", groupPK)
		return
	}

	if len(subjectRelations) == 0 {
		return event, errorWrapf(errors.New("subjectPKs is empty"), "")
	}

	subjectPKs := make([]int64, 0, len(subjectRelations))
	for _, r := range subjectRelations {
		subjectPKs = append(subjectPKs, r.SubjectPK)
	}

	event = types.GroupAlterEvent{
		GroupPK:    groupPK,
		SubjectPKs: subjectPKs,
		ActionPKs:  actionPKs,
	}

	err = s.createEvent(event)
	if err != nil {
		err = errorWrapf(err, "createEvent event=`%+v` fail", event)
		return
	}

	return event, nil
}

// CreateByGroupSubject ...
func (s *groupAlterEventService) CreateByGroupSubject(
	groupPK int64,
	subjectPKs []int64,
) (event types.GroupAlterEvent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateByGroupSubject")

	if len(subjectPKs) == 0 {
		return event, errorWrapf(errors.New("subjectPKs is empty"), "")
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

		for _, actionPK := range actionPKs {
			actionPKSet.Add(actionPK)
		}
	}

	if actionPKSet.Size() == 0 {
		return event, errorWrapf(errors.New("actionPKs is empty"), "")
	}

	event = types.GroupAlterEvent{
		GroupPK:    groupPK,
		SubjectPKs: subjectPKs,
		ActionPKs:  actionPKSet.ToSlice(),
	}

	err = s.createEvent(event)
	if err != nil {
		err = errorWrapf(err, "createEvent event=`%+v` fail", event)
		return
	}

	return event, nil
}

func (s *groupAlterEventService) createEvent(event types.GroupAlterEvent) (err error) {
	subjectPKs, err := jsoniter.MarshalToString(event.SubjectPKs)
	if err != nil {
		return
	}

	actionPKs, err := jsoniter.MarshalToString(event.ActionPKs)
	if err != nil {
		return
	}

	daoEvent := dao.GroupAlterEvent{
		GroupPK:    event.GroupPK,
		SubjectPKs: subjectPKs,
		ActionPKs:  actionPKs,
	}

	return s.manager.Create(daoEvent)
}

func (s *groupAlterEventService) ListByGroup(groupPK int64) (events []types.GroupAlterEvent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "ListByGroupStatus")

	daoEvents, err := s.manager.ListByGroupStatus(groupPK, 0)
	if err != nil {
		err = errorWrapf(err, "manager.ListByGroupStatus groupPK=`%d` status=`%d` fail", groupPK, 0)
		return nil, err
	}

	events = make([]types.GroupAlterEvent, 0, len(daoEvents))
	for _, daoEvent := range daoEvents {
		var (
			subjectPKs []int64
			actionPKs  []int64
		)

		if err = jsoniter.UnmarshalFromString(daoEvent.SubjectPKs, &subjectPKs); err != nil {
			err = errorWrapf(err, "jsoniter.UnmarshalFromString subjectPKs=`%s` fail", daoEvent.SubjectPKs)
			return nil, err
		}

		if err = jsoniter.UnmarshalFromString(daoEvent.ActionPKs, &actionPKs); err != nil {
			err = errorWrapf(err, "jsoniter.UnmarshalFromString actionPKs=`%s` fail", daoEvent.ActionPKs)
			return nil, err
		}

		events = append(events, types.GroupAlterEvent{
			GroupPK:    daoEvent.GroupPK,
			ActionPKs:  actionPKs,
			SubjectPKs: subjectPKs,
		})
	}

	return events, nil
}
