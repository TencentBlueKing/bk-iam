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
	"encoding/hex"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gofrs/uuid"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// NumEventShard 事件分片数
const NumEventShard = 100 // TODO 改成配置输入

// GroupAlterEventSVC ...
const GroupAlterEventSVC = "GroupAlterEventSVC"

// GroupAlterEventService ...
type GroupAlterEventService interface {
	Get(pk int64) (event types.GroupAlterEvent, err error)
	ListPKByCheckTimesBeforeCreateAt(checkTimes int64, createdAt int64) ([]int64, error)

	IncrCheckTimes(pk int64) (err error)
	CreateByGroupAction(groupPK int64, actionPKs []int64) ([]int64, error)
	CreateByGroupSubject(groupPK int64, subjectPKs []int64) ([]int64, error)

	Delete(pk int64) (err error)
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

// Get ...
func (s *groupAlterEventService) Get(pk int64) (event types.GroupAlterEvent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(GroupAlterEventSVC, "Get")

	daoEvent, err := s.manager.Get(pk)
	if err != nil {
		err = errorWrapf(err, "manager.Get pk=`%d` fail", pk)
		return
	}

	event, err = convertToSvcGroupAlterEvent(daoEvent)
	if err != nil {
		err = errorWrapf(err, "convertToSvcGroupAlterEvent fail event=`%+v`", daoEvent)
		return event, err
	}

	return
}

func convertToSvcGroupAlterEvent(daoEvent dao.GroupAlterEvent) (types.GroupAlterEvent, error) {
	event := types.GroupAlterEvent{
		PK:         daoEvent.PK,
		GroupPK:    daoEvent.GroupPK,
		CheckTimes: daoEvent.CheckTimes,
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

// Delete ...
func (s *groupAlterEventService) Delete(pk int64) (err error) {
	return s.manager.Delete(pk)
}

// IncrCheckTimes ...
func (s *groupAlterEventService) IncrCheckTimes(pk int64) (err error) {
	return s.manager.IncrCheckTimes(pk)
}

// CreateByGroupAction ...
func (s *groupAlterEventService) CreateByGroupAction(
	groupPK int64,
	actionPKs []int64,
) (pks []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateByGroupAction")

	if len(actionPKs) == 0 {
		return nil, nil
	}

	subjectRelations, err := s.subjectGroupManager.ListGroupMember(groupPK)
	if err != nil {
		err = errorWrapf(err, "subjectGroupManager.ListGroupMember groupPK=`%d` fail", groupPK)
		return
	}

	if len(subjectRelations) == 0 {
		return nil, nil
	}

	subjectPKs := make([]int64, 0, len(subjectRelations))
	for _, r := range subjectRelations {
		subjectPKs = append(subjectPKs, r.SubjectPK)
	}

	pks, err = s.bulkCreate(groupPK, actionPKs, subjectPKs)
	if err != nil {
		err = errorWrapf(err, "bulkCreate fail groupPK=`%d` actionPKs=`%+v` subjectPKs=`%+v`", actionPKs, subjectPKs)
		return nil, err
	}

	return pks, nil
}

// CreateByGroupSubject ...
func (s *groupAlterEventService) CreateByGroupSubject(
	groupPK int64,
	subjectPKs []int64,
) (pks []int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateByGroupSubject")

	if len(subjectPKs) == 0 {
		return nil, nil
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
		return nil, nil
	}

	pks, err = s.bulkCreate(groupPK, actionPKSet.ToSlice(), subjectPKs)
	if err != nil {
		err = errorWrapf(
			err,
			"bulkCreate fail groupPK=`%d` actionPKs=`%+v` subjectPKs=`%+v`",
			actionPKSet.ToSlice(),
			subjectPKs,
		)
		return nil, err
	}

	return pks, nil
}

func (s *groupAlterEventService) bulkCreate(groupPK int64, actionPKs, subjectPKs []int64) (pks []int64, err error) {
	actionPKStr, err := jsoniter.MarshalToString(actionPKs)
	if err != nil {
		return nil, err
	}

	// 分片批量创建
	step := NumEventShard / len(actionPKs)
	if step < 1 {
		step = 1
	}

	uuid := hex.EncodeToString(uuid.Must(uuid.NewV4()).Bytes())
	events := make([]dao.GroupAlterEvent, 0, len(subjectPKs)/step+1)
	for _, part := range chunks(len(subjectPKs), step) {
		subjectPKStr, err := jsoniter.MarshalToString(subjectPKs[part[0]:part[1]])
		if err != nil {
			return nil, err
		}

		events = append(events, dao.GroupAlterEvent{
			UUID:       uuid,
			GroupPK:    groupPK,
			ActionPKs:  actionPKStr,
			SubjectPKs: subjectPKStr,
		})
	}

	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		return nil, err
	}

	pks, err = s.manager.BulkCreateWithTx(tx, events)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return pks, nil
}

func (s *groupAlterEventService) ListPKByCheckTimesBeforeCreateAt(
	checkTimes int64,
	createdAt int64,
) ([]int64, error) {
	return s.manager.ListPKByCheckTimesBeforeCreateAt(checkTimes, createdAt)
}
