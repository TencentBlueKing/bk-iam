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

	"iam/pkg/task"
	"iam/pkg/task/producer"
)

const PapEvent = "PapEvent"

type PolicyEventProducer interface {
	PublishRBACDeletePolicyEvent(deletedPolicyPKs []int64)
	PublishABACDeletePolicyEvent(deletePolicyIDs []int64)
}

type policyEventProducer struct {
	deletePolicyEventProducer producer.Producer
}

func NewPolicyEventProducer() PolicyEventProducer {
	return &policyEventProducer{
		deletePolicyEventProducer: producer.NewRedisProducer(task.GetEngineDeletionEventQueue()),
	}
}

// FIXME: duplicated with prp/engine.go
const rbacIDBegin = 500000000

func (p *policyEventProducer) PublishRBACDeletePolicyEvent(deletedPolicyPKs []int64) {
	deletePolicyIDs := make([]int64, 0, len(deletedPolicyPKs))
	// NOTE: policy delete event, rbac policyID should add 50000000
	for _, pk := range deletedPolicyPKs {
		deletePolicyIDs = append(deletePolicyIDs, pk+rbacIDBegin)
	}
	p.publishDeletePolicyEvent(deletePolicyIDs)
}

func (p *policyEventProducer) PublishABACDeletePolicyEvent(deletePolicyIDs []int64) {
	p.publishDeletePolicyEvent(deletePolicyIDs)
}

func (p *policyEventProducer) publishDeletePolicyEvent(deletePolicyIDs []int64) {
	data := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"type":      EngineDeletionTypePolicy,
		"data":      map[string]any{"policy_ids": deletePolicyIDs},
	}
	message, err := jsoniter.MarshalToString(data)
	if err != nil {
		log.WithError(err).
			Errorf("pap.createDeletePolicyEvent marshal message data=`%+v` fail", data)
		return
	}

	go p.deletePolicyEventProducer.Publish(message)
}
