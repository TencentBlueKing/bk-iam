/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package handler

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/TencentBlueKing/gopkg/errorx"

	"iam/pkg/abac/prp/rbac/convert"
	"iam/pkg/cacheimpls"
	"iam/pkg/config"
	"iam/pkg/database"
	"iam/pkg/locker"
	"iam/pkg/logging"
	"iam/pkg/service"
	"iam/pkg/service/types"
)

const handlerLayer = "handler"

type groupAlterMessageHandler struct {
	groupService                      service.GroupService
	subjectActionAlterEventService    service.SubjectActionAlterEventService
	subjectActionGroupResourceService service.SubjectActionGroupResourceService
	subjectActionExpressionService    service.SubjectActionExpressionService

	locker *locker.SubjectDistributedActionLocker
}

// NewGroupAlterMessageHandler ...
func NewGroupAlterMessageHandler() MessageHandler {
	return &groupAlterMessageHandler{
		groupService:                      service.NewGroupService(),
		subjectActionAlterEventService:    service.NewSubjectActionAlterEventService(),
		subjectActionGroupResourceService: service.NewSubjectActionGroupResourceService(),
		subjectActionExpressionService:    service.NewSubjectActionExpressionService(),
		locker:                            locker.NewDistributedSubjectActionLocker(),
	}
}

// Handle ...
func (h *groupAlterMessageHandler) Handle(uuid string) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(handlerLayer, "handleEvent")

	event, err := h.subjectActionAlterEventService.Get(uuid)
	if err != nil {
		err = errorWrapf(err, "subjectActionAlterEventService.Get event fail, uuid=`%s`", uuid)
		return err
	}

	// 判断event check times超限，不再处理
	maxCheckCount := int64(config.MaxSubjectActionAlterEventCheckCount)
	if event.CheckCount > maxCheckCount {
		logger := logging.GetWorkerLogger()
		logger.Errorf(
			"subject action alter event uuid=`%s` check times exceed limit, check times=`%d`",
			uuid,
			event.CheckCount,
		)
		return nil
	}

	// update message status to processing
	err = h.subjectActionAlterEventService.BulkUpdateStatus(
		[]string{uuid},
		types.SubjectActionAlterEventStatusProcessing,
	)
	if err != nil {
		err = errorWrapf(
			err,
			"subjectActionAlterEventService.BulkUpdateStatus event fail, uuid=`%s`, status=`%d`",
			uuid,
			types.SubjectActionAlterEventStatusProcessing,
		)
		return err
	}

	// 循环处理所有事件
	for _, m := range event.Messages {
		err = h.alterSubjectActionGroupResource(m.SubjectPK, m.ActionPK, m.GroupPKs)
		if err != nil {
			err = errorWrapf(err, "alterSubjectActionGroupResource fail, message=`%v`", m)
			return err
		}
	}

	err = h.subjectActionAlterEventService.Delete(uuid)
	if err != nil {
		err = errorWrapf(err, "subjectActionAlterEventService.Delete event fail, uuid=`%d`", uuid)
		return err
	}

	return nil
}

// alterSubjectActionGroupResource 处理独立的事件
func (h *groupAlterMessageHandler) alterSubjectActionGroupResource(subjectPK, actionPK int64, groupPKs []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(handlerLayer, "alterSubjectActionGroupResource")

	// 分布式锁, subject_pk, action_pk
	// 请求锁最多3分钟
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Minute))
	defer cancel()

	lock, err := h.locker.Acquire(ctx, subjectPK, actionPK)
	if err != nil {
		return errorWrapf(err, "acquire lock fail, subjectPK=`%d`, actionPK`%d`", subjectPK, actionPK)
	}
	defer lock.Release(ctx)

	tx, err := database.GenerateDefaultDBTx()
	if err != nil {
		return err
	}
	defer database.RollBackWithLog(tx)

	obj, err := h.subjectActionGroupResourceService.Get(subjectPK, actionPK)
	if errors.Is(err, sql.ErrNoRows) {
		obj.SubjectPK = subjectPK
		obj.ActionPK = actionPK
		obj.GroupResource = map[int64]types.ResourceExpiredAt{}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errorWrapf(err,
			"subjectActionGroupResourceService.Get fail, subjectPK=`%d`, actionPK=`%d`",
			subjectPK, actionPK,
		)
	}

	// 遍历更新group resource
	for _, groupPK := range groupPKs {
		// groupPK == 0, 只更新expression
		if groupPK == 0 {
			continue
		}

		// 查询subject group关系过期时间
		expiredAt, err := h.groupService.GetExpiredAtBySubjectGroup(subjectPK, groupPK)
		if err != nil && !errors.Is(err, service.ErrGroupMemberNotFound) {
			return errorWrapf(err,
				"groupService.GetExpiredAtBySubjectGroup fail, subjectPK=`%d`, groupPK=`%d`",
				subjectPK, groupPK,
			)
		}
		found := !errors.Is(err, service.ErrGroupMemberNotFound)

		// 查询group action授权资源实例
		resourceMap, err := cacheimpls.GetGroupActionAuthorizedResource(
			groupPK,
			actionPK,
		)
		// NOTE: action如果被删除, rbac_group_resource_policy中action_pks并没有清理, 这里可能出现操作查询不到的错误, 如果查询不到, 直接删除
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return errorWrapf(err,
				"cacheimpls.GetGroupActionAuthorizedResource fail, groupPK=`%d`, actionPK=`%d`",
				groupPK, actionPK,
			)
		}

		if found && len(resourceMap) != 0 {
			// 更新group resource
			obj.UpdateGroupResource(groupPK, resourceMap, expiredAt)
		} else {
			// 关系不存在，移除用户组
			obj.DeleteGroupResource(groupPK)
		}
	}

	if len(obj.GroupResource) == 0 {
		// 表达式为空，删除subject action group resource/subject action expression
		err := h.subjectActionGroupResourceService.DeleteBySubjectActionWithTx(tx, subjectPK, actionPK)
		if err != nil {
			return errorWrapf(
				err,
				"subjectActionGroupResourceService.DeleteBySubjectActionWithTx fail, subjectPK=`%d`, actionPK=`%d`",
				subjectPK,
				actionPK,
			)
		}

		err = h.subjectActionExpressionService.DeleteBySubjectActionWithTx(tx, subjectPK, actionPK)
		if err != nil {
			return errorWrapf(
				err,
				"subjectActionExpressionService.DeleteBySubjectActionWithTx fail, subjectPK=`%d`, actionPK=`%d`",
				subjectPK,
				actionPK,
			)
		}
	} else {
		// 创建或更新subject action group resource
		err := h.subjectActionGroupResourceService.CreateOrUpdateWithTx(tx, obj)
		if err != nil {
			return errorWrapf(err, "subjectActionGroupResourceService.CreateOrUpdateWithTx fail, obj=`%+v`", obj)
		}

		// subject action resource group -> subject action expression
		expression, err := convert.SubjectActionGroupResourceToExpression(obj)
		if err != nil {
			return errorWrapf(err,
				"convertToSubjectActionExpression fail, subjectActionResourceGroup=`%+v`",
				obj,
			)
		}

		// 更新subject action expression
		err = h.subjectActionExpressionService.CreateOrUpdateWithTx(tx, expression)
		if err != nil {
			return errorWrapf(err,
				"subjectActionExpressionService.CreateOrUpdateWithTx fail, subjectActionExpression=`%+v`",
				expression,
			)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx.Commit fail")
	}

	// 清理 subject action expression缓存
	cacheimpls.DeleteSubjectActionExpressionCache(subjectPK, actionPK)

	return nil
}
