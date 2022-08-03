/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package event

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock
import (
	"time"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"

	"iam/pkg/service/types"
	"iam/pkg/task"
	"iam/pkg/task/producer"
)

type SubjectEventProducer interface {
	PublishDeleteEvent(deleteSubjects []types.Subject)
}

type subjectEventProducer struct {
	deleteSubjectEventProducer producer.Producer
}

func NewSubjectEventProducer() SubjectEventProducer {
	return &subjectEventProducer{
		deleteSubjectEventProducer: producer.NewRedisProducer(task.GetEngineDeletionEventQueue()),
	}
}

func (p *subjectEventProducer) PublishDeleteEvent(deleteSubjects []types.Subject) {
	data := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"type":      EngineDeletionTypeSubject,
		"data":      map[string]any{"subjects": deleteSubjects},
	}
	message, err := jsoniter.MarshalToString(data)
	if err != nil {
		log.WithError(err).
			Errorf("pap.createDeletePolicyEvent marshal message data=`%+v` fail", data)
		return
	}

	go p.deleteSubjectEventProducer.Publish(message)
}
