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
	"iam/pkg/database/dao"
	"iam/pkg/service/types"

	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// SubjectActionAlterMessageSVC ...
const SubjectActionAlterMessageSVC = "SubjectActionAlterMessageSVC"

// SubjectActionAlterMessageService ...
type SubjectActionAlterMessageService interface {
	Get(uuid string) (types.SubjectActionAlterMessage, error)
	BulkCreateWithTx(tx *sqlx.Tx, messages []types.SubjectActionAlterMessage) error
	BulkUpdateStatus(uuids []string, status int64) error
	Delete(uuid string) error
}

type subjectActionAlterMessageService struct {
	manager dao.SubjectActionAlterMessageManager
}

// NewSubjectActionAlterMessageService ...
func NewSubjectActionAlterMessageService() SubjectActionAlterMessageService {
	return &subjectActionAlterMessageService{
		manager: dao.NewSubjectActionAlterMessageManager(),
	}
}

func convertToSvcSubjectActionAlterMessage(
	daoEvent dao.SubjectActionAlterMessage,
) (types.SubjectActionAlterMessage, error) {
	message := types.SubjectActionAlterMessage{
		UUID:       daoEvent.UUID,
		Status:     daoEvent.Status,
		CheckCount: daoEvent.CheckCount,
	}

	err := jsoniter.UnmarshalFromString(daoEvent.Data, &message.Messages)
	if err != nil {
		return message, err
	}
	return message, nil
}

// Get ...
func (s *subjectActionAlterMessageService) Get(uuid string) (message types.SubjectActionAlterMessage, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionAlterMessageSVC, "Get")

	daoMessage, err := s.manager.Get(uuid)
	if err != nil {
		err = errorWrapf(err, "manager.Get uuid=`%s` fail", uuid)
		return message, err
	}

	message, err = convertToSvcSubjectActionAlterMessage(daoMessage)
	if err != nil {
		err = errorWrapf(err, "convertToSvcSubjectActionAlterMessage message=`%v` fail", daoMessage)
		return message, err
	}

	return message, nil
}

// BulkUpdateStatus ...
func (s *subjectActionAlterMessageService) BulkUpdateStatus(uuids []string, status int64) error {
	return s.manager.BulkUpdateStatus(uuids, status)
}

// Delete ...
func (s *subjectActionAlterMessageService) Delete(uuid string) error {
	return s.manager.Delete(uuid)
}

// BulkCreateWithTx ...
func (s *subjectActionAlterMessageService) BulkCreateWithTx(
	tx *sqlx.Tx,
	messages []types.SubjectActionAlterMessage,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(SubjectActionAlterMessageSVC, "BulkCreateWithTx")

	daoMessages := make([]dao.SubjectActionAlterMessage, 0, len(messages))
	for _, message := range messages {
		daoMessage := dao.SubjectActionAlterMessage{
			UUID:       message.UUID,
			Status:     message.Status,
			CheckCount: message.CheckCount,
		}

		daoMessage.Data, err = jsoniter.MarshalToString(message.Messages)
		if err != nil {
			err = errorWrapf(err, "jsoniter.MarshalToString message=`%v` fail", message.Messages)
			return err
		}

		daoMessages = append(daoMessages, daoMessage)
	}

	err = s.manager.BulkCreateWithTx(tx, daoMessages)
	if err != nil {
		err = errorWrapf(err, "manager.BulkCreateWithTx messages=`%v` fail", daoMessages)
		return
	}

	return nil
}
