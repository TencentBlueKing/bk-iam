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
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/config"
	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
	"iam/pkg/util"
)

// GroupAlterEventSVC ...
const GroupAlterEventSVC = "GroupAlterEventSVC"

// GroupAlterEventService ...
type GroupAlterEventService interface {
	Get(pk int64) (event types.GroupAlterEvent, err error)
	ListPKLtCheckCountBeforeCreateAt(CheckCount int64, createdAt int64) ([]int64, error)

	IncrCheckCount(pk int64) (err error)
	CreateByGroupAction(groupPK int64, actionPKs []int64) ([]int64, error)
	CreateByGroupSubject(groupPK int64, subjectPKs []int64) ([]int64, error)
	CreateBySubjectActionGroup(subjectPK, actionPK, groupPK int64) (pk int64, err error)

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
		CheckCount: daoEvent.CheckCount,
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

// IncrCheckCount ...
func (s *groupAlterEventService) IncrCheckCount(pk int64) (err error) {
	return s.manager.IncrCheckCount(pk)
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

		actionPKSet.Append(actionPKs...)
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

/*
举例: 5 个操作, 20 个用户, 总共会产生 100 个消息

maxMessageGenerationCountPerEvent 是 200, chunkSize=200/5=40; 20个用户被切分成 1 段, 生成 1 个event
maxMessageGenerationCountPerEvent 是 100, chunkSize=100/5=20, 20个用户被切分成 1 段, 生成 1 个event
maxMessageGenerationCountPerEvent 是 50, chunkSize=50/5=10, 那么20个用户被切分成 2 段, 生成 2 个event, 每段产生 50 个消息
*/
func (s *groupAlterEventService) bulkCreate(groupPK int64, actionPKs, subjectPKs []int64) (pks []int64, err error) {
	actionPKStr, err := jsoniter.MarshalToString(actionPKs)
	if err != nil {
		return nil, err
	}

	// 每个event最多能生成message的数量
	maxMessageGeneratedCountPerEvent := config.MaxMessageGeneratedCountPreGroupAlterEvent

	// 生成用户subjectPKs分片大小, 每个event的actionPKs都是相同, actionPKs不会被分片, 只分片subjectPKs
	// 一般actionPKs的数量不会太多, subjectPKs的数量可能会很多, 所以需要使用subjectPKs分片
	chunkSize := maxMessageGeneratedCountPerEvent / len(actionPKs)
	if chunkSize < 1 {
		chunkSize = 1
	}

	taskID := util.GenUUID4()
	events := make([]dao.GroupAlterEvent, 0, len(subjectPKs)/chunkSize+1)

	// 使用subjectPKs分片, 每个event能生成的消息数量为 len(actionPKs) * chunkSize
	for _, part := range chunks(len(subjectPKs), chunkSize) {
		subjectPKStr, err := jsoniter.MarshalToString(subjectPKs[part[0]:part[1]])
		if err != nil {
			return nil, err
		}

		events = append(events, dao.GroupAlterEvent{
			TaskID:     taskID,
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

// ListPKLtCheckCountBeforeCreateAt ...
func (s *groupAlterEventService) ListPKLtCheckCountBeforeCreateAt(
	checkCount int64,
	createdAt int64,
) ([]int64, error) {
	return s.manager.ListPKLtCheckCountBeforeCreateAt(checkCount, createdAt)
}

func (s *groupAlterEventService) CreateBySubjectActionGroup(subjectPK, actionPK, groupPK int64) (pk int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(ActionSVC, "CreateBySubjectActionGroup")
	pks, err := s.bulkCreate(groupPK, []int64{actionPK}, []int64{subjectPK})
	if err != nil {
		err = errorWrapf(err, "bulkCreate fail groupPK=`%d` actionPK=`%d` subjectPK=`%d`", groupPK, actionPK, subjectPK)
		return 0, err
	}

	return pks[0], nil
}
