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

	"github.com/TencentBlueKing/gopkg/collection/set"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"

	"iam/pkg/cacheimpls"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/task"
	"iam/pkg/task/producer"
	"iam/pkg/util"
)

const PapEvent = "PapEvent"

type PolicyEventProducer interface {
	PublishRBACGroupAlterEvent(groupPK int64, resourceChangedContents []svctypes.ResourceChangedContent)

	PublishRBACDeletePolicyEvent(deletedPolicyPKs []int64)
	PublishABACDeletePolicyEvent(deletePolicyIDs []int64)
}

type policyEventProducer struct {
	groupAlterEventService service.GroupAlterEventService

	alterEventProducer        producer.Producer
	deletePolicyEventProducer producer.Producer
}

func NewPolicyEventProducer() PolicyEventProducer {
	return &policyEventProducer{
		groupAlterEventService: service.NewGroupAlterEventService(),

		alterEventProducer:        producer.NewRedisProducer(task.GetRbacEventQueue()),
		deletePolicyEventProducer: producer.NewRedisProducer(task.GetEngineDeletionEventQueue()),
	}
}

// publishRBACGroupAlterEvent 创建用户组变更事件
func (p *policyEventProducer) PublishRBACGroupAlterEvent(
	groupPK int64,
	resourceChangedContents []svctypes.ResourceChangedContent,
) {
	actionPKSet := set.NewInt64Set()
	for _, rcc := range resourceChangedContents {
		actionPKSet.Append(rcc.CreatedActionPKs...)
		actionPKSet.Append(rcc.DeletedActionPKs...)
	}

	actionPKs := actionPKSet.ToSlice()

	// 清group action resource 缓存
	cacheimpls.BatchDeleteGroupActionAuthorizedResourceCache(groupPK, actionPKs)

	// 创建 group_alter_event
	pks, err := p.groupAlterEventService.CreateByGroupAction(groupPK, actionPKs)
	if err != nil {
		log.WithError(err).
			Errorf("groupAlterEventService.CreateByGroupAction groupPK=%d actionPKs=%v fail", groupPK, actionPKs)

		// report to sentry
		util.ReportToSentry("createRBACGroupAlterEvent groupAlterEventService.CreateByGroupAction fail",
			map[string]interface{}{
				"layer":     PapEvent,
				"groupPK":   groupPK,
				"actionPKs": actionPKs,
				"error":     err.Error(),
			},
		)
		return
	}

	// 发送event 消息
	if len(pks) == 0 {
		return
	}

	messages := util.Int64SliceToStringSlice(pks)
	go p.alterEventProducer.Publish(messages...)
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
