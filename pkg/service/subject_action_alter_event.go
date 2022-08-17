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
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"

	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

// SubjectActionAlterEventSVC ...
const SubjectActionAlterEventSVC = "SubjectActionAlterEventSVC"

// SubjectActionAlterEventService ...
type SubjectActionAlterEventService interface {
	Get(uuid string) (types.SubjectActionAlterEvent, error)
	BulkCreateWithTx(tx *sqlx.Tx, events []types.SubjectActionAlterEvent) error
	BulkUpdateStatus(uuids []string, status int64) error
	Delete(uuid string) error

	// for task checker
	ListUUIDByStatusBeforeUpdatedAt(status, updateAt int64) ([]string, error)
	ListUUIDGreaterThanStatusLessThanCheckCountBeforeUpdatedAt(
		status, checkCount, updateAt int64,
	) ([]string, error)
	BulkIncrCheckCount(uuids []string) error
}

type subjectActionAlterEventService struct {
	manager dao.SubjectActionAlterEventManager
}

// NewSubjectActionAlterEventService ...
func NewSubjectActionAlterEventService() SubjectActionAlterEventService {
	return &subjectActionAlterEventService{
		manager: dao.NewSubjectActionAlterEventManager(),
	}
}

func convertToSvcSubjectActionAlterEvent(
	daoEvent dao.SubjectActionAlterEvent,
) (types.SubjectActionAlterEvent, error) {
	event := types.SubjectActionAlterEvent{
		UUID:       daoEvent.UUID,
		Status:     daoEvent.Status,
		CheckCount: daoEvent.CheckCount,
	}

	err := jsoniter.UnmarshalFromString(daoEvent.Data, &event.Messages)
	if err != nil {
		return event, err
	}
	return event, nil
}

// Get ...
func (s *subjectActionAlterEventService) Get(uuid string) (event types.SubjectActionAlterEvent, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionAlterEventSVC, "Get")

	daoEvent, err := s.manager.Get(uuid)
	if err != nil {
		err = errorWrapf(err, "manager.Get uuid=`%s` fail", uuid)
		return event, err
	}

	event, err = convertToSvcSubjectActionAlterEvent(daoEvent)
	if err != nil {
		err = errorWrapf(err, "convertToSvcSubjectActionAlterEvent event=`%v` fail", daoEvent)
		return event, err
	}

	return event, nil
}

// BulkUpdateStatus ...
func (s *subjectActionAlterEventService) BulkUpdateStatus(uuids []string, status int64) error {
	return s.manager.BulkUpdateStatus(uuids, status)
}

// Delete ...
func (s *subjectActionAlterEventService) Delete(uuid string) error {
	return s.manager.Delete(uuid)
}

// BulkCreateWithTx ...
func (s *subjectActionAlterEventService) BulkCreateWithTx(
	tx *sqlx.Tx,
	events []types.SubjectActionAlterEvent,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionAlterEventSVC, "BulkCreateWithTx")

	daoEvents := make([]dao.SubjectActionAlterEvent, 0, len(events))
	for _, event := range events {
		daoMessage := dao.SubjectActionAlterEvent{
			UUID:       event.UUID,
			Status:     event.Status,
			CheckCount: event.CheckCount,
		}

		daoMessage.Data, err = jsoniter.MarshalToString(event.Messages)
		if err != nil {
			err = errorWrapf(err, "jsoniter.MarshalToString event=`%v` fail", event.Messages)
			return err
		}

		daoEvents = append(daoEvents, daoMessage)
	}

	err = s.manager.BulkCreateWithTx(tx, daoEvents)
	if err != nil {
		err = errorWrapf(err, "manager.BulkCreateWithTx events=`%v` fail", daoEvents)
		return
	}

	return nil
}

// ListUUIDByStatusBeforeUpdatedAt ...
func (s *subjectActionAlterEventService) ListUUIDByStatusBeforeUpdatedAt(status, updateAt int64) ([]string, error) {
	return s.manager.ListUUIDByStatusBeforeUpdatedAt(status, updateAt)
}

// ListUUIDGreaterThanStatusLessThanCheckCountBeforeUpdatedAt ...
func (s *subjectActionAlterEventService) ListUUIDGreaterThanStatusLessThanCheckCountBeforeUpdatedAt(
	status, checkCount, updateAt int64,
) ([]string, error) {
	return s.manager.ListUUIDGreaterThanStatusLessThanCheckCountBeforeUpdatedAt(status, checkCount, updateAt)
}

// BulkIncrCheckCount ...
func (s *subjectActionAlterEventService) BulkIncrCheckCount(uuids []string) error {
	return s.manager.BulkIncrCheckCount(uuids)
}
